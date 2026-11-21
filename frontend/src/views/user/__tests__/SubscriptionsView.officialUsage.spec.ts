import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import SubscriptionsView from '../SubscriptionsView.vue'

const { getMySubscriptions, getSubscriptionUsage, showError } = vi.hoisted(() => ({
  getMySubscriptions: vi.fn(),
  getSubscriptionUsage: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api/subscriptions', () => ({
  default: { getMySubscriptions, getSubscriptionUsage },
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, cachedPublicSettings: null }),
}))
vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn() }),
}))
vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string, params?: Record<string, unknown>) =>
        params ? `${key}(${Object.values(params).join(',')})` : key,
    }),
  }
})

function sub(id: number, groupId: number, limits: Record<string, number> = {}, platform = 'anthropic') {
  return {
    id,
    user_id: 1,
    group_id: groupId,
    status: 'active',
    starts_at: '2026-11-01T00:00:00Z',
    daily_usage_usd: 1,
    weekly_usage_usd: 2,
    monthly_usage_usd: 3,
    daily_window_start: null,
    weekly_window_start: null,
    monthly_window_start: null,
    created_at: '2026-11-01T00:00:00Z',
    updated_at: '2026-11-01T00:00:00Z',
    expires_at: null,
    group: { id: groupId, name: `G${groupId}`, platform, rate_multiplier: 1, ...limits },
  }
}

const localISO = (y: number, mo: number, d: number, h: number, mi: number) =>
  new Date(y, mo - 1, d, h, mi).toISOString()

function mountView() {
  return mount(SubscriptionsView, {
    global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Icon: true } },
  })
}

