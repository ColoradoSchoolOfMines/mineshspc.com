# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What This Is

Source code for [mineshspc.com](https://mineshspc.com), the registration and event management website for the Mines High School Programming Competition (MinesHSPC). It handles teacher/team registration, student confirmation, parent form signing, volunteer check-in, and admin tooling.

## Development

**Run with auto-reload** (requires Go and [gow](https://github.com/mitranim/gow)):
```
LOG_CONSOLE=1 gow -e=yaml,go,html,css run ./cmd/mineshspc/
```

**Run normally:**
```
go run ./cmd/mineshspc/
```

**Config:** Copy `config.sample.yaml` to `config.yaml` and fill in values. Set `dev_mode: true` to print emails to stdout instead of sending via SendGrid. Set `jwt_secret_key` to any string for local development.

The server listens on port **8090**.

## Architecture

Single Go binary with no external service dependencies beyond SQLite and SendGrid.

### Package Layout

- **`cmd/mineshspc/`** — Entry point. Parses YAML config (supports multiple `-config` flags), sets up zerolog, opens the SQLite DB, runs migrations, starts the HTTP server.
- **`internal/`** — All HTTP handler logic. `Application` struct (in `application.go`) is the central type, holding the DB, config, SendGrid client, and pre-compiled template renderers.
- **`internal/config/`** — `Configuration` struct parsed from YAML. Key fields: `RegistrationEnabled`, `DevMode`, `Domain`, JWT secret, reCAPTCHA keys.
- **`database/`** — SQLite via `go.mau.fi/util/dbutil`. Schema migrations are numbered SQL files (`v01_*.sql`, `v02_*.sql`, …) auto-registered via `//go:embed *.sql`. Separate files for each entity: `teacher.go`, `student.go`, `teams.go`, `admins.go`, `volunteers.go`.
- **`website/`** — Static assets and HTML templates embedded via `//go:embed`. `base.html` is the layout; all pages extend it. Template data is always wrapped as `{"PageName": ..., "Data": ..., "HostedByHTML": ..., "RegistrationEnabled": ...}`.

### Authentication

Three separate auth flows, all JWT-based (HMAC, stored in `tok` cookie):
1. **Teachers** — email magic-link login → `tok` cookie → session validated in `GetLoggedInTeacher()`
2. **Admins** — same flow, separate issuer checked in `AdminAuthMiddleware`
3. **Volunteers** — same flow, separate issuer

Students and parents authenticate via one-time links (no persistent session).

### Adding a Migration

Add a new `vNN_description.sql` file in `database/`. It will be picked up automatically by the embedded FS registration on next startup.

### Registration Flow

1. Teacher creates account → confirms email → fills in school info → creates teams → adds members
2. Students receive confirmation email → confirm personal info
3. Parents receive email → sign forms
4. Admins can resend emails, export to Kattis/Zoom CSV formats, send QR codes, and manually check in participants
