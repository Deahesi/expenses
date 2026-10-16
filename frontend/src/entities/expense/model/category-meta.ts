import type { ExpenseCategory } from './types'

export const categoryMeta = {
  еда: { label: 'Еда' },
  транспорт: { label: 'Транспорт' },
  жильё: { label: 'Жильё' },
  развлечения: { label: 'Развлечения' },
  здоровье: { label: 'Здоровье' },
  покупки: { label: 'Покупки' },
  другое: { label: 'Другое' },
} satisfies Record<ExpenseCategory, { label: string }>
