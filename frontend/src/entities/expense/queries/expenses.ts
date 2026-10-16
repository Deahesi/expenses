import { queryOptions } from '@tanstack/react-query'
import { getExpenses } from '../api/expenses'
import { emptyFilters, type ExpenseFilters } from '../model/types'

export const expenseKeys = { all: ['expenses'] as const }

export function expensesQuery(filters: ExpenseFilters = emptyFilters) {
  return queryOptions({
    queryKey: [...expenseKeys.all, filters],
    queryFn: ({ signal }) => getExpenses(filters, signal),
  })
}
