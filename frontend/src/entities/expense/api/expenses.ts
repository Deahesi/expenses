import { api } from '@/shared/api/client'
import type { Expense, ExpenseFilters, ExpenseValues } from '../model/types'

export async function getExpenses(
  filters: Partial<ExpenseFilters>,
  signal?: AbortSignal,
) {
  const params = Object.fromEntries(
    Object.entries(filters).filter(([, value]) => value),
  )
  const { data } = await api.get<Expense[]>('/expenses/', { params, signal })

  return data
}

export async function createExpense(values: ExpenseValues) {
  const { user_id, ...body } = values
  const { data } = await api.post<Expense>(`/expenses/${user_id}`, body)

  return data
}

export async function updateExpense(id: string, values: ExpenseValues) {
  const { user_id, ...body } = values
  const { data } = await api.put<Expense>(`/expenses/${user_id}/${id}`, body)

  return data
}

export async function deleteExpense(expense: Expense) {
  await api.delete(`/expenses/${expense.user_id}/${expense.id}`)
}
