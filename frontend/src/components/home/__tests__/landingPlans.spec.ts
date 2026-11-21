import { describe, expect, it } from 'vitest'
import { toDisplayPlans } from '../landingPlans'

const base = { description: '', features: '', sort_order: 0, validity_days: 30, validity_unit: 'days' }

describe('toDisplayPlans', () => {
  it('returns an empty list for no plans', () => {
    expect(toDisplayPlans([])).toEqual([])
  })

  it('maps currency symbols and keeps decimals only when needed', () => {
    const [usd, cny, other] = toDisplayPlans([
      { ...base, id: 1, name: 'A', price: 45, currency: 'USD', sort_order: 1 },
      { ...base, id: 2, name: 'B', price: 99.5, currency: 'cny', sort_order: 2 },
      { ...base, id: 3, name: 'C', price: 10, currency: 'SGD', sort_order: 3 },
    ])
    expect([usd.symbol, usd.priceText]).toEqual(['$', '45'])
    expect([cny.symbol, cny.priceText]).toEqual(['¥', '99.50'])
    expect(other.symbol).toBe('SGD ')
  })

  it('shows the original price only when it is higher than the price', () => {
    const [a, b] = toDisplayPlans([
      { ...base, id: 1, name: 'A', price: 45, original_price: 60, sort_order: 1 },
      { ...base, id: 2, name: 'B', price: 50, original_price: 50, sort_order: 2 },
    ])
    expect(a.originalText).toBe('60')
    expect(b.originalText).toBe('')
  })

  it.each([
    [30, 'days', '月'], [7, 'days', '周'], [365, 'days', '年'], [90, 'days', '90 天'],
    [1, 'months', '月'], [3, 'months', '3 个月'], [1, 'years', '年'],
  ])('formats %i %s as %s', (days, unit, expected) => {
    expect(toDisplayPlans([{ ...base, id: 1, name: 'A', price: 1, validity_days: days, validity_unit: unit }])[0].period)
      .toBe(expected)
  })

  it('splits features by line and strips bullets', () => {
    const [p] = toDisplayPlans([{ ...base, id: 1, name: 'A', price: 1, features: '- 一\n• 二\n\n三' }])
    expect(p.features).toEqual(['一', '二', '三'])
  })

  it('does not feature a lone plan', () => {
    expect(toDisplayPlans([{ ...base, id: 1, name: 'A', price: 1 }])[0].featured).toBe(false)
  })
})
