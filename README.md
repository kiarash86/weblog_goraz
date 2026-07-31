# Marginalia — Weblog App

A shared weblog built with **Go**, **Echo**, and **PostgreSQL**, with a vanilla JS frontend. Users can post public or private entries, share private entries with specific people, and leave comments ("notes in the margin").

**Live app:** https://webloggoraz-production.up.railway.app (API) — frontend deployed on Netlify
**Repo:** https://github.com/kiarash86/weblog_goraz

---

## Features

- Sign up / log in with username + password (JWT-based auth)
- Create posts (title, content, optional image, public/private)
- Public feed with search (by title) and pagination
- Private posts, shareable with specific usernames
- Comments on any post you have access to
- Delete your own posts and comments (no editing — by design)
- Image upload with size/type validation

---

## Tech Stack

- **Backend:** Go 1.26, [Echo v5](https://echo.labstack.com/)
- **Database:** PostgreSQL, accessed via [pgx](https://github.com/jackc/pgx)
- **Migrations:** [golang-migrate](https://github.com/golang-migrate/migrate)
- **Auth:** JWT ([golang-jwt](https://github.com/golang-jwt/jwt)) + bcrypt
- **Config:** [cleanenv](https://github.com/ilyakaznacheev/cleanenv) + `.env` via [godotenv](https://github.com/joho/godotenv)
- **Frontend:** Vanilla HTML/CSS/JS (no build step), in `frontend/`

---

## Getting Started

### Prerequisites
- Go 1.26+
- Docker (for running Postgres locally) or a running Postgres instance

### 1. Clone and configure
```bash
git clone git@github.com:kiarash86/weblog_goraz.git
cd weblog_goraz
cp .env.example .env   # then edit values as needed
```
> Note: `.env.example` doesn't exist in the repo yet even though `.gitignore` references it — add one with the three variables below before this step works as written, or just create `.env` directly.

Environment variables (see `internal/config/config.go`):

| Variable       | Default (if unset)                                                    | Description                  |
|----------------|-------------------------------------------------------------------------|-------------------------------|
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/weblog?sslmode=disable`   | Postgres connection string   |
| `JWT_KEY`      | *(insecure placeholder)*                                                 | Secret used to sign JWTs — **set a real value in production** |
| `PORT`         | `8080`                                                                   | HTTP port the server listens on |

> ⚠️ Don't rely on the built-in defaults for `JWT_KEY` outside local dev — always set a strong secret via `.env` or your host's environment variables before deploying.

### 2. Start Postgres
```bash
docker-compose up -d
```
This starts a Postgres 16 container on `localhost:5432` with database `weblog`.

### 3. Run the server
```bash
go run main.go
```
Database migrations (`internal/migration/*.sql`) run automatically on startup. The API will be available at `http://localhost:8080`.

### 4. Run the frontend
The frontend is static — no build step needed. Point `frontend/config.js` at your API URL:
```js
window.MARGINALIA_CONFIG = {
  API_BASE: 'http://localhost:8080',
};
```
Then serve the `frontend/` folder with any static file server, e.g.:
```bash
npx serve frontend
```

---

## API Reference

All protected routes require an `Authorization: Bearer <token>` header.

| Method | Path                          | Auth | Description                                  |
|--------|-------------------------------|------|-----------------------------------------------|
| GET    | `/health`                     | —    | Health check                                  |
| POST   | `/signup`                     | —    | Create an account, returns `{ user, token }`  |
| POST   | `/login`                      | —    | Log in, returns `{ user, token }`             |
| GET    | `/weblog`                     | ✅   | Feed — supports `?page=` and `?search=` (title, partial match). `Next-Page` response header indicates if another page exists. |
| POST   | `/weblog`                     | ✅   | Create a post — `{ title, content, is_private, img_path }` |
| GET    | `/weblog/:id`                 | ✅   | Get a single post (403 if private and no access) |
| DELETE | `/weblog/:id`                 | ✅   | Delete a post (owner only)                    |
| POST   | `/weblog/:id/share`           | ✅   | Share a private post — `{ usernames: [...] }` (owner only) |
| GET    | `/weblog/:id/comment`         | ✅   | List comments on a post you can access        |
| POST   | `/weblog/:id/comment`         | ✅   | Add a comment — `{ content }`                 |
| DELETE | `/weblog/:id/comment/:commentId` | ✅ | Delete a comment (author only)              |
| POST   | `/upload`                     | ✅   | Upload an image (multipart `image` field, max 5MB, jpg/jpeg/png/gif/webp) — returns `{ path }` to use as `img_path` |
| GET    | `/uploads/*`                  | —    | Serves uploaded images as static files        |

---

## Project Structure

```
weblog/
├── main.go                      # entrypoint, routes, middleware
├── internal/
│   ├── auth/                    # JWT creation/parsing
│   ├── config/                  # env-based config loading
│   ├── db/                      # Postgres connection pool
│   ├── handlers/                # HTTP handlers (auth, board, comment, share, upload)
│   ├── middlewares/             # auth middleware
│   ├── migration/                # embedded SQL migrations, run on startup
│   ├── models/                  # data structs
│   └── repository/              # SQL queries
├── frontend/                    # static HTML/CSS/JS client
├── Dockerfile
└── docker-compose.yml            # local Postgres for development
```

---

## Notes / Known Limitations

- No post editing — deleting and re-creating is the only way to change a post (matches the assignment spec).
- No email verification or password reset flow.
- Uploaded file type is validated by extension only, not file content — good enough for this project, not hardened for adversarial input.
- No automated tests yet.

---

## Deployment

The API is deployed on [Railway](https://railway.app) via the included `Dockerfile`; the frontend is deployed on [Netlify](https://netlify.com) as a static site. `docker-compose.yml` is for local development only (Postgres container) — it does not run the app itself.
