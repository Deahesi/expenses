import type { Expense } from '@/entities/expense/model/types'
import { CategoryBadge } from '@/entities/expense/ui/category-badge'
import { formatMoney, formatDate } from '@/shared/lib/format'
import { Button } from '@/shared/ui/button'

interface ExpenseRowProps {
  expense: Expense
  onEdit: (expense: Expense) => void
  onDelete: (expense: Expense) => void
}

export function ExpenseRow({ expense, onEdit, onDelete }: ExpenseRowProps) {
  const label = expense.comment || expense.category

  return (
    <tr className="border-b transition-colors hover:bg-muted/50">
      <td className="p-2 align-middle whitespace-nowrap">
        {formatDate(expense.spent_at)}
      </td>
      <td className="p-2 align-middle">
        <div className="max-w-64 break-words">{expense.comment || '—'}</div>
      </td>
      <td className="p-2 align-middle">
        <CategoryBadge category={expense.category} />
      </td>
      <td
        className="p-2 align-middle whitespace-nowrap"
        title={expense.user_id}
      >
        {expense.user.first_name + ' ' + expense.user.last_name}
      </td>
      <td className="p-2 align-middle text-right whitespace-nowrap tabular-nums">
        {formatMoney(expense.amount)}
      </td>
      <td className="p-2 align-middle">
        <div className="flex justify-end gap-2">
          <Button
            variant="outline"
            size="sm"
            aria-label={`Редактировать: ${label}`}
            onClick={() => onEdit(expense)}
          >
            Изменить
          </Button>
          <Button
            variant="outline"
            size="sm"
            aria-label={`Удалить: ${label}`}
            onClick={() => onDelete(expense)}
          >
            Удалить
          </Button>
        </div>
      </td>
    </tr>
  )
}
