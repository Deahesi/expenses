import { useId } from 'react'
import { useForm } from 'react-hook-form'
import type { Expense, ExpenseValues } from '@/entities/expense/model/types'
import { categories } from '@/entities/expense/model/types'
import { categoryMeta } from '@/entities/expense/model/category-meta'
import { today } from '@/shared/lib/format'
import { Button } from '@/shared/ui/button'
import { Input } from '@/shared/ui/input'
import { Textarea } from '@/shared/ui/textarea'
import { Select } from '@/shared/ui/select'
import { FormField } from '@/shared/ui/form-field'
import { ErrorNotice } from '@/shared/ui/error-notice'
import { expenseResolver } from '../model/resolver'
import { useSaveExpense } from '../model/use-save-expense'
import type { User } from '@/entities/user/model/types'

interface ExpenseFormProps {
  expense?: Expense
  members: User[]
  onClose: () => void
  onSaved: () => void
  onBusyChange: (busy: boolean) => void
}

export function ExpenseForm({
  expense,
  members,
  onClose,
  onSaved,
  onBusyChange,
}: ExpenseFormProps) {
  const id = useId()
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ExpenseValues>({
    resolver: expenseResolver,
    defaultValues: {
      user_id: expense?.user_id ?? '',
      amount: expense?.amount,
      category: expense?.category ?? 'еда',
      spent_at: expense?.spent_at ?? today(),
      comment: expense?.comment ?? '',
    },
  })
  const save = useSaveExpense(expense?.id)

  async function submit(values: ExpenseValues) {
    onBusyChange(true)
    try {
      await save.mutateAsync(values)
      onSaved()
    } catch {
      // Mutation error is displayed below; keep the entered values for retry.
    } finally {
      onBusyChange(false)
    }
  }

  function accessibility(field: keyof ExpenseValues) {
    return {
      'aria-invalid': Boolean(errors[field]),
      'aria-describedby': errors[field]
        ? `${id}-${field}-error`
        : field === 'user_id' || field === 'comment'
          ? `${id}-${field}-hint`
          : undefined,
    }
  }

  return (
    <form
      className="space-y-4"
      onSubmit={handleSubmit(submit)}
      onChange={() => save.reset()}
      noValidate
    >
      <fieldset disabled={isSubmitting} className="grid gap-4 border-0 p-0">
        <FormField
          label="Участник"
          htmlFor={`${id}-user_id`}
          error={errors.user_id?.message}
          hint="UUID участника, который уже зарегистрирован в Telegram-боте."
        >
          <Input
            id={`${id}-user_id`}
            {...register('user_id')}
            list={`${id}-members`}
            placeholder="Выберите или введите UUID"
            readOnly={Boolean(expense)}
            required
            {...accessibility('user_id')}
          />
          <datalist id={`${id}-members`}>
            {members.map((member) => (
              <option key={member.id} value={member.id}>
                {member.first_name + ' ' + member.last_name}
              </option>
            ))}
          </datalist>
        </FormField>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <FormField
            label="Сумма, ₽"
            htmlFor={`${id}-amount`}
            error={errors.amount?.message}
          >
            <Input
              id={`${id}-amount`}
              type="number"
              min="1"
              max="9999999999"
              step="1"
              inputMode="numeric"
              placeholder="0"
              {...register('amount', { valueAsNumber: true })}
              required
              {...accessibility('amount')}
            />
          </FormField>
          <FormField
            label="Дата"
            htmlFor={`${id}-spent_at`}
            error={errors.spent_at?.message}
          >
            <Input
              id={`${id}-spent_at`}
              type="date"
              {...register('spent_at')}
              required
              {...accessibility('spent_at')}
            />
          </FormField>
        </div>
        <FormField
          label="Категория"
          htmlFor={`${id}-category`}
          error={errors.category?.message}
        >
          <Select
            id={`${id}-category`}
            {...register('category')}
            {...accessibility('category')}
          >
            {categories.map((category) => (
              <option key={category} value={category}>
                {categoryMeta[category].label}
              </option>
            ))}
          </Select>
        </FormField>
        <FormField
          label="Комментарий"
          htmlFor={`${id}-comment`}
          error={errors.comment?.message}
          hint="Необязательно · до 500 символов"
        >
          <Textarea
            id={`${id}-comment`}
            rows={3}
            placeholder="Например, продукты на неделю"
            {...register('comment')}
            {...accessibility('comment')}
          />
        </FormField>
      </fieldset>
      {save.error && <ErrorNotice message={save.error.message} />}
      <div className="mt-6 flex flex-wrap justify-end gap-2">
        <Button
          type="button"
          variant="outline"
          onClick={onClose}
          disabled={isSubmitting}
        >
          Отмена
        </Button>
        <Button type="submit" disabled={isSubmitting}>
          {isSubmitting
            ? 'Сохраняем…'
            : expense
              ? 'Сохранить изменения'
              : 'Добавить расход'}
        </Button>
      </div>
    </form>
  )
}
