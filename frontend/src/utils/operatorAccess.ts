// 运营管理员（role=operator）可进入的后台页面：账号、分组、订阅。
// 后端以接口白名单兜底（middleware.OperatorRouteAllowed），这里只负责菜单与路由的展示一致。
export const OPERATOR_ADMIN_PATHS = ['/admin/accounts', '/admin/groups', '/admin/subscriptions'] as const

export const OPERATOR_HOME_PATH = '/admin/accounts'

export function isOperatorAdminPath(path: string): boolean {
  return OPERATOR_ADMIN_PATHS.some((p) => path === p || path.startsWith(p + '/'))
}
