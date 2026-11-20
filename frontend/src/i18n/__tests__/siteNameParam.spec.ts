import { afterEach, beforeAll, beforeEach, describe, expect, it } from 'vitest'
import { defineComponent, nextTick } from 'vue'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { useI18n } from 'vue-i18n'
import { i18n, loadLocaleMessages } from '@/i18n'
import { getAdminSteps, getUserSteps } from '@/components/Guide/steps'
import { useAppStore } from '@/stores/app'
import { DEFAULT_SITE_NAME, setBrandSiteName } from '@/utils/branding'
import type { PublicSettings } from '@/types'

const t = (...args: unknown[]) => (i18n.global.t as (...a: unknown[]) => string)(...args)

describe('i18n {siteName} default parameter', () => {
  beforeAll(async () => {
    await loadLocaleMessages('en')
    await loadLocaleMessages('zh')
  })

  beforeEach(() => {
    i18n.global.locale.value = 'en'
    setBrandSiteName(undefined)
  })

  afterEach(() => {
    setBrandSiteName(undefined)
    delete window.__APP_CONFIG__
  })

  it('falls back to Sub2API when no site name is configured', () => {
    expect(DEFAULT_SITE_NAME).toBe('Sub2API')
    expect(t('onboarding.user.welcome.title')).toBe('👋 Welcome to Sub2API')
    expect(t('setup.title')).toBe('Sub2API Setup')
    i18n.global.locale.value = 'zh'
    expect(t('onboarding.admin.welcome.title')).toBe('👋 欢迎使用 Sub2API')
  })

  it('renders onboarding strings with the configured site name (en + zh)', () => {
    setBrandSiteName('goreach.es')
    expect(t('onboarding.admin.welcome.title')).toBe('👋 Welcome to goreach.es')
    expect(t('onboarding.user.welcome.description')).toContain('Welcome to the goreach.es AI service platform.')
    i18n.global.locale.value = 'zh'
    expect(t('onboarding.user.welcome.title')).toBe('👋 欢迎使用 goreach.es')
    expect(t('onboarding.user.welcome.description')).toContain('欢迎来到 goreach.es AI 服务平台')
    expect(t('admin.settings.linuxdo.description')).toBe('配置 LinuxDo Connect OAuth，用于 goreach.es 用户登录')
  })

  it('never leaks a literal {siteName} or the old brand into tour steps', () => {
    setBrandSiteName('goreach.es')
    for (const locale of ['en', 'zh'] as const) {
      i18n.global.locale.value = locale
      const steps = [...getAdminSteps(t as (key: string) => string), ...getUserSteps(t as (key: string) => string)]
      const text = steps.map((step) => `${step.popover?.title ?? ''}${step.popover?.description ?? ''}`).join('\n')
      expect(text).not.toContain('{siteName}')
      expect(text).not.toContain('Habifying')
      expect(text).toContain('goreach.es')
    }
  })

  it('keeps explicit siteName params and other call signatures working', () => {
    setBrandSiteName('goreach.es')
    expect(t('auth.signUpToStart', { siteName: 'Explicit' })).toBe('Sign up to start using Explicit')
    expect(t('auth.signUpToStart')).toBe('Sign up to start using goreach.es')
    expect(t('common.time.daysAgo', { n: 3 })).toBe(t('common.time.daysAgo', { n: 3, siteName: 'x' }))
    expect(t('missing.key.for.default', 'Default message')).toBe('Default message')
  })

  it('HTML-escapes the site name in onboarding tour titles and descriptions (driver.js innerHTML)', () => {
    setBrandSiteName('A&B <i>')
    expect(t('onboarding.user.welcome.title')).toBe('👋 Welcome to A&amp;B &lt;i&gt;')
    expect(t('onboarding.admin.welcome.title')).toBe('👋 Welcome to A&amp;B &lt;i&gt;')
    expect(t('onboarding.user.welcome.description')).toContain('Welcome to the A&amp;B &lt;i&gt; AI service platform.')

    const [welcome] = getUserSteps(t as (key: string) => string)
    expect(welcome.popover?.title).toBe('👋 Welcome to A&amp;B &lt;i&gt;')
    expect(welcome.popover?.title).not.toContain('<i>')
    expect(welcome.popover?.description).not.toContain('A&B <i>')
  })

  it('keeps the raw site name in non-onboarding plain-text messages', () => {
    setBrandSiteName('A&B <i>')
    expect(t('setup.title')).toBe('A&B <i> Setup')
    expect(t('auth.signUpToStart')).toBe('Sign up to start using A&B <i>')
    expect(t('admin.settings.linuxdo.description')).toBe('Configure LinuxDo Connect OAuth for A&B <i> end-user login')
    expect(t('setup.title')).not.toContain('&amp;')
  })

  it('updates rendered translations reactively when public settings load', async () => {
    setActivePinia(createPinia())
    const Probe = defineComponent({
      setup() {
        const { t: localT } = useI18n()
        return { localT }
      },
      template: '<p class="a">{{ localT("onboarding.user.welcome.title") }}</p><p class="b">{{ $t("setup.title") }}</p>'
    })
    const wrapper = mount(Probe, { global: { plugins: [i18n] } })
    expect(wrapper.find('.a').text()).toBe('👋 Welcome to Sub2API')
    expect(wrapper.find('.b').text()).toBe('Sub2API Setup')

    window.__APP_CONFIG__ = { site_name: 'goreach.es' } as PublicSettings
    const appStore = useAppStore()
    appStore.initFromInjectedConfig()
    await nextTick()

    expect(appStore.siteName).toBe('goreach.es')
    expect(wrapper.find('.a').text()).toBe('👋 Welcome to goreach.es')
    expect(wrapper.find('.b').text()).toBe('goreach.es Setup')
    wrapper.unmount()
  })
})
