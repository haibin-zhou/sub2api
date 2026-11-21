import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import SubscriptionsView from '../SubscriptionsView.vue'

const { getMySubscriptions, getSubscriptionUsage, getSubscriptionPeriodUsage, showError } = vi.hoisted(() => ({
  getMySubscriptions: vi.fn(),
  getSubscriptionUsage: vi.fn(),
  getSubscriptionPeriodUsage: vi.fn(),
  showError: vi.fn(),
}))

vi.mock('@/api/subscriptions', () => ({
  default: { getMySubscriptions, getSubscriptionUsage, getSubscriptionPeriodUsage },
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
        params ? `${key}(${Object.values(params).join('|')})` : key,
    }),
  }
})

function sub(id: number, groupId: number, platform = 'openai', status = 'active') {
  return {
    id,
    user_id: 1,
    group_id: groupId,
    status,
    starts_at: '2026-09-01T00:00:00Z',
    daily_usage_usd: 0,
    weekly_usage_usd: 0,
    monthly_usage_usd: 0,
    daily_window_start: null,
    weekly_window_start: null,
    monthly_window_start: null,
    created_at: '2026-09-01T00:00:00Z',
    updated_at: '2026-09-01T00:00:00Z',
    expires_at: null,
    group: { id: groupId, name: `G${groupId}`, platform, rate_multiplier: 1 },
  }
}

// Local-time instants so MM-DD assertions are timezone independent.
const localISO = (y: number, mo: number, d: number, h = 0, mi = 0) => new Date(y, mo - 1, d, h, mi).toISOString()

function period(subscriptionId: number, groupId: number, extra: Record<string, number> = {}) {
  return {
    subscription_id: subscriptionId,
    group_id: groupId,
    period_index: 3,
    period_start: localISO(2026, 10, 31, 8, 0),
    period_end: localISO(2026, 11, 30, 8, 0),
    request_count: 0,
    input_tokens: 0,
    output_tokens: 0,
    cache_creation_tokens: 0,
    cache_read_tokens: 0,
    total_cost_usd: 0,
    ...extra,
  }
}

function mountView() {
  return mount(SubscriptionsView, {
    global: { stubs: { AppLayout: { template: '<div><slot /></div>' }, Icon: true } },
  })
}

