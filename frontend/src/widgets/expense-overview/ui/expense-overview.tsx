import type { Expense } from '@/entities/expense/model/types'
import { formatMoney } from '@/shared/lib/format'

interface ExpenseOverviewProps {
  expenses: Expense[]
  loading: boolean
}

export function ExpenseOverview({ expenses, loading }: ExpenseOverviewProps) {
  const total = expenses.reduce((sum, expense) => sum + expense.amount, 0)

  return (
    <dl
      className="flex flex-wrap items-center gap-x-6 gap-y-2 text-sm"
      aria-label="Итоги по выбранным фильтрам"
      aria-busy={loading}
    >
      <div className="flex items-center gap-2 border-r pr-6">
        <dt className="text-muted-foreground">Всего потрачено:</dt>
        <dd className="font-medium tabular-nums">
          {loading ? '—' : formatMoney(total)}
        </dd>
      </div>
      <div className="flex items-center gap-2">
        <dt className="text-muted-foreground">Всего записей:</dt>
        <dd className="font-medium tabular-nums">
          {loading ? '—' : expenses.length}
        </dd>
      </div>
    </dl>
  )
}
