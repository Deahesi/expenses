import { useDeleteExpense } from '../model/use-delete-expense'
import type { Expense } from '@/entities/expense/model/types'
import { formatMoney, formatDate } from '@/shared/lib/format'
import { Button } from '@/shared/ui/button'
import { Modal } from '@/shared/ui/modal'
import { ErrorNotice } from '@/shared/ui/error-notice'

interface DeleteExpenseDialogProps {
  expense: Expense
  onClose: () => void
  onDeleted: () => void
}

export function DeleteExpenseDialog({
  expense,
  onClose,
  onDeleted,
}: DeleteExpenseDialogProps) {
  const mutation = useDeleteExpense(expense, onDeleted)

  return (
    <Modal
      title="Удалить расход?"
      description="Эту запись нельзя будет восстановить."
      onClose={onClose}
      busy={mutation.isPending}
    >
      <div className="flex flex-col gap-2 break-words text-sm">
        <strong>{formatMoney(expense.amount)}</strong>
        <span>
          {expense.comment || 'Без комментария'} ·{' '}
          {formatDate(expense.spent_at)}
        </span>
      </div>
      {mutation.error && <ErrorNotice message={mutation.error.message} />}
      <div className="flex flex-wrap justify-end gap-2">
        <Button
          variant="outline"
          onClick={onClose}
          disabled={mutation.isPending}
        >
          Отмена
        </Button>
        <Button
          variant="default"
          onClick={() => mutation.mutate()}
          disabled={mutation.isPending}
        >
          {mutation.isPending ? 'Удаляем…' : 'Удалить расход'}
        </Button>
      </div>
    </Modal>
  )
}
