# ManyACG Copilot Instructions

These notes guide AI coding agents working on this repo. Focus on existing patterns; do not introduce new architectures unless requested.

## Big Picture
- **Purpose**: Backend for ManyACG – collect, download, organize and serve anime illustrations, mainly driven by a Telegram bot and optional REST API.
- **Entry / CLI**: `main.go` → `cmd/root.go` → `cmd/run.go` (`manyacg` binary). Runtime wiring (config, infra, services, interfaces) is done in `cmd/run.go`.
- **Layering**:
  - `internal/infra`: low-level infrastructure (config, database, kvdb, search, storage backends, tagging, sources).
  - `internal/model`: domain layer – `entity` (GORM models), `dto` (API/search/Telegram shapes), `command` (write commands), `query` (read filters) and `converter` (entity/DTO mapping).
  - `internal/repo`: repository interfaces + implementations on top of `infra/database.DB`, sometimes decorated with event buses.
  - `internal/service`: business logic orchestrating multiple repos + storage + search + tagging.
  - `internal/interface`: integration interfaces – `telegram` bot, `rest` HTTP API (Fiber v3), `scheduler` for automatic posting.
  - `internal/shared`: shared enums/interfaces/errors used across layers.
  - `pkg/*`: self-contained utilities (logging, string/url utils, OS helpers, etc.). Prefer them over adding new helpers.

## Key Runtime Patterns
- **Configuration**: Loaded via `internal/infra/config/runtimecfg` from `config.toml`. New config knobs should be added there and plumbed into `cmd/run.go` and relevant infra/service code, not read directly from env.
- **Startup Flow** (see `cmd/run.go`):
  - Read runtime config, initialize logging with `pkg/log`.
  - Call `infra.Init` (KV store, storage, database, etc.).
  - Initialize repository bundle via `internal/infra/database` and `internal/repo` (`repo.NewWithArtworkEventImpl` when search is enabled).
  - Build a single `service.Service` with repos/search/tagger/storages/sources and set it as default (`service.SetDefault`).
  - Conditionally start: Telegram bot (`internal/interface/telegram`), scheduler poster, REST API server (`internal/interface/rest`).
- **Database**: `internal/infra/database/db.go` manages a global GORM `DB`. It auto-migrates all entities and applies PostgreSQL-specific index optimizations (PGroonga or `pg_trgm`). New tables/indexes should follow this pattern and be added to `AutoMigrate` and, where appropriate, optimized in `optimizePgSQL`.
- **ID / Types**: Domain IDs use `ouid.OUID` (UUID-like) with JSON and GORM integration. Many entities implement interfaces from `internal/shared` (e.g. `ArtworkLike`, `MediaLike`) – new entities should implement existing interfaces instead of inventing parallel ones.

## Repos, Services, Events
- **Repositories**: Define behaviour in `internal/repo/*.go` (e.g. `Artwork`, `Artist`, `Tag`). Concrete DB operations live in `internal/infra/database/*.go`. When adding a new capability, first extend the repo interface, then implement it in the DB layer.
- **Transactions**: Use `repos.Transaction(ctx, func(repos repo.Repositories) error { ... })` from `internal/repo` (see `service/artwork.go`) instead of hand-rolling GORM transactions.
- **Events for Search**: Artwork-related repos can be wrapped in `ArtworkWithEvent` / `ArtworkWithRecorder` and plugged into `repo.NewWithArtworkEventImpl`. These publish `dto.ArtworkEventItem` to an `eventbus`, which `cmd/run.go` wires into `infra/search`. For changes that should update MeiliSearch, go through the event-aware repos instead of calling search directly.
- **Services**: Service methods are thin orchestrators that:
  - Validate / normalize commands (`internal/model/command`).
  - Fetch/update via repos (and their decorators).
  - Coordinate tag/artist creation, cached-artwork status, deletion tracking, etc.
  - Convert to DTOs as needed using `internal/model/converter`.

## Interfaces (REST, Telegram, Scheduler)
- **REST API**: `internal/interface/rest` uses Fiber v3.
  - `api.go` constructs the app, middlewares (CORS, ETag, compress, limiter, logger), and a common error handler mapping domain errors (e.g. `errs.ErrRecordNotFound`) to proper HTTP codes.
  - `handlers/` contains route handlers and should depend on `service.Service` and DTO/command types, not directly on GORM or infra.
  - When adding endpoints, register them via `handlers.Register` and reuse validator/DTO patterns in existing handlers.
- **Error Handling**: Surface errors as `common.Error` values or `fiber.Error` in REST, and let the central error handler map them. In services/repos, prefer returning `internal/shared/errs` values for shared conditions.
- **Telegram/Scheduler**: `internal/interface/telegram` owns all bot-specific logic; `internal/interface/scheduler` handles periodic posting using `scheduler.ArtworkPoster`. New bot behaviours should go there rather than into services directly.

## Workflows & Commands
- **Build / Run**:
  - Local run: `go run ./cmd` (or simply build `manyacg` via `go build -o manyacg .` and run).
  - Docker: use the provided `Dockerfile` / scripts under `scripts/` when adjusting container behaviour.
- **Config for Dev**: Minimal `config.toml` as shown in `README.md` (Telegram, storage, Pixiv cookies, DB `dsn`, search config). Avoid hard-coding secrets in code; keep them in config.
- **Tests**: Existing tests live mainly under `pkg/*` and `internal/*`. When modifying a package with tests, extend those patterns instead of adding parallel test helpers elsewhere.

## Conventions for New Code
- Prefer adding new behaviour by extending existing layers:
  - New data source → `internal/infra/source` + `internal/shared/source_type_enum.go` + relevant service methods.
  - New storage backend → `internal/infra/storage` + `shared.StorageType` + wiring in `cmd/run.go`.
  - New search feature → extend `infra/search`, repo events, and DTOs.
- Follow existing logging style (`pkg/log` or `charm` logger wrappers) and avoid creating ad-hoc loggers.
- Keep public APIs (REST, Telegram commands) thin and delegated: parse/validate → call service → convert to DTO/response.
