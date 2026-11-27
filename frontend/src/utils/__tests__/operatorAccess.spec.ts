import { readFileSync } from 'node:fs'
import { dirname, resolve } from 'node:path'
import { fileURLToPath } from 'node:url'

import { describe, expect, it } from 'vitest'

import { isOperatorAdminPath, OPERATOR_HOME_PATH } from '@/utils/operatorAccess'

const here = dirname(fileURLToPath(import.meta.url))
const routerSource = readFileSync(resolve(here, '../../router/index.ts'), 'utf8')
const sidebarSource = readFileSync(resolve(here, '../../components/layout/AppSidebar.vue'), 'utf8')

describe('operator admin path access', () => {
  it('allows only accounts, groups and subscriptions pages', () => {
    for (const p of ['/admin/accounts', '/admin/groups', '/admin/subscriptions', '/admin/accounts/123']) {
      expect(isOperatorAdminPath(p)).toBe(true)
    }
    for (const p of ['/admin/settings', '/admin/dashboard', '/admin/users', '/admin/proxies', '/admin/backups', '/admin/accountsX', '/admin/ops']) {
      expect(isOperatorAdminPath(p)).toBe(false)
    }
  })

  it('lands operators on account management', () => {
    expect(OPERATOR_HOME_PATH).toBe('/admin/accounts')
    expect(isOperatorAdminPath(OPERATOR_HOME_PATH)).toBe(true)
  })

  it('router guard redirects operators away from non-whitelisted admin pages', () => {
    expect(routerSource).toMatch(/if \(authStore\.isOperator\) \{\n\s*if \(!isOperatorAdminPath\(to\.path\)\) \{\n\s*next\(OPERATOR_HOME_PATH\)/)
  })

  it('sends operators from the user dashboard (default login target) to account management', () => {
    expect(routerSource).toMatch(/if \(authStore\.isOperator && to\.path === '\/dashboard'\) \{\n\s*next\(OPERATOR_HOME_PATH\)/)
  })

  it('sidebar shows the admin section to operators but filters it to whitelisted pages', () => {
    expect(sidebarSource).toContain('<template v-if="hasAdminPanel">')
    expect(sidebarSource).toMatch(/if \(isOperator\.value\) \{\n\s*return visible\.filter\(item => isOperatorAdminPath\(item\.path\)\)/)
  })
})
