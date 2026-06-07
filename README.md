# SnackMates

Snack pen pals — build snack wishlists, get randomly matched with another user, and send care packages.

## Stack

| Layer | Technology |
|-------|------------|
| Frontend | Next.js, Adobe React Spectrum (Aria) |
| API | Go (Fiber) |
| Database | PostgreSQL 18 |
| Search | Meilisearch |
| Cache | Valkey |
| Object storage | S3-compatible (OVH Object Storage in prod; MinIO locally) |
| Auth | Email/password, Discord OAuth, TOTP, WebAuthn |

## Quick start

### 1. Infrastructure

```bash
cp .env.example .env
docker compose up -d
```

Services:

- Postgres: `localhost:5432`
- Valkey: `localhost:6379`
- Meilisearch: `localhost:7700`
- MinIO (S3): `localhost:9000` (console `:9001`)
- Mailpit (dev email): UI `localhost:8025`, SMTP `localhost:1025`

### 2. API

```bash
cd api
go run ./cmd/server
```

The API runs pending database migrations automatically on startup (from the `migrations/` directory). API listens on `http://localhost:8080`.

### 3. Web

```bash
cd web
cp ../.env.example .env.local   # or set NEXT_PUBLIC_API_URL
npm run dev
```

App: `http://localhost:3000`

## Environment

Copy `.env.example` to `.env` for the API and `.env.local` for the web app.

### Config file (optional)

By default the API loads settings from environment variables with the same built-in dev defaults as `.env.example`. You can optionally use a TOML config file instead:

```bash
cp api/config.example.toml api/config.toml
CONFIG_FILE=./config.toml go run ./cmd/server
# or
go run ./cmd/server --config ./config.toml
```

Environment variables always override values from the TOML file, so secrets can stay in env even when using a config file.

Discord OAuth requires `DISCORD_CLIENT_ID` and `DISCORD_CLIENT_SECRET` from the [Discord Developer Portal](https://discord.com/developers/applications). Set the redirect URI to:

`http://localhost:8080/api/v1/auth/discord/callback`

## Features

### Wishlists
Registered users create wishlists and add snack items (name, brand, notes, link, priority). Public items are indexed in Meilisearch.

### Snack matching
The `/api/v1/matches/run` endpoint pairs verified users who:
- Have a country set on their profile
- Have at least one item on a public wishlist
- Do not already have an active/pending match

Users are only matched with snack mates from a **different country**. Pairing uses a randomized cross-country algorithm.

### Authentication
- **Email**: register → verification email (Mailpit in dev) → login
- **Discord OAuth**: optional when credentials are configured
- **MFA**: TOTP (authenticator apps) and WebAuthn (security keys) via Settings

### Storage

S3-compatible object storage (OVH, MinIO, AWS, etc.) via the AWS SDK with a custom endpoint — no AWS-specific services required.

- `client-assets` bucket: user uploads (profile avatars, typically private)
- `static-assets` bucket: first-party branding/static files

**OVH Object Storage example:**

```bash
S3_ENDPOINT=https://s3.gra.io.cloud.ovh.net
S3_REGION=gra
S3_USE_PATH_STYLE=true
S3_PRESIGN_PRIVATE_OBJECTS=true
```

Private avatar objects are served via presigned URLs by default (recommended for OVH). For local MinIO with public buckets, set `S3_PRESIGN_PRIVATE_OBJECTS=false`.

## Project layout

```
SnackMates/
├── api/                 Go Fiber API
│   ├── cmd/server/      Entry point
│   ├── internal/        Auth, handlers, matching, search, cache, storage
│   └── migrations/      Postgres schema (applied automatically on API startup)
├── web/                 Next.js + Adobe Spectrum frontend
├── docker-compose.yml   Local infrastructure
└── .env.example
```

## API overview

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/auth/register` | Create account |
| POST | `/api/v1/auth/login` | Sign in (supports TOTP step-up) |
| GET | `/api/v1/auth/discord` | Start Discord OAuth |
| GET/POST | `/api/v1/auth/mfa/*` | TOTP & WebAuthn setup |
| CRUD | `/api/v1/wishlists/*` | Wishlists & items |
| GET | `/api/v1/search?q=` | Meilisearch snack search |
| GET/POST | `/api/v1/matches/*` | View matches / run pairing |

## Development notes

- Email verification and password reset links point at the web app (`WEB_ORIGIN`).
- Sessions are stored in Postgres with HTTP-only cookies and Bearer token support.
- Valkey caches OAuth state during Discord login.
- For production, use OVH Object Storage (or another S3-compatible provider) and configure TLS, secrets, and SMTP.
# SnackMates
