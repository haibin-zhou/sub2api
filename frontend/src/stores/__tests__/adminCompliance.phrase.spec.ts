import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import backendComplianceSource from '../../../../backend/internal/service/admin_compliance.go?raw'
import { useAdminComplianceStore } from '../adminCompliance'

const locale = vi.hoisted(() => ({ current: 'en' }))
vi.mock('@/api/admin/compliance', () => ({ default: { getStatus: vi.fn(), accept: vi.fn() } }))
vi.mock('@/i18n', () => ({ getLocale: () => locale.current }))

function backendConstant(name: string): string {
  const match = backendComplianceSource.match(new RegExp(`${name}\\s*=\\s*"([^"]+)"`))
  if (!match) throw new Error(`backend constant ${name} not found`)
  return match[1]
}

describe('admin compliance fallback acknowledgement phrase', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('is brand-neutral and byte-identical to the backend constants (EN)', () => {
    locale.current = 'en'
    const store = useAdminComplianceStore()
    expect(store.expectedPhrase).toBe(backendConstant('AdminComplianceAckPhraseEN'))
    expect(store.expectedPhrase).toContain('this platform')
    expect(store.expectedPhrase).not.toContain('Habifying')
  })

  it('is brand-neutral and byte-identical to the backend constants (ZH)', () => {
    locale.current = 'zh'
    const store = useAdminComplianceStore()
    expect(store.expectedPhrase).toBe(backendConstant('AdminComplianceAckPhraseZH'))
    expect(store.expectedPhrase).toContain('本平台')
    expect(store.expectedPhrase).not.toContain('Habifying')
  })
})
