package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeGroupUsage struct {
	mu     sync.Mutex
	byID   map[int64]*service.SubscriptionUsage
	errs   map[int64]error
	called map[int64]int
}

func (f *fakeGroupUsage) GetForAPIKey(context.Context, *service.APIKey) (*service.SubscriptionUsage, error) {
	return nil, errors.New("unused")
}

func (f *fakeGroupUsage) GetForGroup(_ context.Context, groupID int64) (*service.SubscriptionUsage, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.called == nil {
		f.called = map[int64]int{}
	}
	f.called[groupID]++
	if err := f.errs[groupID]; err != nil {
		return nil, err
	}
	return f.byID[groupID], nil
}

type fakeActiveSubs struct {
	subs   []service.UserSubscription
	err    error
	gotUID int64
}

func (f *fakeActiveSubs) ListActiveUserSubscriptions(_ context.Context, userID int64) ([]service.UserSubscription, error) {
	f.gotUID = userID
	return f.subs, f.err
}

type userUsageResp struct {
	Code int `json:"code"`
	Data []struct {
		GroupID int64                      `json:"group_id"`
		Usage   *service.SubscriptionUsage `json:"usage"`
		Error   *struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	} `json:"data"`
}

func runUserSubUsage(t *testing.T, g *fakeGroupUsage, subs *fakeActiveSubs, userID int64) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/subscriptions/usage?group_id=999", nil)
	if userID > 0 {
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: userID})
	}
	(&SubscriptionUsageHandler{svc: g, subs: subs}).ListForUser(c)
	return w
}

func activeSub(id, groupID int64) service.UserSubscription {
	return service.UserSubscription{ID: id, GroupID: groupID, Status: service.SubscriptionStatusActive, ExpiresAt: time.Now().Add(24 * time.Hour),
		Group: &service.Group{ID: groupID, Platform: service.PlatformAnthropic}}
}

func TestSubscriptionUsageUser_MixedResults(t *testing.T) {
	g := &fakeGroupUsage{
		byID: map[int64]*service.SubscriptionUsage{1: {Plan: "max", FiveHour: &service.SubscriptionUsageWindow{Utilization: 42}}},
		errs: map[int64]error{
			2: service.ErrSubscriptionUsageNotSubscriptionGroup,
			3: fmt.Errorf("%w: 401 for account secret-name", service.ErrSubscriptionUsageUpstream),
			4: service.ErrSubscriptionUsageNoAvailableAccount,
			5: errors.New("db down secret-name"),
		},
	}
	subs := &fakeActiveSubs{subs: []service.UserSubscription{activeSub(10, 1), activeSub(11, 2), activeSub(12, 3), activeSub(13, 4), activeSub(14, 5)}}
	w := runUserSubUsage(t, g, subs, 77)
	require.Equal(t, http.StatusOK, w.Code)
	require.Equal(t, int64(77), subs.gotUID)
	require.NotContains(t, w.Body.String(), "secret-name")
	var resp userUsageResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, 0, resp.Code)
	require.Len(t, resp.Data, 5)

	require.Equal(t, int64(1), resp.Data[0].GroupID)
	require.NotNil(t, resp.Data[0].Usage)
	require.Equal(t, "max", resp.Data[0].Usage.Plan)
	require.Equal(t, 42.0, resp.Data[0].Usage.FiveHour.Utilization)
	require.Nil(t, resp.Data[0].Error)

	want := []struct{ typ, msg string }{
		{"not_subscription_group", "This group is not backed by a Claude subscription"},
		{"upstream_usage_error", "Failed to query subscription usage upstream"},
		{"no_available_account", "No subscription account is currently available in this group"},
		{"api_error", "Internal error"},
	}
	for i, tc := range want {
		item := resp.Data[i+1]
		require.Nil(t, item.Usage, tc.typ)
		require.NotNil(t, item.Error, tc.typ)
		require.Equal(t, tc.typ, item.Error.Type)
		require.Equal(t, tc.msg, item.Error.Message)
	}
	require.NotContains(t, g.called, int64(999), "client-supplied group_id must be ignored")
}

func TestSubscriptionUsageUser_OnlyActiveAndDeduped(t *testing.T) {
	g := &fakeGroupUsage{byID: map[int64]*service.SubscriptionUsage{1: {}, 2: {}, 3: {}}}
	expired := activeSub(20, 2)
	expired.ExpiresAt = time.Now().Add(-time.Hour)
	revoked := activeSub(21, 3)
	revoked.Status = "revoked"
	subs := &fakeActiveSubs{subs: []service.UserSubscription{activeSub(10, 1), activeSub(11, 1), expired, revoked}}
	w := runUserSubUsage(t, g, subs, 5)
	require.Equal(t, http.StatusOK, w.Code)
	var resp userUsageResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, int64(1), resp.Data[0].GroupID)
	require.Equal(t, 1, g.called[1])
	require.NotContains(t, g.called, int64(2))
	require.NotContains(t, g.called, int64(3))
}

func TestSubscriptionUsageUser_EmptyIsArray(t *testing.T) {
	w := runUserSubUsage(t, &fakeGroupUsage{}, &fakeActiveSubs{}, 5)
	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"data":[]`)
}

func TestSubscriptionUsageUser_Unauthenticated(t *testing.T) {
	g := &fakeGroupUsage{}
	w := runUserSubUsage(t, g, &fakeActiveSubs{subs: []service.UserSubscription{activeSub(1, 1)}}, 0)
	require.Equal(t, http.StatusUnauthorized, w.Code)
	require.Empty(t, g.called)
}

func TestSubscriptionUsageUser_ListError(t *testing.T) {
	w := runUserSubUsage(t, &fakeGroupUsage{}, &fakeActiveSubs{err: errors.New("db secret-name")}, 5)
	require.GreaterOrEqual(t, w.Code, 500)
	require.NotContains(t, w.Body.String(), "secret-name")
}

func TestSubscriptionUsageUser_OnlyAnthropicGroups(t *testing.T) {
	g := &fakeGroupUsage{byID: map[int64]*service.SubscriptionUsage{1: {}, 2: {}, 3: {}}}
	openai := activeSub(11, 2)
	openai.Group = &service.Group{ID: 2, Platform: service.PlatformOpenAI}
	noGroup := activeSub(12, 3)
	noGroup.Group = nil
	w := runUserSubUsage(t, g, &fakeActiveSubs{subs: []service.UserSubscription{activeSub(10, 1), openai, noGroup}}, 5)
	require.Equal(t, http.StatusOK, w.Code)
	var resp userUsageResp
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	require.Len(t, resp.Data, 1)
	require.Equal(t, int64(1), resp.Data[0].GroupID)
	require.NotContains(t, g.called, int64(2))
	require.NotContains(t, g.called, int64(3))
}
