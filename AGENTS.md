# Repository Guidelines

## Project Structure & Module Organization

This is the Mines High School Programming Competition registration site. `cmd/mineshspc/` loads YAML configuration, opens SQLite, runs migrations, and starts the Go server. HTTP handlers live in `internal/`; `Application` holds shared dependencies. Configuration types are in `internal/config/`. Database code and embedded SQL migrations live in `database/`; add migrations as `vNN_description.sql`. Embedded templates and assets are in `website/templates/` and `website/static/`. Tests sit beside Go code.

## Build, Test, and Development Commands

- `cp config.sample.yaml config.yaml`: create local configuration. Set `dev_mode: true` to print emails and use any local `jwt_secret_key` value.
- `go run ./cmd/mineshspc/`: start the server on port 8090.
- `LOG_CONSOLE=1 gow -e=yaml,go,html,css run ./cmd/mineshspc/`: run with auto-reload; requires `gow`.
- `go test ./...`: run all Go tests.
- `go build ./cmd/mineshspc/`: compile the server.
- `pre-commit run --all-files`: run formatting, static analysis, vet, and file checks. CI also runs `nix flake check` and `nix build`.

## Coding Style & Naming Conventions

Follow `.editorconfig`: tabs in Go, two spaces in HTML and CSS, LF line endings, and a final newline. Format imports with `go tool goimports -local github.com/ColoradoSchoolOfMines/mineshspc.com -w .`; hooks run `staticcheck` and `go vet`. Name files by feature or entity and number SQL migrations sequentially.

## Application Flows

Teachers use email magic links and a JWT in the `tok` cookie; admins and volunteers have separate issuers and auth checks. Students and parents use one-time links. Registration runs from teacher account and team creation to student confirmation and parent forms. Admin tools handle resends, exports, QR codes, and check-in. Extend `base.html` and preserve template data keys `PageName`, `Data`, `HostedByHTML`, and `RegistrationEnabled`.

## Testing Guidelines

Use Go's `testing` package, `*_test.go` files, and `TestXxx` functions. Add focused tests beside changed handlers or database code. Run `go test ./...` before a pull request. No coverage threshold is documented.

## Commit & Pull Request Guidelines

Use [Scoped Commits](https://scopedcommits.com/): `<scope>: <short description>`, with optional body and trailers. Name the changed area, as in `admin: add manual uncheck-in`, `ci: ...`, or `deps/go: ...`. Keep commits focused. In pull requests, describe behavior, configuration or migration steps, and test results; link relevant issues and include screenshots for visible changes.

## Security & Configuration

Keep `config.yaml`, JWT secrets, and service credentials out of commits. Start from `config.sample.yaml` and use development-only values locally. Review authentication and email behavior when changing registration flows.
