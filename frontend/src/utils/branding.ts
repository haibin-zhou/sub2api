import { ref } from 'vue'
import { sanitizeUrl } from '@/utils/url'

/**
 * 未配置站点名称（后台「站点名称」/ public setting `site_name`）时使用的中性品牌名。
 * 产品对外品牌一律以后台配置为准，这里只是兜底。
 */
export const DEFAULT_SITE_NAME = 'Sub2API'

/** 规范化站点名称：去掉首尾空白，空值回退到 DEFAULT_SITE_NAME。 */
export function normalizeSiteName(name: unknown): string {
  return typeof name === 'string' && name.trim() ? name.trim() : DEFAULT_SITE_NAME
}

// 当前生效的站点名称（响应式）。i18n 的 t() 会读取它作为默认的 {siteName} 参数，
// 因此在渲染函数中翻译的文案会随公共设置加载/变更自动刷新。
const brandSiteName = ref<string>(DEFAULT_SITE_NAME)

export function setBrandSiteName(name: unknown): void {
  brandSiteName.value = normalizeSiteName(name)
}

export function getBrandSiteName(): string {
  return brandSiteName.value
}

export function updateFavicon(logoUrl: string): void {
  const sanitizedLogoUrl = sanitizeUrl(logoUrl, {
    allowRelative: true,
    allowDataUrl: true,
  })
  if (!sanitizedLogoUrl) {
    return
  }

  let link = document.querySelector<HTMLLinkElement>('link[rel="icon"]')
  if (!link) {
    link = document.createElement('link')
    link.rel = 'icon'
    document.head.appendChild(link)
  }

  link.type = sanitizedLogoUrl.endsWith('.svg') ? 'image/svg+xml' : 'image/x-icon'
  link.href = sanitizedLogoUrl
}
