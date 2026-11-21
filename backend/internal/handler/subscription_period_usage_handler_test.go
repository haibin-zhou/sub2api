package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakePeriodUsageLister struct {
	items  []service.SubscriptionPeriodUsage
	err    error
	gotUID int64
	calls  int
}

func (f *fakePeriodUsageLister) ListForUser(_ context.Context, userID int64) ([]service.SubscriptionPeriodUsage, error) {
	f.calls++
	f.gotUID = userID
	return f.items, f.err
}

func runUserPeriodUsage(t *testing.T, p subscriptionPeriodUsageLister, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/period-usage?subscription_id=999&user_id=1", nil)
	if userID > 0 {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID})
	}
	(&SubscriptionUsageHandler{periods: p}).ListPeriodUsageForUser(c)
	return w
}

func TestSubscriptionPeriodUsageUser_MapsValues(t *testing.T) {
	start := time.Date(2026, 10, 31, 8, 0, 0, 0, time.UTC)
	p := &fakePeriodUsageLister{items: []service.SubscriptionPeriodUsage{{
		SubscriptionID: 2, GroupID: 7, PeriodIndex: 3,
		PeriodStart: start, PeriodEnd: start.Add(30 * 24 * time.Hour),
		RequestCount: 1753, InputTokens: 11, OutputTokens: 22, CacheCreationTokens: 33, CacheReadTokens: 44,
		TotalCostUSD: 180.51,
	}}}
	w := runUserPeriodUsage(t, p, 77)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int64(77), p.gotUID, "user id must come from the JWT subject, not the query string")

	var resp struct {
		Code int              `json:"code"`
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data, 1)
	item := resp.Data[0]
	require.Equal(t, 2.0, item["subscription_id"])
	require.Equal(t, 7.0, item["group_id"])
	require.Equal(t, 3.0, item["period_index"])
	require.Equal(t, "2026-10-31T08:00:00Z", item["period_start"])
	require.Equal(t, "2026-11-30T08:00:00Z", item["period_end"])
	require.Equal(t, 1753.0, item["request_count"])
	require.Equal(t, 11.0, item["input_tokens"])
	require.Equal(t, 22.0, item["output_tokens"])
	require.Equal(t, 33.0, item["cache_creation_tokens"])
	require.Equal(t, 44.0, item["cache_read_tokens"])
	require.Equal(t, 180.51, item["total_cost_usd"])
	require.Len(t, item, 11, "no extra fields (e.g. account identity) may leak")
}

func TestSubscriptionPeriodUsageUser_EmptyIsArray(t *testing.T) {
	w := runUserPeriodUsage(t, &fakePeriodUsageLister{items: []service.SubscriptionPeriodUsage{}}, 5)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"data":[]`)

	w = runUserPeriodUsage(t, &fakePeriodUsageLister{}, 5)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"data":[]`)
}

func TestSubscriptionPeriodUsageUser_ErrorIsFixedMessage(t *testing.T) {
	w := runUserPeriodUsage(t, &fakePeriodUsageLister{err: errors.New("pq: relation secret-table does not exist")}, 5)
	require.Equal(t, http.StatusInternalServerError, w.Code)
	require.NotContains(t, w.Body.String(), "secret-table")
	require.Contains(t, w.Body.String(), "Failed to load subscription period usage")
}

func TestSubscriptionPeriodUsageUser_Unauthorized(t *testing.T) {
	p := &fakePeriodUsageLister{}
	w := runUserPeriodUsage(t, p, 0)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Equal(t, 0, p.calls)
}

// 端到端（handler + 真实 service）：只统计有效订阅，忽略客户端参数。
func TestSubscriptionPeriodUsageUser_WithServiceActiveOnly(t *testing.T) {
	now := time.Now().UTC()
	active := activeSub(10, 1)
	active.StartsAt = now.Add(-24 * time.Hour)
	expired := activeSub(11, 2)
	expired.StartsAt = now.Add(-48 * time.Hour)
	expired.ExpiresAt = now.Add(-time.Hour)
	revoked := activeSub(12, 3)
	revoked.Status = "revoked"
	subs := &fakeActiveSubs{subs: []service.UserSubscription{active, expired, revoked}}
	repo := &fakeSubPeriodRepo{out: map[int64]service.SubscriptionPeriodUsageTotals{10: {RequestCount: 4, TotalCostUSD: 1.5}, 11: {RequestCount: 99}}}
	svc := service.NewSubscriptionPeriodUsageServiceWith(subs, repo, time.Now)

	w := runUserPeriodUsage(t, svc, 9)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int64(9), subs.gotUID)
	require.Equal(t, int64(9), repo.gotUID)
	require.Len(t, repo.windows, 1)
	require.Equal(t, int64(10), repo.windows[0].SubscriptionID)
	var resp struct {
		Data []service.SubscriptionPeriodUsage `json:"data"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, int64(10), resp.Data[0].SubscriptionID)
	require.Equal(t, int64(4), resp.Data[0].RequestCount)
	require.Equal(t, 1, resp.Data[0].PeriodIndex)
}

type fakeSubPeriodRepo struct {
	gotUID  int64
	windows []service.SubscriptionPeriodWindow
	out     map[int64]service.SubscriptionPeriodUsageTotals
}

func (f *fakeSubPeriodRepo) SumSubscriptionPeriodUsage(_ context.Context, userID int64, windows []service.SubscriptionPeriodWindow) (map[int64]service.SubscriptionPeriodUsageTotals, error) {
	f.gotUID = userID
	f.windows = windows
	return f.out, nil
}
