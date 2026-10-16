import { useState } from 'react'
import type { Expense } from '@/entities/expense/model/types'
import { Modal } from '@/shared/ui/modal'
import { ExpenseForm } from './expense-form'
import type { User } from '@/entities/user/model/types'

interface ExpenseDialogProps {
  expense?: Expense
  members: User[]
  onClose: () => void
  onSaved: () => void
}

export function ExpenseDialog({
  expense,
  members,
  onClose,
  onSaved,
}: ExpenseDialogProps) {
  const [busy, setBusy] = useState(false)

  return (
    <Modal
      title={expense ? 'Редактировать расход' : 'Новый расход'}
      onClose={onClose}
      busy={busy}
    >
      <ExpenseForm
        expense={expense}
        members={members}
        onClose={onClose}
        onSaved={onSaved}
        onBusyChange={setBusy}
      />
    </Modal>
  )
}
