import { Button } from '@/shared/ui/button'

interface EmptyExpensesProps {
  filtered: boolean
  onCreate: () => void
  onReset: () => void
}

export function EmptyExpenses({
  filtered,
  onCreate,
  onReset,
}: EmptyExpensesProps) {
  return (
    <div className="flex flex-wrap items-center gap-4 py-8 text-sm">
      <p className="text-muted-foreground">
        {filtered ? 'Расходы по фильтрам не найдены.' : 'Расходов пока нет.'}
      </p>
      <Button variant="outline" onClick={filtered ? onReset : onCreate}>
        {filtered ? 'Сбросить фильтры' : 'Добавить первый расход'}
      </Button>
    </div>
  )
}
