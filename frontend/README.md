# Семейные расходы

Веб-панель главы семьи без авторизации. React + Vite + TypeScript, React Compiler, shadcn/ui на Base UI, Tailwind CSS, TanStack Query, axios и React Hook Form.

## Локальный запуск

Рекомендуется Node.js 24 LTS (минимальные версии указаны в `package.json`).

```sh
npm ci
cp .env.example .env
npm run dev
```

Открыть http://localhost:5173. Бэкенд должен работать на http://localhost:8080. Другой адрес задаётся через `API_PROXY_TARGET` в `.env`. Запросы идут через Vite-прокси `/api`, поэтому настройка CORS в Go не нужна.

`VITE_API_BASE_URL` по умолчанию `/api`. Это переменная времени сборки; для прямого обращения к другому origin на сервере понадобится CORS.

## Dockerfile фронтенда

```sh
docker build -t expenses-frontend .
docker run --rm -p 8081:80 expenses-frontend
```

Панель доступна на http://localhost:8081. Бэкенд запускается отдельно; по умолчанию nginx проксирует `/api` на `http://host.docker.internal:8080`. Другой адрес задаётся переменной `API_UPSTREAM` при запуске контейнера.

## Что реализовано

- Список расходов всей семьи, сортировка по дате и постраничный вывод.
- Серверные фильтры по участнику, категории и обеим границам периода.
- Создание, изменение и удаление с подтверждением.
- Всего потрачено и количество записей для текущей выборки — текстом в одну строку.
- Нейтральная чёрно-белая тема shadcn; категории выделяются цветом через `colorByTextHash`.
- Формы расходов и фильтров на React Hook Form; валидация, ошибки запросов, состояния загрузки и пустого списка.
- Обновление кэша после изменений, автообновление каждые 60 секунд и при возвращении к вкладке.
- Адаптация к мобильному экрану, клавиатурная навигация, доступные диалоги Base UI.

## Структура

Компактная FSD без лишних слоёв и barrel-файлов:

```text
src/
  app/                         # провайдеры, конфигурация QueryClient, стили
  pages/expenses/              # сборка экрана и состояние открытого диалога
  widgets/                     # таблица и текстовые итоги
  features/
    filter-expenses/           # фильтры списка
    manage-expense/            # форма, валидация, create/update mutation
    delete-expense/            # подтверждение и delete mutation
  entities/expense/
    api/                       # REST-запросы через axios
    queries/                   # TanStack Query options и ключи
    model/                     # типы и категории
    ui/                        # отображение категории
  entities/user/
    model/                     # тип User
    api/                       # getUsers
    queries/                   # usersQuery и ключи пользователей
  shared/                      # axios instance, форматирование, базовый UI
```

Импорты идут вниз по слоям. Каждый компонент находится в отдельном `.tsx`, его props объявлены в том же файле. Модели, API и хуки находятся в `.ts` рядом со своим функциональным блоком. Компоненты shadcn установлены официальным CLI со стилем `base-nova`; базовые примитивы — Base UI. Для простой фильтрации используются нативные select и date input.

React Compiler подключён через `reactCompilerPreset` и `@rolldown/plugin-babel` в `vite.config.ts`. Правила React и Compiler проверяет ESLint. Ручные `useMemo` и `useCallback` не нужны.

## Контракт API

| Действие  | Запрос                            |
| --------- | --------------------------------- |
| Список    | `GET /expenses/`                  |
| Создание  | `POST /expenses/{user_id}`        |
| Изменение | `PUT /expenses/{user_id}/{id}`    |
| Удаление  | `DELETE /expenses/{user_id}/{id}` |

Фильтры: `user_id`, `category`, `date_from`, `date_to`. Даты — `YYYY-MM-DD`, обе границы включаются. Пустые фильтры не передаются.

`amount` — **целые рубли**, 1 ≤ amount < 10¹⁰. API и его тест `TestExpenseDTOUsesWholeRubles` подтверждают это, хотя комментарий модели Go ещё упоминает копейки. Фронтенд не умножает и не делит сумму на 100.

`user_id` — **внутренний UUID**, а не числовой Telegram ID. API списка пользователей и имён пока нет: фильтр строится из всех расходов, участники показываются по короткому UUID, полный UUID доступен в подсказке строки. В форме можно выбрать известный UUID или ввести UUID уже зарегистрированного участника. Если участник ещё не имеет расходов, его UUID нужно получить из бэкенда. Фронтенд не создаёт пользователей и не придумывает им имена.

## Проверки

```sh
npm run check             # ESLint + unit-тесты + TypeScript + production build
npm run format:check
npx playwright install chromium
npm run test:e2e          # desktop + mobile
```

Можно использовать уже установленный Google Chrome:

```sh
PLAYWRIGHT_CHANNEL=chrome npm run test:e2e
```

Unit-тесты проверяют границы суммы, UUID, реальные календарные даты, Unicode-комментарии, URL и JSON API, ответ DELETE 204 и ошибки сети. Браузерные тесты проверяют полный CRUD, фильтры, ошибки формы и сервера, сохранение данных при неудаче, отмену удаления, клавишу Escape и мобильную ширину. API в браузерных тестах подменяется на уровне HTTP; отдельный тестовый сервер на порту 4173 использует `/api` независимо от локальной `.env`. Реальная БД и Telegram не изменяются.

Результаты самопроверки и исправления записаны в `REVIEW.md`.

`getUsers(signal?)` запрашивает `GET /users/`; `usersQuery()` предоставляет query options для TanStack Query. В текущем бэкенде маршрут `/users/` ещё не зарегистрирован.
