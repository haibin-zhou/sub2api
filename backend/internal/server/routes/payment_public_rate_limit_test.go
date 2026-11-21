package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler"
	"github.com/Wei-Shaw/sub2api/internal/handler/admin"
	servermiddleware "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newPaymentRoutesTestRouter(redisClient *redis.Client) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	noop := func(c *gin.Context) { c.Next() }

	RegisterPaymentRoutes(
		v1,
		&handler.PaymentHandler{},
		&handler.PaymentWebhookHandler{},
		&admin.PaymentHandler{},
		servermiddleware.JWTAuthMiddleware(noop),
		servermiddleware.AdminAuthMiddleware(noop),
		servermiddleware.AuditLogMiddleware(noop),
		nil,
		nil,
		redisClient,
	)
	return router
}

func postPublicOrderVerify(router *gin.Engine, remoteAddr string) *httptest.ResponseRecorder {
	// Empty body fails binding in the handler, so no service is touched and a
	// non-429 response proves the request passed the limiter.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payment/public/orders/verify", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestPublicOrderVerifyRateLimitedPerIP(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	router := newPaymentRoutesTestRouter(rdb)

	for i := 1; i <= publicOrderVerifyRateLimit; i++ {
		w := postPublicOrderVerify(router, "198.51.100.20:1234")
		require.Equal(t, http.StatusBadRequest, w.Code, "request %d should reach the handler", i)
	}

	w := postPublicOrderVerify(router, "198.51.100.20:1234")
	require.Equal(t, http.StatusTooManyRequests, w.Code)
	require.Contains(t, w.Body.String(), "rate limit exceeded")

	// A different client IP has its own budget.
	w = postPublicOrderVerify(router, "198.51.100.21:1234")
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestPublicOrderVerifyRateLimitFailsOpenWhenRedisUnavailable(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         "127.0.0.1:1",
		DialTimeout:  50 * time.Millisecond,
		ReadTimeout:  50 * time.Millisecond,
		WriteTimeout: 50 * time.Millisecond,
	})
	t.Cleanup(func() { _ = rdb.Close() })

	router := newPaymentRoutesTestRouter(rdb)
	w := postPublicOrderVerify(router, "203.0.113.30:1234")
	require.Equal(t, http.StatusBadRequest, w.Code, "users mid-payment must not be blocked by a Redis outage")
}

func TestPublicPlansRouteIsAnonymousAndRateLimitedPerIP(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	gin.SetMode(gin.TestMode)
	router := gin.New()
	// The zero-value handler has no services; Recovery turns its nil deref into
	// a 500, which still proves the request got past auth and the limiter.
	router.Use(gin.Recovery())
	v1 := router.Group("/api/v1")
	noop := func(c *gin.Context) { c.Next() }
	deny := func(c *gin.Context) { c.AbortWithStatus(http.StatusUnauthorized) }
	RegisterPaymentRoutes(v1, &handler.PaymentHandler{}, &handler.PaymentWebhookHandler{}, &admin.PaymentHandler{},
		servermiddleware.JWTAuthMiddleware(deny), servermiddleware.AdminAuthMiddleware(deny),
		servermiddleware.AuditLogMiddleware(noop), nil, nil, rdb)

	get := func(remoteAddr string) int {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payment/public/plans", nil)
		req.RemoteAddr = remoteAddr
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	for i := 1; i <= publicPlansRateLimit; i++ {
		code := get("198.51.100.30:1234")
		require.NotEqual(t, http.StatusUnauthorized, code, "request %d must not require JWT", i)
		require.NotEqual(t, http.StatusTooManyRequests, code, "request %d should pass the limiter", i)
	}
	require.Equal(t, http.StatusTooManyRequests, get("198.51.100.30:1234"))
	require.NotEqual(t, http.StatusTooManyRequests, get("198.51.100.31:1234"), "other IPs keep their own budget")
}
