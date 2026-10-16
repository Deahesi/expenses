import type { User } from '@/entities/user/model/types'

export const categories = [
  'еда',
  'транспорт',
  'жильё',
  'развлечения',
  'здоровье',
  'покупки',
  'другое',
] as const
export type ExpenseCategory = (typeof categories)[number]

export interface Expense {
  id: string
  user: User
  user_id: string
  amount: number
  category: ExpenseCategory
  spent_at: string
  comment?: string
  created_at: string
  updated_at: string
}

export interface ExpenseValues {
  user_id: string
  amount: number
  category: ExpenseCategory
  spent_at: string
  comment: string
}

export interface ExpenseFilters {
  user_id: string
  category: ExpenseCategory | ''
  date_from: string
  date_to: string
}

export const emptyFilters: ExpenseFilters = {
  user_id: '',
  category: '',
  date_from: '',
  date_to: '',
}
