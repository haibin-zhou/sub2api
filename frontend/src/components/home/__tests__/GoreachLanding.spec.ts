import { describe, expect, it } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'

import GoreachLanding from '../GoreachLanding.vue'

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

    it('links pricing buttons to register with the chosen plan', () => {
      const plans = wrapper.findAll('.plan').map((plan) => plan.findComponent(RouterLinkStub).props('to'))
      expect(plans).toEqual(['/register?plan=standard', '/register?plan=premium'])
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

    it('links pricing buttons to the dashboard', () => {
      const plans = wrapper.findAll('.plan').map((plan) => plan.findComponent(RouterLinkStub))
      expect(plans.map((p) => p.props('to'))).toEqual(['/admin/dashboard', '/admin/dashboard'])
      expect(plans.map((p) => p.text())).toEqual(['选择 Standard', '选择 Premium'])
    })

    it('shows the signed-in hint in the hero and dashboard CTA in the closing box', () => {
      expect(wrapper.get('.hero .signin').text()).toContain('已登录')
      expect(linkTargets(wrapper, '.hero .signin')).toEqual(['/admin/dashboard'])
      expect(linkTargets(wrapper, '.closing')).toEqual(['/admin/dashboard'])
      expect(wrapper.get('.closing').text()).toContain('进入控制台')
    })

    it('has no login or register links anywhere', () => {
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
})
