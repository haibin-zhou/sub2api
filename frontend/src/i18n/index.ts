import { createI18n } from 'vue-i18n'
import { getBrandSiteName } from '@/utils/branding'

type LocaleCode = 'en' | 'zh'

type LocaleMessages = Record<string, any>

const LOCALE_KEY = 'sub2api_locale'
const DEFAULT_LOCALE: LocaleCode = 'en'

const localeLoaders: Record<LocaleCode, () => Promise<{ default: LocaleMessages }>> = {
  en: () => import('./locales/en'),
  zh: () => import('./locales/zh')
}

function isLocaleCode(value: string): value is LocaleCode {
  return value === 'en' || value === 'zh'
}

function getDefaultLocale(): LocaleCode {
  const saved = localStorage.getItem(LOCALE_KEY)
  if (saved && isLocaleCode(saved)) {
    return saved
  }

  const browserLang = navigator.language.toLowerCase()
  if (browserLang.startsWith('zh')) {
    return 'zh'
  }

  return DEFAULT_LOCALE
}

export const i18n = createI18n({
  legacy: false,
  locale: getDefaultLocale(),
  fallbackLocale: DEFAULT_LOCALE,
  messages: {},
  // 禁用 HTML 消息警告 - 引导步骤使用富文本内容（driver.js 支持 HTML）
  // 这些内容是内部定义的，不存在 XSS 风险
  warnHtmlMessage: false
})

const HTML_TAG_PATTERN = /<[a-z][^>]*>/i

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

function isPlainParams(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value)
}

/**
 * 让所有 t() 调用自动获得命名参数 {siteName}（后台配置的站点名称，缺省 Sub2API）。
 *
 * - 调用方显式传入 siteName 时以调用方为准；
 * - 读取的是响应式 ref，模板 / computed 中的 t() 会在公共设置加载后自动重新渲染；
 * - 富文本（HTML）文案中插入的站点名会做 HTML 转义，避免站点名破坏引导弹窗结构。
 *
 * 必须在 app.use(i18n) 之前执行：全局 $t 在安装时复制的是 composer.t 的属性描述符。
 */
export function installSiteNameParam(composer: { t: unknown; tm: unknown }): void {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const originalT = composer.t as (...args: any[]) => string
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const tm = composer.tm as (key: string) => unknown
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const patched = (...args: any[]): string => {
    const [key, arg2, arg3] = args
    const siteNameFor = () => {
      const name = getBrandSiteName()
      if (typeof key !== 'string') return name
      const raw = tm(key)
      return typeof raw === 'string' && HTML_TAG_PATTERN.test(raw) ? escapeHtml(name) : name
    }

    if (args.length <= 1) {
      return originalT(key, { siteName: siteNameFor() })
    }
    if (isPlainParams(arg2)) {
      if ('siteName' in arg2) {
        return originalT(...args)
      }
      const rest = args.slice(2)
      return originalT(key, { siteName: siteNameFor(), ...arg2 }, ...rest)
    }
    if ((typeof arg2 === 'number' || typeof arg2 === 'string') && args.length === 2) {
      // t(key, plural) / t(key, defaultMessage) → t(key, named, plural|defaultMessage)
      return originalT(key, { siteName: siteNameFor() }, arg2)
    }
    if ((typeof arg2 === 'number' || typeof arg2 === 'string') && isPlainParams(arg3)) {
      // t(key, plural|default, options)：把站点名并入 options.named
      const options = arg3 as { named?: Record<string, unknown> }
      const named = { siteName: siteNameFor(), ...(options.named ?? {}) }
      return originalT(key, arg2, { ...options, named })
    }
    return originalT(...args)
  }
  composer.t = patched
}

installSiteNameParam(i18n.global as unknown as { t: unknown; tm: unknown })

const loadedLocales = new Set<LocaleCode>()

export async function loadLocaleMessages(locale: LocaleCode): Promise<void> {
  if (loadedLocales.has(locale)) {
    return
  }

  const loader = localeLoaders[locale]
  const module = await loader()
  i18n.global.setLocaleMessage(locale, module.default)
  loadedLocales.add(locale)
}

export async function initI18n(): Promise<void> {
  const current = getLocale()
  await loadLocaleMessages(current)
  document.documentElement.setAttribute('lang', current)
}

export async function setLocale(locale: string): Promise<void> {
  if (!isLocaleCode(locale)) {
    return
  }

  await loadLocaleMessages(locale)
  i18n.global.locale.value = locale
  localStorage.setItem(LOCALE_KEY, locale)
  document.documentElement.setAttribute('lang', locale)

  // 同步更新浏览器页签标题，使其跟随语言切换
  const { resolveRouteDocumentTitle } = await import('@/router/title')
  const { default: router } = await import('@/router')
  const { useAppStore } = await import('@/stores/app')
  const { useAuthStore } = await import('@/stores/auth')
  const { useAdminSettingsStore } = await import('@/stores/adminSettings')
  const route = router.currentRoute.value
  const appStore = useAppStore()
  const authStore = useAuthStore()
  const adminSettingsStore = useAdminSettingsStore()
  const customMenuItems = [
    ...(appStore.cachedPublicSettings?.custom_menu_items ?? []),
    ...(authStore.isAdmin ? adminSettingsStore.customMenuItems : []),
  ]
  document.title = resolveRouteDocumentTitle(route, appStore.siteName, customMenuItems)
}

export function getLocale(): LocaleCode {
  const current = i18n.global.locale.value
  return isLocaleCode(current) ? current : DEFAULT_LOCALE
}

export const availableLocales = [
  { code: 'en', name: 'English', flag: '🇺🇸' },
  { code: 'zh', name: '中文', flag: '🇨🇳' }
] as const

export default i18n