describe('SubscriptionsView current-period usage', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers({ toFake: ['Date'] })
    vi.setSystemTime(new Date(2026, 10, 21, 10, 0))
    vi.spyOn(console, 'error').mockImplementation(() => {})
    getSubscriptionUsage.mockResolvedValue([])
  })

  afterEach(() => {
    vi.useRealTimers()
    vi.restoreAllMocks()
  })

  it('renders amount, period and compact counts for every active subscription (all platforms)', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1, 'anthropic'), sub(11, 2, 'openai')])
    getSubscriptionPeriodUsage.mockResolvedValue([
      period(10, 1, {
        request_count: 1753,
        input_tokens: 1_234_567,
        output_tokens: 35_420,
        cache_creation_tokens: 999,
        cache_read_tokens: 2_500_000_000,
        total_cost_usd: 1234.5,
      }),
      period(11, 2, { request_count: 2, input_tokens: 10, total_cost_usd: 0.004 }),
    ])
    const wrapper = mountView()
    await flushPromises()

    const block = wrapper.get('[data-testid="period-usage-10"]')
    const text = block.text()
    expect(text).toContain('userSubscriptions.periodUsage.title')
    // end date is shown as the inclusive last day (period_end - 1 day)
    expect(text).toContain('userSubscriptions.periodUsage.period(3|10-31|11-29)')
    expect(text).toContain('userSubscriptions.periodUsage.amount($1,234.50)')
    expect(text).toContain('userSubscriptions.periodUsage.summary(1753|1.2M|35.4K|999|2.5B)')
    expect(text).not.toContain('userSubscriptions.periodUsage.empty')

    const other = wrapper.get('[data-testid="period-usage-11"]').text()
    expect(other).toContain('userSubscriptions.periodUsage.amount($0.00)')
    expect(other).toContain('userSubscriptions.periodUsage.summary(2|10|0|0|0)')
    expect(getSubscriptionPeriodUsage).toHaveBeenCalledTimes(1)
  })

  it('shows $0.00 and the empty hint when there is no usage this period', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1)])
    getSubscriptionPeriodUsage.mockResolvedValue([period(10, 1)])
    const wrapper = mountView()
    await flushPromises()
    const text = wrapper.get('[data-testid="period-usage-10"]').text()
    expect(text).toContain('userSubscriptions.periodUsage.amount($0.00)')
    expect(text).toContain('userSubscriptions.periodUsage.empty')
    expect(text).not.toContain('userSubscriptions.periodUsage.summary')
  })

  it('does not render the block for inactive subscriptions and skips the request when none are active', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1, 'openai', 'expired')])
    const wrapper = mountView()
    await flushPromises()
    expect(getSubscriptionPeriodUsage).not.toHaveBeenCalled()
    expect(wrapper.find('[data-testid="period-usage-10"]').exists()).toBe(false)
  })

  it('shows a single notice on failure while subscriptions still render', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1), sub(11, 2)])
    getSubscriptionPeriodUsage.mockRejectedValue(new Error('boom'))
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.text()).toContain('G1')
    expect(wrapper.text()).toContain('G2')
    expect(wrapper.findAll('[data-testid="period-usage-notice"]')).toHaveLength(1)
    expect(wrapper.find('[data-testid="period-usage-10"]').exists()).toBe(false)
    expect(showError).not.toHaveBeenCalled()
  })

  it('shows per-card loading, loads both datasets in parallel, and refresh re-requests both', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1, 'anthropic')])
    let resolvePeriod: (v: unknown) => void = () => {}
    getSubscriptionPeriodUsage.mockReturnValueOnce(new Promise((r) => { resolvePeriod = r }))
    getSubscriptionUsage.mockReturnValueOnce(new Promise(() => {}))
    const wrapper = mountView()
    await flushPromises()
    // both requests are in flight at the same time
    expect(getSubscriptionUsage).toHaveBeenCalledTimes(1)
    expect(getSubscriptionPeriodUsage).toHaveBeenCalledTimes(1)
    expect(wrapper.find('[data-testid="period-usage-loading-10"]').exists()).toBe(true)

    resolvePeriod([period(10, 1, { request_count: 1, total_cost_usd: 1 })])
    await flushPromises()
    expect(wrapper.get('[data-testid="period-usage-10"]').text()).toContain('$1.00')
  })

  it('refresh triggers both requests and clears a previous failure', async () => {
    getMySubscriptions.mockResolvedValue([sub(10, 1, 'anthropic')])
    getSubscriptionPeriodUsage.mockRejectedValueOnce(new Error('boom'))
    const wrapper = mountView()
    await flushPromises()
    expect(wrapper.find('[data-testid="period-usage-notice"]').exists()).toBe(true)

    getSubscriptionPeriodUsage.mockResolvedValueOnce([period(10, 1, { request_count: 5, total_cost_usd: 12.345 })])
    await wrapper.get('[data-testid="official-usage-refresh"]').trigger('click')
    await flushPromises()
    expect(getSubscriptionUsage).toHaveBeenCalledTimes(2)
    expect(getSubscriptionPeriodUsage).toHaveBeenCalledTimes(2)
    expect(wrapper.find('[data-testid="period-usage-notice"]').exists()).toBe(false)
    expect(wrapper.get('[data-testid="period-usage-10"]').text()).toContain('$12.35')
  })

  it('shows refresh for non-anthropic active subscriptions and only refreshes period usage', async () => {
    getMySubscriptions.mockResolvedValue([sub(11, 2, 'openai')])
    getSubscriptionPeriodUsage.mockResolvedValue([period(11, 2)])
    const wrapper = mountView()
    await flushPromises()
    await wrapper.get('[data-testid="official-usage-refresh"]').trigger('click')
    await flushPromises()
    expect(getSubscriptionUsage).not.toHaveBeenCalled()
    expect(getSubscriptionPeriodUsage).toHaveBeenCalledTimes(2)
  })
})
