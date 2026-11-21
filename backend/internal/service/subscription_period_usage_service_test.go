package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSubscriptionPeriod(t *testing.T) {
	starts := time.Date(2026, 9, 1, 8, 30, 0, 0, time.UTC)
	day := 24 * time.Hour
	cases := []struct {
		name      string
		now       time.Time
		wantIndex int
		wantStart time.Time
	}{
		{"exactly at starts_at", starts, 1, starts},
		{"29d23h", starts.Add(29*day + 23*time.Hour), 1, starts},
		{"just before 30d", starts.Add(30*day - time.Nanosecond), 1, starts},
		{"exactly 30d", starts.Add(30 * day), 2, starts.Add(30 * day)},
		{"61d", starts.Add(61 * day), 3, starts.Add(60 * day)},
		{"now before starts_at", starts.Add(-5 * day), 1, starts},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			idx, start, end := subscriptionPeriod(starts, tc.now)
			require.Equal(t, tc.wantIndex, idx)
			require.True(t, tc.wantStart.Equal(start), "start=%s", start)
			require.True(t, tc.wantStart.Add(30*day).Equal(end), "end=%s", end)
		})
	}
}

func TestSubscriptionPeriod_NonUTCInputReturnsUTC(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	starts := time.Date(2026, 9, 1, 8, 0, 0, 0, loc)
	_, start, end := subscriptionPeriod(starts, starts.Add(time.Hour))
	require.Equal(t, time.UTC, start.Location())
	require.Equal(t, time.UTC, end.Location())
	require.True(t, starts.Equal(start))
}

type fakePeriodSubLister struct {
	subs   []UserSubscription
	err    error
	gotUID int64
}

func (f *fakePeriodSubLister) ListActiveUserSubscriptions(_ context.Context, userID int64) ([]UserSubscription, error) {
	f.gotUID = userID
	return f.subs, f.err
}

type fakePeriodUsageRepo struct {
	calls   int
	gotUID  int64
	windows []SubscriptionPeriodWindow
	out     map[int64]SubscriptionPeriodUsageTotals
	err     error
}

func (f *fakePeriodUsageRepo) SumSubscriptionPeriodUsage(_ context.Context, userID int64, windows []SubscriptionPeriodWindow) (map[int64]SubscriptionPeriodUsageTotals, error) {
	f.calls++
	f.gotUID = userID
	f.windows = windows
	return f.out, f.err
}

func TestSubscriptionPeriodUsageService_ListForUser(t *testing.T) {
	now := time.Date(2026, 11, 21, 12, 0, 0, 0, time.UTC)
	day := 24 * time.Hour
	subA := UserSubscription{ID: 2, GroupID: 7, Status: SubscriptionStatusActive, StartsAt: now.Add(-45 * day), ExpiresAt: now.Add(30 * day)}
	subB := UserSubscription{ID: 3, GroupID: 8, Status: SubscriptionStatusActive, StartsAt: now.Add(-10 * day), ExpiresAt: now.Add(20 * day)}
	expired := UserSubscription{ID: 4, GroupID: 9, Status: SubscriptionStatusActive, StartsAt: now.Add(-40 * day), ExpiresAt: time.Now().Add(-time.Hour)}
	revoked := UserSubscription{ID: 5, GroupID: 10, Status: "revoked", StartsAt: now.Add(-40 * day), ExpiresAt: time.Now().Add(time.Hour)}
	lister := &fakePeriodSubLister{subs: []UserSubscription{subA, subB, expired, revoked}}
	repo := &fakePeriodUsageRepo{out: map[int64]SubscriptionPeriodUsageTotals{
		2: {RequestCount: 12, InputTokens: 100, OutputTokens: 200, CacheCreationTokens: 300, CacheReadTokens: 400, TotalCostUSD: 180.51},
	}}
	svc := &SubscriptionPeriodUsageService{subs: lister, repo: repo, now: func() time.Time { return now }}

	items, err := svc.ListForUser(context.Background(), 42)
	require.NoError(t, err)
	require.Equal(t, int64(42), lister.gotUID)
	require.Equal(t, int64(42), repo.gotUID)
	require.Equal(t, 1, repo.calls)

	require.Len(t, repo.windows, 2)
	require.Equal(t, int64(2), repo.windows[0].SubscriptionID)
	require.True(t, subA.StartsAt.Add(30*day).Equal(repo.windows[0].Start))
	require.True(t, now.Equal(repo.windows[0].End), "query upper bound must be min(now, period_end)")
	require.Equal(t, int64(3), repo.windows[1].SubscriptionID)

	require.Len(t, items, 2)
	a := items[0]
	require.Equal(t, int64(2), a.SubscriptionID)
	require.Equal(t, int64(7), a.GroupID)
	require.Equal(t, 2, a.PeriodIndex)
	require.True(t, subA.StartsAt.Add(30*day).Equal(a.PeriodStart))
	require.True(t, subA.StartsAt.Add(60*day).Equal(a.PeriodEnd))
	require.Equal(t, int64(12), a.RequestCount)
	require.Equal(t, int64(100), a.InputTokens)
	require.Equal(t, int64(200), a.OutputTokens)
	require.Equal(t, int64(300), a.CacheCreationTokens)
	require.Equal(t, int64(400), a.CacheReadTokens)
	require.InDelta(t, 180.51, a.TotalCostUSD, 1e-9)

	b := items[1]
	require.Equal(t, int64(3), b.SubscriptionID)
	require.Equal(t, 1, b.PeriodIndex)
	require.Equal(t, int64(0), b.RequestCount)
	require.Equal(t, 0.0, b.TotalCostUSD)
}

func TestSubscriptionPeriodUsageService_NoActiveSkipsQuery(t *testing.T) {
	repo := &fakePeriodUsageRepo{}
	svc := &SubscriptionPeriodUsageService{subs: &fakePeriodSubLister{}, repo: repo, now: time.Now}
	items, err := svc.ListForUser(context.Background(), 1)
	require.NoError(t, err)
	require.NotNil(t, items)
	require.Empty(t, items)
	require.Equal(t, 0, repo.calls)
}

func TestSubscriptionPeriodUsageService_Errors(t *testing.T) {
	boom := errors.New("boom")
	svc := &SubscriptionPeriodUsageService{subs: &fakePeriodSubLister{err: boom}, repo: &fakePeriodUsageRepo{}, now: time.Now}
	_, err := svc.ListForUser(context.Background(), 1)
	require.ErrorIs(t, err, boom)

	sub := UserSubscription{ID: 1, Status: SubscriptionStatusActive, StartsAt: time.Now().Add(-time.Hour), ExpiresAt: time.Now().Add(time.Hour)}
	svc = &SubscriptionPeriodUsageService{subs: &fakePeriodSubLister{subs: []UserSubscription{sub}}, repo: &fakePeriodUsageRepo{err: boom}, now: time.Now}
	_, err = svc.ListForUser(context.Background(), 1)
	require.ErrorIs(t, err, boom)
}
