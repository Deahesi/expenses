import type { FieldErrors, Resolver } from 'react-hook-form'
import type { ExpenseValues } from '@/entities/expense/model/types'
import { validateExpense } from './validation'

export const expenseResolver: Resolver<ExpenseValues> = (input) => {
  const values = {
    ...input,
    user_id: input.user_id.trim(),
    comment: input.comment.trim(),
  }
  const messages = validateExpense(values)
  const errors: FieldErrors<ExpenseValues> = {}

  for (const field of Object.keys(messages) as (keyof ExpenseValues)[]) {
    errors[field] = { type: 'validate', message: messages[field] }
  }

  return Object.keys(errors).length
    ? { values: {}, errors }
    : { values, errors: {} }
}
