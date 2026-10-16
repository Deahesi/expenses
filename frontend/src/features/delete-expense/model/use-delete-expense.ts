import { useMutation, useQueryClient } from '@tanstack/react-query'
import { deleteExpense } from '@/entities/expense/api/expenses'
import { expenseKeys } from '@/entities/expense/queries/expenses'
import type { Expense } from '@/entities/expense/model/types'

export function useDeleteExpense(expense: Expense, onDeleted: () => void) {
  const client = useQueryClient()

  return useMutation({
    mutationFn: () => deleteExpense(expense),
    onSuccess: async () => {
      await client.invalidateQueries({ queryKey: expenseKeys.all })
      onDeleted()
    },
  })
}
