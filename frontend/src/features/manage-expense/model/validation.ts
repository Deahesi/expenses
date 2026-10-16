import { categories, type ExpenseValues } from '@/entities/expense/model/types'

export type FormErrors = Partial<Record<keyof ExpenseValues, string>>

export function validateExpense(values: ExpenseValues): FormErrors {
  const errors: FormErrors = {}
  const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i

  if (
    !uuid.test(values.user_id) ||
    values.user_id === '00000000-0000-0000-0000-000000000000'
  ) {
    errors.user_id = 'Укажите корректный UUID участника.'
  }
  if (
    !Number.isInteger(values.amount) ||
    values.amount < 1 ||
    values.amount >= 1e10
  ) {
    errors.amount = 'Введите целую сумму от 1 до 9 999 999 999 ₽.'
  }
  if (!categories.includes(values.category))
    errors.category = 'Выберите категорию.'

  const date = new Date(`${values.spent_at}T00:00:00Z`)
  if (
    !/^\d{4}-\d{2}-\d{2}$/.test(values.spent_at) ||
    Number.isNaN(date.getTime()) ||
    date.toISOString().slice(0, 10) !== values.spent_at
  ) {
    errors.spent_at = 'Укажите корректную дату.'
  }
  if ([...values.comment.trim()].length > 500)
    errors.comment = 'Не более 500 символов.'

  return errors
}
