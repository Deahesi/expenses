import { describe, expect, it } from 'vitest'
import { validateExpense } from './validation'
import type { ExpenseValues } from '@/entities/expense/model/types'

const valid: ExpenseValues = {
  user_id: '11111111-1111-4111-8111-111111111111',
  amount: 350,
  category: 'еда',
  spent_at: '2026-10-16',
  comment: 'Кофе',
}

describe('expense validation matches HTTP DTO', () => {
  it('accepts integer rubles and leap dates', () => {
    expect(validateExpense(valid)).toEqual({})
    expect(validateExpense({ ...valid, spent_at: '2024-02-29' })).toEqual({})
  })

  it.each([0, -1, 1.5, 1e10, NaN, Infinity])(
    'rejects invalid amount %s',
    (amount) => {
      expect(validateExpense({ ...valid, amount }).amount).toBeTruthy()
    },
  )

  it('accepts the maximum valid amount', () => {
    expect(validateExpense({ ...valid, amount: 9999999999 })).toEqual({})
  })

  it.each(['', '1234567', '00000000-0000-0000-0000-000000000000'])(
    'rejects invalid UUID %s',
    (user_id) => {
      expect(validateExpense({ ...valid, user_id }).user_id).toBeTruthy()
    },
  )

  it.each(['2026-02-29', '2026-02-30', '2026-13-01', '2026-1-1', ''])(
    'rejects invalid date %s',
    (spent_at) => {
      expect(validateExpense({ ...valid, spent_at }).spent_at).toBeTruthy()
    },
  )

  it('counts Unicode code points like Go RuneLength', () => {
    expect(validateExpense({ ...valid, comment: '💚'.repeat(500) })).toEqual({})
    expect(
      validateExpense({ ...valid, comment: '💚'.repeat(501) }).comment,
    ).toBeTruthy()
  })
})
