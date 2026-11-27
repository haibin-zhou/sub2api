//go:build unit

package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestOperatorRouteAllowed(t *testing.T) {
	allowed := []string{
		"GET /api/v1/admin/accounts",
		"POST /api/v1/admin/accounts",
		"PUT /api/v1/admin/accounts/:id",
		"DELETE /api/v1/admin/accounts/:id",
		"POST /api/v1/admin/accounts/generate-auth-url",
		"POST /api/v1/admin/accounts/exchange-code",
		"POST /api/v1/admin/accounts/:id/refresh",
		"GET /api/v1/admin/groups",
		"GET /api/v1/admin/groups/all",
		"POST /api/v1/admin/groups",
		"PUT /api/v1/admin/groups/:id",
		"DELETE /api/v1/admin/groups/:id",
		"GET /api/v1/admin/groups/:id/subscriptions",
		"GET /api/v1/admin/subscriptions",
		"POST /api/v1/admin/subscriptions/assign",
		"POST /api/v1/admin/subscriptions/:id/revoke",
		"DELETE /api/v1/admin/subscriptions/:id",
		"GET /api/v1/admin/users",
		"GET /api/v1/admin/users/:id",
		"GET /api/v1/admin/users/:id/subscriptions",
		"GET /api/v1/admin/proxies",
		"GET /api/v1/admin/proxies/all",
		"GET /api/v1/admin/proxies/:id",
		"GET /api/v1/admin/compliance",
		"POST /api/v1/admin/compliance/accept",
		"POST /api/v1/admin/openai/generate-auth-url",
		"POST /api/v1/admin/gemini/oauth/auth-url",
	}
	for _, route := range allowed {
		method, path := splitRoute(route)
		require.Truef(t, OperatorRouteAllowed(method, path), "operator should be allowed: %s", route)
	}

	denied := []string{
		// 系统设置及其子接口
		"GET /api/v1/admin/settings",
		"PUT /api/v1/admin/settings",
		"GET /api/v1/admin/settings/web-search-emulation",
		"GET /api/v1/admin/settings/admin-api-keys",
		// 用户：只读，禁止改资料/角色/余额
		"POST /api/v1/admin/users",
		"PUT /api/v1/admin/users/:id",
		"DELETE /api/v1/admin/users/:id",
		"POST /api/v1/admin/users/:id/balance",
		"GET /api/v1/admin/users/:id/balance-history",
		// 代理：只读
		"POST /api/v1/admin/proxies",
		"PUT /api/v1/admin/proxies/:id",
		"DELETE /api/v1/admin/proxies/:id",
		"GET /api/v1/admin/proxies/data",
		// 账号：批量导出/导入凭据、全局设置类接口
		"GET /api/v1/admin/accounts/data",
		"POST /api/v1/admin/accounts/data",
		"POST /api/v1/admin/accounts/batch-update-credentials",
		"POST /api/v1/admin/accounts/sync/crs",
		"POST /api/v1/admin/accounts/sync/crs/preview",
		"PUT /api/v1/admin/accounts/upstream-billing-probe/settings",
		"PUT /api/v1/admin/accounts/ollama-cloud-usage/settings",
		"PUT /api/v1/admin/accounts/opencode-go-usage/settings",
		// 其余后台模块
		"GET /api/v1/admin/dashboard/stats",
		"GET /api/v1/admin/backups",
		"POST /api/v1/admin/data-management/export",
		"GET /api/v1/admin/system/version",
		"GET /api/v1/admin/ops/concurrency",
		"GET /api/v1/admin/audit-logs",
		"GET /api/v1/admin/payment/orders",
		"GET /api/v1/admin/plugins",
		"GET /api/v1/admin/redeem-codes",
		// 前缀不能误匹配
		"GET /api/v1/admin/accountsX",
		"GET /api/v1/admin/groupsettings",
		"",
	}
	for _, route := range denied {
		method, path := splitRoute(route)
		require.Falsef(t, OperatorRouteAllowed(method, path), "operator should be denied: %s", route)
	}
}

func splitRoute(route string) (string, string) {
	for i, ch := range route {
		if ch == ' ' {
			return route[:i], route[i+1:]
		}
	}
	return route, ""
}

func TestAdminAuthOperatorRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &config.Config{JWT: config.JWTConfig{Secret: "test-secret", ExpireHour: 1}}
	authService := service.NewAuthService(nil, nil, nil, nil, cfg, nil, nil, nil, nil, nil, nil, nil, nil)

	users := map[int64]*service.User{
		1: {ID: 1, Email: "admin@example.com", Role: service.RoleAdmin, Status: service.StatusActive, Concurrency: 1},
		2: {ID: 2, Email: "op@example.com", Role: service.RoleOperator, Status: service.StatusActive, Concurrency: 1},
		3: {ID: 3, Email: "user@example.com", Role: service.RoleUser, Status: service.StatusActive, Concurrency: 1},
	}
	userRepo := &stubUserRepo{
		getByID: func(ctx context.Context, id int64) (*service.User, error) {
			u, ok := users[id]
			if !ok {
				return nil, service.ErrUserNotFound
			}
			clone := *u
			return &clone, nil
		},
	}
	userService := service.NewUserService(userRepo, nil, nil, nil)

	router := gin.New()
	admin := router.Group("/api/v1/admin")
	admin.Use(gin.HandlerFunc(NewAdminAuthMiddleware(authService, userService, nil, nil)))
	ok := func(c *gin.Context) {
		role, _ := GetUserRoleFromContext(c)
		c.JSON(http.StatusOK, gin.H{"role": role})
	}
	admin.GET("/accounts", ok)
	admin.GET("/settings", ok)
	admin.PUT("/users/:id", ok)

	call := func(t *testing.T, userID int64, method, path string) *httptest.ResponseRecorder {
		t.Helper()
		token, err := authService.GenerateToken(context.Background(), users[userID])
		require.NoError(t, err)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		router.ServeHTTP(w, req)
		return w
	}

	t.Run("operator_allowed_route", func(t *testing.T) {
		w := call(t, 2, http.MethodGet, "/api/v1/admin/accounts")
		require.Equal(t, http.StatusOK, w.Code)
		require.Contains(t, w.Body.String(), `"role":"operator"`)
	})
	t.Run("operator_denied_settings", func(t *testing.T) {
		w := call(t, 2, http.MethodGet, "/api/v1/admin/settings")
		require.Equal(t, http.StatusForbidden, w.Code)
		require.Contains(t, w.Body.String(), "OPERATOR_FORBIDDEN")
	})
	t.Run("operator_denied_user_update", func(t *testing.T) {
		w := call(t, 2, http.MethodPut, "/api/v1/admin/users/3")
		require.Equal(t, http.StatusForbidden, w.Code)
	})
	t.Run("admin_still_allowed_settings", func(t *testing.T) {
		w := call(t, 1, http.MethodGet, "/api/v1/admin/settings")
		require.Equal(t, http.StatusOK, w.Code)
	})
	t.Run("regular_user_still_denied", func(t *testing.T) {
		w := call(t, 3, http.MethodGet, "/api/v1/admin/accounts")
		require.Equal(t, http.StatusForbidden, w.Code)
	})
}
