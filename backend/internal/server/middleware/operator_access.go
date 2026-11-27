package middleware

import "strings"

// 运营管理员（role=operator）只能操作账号、分组、订阅三个模块，外加完成这些工作所需的只读查询。
// 采用白名单：未登记的后台模块（系统设置、备份、数据管理、插件、支付、运维、审计等）一律拒绝，
// 上游新增的后台模块默认也不对运营管理员开放。

const adminRoutePrefix = "/api/v1/admin"

// operatorModulePrefixes 整个模块（任意方法）开放给运营管理员。
var operatorModulePrefixes = []string{
	"/accounts",
	"/groups",
	"/subscriptions",
	"/compliance",
	// 各平台 OAuth 建号与账号级操作
	"/openai",
	"/gemini",
	"/antigravity",
	"/grok",
}

// operatorReadOnlyRoutes 仅开放的只读接口（方法 + Gin 完整路径，精确匹配）。
var operatorReadOnlyRoutes = map[string]struct{}{
	// 绑定订阅时查找用户；不开放用户修改，避免运营管理员改角色/余额。
	"GET /users":                   {},
	"GET /users/:id":               {},
	"GET /users/:id/subscriptions": {},
	// 建号时选择代理
	"GET /proxies":     {},
	"GET /proxies/all": {},
	"GET /proxies/:id": {},
	// 国产供应商账号额度查询
	"GET /cn-providers/accounts/:id/quota":   {},
	"GET /cn-providers/accounts/:id/balance": {},
}

// operatorDeniedRoutes 在已开放模块内仍禁止的接口：批量导出/导入凭据及全局设置类开关。
var operatorDeniedRoutes = map[string]struct{}{
	"GET /accounts/data":                            {},
	"POST /accounts/data":                           {},
	"POST /accounts/batch-update-credentials":       {},
	"POST /accounts/sync/crs":                       {},
	"POST /accounts/sync/crs/preview":               {},
	"PUT /accounts/upstream-billing-probe/settings": {},
	"PUT /accounts/ollama-cloud-usage/settings":     {},
	"PUT /accounts/opencode-go-usage/settings":      {},
}

// OperatorRouteAllowed 判断运营管理员能否访问某个后台接口。fullPath 为 Gin 的 c.FullPath()。
func OperatorRouteAllowed(method, fullPath string) bool {
	if !strings.HasPrefix(fullPath, adminRoutePrefix+"/") {
		return false
	}
	rel := strings.TrimPrefix(fullPath, adminRoutePrefix)
	key := strings.ToUpper(method) + " " + rel
	if _, denied := operatorDeniedRoutes[key]; denied {
		return false
	}
	if _, ok := operatorReadOnlyRoutes[key]; ok {
		return true
	}
	for _, prefix := range operatorModulePrefixes {
		if rel == prefix || strings.HasPrefix(rel, prefix+"/") {
			return true
		}
	}
	return false
}
