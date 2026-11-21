import type { PublicPlan } from '@/api/payment'

// ---- 价格区：优先读取后台「订阅套餐」中在售的套餐，读取失败时回退到内置文案 ----
export interface DisplayPlan {
  key: string
  id: number | null
  name: string
  description: string
  symbol: string
  priceText: string
  originalText: string
  period: string
  features: string[]
  featured: boolean
}

export const FALLBACK_PLANS: DisplayPlan[] = [
  { key: 'fallback-standard', id: null, name: 'Standard', description: '约 1.25× Pro 用量', symbol: '$', priceText: '45',
    originalText: '60', period: '月', features: ['适合日常单项目，常规调用频次'], featured: false },
  { key: 'fallback-premium', id: null, name: 'Premium', description: '约 6.25× Pro 用量', symbol: '$', priceText: '180',
    originalText: '220', period: '月', features: ['适合高频调用，多项目并行；优先开通，优先异常协助排查'], featured: true },
]

const CURRENCY_SYMBOLS: Record<string, string> = { USD: '$', CNY: '¥', EUR: '€', GBP: '£', HKD: 'HK$', JPY: '¥' }

function formatAmount(n: number): string {
  return Number.isInteger(n) ? String(n) : n.toFixed(2)
}

function formatPeriod(days: number, unit: string): string {
  const u = (unit || 'days').toLowerCase()
  if (u.startsWith('month')) return days === 1 ? '月' : `${days} 个月`
  if (u.startsWith('year')) return days === 1 ? '年' : `${days} 年`
  if (u.startsWith('week')) return days === 1 ? '周' : `${days} 周`
  if (days === 30 || days === 31) return '月'
  if (days === 365) return '年'
  if (days === 7) return '周'
  return `${days} 天`
}

function splitFeatures(raw: string): string[] {
  return (raw || '')
    .split(/\r?\n/)
    .map((line) => line.replace(/^[-•·*\s]+/, '').trim())
    .filter(Boolean)
}

export function toDisplayPlans(plans: PublicPlan[]): DisplayPlan[] {
  if (!plans.length) return []
  const maxPrice = Math.max(...plans.map((p) => p.price))
  const sorted = [...plans].sort((a, b) => a.sort_order - b.sort_order || a.price - b.price)
  return sorted.map((p) => {
    const code = (p.currency || 'USD').toUpperCase()
    return {
      key: `plan-${p.id}`,
      id: p.id,
      name: p.name,
      description: p.description || '',
      symbol: CURRENCY_SYMBOLS[code] ?? `${code} `,
      priceText: formatAmount(p.price),
      originalText: p.original_price && p.original_price > p.price ? formatAmount(p.original_price) : '',
      period: formatPeriod(p.validity_days, p.validity_unit),
      features: splitFeatures(p.features),
      featured: plans.length > 1 && p.price === maxPrice,
    }
  })
}

