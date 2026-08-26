# Friendly API

Backend service for [friendly](https://github.com/sayze/friendly): a small CRUD API for a
"friends" roster, with avatar images hosted on Cloudinary.

## Architecture

Go, [chi](https://github.com/go-chi/chi) router, cmd-pattern composition
root, and a lightweight DDD split for the friend roster logic:

```
cmd/api/main.go              entrypoint / composition root
internal/config               env var configuration
internal/server                router + middleware
internal/handler                HTTP transport (requests in, JSON out)
internal/friend/domain           entity + ports (FriendRepository, ImageStore)
internal/friend/service           business logic (roster cap, partial updates, image orchestration)
internal/friend/infra/memory       in-memory FriendRepository
internal/friend/infra/cloudinary    Cloudinary-backed ImageStore
```

`service.Service` depends only on the two domain interfaces, not on the
memory or Cloudinary implementations directly, so either can be swapped
(e.g. for a database-backed repository, or a test double) without touching
business logic. See [CLAUDE.md](./CLAUDE.md) for the full design rationale.

## Running locally

Requires Go 1.25+.

```
go run ./cmd/api
```

Or via Docker Compose (copy `.env.example` to `.env` first):

```
docker compose up --build
```

The API listens on `:4040` by default (`ADDR` env var). Cloudinary
credentials (`CDN_API_KEY` / `CDN_API_SECRET`) are optional for local dev —
everything works except actual image upload/delete without them.

## Testing

```
make unit
```

Runs the full unit test suite (`go test --tags unit ./...`) through
[tparse](https://github.com/mfridman/tparse) for readable output. Run a
single package or test with `make unit path=internal/friend/service` or
`make unit path=internal/friend/service test=TestService_Create`.

## API

- `GET /` — health check.
- `GET /friend?search=<term>` — list friends, optionally filtered by name.
- `GET /friend/{id}` — get a single friend.
- `POST /friend` — create (`multipart/form-data`: `name`, optional `image`).
- `PATCH /friend` — update (`multipart/form-data`: `id`, `name`, optional
  `image`).
- `DELETE /friend/{id}` — delete.

See [CLAUDE.md](./CLAUDE.md) for the full response envelope and error
shapes.

## Deployment

Deploys as a Docker image to a [Nomad](https://www.nomadproject.io/) cluster
via `friendly-api.nomad.hcl`, fronted by Traefik. See
`.github/workflows/deploy.yml` for the CI pipeline (test → build/push →
`nomad job run`).
