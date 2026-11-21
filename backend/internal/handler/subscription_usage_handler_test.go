package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type fakeSubUsageGetter struct {
	out *service.SubscriptionUsage
	err error
}

func (f fakeSubUsageGetter) GetForAPIKey(context.Context, *service.APIKey) (*service.SubscriptionUsage, error) {
	return f.out, f.err
}

func (f fakeSubUsageGetter) GetForGroup(context.Context, int64) (*service.SubscriptionUsage, error) {
	return f.out, f.err
}

func runSubUsage(t *testing.T, g subscriptionUsageGetter, withKey bool) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/subscription-usage", nil)
	if withKey {
		c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{})
	}
	(&SubscriptionUsageHandler{svc: g}).Get(c)
	return w
}

func TestSubscriptionUsageHandler_OK(t *testing.T) {
	w := runSubUsage(t, fakeSubUsageGetter{out: &service.SubscriptionUsage{Plan: "max",
		FiveHour: &service.SubscriptionUsageWindow{Utilization: 42}}}, true)
	require.Equal(t, http.StatusOK, w.Code)
	require.JSONEq(t, `{"plan":"max","five_hour":{"utilization":42,"resets_at":null,"remaining_seconds":0}}`, w.Body.String())
}

func TestSubscriptionUsageHandler_ErrorMapping(t *testing.T) {
	cases := []struct {
		err    error
		status int
		typ    string
	}{
		{service.ErrSubscriptionUsageNoGroup, 404, "no_group"},
		{service.ErrSubscriptionUsageNotSubscriptionGroup, 404, "not_subscription_group"},
		{service.ErrSubscriptionUsageNoAvailableAccount, 503, "no_available_account"},
		{fmt.Errorf("%w: account secret-name 401", service.ErrSubscriptionUsageUpstream), 502, "upstream_usage_error"},
		{errors.New("db down"), 500, "api_error"},
	}
	for _, tc := range cases {
		w := runSubUsage(t, fakeSubUsageGetter{err: tc.err}, true)
		require.Equal(t, tc.status, w.Code, tc.typ)
		require.Contains(t, w.Body.String(), `"type":"`+tc.typ+`"`)
		require.NotContains(t, w.Body.String(), "secret-name")
	}
}

func TestSubscriptionUsageHandler_NoAPIKey(t *testing.T) {
	w := runSubUsage(t, fakeSubUsageGetter{}, false)
	require.Equal(t, http.StatusUnauthorized, w.Code)
}
