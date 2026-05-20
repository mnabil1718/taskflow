# TaskFlow — Agent Context

## Project goal

Build a fullstack task management app for a technical test.
Evaluation: backend quality (30pts), frontend (25pts), Docker (20pts),
code quality (15pts), bonus (10pts).

## Hard constraints (instant fail if violated)

- Never store plaintext passwords — always bcrypt
- Never hardcode secrets — use env vars only
- Always validate API inputs
- docker compose up must work first try

## Tech stack

- Backend: Go with Fiber, JWT auth, PostgreSQL 15
- Frontend: Next.js 14 App Router, TypeScript strict, TanStack Query, Tailwind
- Infra: Docker Compose (BE + FE + DB in one command)

## Folder structure

backend/
cmd/main.go
internal/handler/
internal/service/
internal/repository/
internal/middleware/
internal/model/
migrations/
frontend/
app/
components/
lib/
docker-compose.yml
.env.example

## API conventions

- REST: /api/v1/...
- Always return { data, error, message }
- Use proper HTTP status codes
- JWT: access token (15min) + refresh token (7d)

## Git Workflow

### Branch strategy

- `main` — never commit directly
- Feature branches: `feat/<short-description>`
- Bug branches: `fix/<short-description>`
- Always branch off `main`

### Before making any changes

1. `git pull origin main`
2. `git checkout -b feat/<relevant-name>`
3. Only then start implementing

### Conventional commits (enforce strictly)

Format: `<type>(<scope>): <short description>`

Types:

- `feat` — new feature
- `fix` — bug fix
- `chore` — tooling, deps, config
- `refactor` — no behavior change
- `test` — adding/fixing tests
- `docs` — documentation only
- `ci` — CI/CD changes

Examples:

- `feat(auth): add refresh token rotation`
- `fix(task): resolve assignee validation on member check`
- `chore(docker): add healthcheck to postgres service`

Rules:

- Subject line ≤ 72 chars, lowercase, no period
- Commit one logical change at a time — no "WIP" or "misc fixes"
- Never batch unrelated changes in one commit

### After implementing

1. `git add -p` — stage hunks selectively, review each change
2. `git commit -m "<conventional message>"`
3. `git fetch origin main`
4. `git rebase origin/main` — prefer rebase over merge to keep history clean
5. Resolve any conflicts, then `git rebase --continue`
6. `git push origin feat/<branch-name>`

### Pull Request

- Title must follow conventional commit format
- Description must include:
  - What changed and why
  - How to test it
  - Any design decisions worth noting
- Keep PRs small — one feature or fix per PR
- Never self-merge without review (in real teams)

### What Claude must NEVER do

- `git push --force` on shared branches
- Commit directly to `main`
- Commit secrets, .env files, or build artifacts

## Current phase

Backend: Go + Gin. DB Schema -> migrations -> auth -> project/task APIs -> tests
