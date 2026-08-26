# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Backend service for [friendly](https://github.com/sayze/friendly): a small CRUD API
for a "friends" roster, with avatar images hosted on Cloudinary. Go, chi
router, cmd pattern for wiring, lightweight DDD (domain / service / infra)
for the friend roster logic. Modernized from a pre-modules, Go 1.12 codebase
to match the conventions established in the sibling `qotd-api` repo.

## Architecture (cmd pattern)

- `cmd/api/main.go` — entrypoint/composition root. Loads `config.Load()`,
  builds the in-memory `domain.FriendRepository`, the Cloudinary-backed
  `domain.ImageStore`, wires them into `service.NewService`, builds the
  router via `server.New(svc)`, and starts `http.ListenAndServe` on
  `cfg.Addr` (`ADDR` env var, default `:4040`).
- `internal/config/config.go` — env var loading (`ADDR`, `CDN_UPLOAD_URL`,
  `CDN_API_KEY`, `CDN_API_SECRET`), plain `os.Getenv` with defaults, no
  third-party config library.
- `internal/server/server.go` — chi router setup. Middleware: chi's default
  stack (`RequestID`, `ClientIPFromHeader("CF-Connecting-IP")` — the API
  sits behind Cloudflare/Traefik in production — `Logger`, `Recoverer`),
  permissive CORS (`github.com/go-chi/cors`, all origins), and
  `httprate.LimitByIP` capping each client to 50 requests/minute. No auth —
  this API has never required it and none was added as part of the
  modernization.
- `internal/handler/friend.go` — `FriendHandler`, pure HTTP transport: parses
  requests (including multipart form uploads), calls into `service.Service`,
  encodes JSON responses. Contains no business logic or storage/CDN calls.

### Friend domain (`internal/friend/`)

Lightweight DDD split, mirroring `qotd-api`'s `internal/<product>/<entity>/...`
layout:

- `internal/friend/domain/friend.go` — the `Friend{ID, Name, Image}` entity,
  two ports the service depends on (`FriendRepository` for persistence,
  `ImageStore` for avatar hosting), and the sentinel errors
  `ErrFriendNotFound` / `ErrLimitExceeded`.
- `internal/friend/service/service.go` — `service.Service`, the business
  logic layer handlers call into. Enforces the roster size cap
  (`MaxFriends = 100`), orchestrates image upload/deletion against
  `ImageStore` alongside persistence via `FriendRepository`, and applies
  partial-update semantics (an empty name/image on `Update` leaves the
  existing value untouched). Storage- and CDN-agnostic — depends only on the
  two domain interfaces, not on any concrete implementation.
- `internal/friend/infra/memory/repository.go` — an in-memory, mutex-guarded,
  insertion-ordered implementation of `FriendRepository`. State does not
  survive a restart; it's the only backend this service has (there's no
  database yet). IDs are assigned from a monotonically increasing counter,
  not `len(slice)+1` as the original implementation did — the old scheme
  could reissue an ID that was still in use after a friend in the middle of
  the roster was deleted.
- `internal/friend/infra/cloudinary/cloudinary.go` — the Cloudinary-backed
  `ImageStore` implementation. Signs upload/delete requests per Cloudinary's
  signature scheme (`sign()`), uploads via a multipart POST, and returns the
  `secure_url` from the response. This used to live directly in the HTTP
  handler; it's now an injected infra port so the handler and service never
  talk to Cloudinary directly, and so it can be swapped for a test double (or
  a different provider) without touching business logic.

There's no `nominee`-style "who's currently active" state here — `Friend` is
the only entity in this service, so the domain/service split exists purely
to separate storage-agnostic business rules (the roster cap, partial
updates, image orchestration) from the storage and CDN implementations
themselves.

## API contract

Preserved from the pre-modernization implementation (the `friendly` frontend
depends on it) — response envelope, routes, and status codes are unchanged;
only the internal implementation moved:

- `GET /` → health check, `{"status": "ok"}`.
- `GET /friend?search=<term>` → `{"status": "ok", "data": [...]}`, all
  friends whose name contains `search` (case-insensitive substring; omit for
  the full roster).
- `GET /friend/{id}` → a single friend, or a `404` with
  `{"status": "resource not found", "data": "<message>"}` if `id` doesn't
  exist.
- `POST /friend` (`multipart/form-data`: `name` required, `image` file
  optional) → creates a friend. `400` if the roster is already at
  `MaxFriends` (100) or `name` fails validation (`min=2,max=20`).
- `PATCH /friend` (`multipart/form-data`: `id`, `name` required, `image`
  file optional) → partial update; an omitted `image` leaves the existing
  avatar untouched.
- `DELETE /friend/{id}` → deletes the friend's stored image, then the
  record. If the image delete fails, the record is left in place (matches
  the original behavior).
- All error responses: `{"status": "<text>", "error": "<message>"}`.

## Required env vars

- `ADDR` — listen address, defaults to `:4040`.
- `CDN_UPLOAD_URL` — Cloudinary upload endpoint, defaults to the `sayze`
  Cloudinary account's upload URL.
- `CDN_API_KEY` / `CDN_API_SECRET` — Cloudinary API credentials, used to
  sign upload/delete requests. Empty by default; image upload/delete calls
  will fail against the real Cloudinary API without them, but the rest of
  the service (including creating/updating friends without an image) works
  fine locally without them.

## Docker

Two-stage build, matching `qotd-api`:
1. `golang:1.25-alpine` — installs `ca-certificates`, builds a static binary
   (`CGO_ENABLED=0`).
2. `FROM scratch` — copies in only the binary and
   `/etc/ssl/certs/ca-certificates.crt` (needed for the outbound HTTPS calls
   to Cloudinary).

```
docker build -t friendly-api:dev .
docker run --rm -p 4040:4040 friendly-api:dev
```

For local dev, `docker-compose.yml` builds and runs the same image; copy
`.env.example` to `.env` (all vars optional for local dev — Cloudinary
credentials only matter if you're exercising image upload/delete against a
real account), then `docker compose up --build`.

## Deployment (`friendly-api.nomad.hcl`)

Replaces the previous Kubernetes/Helm deployment (`charts/` +
`.github/workflows/build_deploy.yml`, both removed) with the Nomad+Traefik
pattern used by `qotd-api` and the rest of the homelab. Fronted by Traefik
on its own subdomain (`Host(friendly-api.sayedsadeed.com)`, no path prefix —
unlike `qotd-api`, this service isn't sharing the `api.sayedsadeed.com`
domain). Since the health check endpoint (`GET /`) needs no auth, the Nomad
service check is a real `http` check against `/` rather than a bare `tcp`
check. `CDN_API_KEY`/`CDN_API_SECRET` are pulled from Vault at runtime via a
`template` stanza (`vault { policies = ["nomad"] }`), reading
`secret/data/homelab/friendly-api`'s `cdn_api_key`/`cdn_api_secret` fields —
that Vault path/keys need to exist before this job can start.

`.github/workflows/test.yml` runs `make unit` on every push/PR.
`.github/workflows/deploy.yml` runs on push to `master`: test → build/push
the Docker image (tagged by branch and short SHA) → `nomad job run` against
the homelab Nomad cluster.

## Related

- Frontend lives in a separate sibling repo: `sayze/friendly`.
