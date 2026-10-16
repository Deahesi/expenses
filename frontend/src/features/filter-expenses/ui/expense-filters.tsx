import { Controller, useForm } from 'react-hook-form'
import {
  categories,
  emptyFilters,
  type ExpenseFilters as Filters,
} from '@/entities/expense/model/types'
import { categoryMeta } from '@/entities/expense/model/category-meta'
import { Select } from '@/shared/ui/select'
import { Input } from '@/shared/ui/input'
import { Button } from '@/shared/ui/button'
import type { User } from '@/entities/user/model/types'

interface ExpenseFiltersProps {
  value: Filters
  members: User[]
  onChange: (value: Filters) => void
}

export function ExpenseFilters({
  value,
  members,
  onChange,
}: ExpenseFiltersProps) {
  const { control, handleSubmit } = useForm<Filters>({ values: value })
  const active = Object.values(value).some(Boolean)
  const invalidDates = Boolean(
    value.date_from && value.date_to && value.date_from > value.date_to,
  )

  return (
    <form
      className="space-y-2"
      onChange={() => {
        void handleSubmit(onChange)()
      }}
      onSubmit={handleSubmit(onChange)}
    >
      <div className="grid grid-cols-1 items-end gap-4 sm:grid-cols-2 lg:grid-cols-[1fr_1fr_1fr_1fr_auto]">
        <div className="min-w-0 space-y-2 [&_label]:text-sm [&_label]:font-medium">
          <label htmlFor="filter-member">Участник</label>
          <Controller
            name="user_id"
            control={control}
            render={({ field }) => (
              <Select id="filter-member" {...field}>
                <option value="">Вся семья</option>
                {members.map((member) => (
                  <option key={member.id} value={member.id}>
                    {member.first_name + ' ' + member.last_name}
                  </option>
                ))}
              </Select>
            )}
          />
        </div>
        <div className="min-w-0 space-y-2 [&_label]:text-sm [&_label]:font-medium">
          <label htmlFor="filter-category">Категория</label>
          <Controller
            name="category"
            control={control}
            render={({ field }) => (
              <Select id="filter-category" {...field}>
                <option value="">Все категории</option>
                {categories.map((category) => (
                  <option key={category} value={category}>
                    {categoryMeta[category].label}
                  </option>
                ))}
              </Select>
            )}
          />
        </div>
        <div className="min-w-0 space-y-2 [&_label]:text-sm [&_label]:font-medium">
          <label htmlFor="filter-from">С даты</label>
          <Controller
            name="date_from"
            control={control}
            render={({ field }) => (
              <Input
                id="filter-from"
                type="date"
                {...field}
                aria-invalid={invalidDates}
                aria-describedby={
                  invalidDates ? 'filter-date-error' : undefined
                }
              />
            )}
          />
        </div>
        <div className="min-w-0 space-y-2 [&_label]:text-sm [&_label]:font-medium">
          <label htmlFor="filter-to">По дату</label>
          <Controller
            name="date_to"
            control={control}
            render={({ field }) => (
              <Input
                id="filter-to"
                type="date"
                {...field}
                aria-invalid={invalidDates}
                aria-describedby={
                  invalidDates ? 'filter-date-error' : undefined
                }
              />
            )}
          />
        </div>
        <Button
          type="button"
          variant="outline"
          aria-label="Сбросить фильтры"
          title="Сбросить фильтры"
          disabled={!active}
          onClick={() => {
            onChange({ ...emptyFilters })
          }}
        >
          Сбросить фильтры
        </Button>
      </div>
      {invalidDates && (
        <p
          className="text-sm text-destructive"
          id="filter-date-error"
          role="alert"
        >
          Дата окончания не может быть раньше даты начала.
        </p>
      )}
    </form>
  )
}
