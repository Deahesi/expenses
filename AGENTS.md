# AGENTS.md

## Project

Go **DDD-inspired modular monolith** for family expense tracking: REST API for the web panel and a Telegram bot using an LLM. Both use the same application services. **Follow existing code and conventions; do not restructure the project.**

## Architecture

```text
cmd/server/                     # Startup and dependency wiring
internal/
  domain/                       # Plain entities, types, domain errors
  <feature>/                    # service.go, input.go, repository.go
  httpapi/
    <feature>Handlers/          # <feature>.go, dto.go
    requests/ responses/ middlewares/
    server.go                   # Routing
  bot/                          # Telegram transport
  llm/                          # LLM integration
  storage/postgres/             # Repository implementations
  storage/postgres/gen/         # sqlc-generated code
  app/                          # Optional application setup
migrations/                     # Goose migrations
```

- **Domain:** plain structs, types, and errors; no HTTP, database, or LLM dependencies. Do not introduce rich domain models.
- **Feature services:** all business logic, validations, authorization, and use-case coordination. Keep `input.go` beside the service.
- **Repositories:** small interfaces beside the consuming service; PostgreSQL implementations in `storage/postgres/`.
- **HTTP:** handlers and their `dto.go` grouped in `httpapi/<feature>Handlers/`. DTOs are not service inputs. Use shared `requests`/`responses` helpers.
- **Bot/LLM:** translate messages into validated service calls; never access repositories directly from transport handlers or the LLM.
- **Dependencies:** use explicit constructor injection; wire services, handlers, and routes at startup. No cyclic imports, global registries, or unnecessary abstractions.

## Security and behavior

- Get Telegram `user_id` from the **Telegram update**, never from text or LLM tool arguments.
- Scope bot reads, edits, and deletes to that user **in service/repository logic**.
- The web panel intentionally manages all users' expenses without authentication (per task); keep admin and user-scoped operations separate.
- Treat LLM output as untrusted: validate amounts, categories, dates, and identifiers in Go. Resolve ambiguous expense references safely; ask for clarification if necessary.
- Pass `context.Context` through services, repositories, and external calls. Follow existing error-handling conventions.

## Database

- Use **PostgreSQL + pgx + sqlc**. Edit SQL sources and regenerate sqlc output; never edit generated files manually or leak generated DB types beyond persistence.
- Use **Goose** migrations with Up/Down sections where feasible. Never modify an applied migration or perform destructive DB operations without approval.
- Use parameterized SQL and user-scoped queries where appropriate.

## Agent workflow

1. Read this file, `go.mod`, and relevant existing modules **before editing**. Match actual package names and patterns.
2. Make the smallest complete change: domain → service/input/repository contract → persistence → handler/DTO → wiring/routes, **only where needed**.
3. Reuse existing helpers; do not add dependencies, frameworks, layers, or unrelated refactors without justification.
4. Add focused tests (mock repository/LLM boundaries; no real LLM calls). Format with `gofmt` and, when available, run `go test ./...` and `go build ./...`.
5. Report changed files, verification results, and unresolved issues. Never claim checks passed unless executed; never leave disconnected modules or placeholder implementations.
