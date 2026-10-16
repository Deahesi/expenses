import { useMutation, useQueryClient } from '@tanstack/react-query'
import { createExpense, updateExpense } from '@/entities/expense/api/expenses'
import { expenseKeys } from '@/entities/expense/queries/expenses'
import type { ExpenseValues } from '@/entities/expense/model/types'

export function useSaveExpense(id?: string) {
  const client = useQueryClient()

  return useMutation({
    mutationFn: (values: ExpenseValues) =>
      id ? updateExpense(id, values) : createExpense(values),
    onSuccess: () => client.invalidateQueries({ queryKey: expenseKeys.all }),
  })
}
