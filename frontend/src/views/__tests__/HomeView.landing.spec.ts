import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount, RouterLinkStub } from '@vue/test-utils'

import HomeView from '../HomeView.vue'
import GoreachLanding from '@/components/home/GoreachLanding.vue'

const { appStore, authStore } = vi.hoisted(() => ({
  appStore: {
    cachedPublicSettings: {} as Record<string, unknown>,
    siteName: 'Fallback site',
    siteLogo: '',
    docUrl: '',
    publicSettingsLoaded: true,
    fetchPublicSettings: vi.fn(),
  },
  authStore: {
    isAuthenticated: false,
    isAdmin: false,
    user: null as { email?: string } | null,
    checkAuth: vi.fn(),
  },
}))

vi.mock('@/stores', () => ({
  useAppStore: () => appStore,
  useAuthStore: () => authStore,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => appStore,
}))

vi.mock('vue-i18n', async (importOriginal) => {
  const actual = await importOriginal<typeof import('vue-i18n')>()
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

function mountHome(settings: Record<string, unknown> = {}) {
  appStore.cachedPublicSettings = {
    site_name: 'Test site',
    site_subtitle: 'Test subtitle',
    ...settings,
  }

  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: { template: '<div data-testid="locale-switcher" />' },
        Icon: { template: '<span data-testid="icon" />' },
      },
    },
  })
}

describe('HomeView default mode (Goreach landing)', () => {
  beforeEach(() => {
    authStore.isAuthenticated = false
    authStore.isAdmin = false
    authStore.user = null
    authStore.checkAuth.mockClear()
    appStore.publicSettingsLoaded = true
    appStore.fetchPublicSettings.mockClear()
    localStorage.clear()
    vi.spyOn(window, 'matchMedia').mockReturnValue({ matches: false } as MediaQueryList)
  })

  it('renders GoreachLanding for guests', () => {
    const wrapper = mountHome()
    const landing = wrapper.findComponent(GoreachLanding)

    expect(landing.exists()).toBe(true)
    expect(landing.props()).toEqual({ isAuthenticated: false, dashboardPath: '/dashboard', siteLogo: '' })
    expect(wrapper.find('[data-testid="compact-home"]').exists()).toBe(false)
  })

  it('passes the configured site logo to GoreachLanding', () => {
    const logo = 'data:image/svg+xml;base64,PHN2Zy8+'
    const wrapper = mountHome({ site_logo: logo })

    expect(wrapper.findComponent(GoreachLanding).props('siteLogo')).toBe(logo)
  })

  it('passes the admin dashboard path to authenticated administrators', () => {
    authStore.isAuthenticated = true
    authStore.isAdmin = true

    const landing = mountHome({ compact_home_enabled: false }).findComponent(GoreachLanding)
    expect(landing.props()).toEqual({ isAuthenticated: true, dashboardPath: '/admin/dashboard', siteLogo: '' })
  })

  it('passes the user dashboard path to authenticated users', () => {
    authStore.isAuthenticated = true

    const landing = mountHome().findComponent(GoreachLanding)
    expect(landing.props()).toEqual({ isAuthenticated: true, dashboardPath: '/dashboard', siteLogo: '' })
  })

  it('still checks auth and loads public settings on mount', () => {
    appStore.publicSettingsLoaded = false
    mountHome()

    expect(authStore.checkAuth).toHaveBeenCalledOnce()
    expect(appStore.fetchPublicSettings).toHaveBeenCalledOnce()
  })

  it('does not render the landing in custom content mode', () => {
    const wrapper = mountHome({ home_content: '<section id="custom-home">Custom home</section>' })

    expect(wrapper.get('#custom-home').text()).toBe('Custom home')
    expect(wrapper.findComponent(GoreachLanding).exists()).toBe(false)
  })

  it('does not render the landing in compact mode', () => {
    const wrapper = mountHome({ compact_home_enabled: true })

    expect(wrapper.find('[data-testid="compact-home"]').exists()).toBe(true)
    expect(wrapper.findComponent(GoreachLanding).exists()).toBe(false)
  })
})
