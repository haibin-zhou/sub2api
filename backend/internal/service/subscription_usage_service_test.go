package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeSubUsageLister struct {
	accounts []Account
	err      error
	gotGroup int64
}

func (f *fakeSubUsageLister) ListByGroup(_ context.Context, groupID int64) ([]Account, error) {
	f.gotGroup = groupID
	return f.accounts, f.err
}

type fakeSubUsageFetcher struct {
	info      *UsageInfo
	err       error
	gotID     int64
	gotForced bool
}

func (f *fakeSubUsageFetcher) GetUsageForAccount(_ context.Context, account *Account, force ...bool) (*UsageInfo, error) {
	f.gotID = account.ID
	f.gotForced = len(force) > 0 && force[0]
	return f.info, f.err
}

func subAcc(id int64, groupPrio, accPrio int) Account {
	return Account{
		ID: id, Name: "secret-name", Platform: PlatformAnthropic, Type: AccountTypeOAuth,
		Status: StatusActive, Schedulable: true, Priority: accPrio,
		AccountGroups: []AccountGroup{{AccountID: id, GroupID: 7, Priority: groupPrio}},
	}
}

func subUsageInt64Ptr(v int64) *int64 { return &v }

func newSubSvc(l *fakeSubUsageLister, f *fakeSubUsageFetcher) *SubscriptionUsageService {
	return &SubscriptionUsageService{accounts: l, usage: f}
}

func TestSubscriptionUsage_PicksHighestPriority(t *testing.T) {
	reset := time.Date(2026, 11, 19, 16, 30, 0, 0, time.UTC)
	l := &fakeSubUsageLister{accounts: []Account{subAcc(3, 50, 1), subAcc(5, 10, 99), subAcc(4, 10, 20)}}
	f := &fakeSubUsageFetcher{info: &UsageInfo{
		SubscriptionTier: "MAX",
		FiveHour:         &UsageProgress{Utilization: 42, ResetsAt: &reset, RemainingSeconds: 9000},
	}}
	got, err := newSubSvc(l, f).GetForAPIKey(context.Background(), &APIKey{GroupID: subUsageInt64Ptr(7)})
	require.NoError(t, err)
	require.Equal(t, int64(7), l.gotGroup)
	require.Equal(t, int64(4), f.gotID) // group prio 10 ties → account prio 20 < 99
	require.False(t, f.gotForced)
	require.Equal(t, "max", got.Plan)
	require.Equal(t, 42.0, got.FiveHour.Utilization)
	require.Nil(t, got.SevenDay)
}

func TestSubscriptionUsage_TieBreaksByID(t *testing.T) {
	l := &fakeSubUsageLister{accounts: []Account{subAcc(9, 10, 10), subAcc(2, 10, 10)}}
	f := &fakeSubUsageFetcher{info: &UsageInfo{}}
	_, err := newSubSvc(l, f).GetForAPIKey(context.Background(), &APIKey{GroupID: subUsageInt64Ptr(7)})
	require.NoError(t, err)
	require.Equal(t, int64(2), f.gotID)
}

func TestSubscriptionUsage_SkipsUnschedulableAndNonOAuth(t *testing.T) {
	future := time.Now().Add(time.Hour)
	limited := subAcc(1, 1, 1)
	limited.RateLimitResetAt = &future
	paused := subAcc(2, 1, 1)
	paused.Schedulable = false
	apiKeyAcc := subAcc(3, 1, 1)
	apiKeyAcc.Type = AccountTypeAPIKey
	openai := subAcc(4, 1, 1)
	openai.Platform = PlatformOpenAI
	ok := subAcc(8, 90, 90)
	l := &fakeSubUsageLister{accounts: []Account{limited, paused, apiKeyAcc, openai, ok}}
	f := &fakeSubUsageFetcher{info: &UsageInfo{}}
	_, err := newSubSvc(l, f).GetForAPIKey(context.Background(), &APIKey{GroupID: subUsageInt64Ptr(7)})
	require.NoError(t, err)
	require.Equal(t, int64(8), f.gotID)
}

func TestSubscriptionUsage_NoGroup(t *testing.T) {
	svc := newSubSvc(&fakeSubUsageLister{}, &fakeSubUsageFetcher{})
	_, err := svc.GetForAPIKey(context.Background(), &APIKey{})
	require.ErrorIs(t, err, ErrSubscriptionUsageNoGroup)
	_, err = svc.GetForAPIKey(context.Background(), nil)
	require.ErrorIs(t, err, ErrSubscriptionUsageNoGroup)
}

func TestSubscriptionUsage_NotSubscriptionGroup(t *testing.T) {
	a := subAcc(1, 1, 1)
	a.Type = AccountTypeAPIKey
	_, err := newSubSvc(&fakeSubUsageLister{accounts: []Account{a}}, &fakeSubUsageFetcher{}).
		GetForAPIKey(context.Background(), &APIKey{GroupID: subUsageInt64Ptr(7)})
	require.ErrorIs(t, err, ErrSubscriptionUsageNotSubscriptionGroup)
}

func TestSubscriptionUsage_NoAvailableAccount(t *testing.T) {
	a := subAcc(1, 1, 1)
	a.Schedulable = false
	_, err := newSubSvc(&fakeSubUsageLister{accounts: []Account{a}}, &fakeSubUsageFetcher{}).
		GetForAPIKey(context.Background(), &APIKey{GroupID: subUsageInt64Ptr(7)})
	require.ErrorIs(t, err, ErrSubscriptionUsageNoAvailableAccount)
}

func TestSubscriptionUsage_UpstreamErrorWrapped(t *testing.T) {
	f := &fakeSubUsageFetcher{err: errors.New("401 for account secret-name")}
	_, err := newSubSvc(&fakeSubUsageLister{accounts: []Account{subAcc(1, 1, 1)}}, f).
		GetForAPIKey(context.Background(), &APIKey{GroupID: subUsageInt64Ptr(7)})
	require.ErrorIs(t, err, ErrSubscriptionUsageUpstream)
}

func TestSubscriptionUsage_JSONHasNoAccountIdentity(t *testing.T) {
	reset := time.Now().Add(time.Hour)
	f := &fakeSubUsageFetcher{info: &UsageInfo{
		FiveHour: &UsageProgress{Utilization: 1, ResetsAt: &reset, WindowStats: &WindowStats{}},
		SevenDay: &UsageProgress{Utilization: 2},
	}}
	got, err := newSubSvc(&fakeSubUsageLister{accounts: []Account{subAcc(12345, 1, 1)}}, f).
		GetForAPIKey(context.Background(), &APIKey{GroupID: subUsageInt64Ptr(7)})
	require.NoError(t, err)
	raw, err := json.Marshal(got)
	require.NoError(t, err)
	s := string(raw)
	require.NotContains(t, s, "12345")
	require.NotContains(t, s, "secret-name")
	require.NotContains(t, s, "window_stats")
	require.Contains(t, s, `"five_hour"`)
	require.NotContains(t, s, `"seven_day_sonnet"`)
}
