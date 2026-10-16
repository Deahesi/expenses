import { afterEach, describe, expect, it, vi } from 'vitest'
import { api } from '@/shared/api/client'
import {
  createExpense,
  deleteExpense,
  getExpenses,
  updateExpense,
} from './expenses'
import type { Expense, ExpenseValues } from '../model/types'

const user = '11111111-1111-4111-8111-111111111111'
const values: ExpenseValues = {
  user_id: user,
  amount: 350,
  category: 'еда',
  spent_at: '2026-10-16',
  comment: 'Кофе',
}

afterEach(() => vi.restoreAllMocks())

describe('expense API contract', () => {
  it('passes filters and cancellation signal to axios, preserving the trailing slash', async () => {
    const get = vi.spyOn(api, 'get').mockResolvedValue({ data: [] })
    const signal = new AbortController().signal
    const filters = {
      user_id: user,
      category: 'еда' as const,
      date_from: '2026-10-01',
      date_to: '',
    }

    await expect(getExpenses(filters, signal)).resolves.toEqual([])
    expect(get).toHaveBeenCalledWith('/expenses/', {
      params: { user_id: user, category: 'еда', date_from: '2026-10-01' },
      signal,
    })
  })

  it('uses owner in route and integer rubles in the body for create and update', async () => {
    const post = vi.spyOn(api, 'post').mockResolvedValue({ data: values })
    const put = vi.spyOn(api, 'put').mockResolvedValue({ data: values })
    const body = {
      amount: 350,
      category: 'еда',
      spent_at: '2026-10-16',
      comment: 'Кофе',
    }

    await expect(createExpense(values)).resolves.toEqual(values)
    await expect(updateExpense('expense-id', values)).resolves.toEqual(values)
    expect(post).toHaveBeenCalledWith(`/expenses/${user}`, body)
    expect(put).toHaveBeenCalledWith(`/expenses/${user}/expense-id`, body)
  })

  it('handles a successful empty DELETE response', async () => {
    const remove = vi
      .spyOn(api, 'delete')
      .mockResolvedValue({ status: 204, data: '' })

    await expect(
      deleteExpense({ ...values, id: 'expense-id' } as Expense),
    ).resolves.toBeUndefined()
    expect(remove).toHaveBeenCalledWith(`/expenses/${user}/expense-id`)
  })

  it('passes axios errors to the caller', async () => {
    const error = new Error('Network Error')
    vi.spyOn(api, 'get').mockRejectedValue(error)

    await expect(getExpenses({})).rejects.toBe(error)
  })
})
