import { useState } from 'react'
import type { Expense } from '@/entities/expense/model/types'
import { Button } from '@/shared/ui/button'
import { ExpenseRow } from './expense-row'
import { EmptyExpenses } from './empty-expenses'

interface ExpensesTableProps {
  expenses: Expense[]
  loading: boolean
  filtered: boolean
  onCreate: () => void
  onReset: () => void
  onEdit: (expense: Expense) => void
  onDelete: (expense: Expense) => void
}

const pageSize = 10

export function ExpensesTable({
  expenses,
  loading,
  filtered,
  onCreate,
  onReset,
  onEdit,
  onDelete,
}: ExpensesTableProps) {
  const [page, setPage] = useState(0)
  const lastPage = Math.max(0, Math.ceil(expenses.length / pageSize) - 1)
  const currentPage = Math.min(page, lastPage)
  const start = currentPage * pageSize
  const sorted = [...expenses].sort(
    (a, b) =>
      b.spent_at.localeCompare(a.spent_at) ||
      b.created_at.localeCompare(a.created_at),
  )

  if (loading)
    return (
      <div className="py-8 text-sm text-muted-foreground" role="status">
        Загружаем расходы…
      </div>
    )
  if (!expenses.length)
    return (
      <EmptyExpenses
        filtered={filtered}
        onCreate={onCreate}
        onReset={onReset}
      />
    )

  return (
    <>
      <div
        className="relative w-full overflow-x-auto"
        role="region"
        aria-label="Таблица расходов, доступна горизонтальная прокрутка"
        tabIndex={0}
      >
        <table className="w-full min-w-[750px] caption-bottom text-sm">
          <caption className="sr-only">Расходы семьи</caption>
          <thead className="[&_tr]:border-b [&_th]:h-10 [&_th]:px-2 [&_th]:text-left [&_th]:align-middle [&_th]:font-medium [&_th]:whitespace-nowrap">
            <tr>
              <th scope="col">Дата</th>
              <th scope="col">Комментарий</th>
              <th scope="col">Категория</th>
              <th scope="col">Пользователь</th>
              <th scope="col" className="!text-right">
                Сумма
              </th>
              <th scope="col" className="!text-right">
                Действия
              </th>
            </tr>
          </thead>
          <tbody className="[&_tr:last-child]:border-0">
            {sorted.slice(start, start + pageSize).map((expense) => (
              <ExpenseRow
                key={expense.id}
                expense={expense}
                onEdit={onEdit}
                onDelete={onDelete}
              />
            ))}
          </tbody>
        </table>
      </div>
      <div className="flex flex-wrap items-center justify-between gap-4 text-sm">
        <span>
          {start + 1}–{Math.min(start + pageSize, expenses.length)} из{' '}
          {expenses.length}
        </span>
        <div className="flex items-center gap-3">
          <Button
            variant="outline"
            size="sm"
            aria-label="Предыдущая страница"
            disabled={currentPage === 0}
            onClick={() => setPage(currentPage - 1)}
          >
            Назад
          </Button>
          <span>
            {currentPage + 1} / {lastPage + 1}
          </span>
          <Button
            variant="outline"
            size="sm"
            aria-label="Следующая страница"
            disabled={currentPage === lastPage}
            onClick={() => setPage(currentPage + 1)}
          >
            Вперёд
          </Button>
        </div>
      </div>
    </>
  )
}
