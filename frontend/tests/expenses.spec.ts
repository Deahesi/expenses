import { test, expect, type Page } from '@playwright/test'

const user = '11111111-1111-4111-8111-111111111111'
const other = '22222222-2222-4222-8222-222222222222'
const initial = [
  {
    id: 'aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa',
    user_id: user,
    user: {
      id: user,
      first_name: 'Иван',
      last_name: 'Иванов',
      telegram_id: 123,
    },
    amount: 350,
    category: 'еда',
    spent_at: '2026-10-16',
    comment: 'Кофе',
    created_at: '2026-10-16T08:00:00Z',
    updated_at: '2026-10-16T08:00:00Z',
  },
  {
    id: 'bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb',
    user_id: other,
    user: {
      id: other,
      first_name: 'Анна',
      last_name: 'Иванова',
      telegram_id: 456,
    },
    amount: 1200,
    category: 'транспорт',
    spent_at: '2026-10-15',
    comment: 'Такси',
    created_at: '2026-10-15T08:00:00Z',
    updated_at: '2026-10-15T08:00:00Z',
  },
]

async function mockApi(
  page: Page,
  options: { empty?: boolean; failPost?: boolean } = {},
) {
  let records = options.empty ? [] : structuredClone(initial)
  await page.route('**/api/expenses/**', async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    if (request.method() === 'GET') {
      const result = records.filter(
        (item) =>
          (!url.searchParams.get('user_id') ||
            item.user_id === url.searchParams.get('user_id')) &&
          (!url.searchParams.get('category') ||
            item.category === url.searchParams.get('category')) &&
          (!url.searchParams.get('date_from') ||
            item.spent_at >= url.searchParams.get('date_from')!) &&
          (!url.searchParams.get('date_to') ||
            item.spent_at <= url.searchParams.get('date_to')!),
      )
      return route.fulfill({ json: result })
    }
    if (request.method() === 'POST') {
      if (options.failPost)
        return route.fulfill({ status: 404, json: { error: 'user not found' } })
      const body = request.postDataJSON()
      expect(Number.isInteger(body.amount)).toBe(true)
      const created = {
        ...initial[0],
        ...body,
        id: 'cccccccc-cccc-4ccc-8ccc-cccccccccccc',
        user_id: url.pathname.split('/')[3],
      }
      records.push(created)
      return route.fulfill({ status: 201, json: created })
    }
    const id = url.pathname.split('/')[4]
    if (request.method() === 'PUT') {
      records = records.map((item) =>
        item.id === id ? { ...item, ...request.postDataJSON() } : item,
      )
      return route.fulfill({ json: records.find((item) => item.id === id) })
    }
    records = records.filter((item) => item.id !== id)
    return route.fulfill({ status: 204 })
  })
}

test('filters, dates, empty state and no horizontal page overflow', async ({
  page,
}) => {
  await mockApi(page)
  await page.goto('/')
  await expect(
    page.getByRole('heading', { name: 'Расходы семьи' }),
  ).toBeVisible()
  await expect(page.getByRole('row')).toHaveCount(3)
  await expect(page.locator('svg')).toHaveCount(0)
  await expect(page.getByLabel('Итоги по выбранным фильтрам')).toContainText(
    '1 550',
  )
  expect(
    await page.evaluate(
      () =>
        document.documentElement.scrollWidth <=
        document.documentElement.clientWidth,
    ),
  ).toBe(true)
  await page.getByLabel('Категория', { exact: true }).selectOption('еда')
  await expect(page.getByRole('row')).toHaveCount(2)
  await expect(page.getByLabel('Итоги по выбранным фильтрам')).toContainText(
    '350',
  )
  await expect(page.getByRole('row').filter({ hasText: 'Кофе' })).toBeVisible()
  await page.getByLabel('С даты').fill('2026-10-17')
  await expect(page.getByText('Расходы по фильтрам не найдены.')).toBeVisible()
  await page.getByLabel('По дату').fill('2026-10-01')
  await expect(
    page.getByText('Дата окончания не может быть раньше даты начала.'),
  ).toBeVisible()
  await page.getByRole('button', { name: 'Сбросить фильтры' }).click()
  await expect(page.getByRole('row')).toHaveCount(3)
  await page.getByLabel('Участник', { exact: true }).selectOption(other)
  await expect(page.getByRole('row')).toHaveCount(2)
  await expect(page.getByRole('row').filter({ hasText: 'Такси' })).toBeVisible()
})

test('creates, edits, cancels deletion and deletes via actual API contract', async ({
  page,
}) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  await mockApi(page)
  await page.goto('/')
  await page
    .getByRole('button', { name: 'Добавить расход', exact: true })
    .click()
  const dialog = page.getByRole('dialog')
  await dialog.getByRole('button', { name: 'Добавить расход' }).click()
  await expect(
    dialog.getByText('Укажите корректный UUID участника.'),
  ).toBeVisible()
  await dialog.getByLabel('Участник', { exact: true }).fill(user)
  await dialog.getByLabel('Сумма, ₽').fill('800')
  await dialog.getByLabel('Комментарий').fill('Продукты')
  await dialog.getByRole('button', { name: 'Добавить расход' }).click()
  await expect(dialog).not.toBeVisible()
  await expect(page.getByRole('row')).toHaveCount(4)
  await page.getByRole('button', { name: 'Редактировать: Продукты' }).click()
  await dialog.getByLabel('Сумма, ₽').fill('950')
  await dialog.getByRole('button', { name: 'Сохранить изменения' }).click()
  await expect(dialog).not.toBeVisible()
  await expect(
    page.getByRole('row').filter({ hasText: 'Продукты' }),
  ).toContainText('950')
  await page.getByRole('button', { name: 'Удалить: Продукты' }).click()
  await dialog.getByRole('button', { name: 'Отмена' }).click()
  await expect(page.getByRole('row')).toHaveCount(4)
  await page.getByRole('button', { name: 'Удалить: Продукты' }).click()
  await dialog
    .getByRole('button', { name: 'Удалить расход', exact: true })
    .click()
  await expect(dialog).not.toBeVisible()
  await expect(page.getByRole('row')).toHaveCount(3)
  expect(errors).toEqual([])
})

test('keeps form values on server failure and supports keyboard dismissal', async ({
  page,
}) => {
  await mockApi(page, { failPost: true })
  await page.goto('/')
  await page
    .getByRole('button', { name: 'Добавить расход', exact: true })
    .click()
  const dialog = page.getByRole('dialog')
  await dialog.getByLabel('Участник', { exact: true }).fill(user)
  await dialog.getByLabel('Сумма, ₽').fill('500')
  await dialog.getByRole('button', { name: 'Добавить расход' }).click()
  await expect(dialog.getByRole('alert')).toContainText('404')
  await expect(dialog.getByLabel('Сумма, ₽')).toHaveValue('500')
  await page.keyboard.press('Escape')
  await expect(dialog).not.toBeVisible()
})

test('shows genuine empty state', async ({ page }) => {
  await mockApi(page, { empty: true })
  await page.goto('/')
  await expect(page.getByText('Расходов пока нет.')).toBeVisible()
  await page.getByRole('button', { name: 'Добавить первый расход' }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
})

test('shows connection error without fictional expenses', async ({ page }) => {
  await page.route('**/api/expenses/**', (route) =>
    route.fulfill({ status: 503, json: { error: 'unavailable' } }),
  )
  await page.goto('/')
  await expect(page.getByRole('alert')).toContainText('503')
  await expect(page.getByRole('table')).not.toBeVisible()
  await expect(page.getByRole('button', { name: 'Повторить' })).toBeVisible()
})
