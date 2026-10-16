# Expenses Service

## Docker

Нужен Docker с Docker Compose v2. Команды выполняются из корня проекта.

## Окружение

```bash
cp .env.docker.example .env.docker
```

Заполните `TELEGRAM_BOT_API_KEY` и `OPENROUTER_API_KEY` в `.env.docker`.

| Переменная | Описание |
| --- | --- |
| `POSTGRES_USER` | Пользователь PostgreSQL. |
| `POSTGRES_PASSWORD` | Пароль PostgreSQL. |
| `POSTGRES_DB` | Название базы. |
| `DATABASE_URL` | Подключение к `postgres:5432`; данные должны совпадать с `POSTGRES_*`. |
| `REDIS_ADDR` | Адрес Redis: `redis:6379`. |
| `HTTP_ADDR` | Адрес бэкенда внутри контейнера: `:8080`. |
| `TELEGRAM_BOT_API_KEY` | Токен бота из [BotFather](https://t.me/BotFather). |
| `OPENROUTER_API_KEY` | [Ключ OpenRouter](https://openrouter.ai/settings/keys). |
| `LLM_MODEL` | Модель с поддержкой tools: `openrouter/<provider>/<model>`. |
| `LLM_TIMEZONE` | Часовой пояс, например `Asia/Novosibirsk`. |

## Запуск

```bash
docker compose --env-file .env.docker up -d --build
```

Миграции применяются автоматически. Бот запускается вместе с бэкендом.

| Сервис | Адрес |
| --- | --- |
| Веб-панель | [http://localhost:8081](http://localhost:8081) |
| Backend API | [http://localhost:8080](http://localhost:8080) |
| PostgreSQL | `localhost:5432` |

## Команды

```bash
# Состояние сервисов
docker compose --env-file .env.docker ps -a

# Логи
docker compose --env-file .env.docker logs -f

# Остановка без удаления данных PostgreSQL
docker compose --env-file .env.docker down
```
