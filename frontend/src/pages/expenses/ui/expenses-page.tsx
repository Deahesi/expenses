import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { emptyFilters, type Expense } from '@/entities/expense/model/types'
import { expensesQuery } from '@/entities/expense/queries/expenses'
import { ExpenseDialog } from '@/features/manage-expense/ui/expense-dialog'
import { DeleteExpenseDialog } from '@/features/delete-expense/ui/delete-expense-dialog'
import { ExpenseFilters } from '@/features/filter-expenses/ui/expense-filters'
import { ExpenseOverview } from '@/widgets/expense-overview/ui/expense-overview'
import { ExpensesTable } from '@/widgets/expenses-table/ui/expenses-table'
import { Button } from '@/shared/ui/button'
import { ErrorNotice } from '@/shared/ui/error-notice'
import { usersQuery } from '@/entities/user/queries/users'

type ActiveDialog =
  | { type: 'create' }
  | { type: 'edit'; expense: Expense }
  | { type: 'delete'; expense: Expense }
  | null

export function ExpensesPage() {
  const [filters, setFilters] = useState({ ...emptyFilters })
  const [dialog, setDialog] = useState<ActiveDialog>(null)
  const [notice, setNotice] = useState('')
  const validDates =
    !filters.date_from ||
    !filters.date_to ||
    filters.date_from <= filters.date_to
  const all = useQuery(expensesQuery())

  const list = useQuery({ ...expensesQuery(filters), enabled: validDates })
  const expenses = validDates ? (list.data ?? []) : []

  const members = useQuery(usersQuery())

  const loading = validDates && list.isPending
  const filtered = Object.values(filters).some(Boolean)
  const unavailable = !validDates || Boolean(list.error && !list.data)

  function completed(message: string) {
    setDialog(null)
    setNotice(message)
  }

  return (
    <>
      <main className="mx-auto flex max-w-7xl flex-col gap-6 px-4 py-8 sm:px-6">
        <div className="flex flex-wrap items-center justify-between gap-4">
          <h1 className="text-2xl font-semibold tracking-tight">
            Расходы семьи
          </h1>
          <Button
            onClick={() => {
              setNotice('')
              setDialog({ type: 'create' })
            }}
          >
            Добавить расход
          </Button>
        </div>
        {notice && (
          <div
            className="flex items-center justify-between gap-4 text-sm"
            role="status"
          >
            <p>{notice}</p>
            <Button variant="ghost" size="sm" onClick={() => setNotice('')}>
              Скрыть
            </Button>
          </div>
        )}
        <ExpenseOverview expenses={expenses} loading={loading || unavailable} />
        <ExpenseFilters
          value={filters}
          members={members.data || []}
          onChange={setFilters}
        />
        {list.error && (
          <ErrorNotice
            message={list.error.message}
            onRetry={() => {
              void list.refetch()
            }}
          />
        )}
        {all.error && !list.error && (
          <ErrorNotice
            message="Не удалось обновить список участников."
            onRetry={() => {
              void all.refetch()
            }}
          />
        )}
        {unavailable ? (
          <p className="text-sm text-muted-foreground">
            {validDates
              ? 'Не удалось загрузить расходы.'
              : 'Выберите корректный период.'}
          </p>
        ) : (
          <ExpensesTable
            key={JSON.stringify(filters)}
            expenses={expenses}
            loading={loading}
            filtered={filtered}
            onCreate={() => setDialog({ type: 'create' })}
            onReset={() => setFilters({ ...emptyFilters })}
            onEdit={(expense) => setDialog({ type: 'edit', expense })}
            onDelete={(expense) => setDialog({ type: 'delete', expense })}
          />
        )}
      </main>
      {(dialog?.type === 'create' || dialog?.type === 'edit') && (
        <ExpenseDialog
          expense={dialog.type === 'edit' ? dialog.expense : undefined}
          members={members.data || []}
          onClose={() => setDialog(null)}
          onSaved={() =>
            completed(
              dialog.type === 'edit'
                ? 'Изменения сохранены.'
                : 'Расход добавлен.',
            )
          }
        />
      )}
      {dialog?.type === 'delete' && (
        <DeleteExpenseDialog
          expense={dialog.expense}
          onClose={() => setDialog(null)}
          onDeleted={() => completed('Расход удалён.')}
        />
      )}
    </>
  )
}