describe('SubscriptionsView official subscription usage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date(2026, 10, 21, 10, 0))
    vi.spyOn(console, 'error').mockImplementation(() => {})
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('renders windows with percentages, reset time, plan and updated time; hides unlimited box', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1)])
    getSubscriptionUsage.mockResolvedValue([
      {
        group_id: 1,
        usage: {
          plan: 'max',
          five_hour: { utilization: 42.4, resets_at: localISO(2026, 11, 21, 14, 30), remaining_seconds: 0 },
          seven_day: { utilization: 130, resets_at: localISO(2026, 11, 25, 9, 5), remaining_seconds: 0 },
          seven_day_fable: { utilization: 7.6, resets_at: null, remaining_seconds: 0 },
          updated_at: localISO(2026, 11, 21, 9, 58),
        },
      },
    ])
    const wrapper = mountView()
    await flushPromises()

    const section = wrapper.get('[data-testid="official-usage-1"]')
    const text = section.text()
    expect(text).toContain('userSubscriptions.officialUsage.title')
    expect(text).toContain('userSubscriptions.officialUsage.fiveHour')
    expect(text).toContain('42%')
    expect(text).toContain('userSubscriptions.officialUsage.resetsAt(14:30)')
    expect(text).toContain('userSubscriptions.officialUsage.sevenDay')
    expect(text).toContain('130%')
    expect(text).toContain('userSubscriptions.officialUsage.resetsAt(11-25 09:05)')
    expect(text).toContain('userSubscriptions.officialUsage.sevenDayFable')
    expect(text).toContain('8%')
    expect(text).not.toContain('userSubscriptions.officialUsage.sevenDaySonnet')
    expect(text).toContain('max')
    expect(text).toContain('userSubscriptions.officialUsage.updatedAt(09:58)')

    const bars = section.findAll('[data-testid="official-usage-bar"]')
    expect(bars).toHaveLength(3)
    expect(bars[0].attributes('style')).toContain('width: 42.4%')
    expect(bars[1].attributes('style')).toContain('width: 100%')
    expect(bars[1].classes()).toContain('bg-red-500')
    expect(bars[0].classes()).toContain('bg-green-500')

    expect(wrapper.text()).not.toContain('userSubscriptions.unlimitedDesc')
  })

  it('shows friendly text for per-item errors and keeps unlimited box when no usage', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1), sub(11, 2), sub(12, 3)])
    getSubscriptionUsage.mockResolvedValue([
      { group_id: 1, usage: null, error: { type: 'not_subscription_group', message: 'x' } },
      { group_id: 2, usage: null, error: { type: 'no_available_account', message: 'x' } },
      { group_id: 3, usage: null, error: { type: 'upstream_usage_error', message: 'x' } },
    ])
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.get('[data-testid="official-usage-1"]').text()).toContain('userSubscriptions.officialUsage.provisioning')
    expect(wrapper.get('[data-testid="official-usage-2"]').text()).toContain('userSubscriptions.officialUsage.noAvailableAccount')
    expect(wrapper.get('[data-testid="official-usage-3"]').text()).toContain('userSubscriptions.officialUsage.unavailable')
    expect(wrapper.findAll('[data-testid="unlimited-box"]')).toHaveLength(3)
  })

  it('keeps USD limit bars alongside official usage', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1, { daily_limit_usd: 10 })])
    getSubscriptionUsage.mockResolvedValue([
      { group_id: 1, usage: { five_hour: { utilization: 5, resets_at: null, remaining_seconds: 0 } } },
    ])
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('userSubscriptions.daily')
    expect(wrapper.find('[data-testid="official-usage-1"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="unlimited-box"]').exists()).toBe(false)
  })

  it('renders official usage only for anthropic groups; others keep the unlimited box', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1), sub(11, 2, {}, 'openai'), sub(12, 3, {}, 'gemini')])
    getSubscriptionUsage.mockResolvedValue([
      { group_id: 1, usage: { five_hour: { utilization: 20, resets_at: null, remaining_seconds: 0 } } },
      { group_id: 2, usage: { five_hour: { utilization: 30, resets_at: null, remaining_seconds: 0 } } },
      { group_id: 3, usage: null, error: { type: 'not_subscription_group', message: 'x' } },
    ])
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('[data-testid="official-usage-1"]').exists()).toBe(true)
    expect(wrapper.find('[data-testid="official-usage-2"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="official-usage-3"]').exists()).toBe(false)
    expect(wrapper.findAll('[data-testid="unlimited-box"]')).toHaveLength(2)
    expect(wrapper.text()).not.toContain('30%')
  })

  it('does not request official usage or show refresh when there are no anthropic subscriptions', async () => {
    getMySubscriptions.mockResolvedValue([sub(11, 2, {}, 'openai')])
    const wrapper = mountView()
    await flushPromises()
    expect(getSubscriptionUsage).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="official-usage-refresh"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="unlimited-box"]').exists()).toBe(true)
  })

  it('shows a single notice when the usage request fails; subscriptions still render', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1), sub(11, 2)])
    getSubscriptionUsage.mockRejectedValue(new Error('boom'))
    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.text()).toContain('G1')
    expect(wrapper.text()).toContain('G2')
    expect(wrapper.findAll('[data-testid="official-usage-notice"]')).toHaveLength(1)
    expect(wrapper.findAll('[data-testid="unlimited-box"]')).toHaveLength(2)
    expect(showError).not.toHaveBeenCalled()
  })

  it('refresh after a failure clears the notice and shows the per-card spinner', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1)])
    getSubscriptionUsage.mockRejectedValueOnce(new Error('boom'))
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('[data-testid="official-usage-notice"]').exists()).toBe(true)

    let resolveUsage: (v: unknown) => void = () => {}
    getSubscriptionUsage.mockReturnValueOnce(new Promise((r) => { resolveUsage = r }))
    await wrapper.get('[data-testid="official-usage-refresh"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-testid="official-usage-notice"]').exists()).toBe(false)
    expect(wrapper.find('[data-testid="official-usage-loading-1"]').exists()).toBe(true)

    resolveUsage([{ group_id: 1, usage: { five_hour: { utilization: 33, resets_at: null, remaining_seconds: 0 } } }])
    await flushPromises()
    expect(wrapper.get('[data-testid="official-usage-1"]').text()).toContain('33%')
  })

  it('does not show error text for non-anthropic cards even if an item is returned', async () => {
    getMySubscriptions.mockResolvedValue([sub(11, 2, {}, 'openai')])
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).not.toContain('userSubscriptions.officialUsage')
  })

  it('shows loading per card, then refresh re-calls the endpoint', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1)])
    let resolveUsage: (v: unknown) => void = () => {}
    getSubscriptionUsage.mockReturnValueOnce(new Promise((r) => { resolveUsage = r }))
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('G1')
    expect(wrapper.find('[data-testid="official-usage-loading-1"]').exists()).toBe(true)

    resolveUsage([{ group_id: 1, usage: { five_hour: { utilization: 10, resets_at: null, remaining_seconds: 0 } } }])
    await flushPromises()
    expect(wrapper.find('[data-testid="official-usage-loading-1"]').exists()).toBe(false)

    getSubscriptionUsage.mockResolvedValueOnce([
      { group_id: 1, usage: { five_hour: { utilization: 55, resets_at: null, remaining_seconds: 0 } } },
    ])
    await wrapper.get('[data-testid="official-usage-refresh"]').trigger('click')
    await flushPromises()
    expect(getSubscriptionUsage).toHaveBeenCalledTimes(2)
    expect(wrapper.get('[data-testid="official-usage-1"]').text()).toContain('55%')
  })
})
