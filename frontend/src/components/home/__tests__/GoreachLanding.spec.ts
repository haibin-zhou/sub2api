import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount, RouterLinkStub } from '@vue/test-utils'

const getPublicPlans = vi.fn()
vi.mock('@/api/payment', () => ({ paymentAPI: { getPublicPlans: (...args: unknown[]) => getPublicPlans(...args) } }))

import GoreachLanding from '../GoreachLanding.vue'

// Default: API unavailable → landing falls back to the built-in Standard/Premium copy.
getPublicPlans.mockRejectedValue(new Error('offline'))

function mountLanding(props: { isAuthenticated: boolean; dashboardPath: string; siteLogo?: string }) {
  return mount(GoreachLanding, {
    props,
    global: {
      stubs: { RouterLink: RouterLinkStub },
    },
  })
}

function linkTargets(wrapper: ReturnType<typeof mountLanding>, selector?: string) {
  const scope = selector ? wrapper.get(selector) : wrapper
  return scope.findAllComponents(RouterLinkStub).map((link) => link.props('to'))
}

describe('GoreachLanding', () => {
  describe('brand logo', () => {
    it('shows the configured site logo in the nav brand', () => {
      const logo = 'data:image/svg+xml;base64,PHN2Zy8+'
      const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard', siteLogo: logo })
      const img = wrapper.get('header.nav .brand img.brand-logo')
      expect(img.attributes('src')).toBe(logo)
      expect(wrapper.find('header.nav .brand .brand-mark').exists()).toBe(false)
    })

    it('falls back to the built-in mark when no site logo is configured', () => {
      const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard' })
      expect(wrapper.find('header.nav .brand img').exists()).toBe(false)
      expect(wrapper.find('header.nav .brand .brand-mark').exists()).toBe(true)
    })
  })

  describe('guest', () => {
    const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard' })

    it('shows login and register entry points in the nav', () => {
      const nav = linkTargets(wrapper, 'header.nav')
      expect(nav).toContain('/login')
      expect(nav).toContain('/register')
      expect(wrapper.get('header.nav').text()).toContain('登录')
      expect(wrapper.get('header.nav').text()).toContain('注册')
    })

    it('links the brand to /home', () => {
      expect(wrapper.get('header.nav').findComponent(RouterLinkStub).props('to')).toBe('/home')
    })

    it('links fallback pricing buttons to register, then the subscription tab', () => {
      const plans = wrapper.findAll('.plan').map((plan) => plan.findComponent(RouterLinkStub).props('to'))
      const target = `/register?redirect=${encodeURIComponent('/purchase?tab=subscription')}`
      expect(plans).toEqual([target, target])
    })

    it('does not show the dashboard entry', () => {
      expect(wrapper.text()).not.toContain('进入控制台')
    })

    it('keeps in-page anchors as plain links', () => {
      const hrefs = wrapper.findAll('a[href^="#"]').map((a) => a.attributes('href'))
      for (const id of ['#compare', '#pricing', '#how', '#faq', '#reading']) {
        expect(hrefs).toContain(id)
      }
    })

    it('opens external reading links in a new tab', () => {
      const external = wrapper.findAll('a[href^="https://"]')
      expect(external.length).toBe(2)
      for (const a of external) {
        expect(a.attributes('target')).toBe('_blank')
        expect(a.attributes('rel')).toBe('noopener')
      }
    })
  })

  describe('authenticated', () => {
    const wrapper = mountLanding({ isAuthenticated: true, dashboardPath: '/admin/dashboard' })

    it('replaces nav login/register with a dashboard button', () => {
      const nav = linkTargets(wrapper, 'header.nav')
      expect(nav).toContain('/admin/dashboard')
      expect(nav).not.toContain('/login')
      expect(nav).not.toContain('/register')
      const navAuth = wrapper.get('.nav-auth')
      expect(navAuth.text()).toBe('进入控制台')
    })

    it('links fallback pricing buttons straight to the subscription tab', () => {
      const plans = wrapper.findAll('.plan').map((plan) => plan.findComponent(RouterLinkStub))
      expect(plans.map((p) => p.props('to'))).toEqual(['/purchase?tab=subscription', '/purchase?tab=subscription'])
      expect(plans.map((p) => p.text())).toEqual(['订阅 Standard', '订阅 Premium'])
    })

    it('shows the signed-in hint in the hero and dashboard CTA in the closing box', () => {
      expect(wrapper.get('.hero .signin').text()).toContain('已登录')
      expect(linkTargets(wrapper, '.hero .signin')).toEqual(['/admin/dashboard'])
      expect(linkTargets(wrapper, '.closing')).toEqual(['/admin/dashboard'])
      expect(wrapper.get('.closing').text()).toContain('进入控制台')
    })

    it('has no login or register links anywhere', () => {
      // (pricing links go to /purchase for signed-in users)
      const all = linkTargets(wrapper)
      expect(all.some((to) => String(to).startsWith('/login'))).toBe(false)
      expect(all.some((to) => String(to).startsWith('/register'))).toBe(false)
    })
  })

  describe('content', () => {
    const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard' })

    it('renders six FAQ entries as details', () => {
      expect(wrapper.findAll('.faq details')).toHaveLength(6)
    })

    it('renders six comparison rows', () => {
      expect(wrapper.findAll('table.compare tbody tr')).toHaveLength(6)
    })

    it('renders two plans priced $45 and $180', () => {
      const plans = wrapper.findAll('.plan')
      expect(plans).toHaveLength(2)
      expect(plans.map((p) => p.get('.price .num').text())).toEqual(['45', '180'])
    })
  })

  describe('pricing from backend plans', () => {
    const apiPlans = [
      { id: 7, name: 'Premium', description: '约 6.25× Pro 用量', price: 180, original_price: 220, currency: 'USD',
        validity_days: 30, validity_unit: 'days', features: '多项目并行\n优先开通', sort_order: 0 },
      { id: 3, name: 'Standard', description: '约 1.25× Pro 用量', price: 45, original_price: 60, currency: 'USD',
        validity_days: 30, validity_unit: 'days', features: '适合日常单项目', sort_order: 0 },
    ]
    beforeEach(() => {
      getPublicPlans.mockReset().mockResolvedValue({ data: apiPlans })
    })

    it('renders plans from the API (sorted by price within the same sort order)', async () => {
      const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard' })
      await flushPromises()
      const plans = wrapper.findAll('.plan')
      expect(plans.map((p) => p.get('h3').text())).toEqual(['Standard', 'Premium'])
      expect(plans.map((p) => p.get('.price .num').text())).toEqual(['45', '180'])
      expect(plans.map((p) => p.get('.ratio').text())).toEqual(['约 1.25× Pro 用量', '约 6.25× Pro 用量'])
      expect(plans[1].findAll('li').map((li) => li.text())).toEqual(['多项目并行', '优先开通'])
      expect(plans[0].get('.was').text()).toBe('原价 $60 / 月')
    })

    it('marks only the highest-priced plan as featured', async () => {
      const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard' })
      await flushPromises()
      const plans = wrapper.findAll('.plan')
      expect(plans.map((p) => p.classes('featured'))).toEqual([false, true])
      expect(plans[1].find('.plan-badge').exists()).toBe(true)
    })

    it('sends signed-in users to the payment page with the plan preselected', async () => {
      const wrapper = mountLanding({ isAuthenticated: true, dashboardPath: '/dashboard' })
      await flushPromises()
      const links = wrapper.findAll('.plan').map((p) => p.findComponent(RouterLinkStub).props('to'))
      expect(links).toEqual(['/purchase?tab=subscription&plan=3', '/purchase?tab=subscription&plan=7'])
    })

    it('sends guests to register, then back to the chosen plan', async () => {
      const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard' })
      await flushPromises()
      const link = wrapper.findAll('.plan')[0].findComponent(RouterLinkStub).props('to')
      expect(link).toBe(`/register?redirect=${encodeURIComponent('/purchase?tab=subscription&plan=3')}`)
    })

    it('keeps the built-in copy when the API returns no plans', async () => {
      getPublicPlans.mockReset().mockResolvedValue({ data: [] })
      const wrapper = mountLanding({ isAuthenticated: false, dashboardPath: '/dashboard' })
      await flushPromises()
      expect(wrapper.findAll('.plan').map((p) => p.get('h3').text())).toEqual(['Standard', 'Premium'])
    })
  })
})
