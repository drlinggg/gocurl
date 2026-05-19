# Changelog

## [0.0.2] — 2026-05-20

### Breaking changes

- Module path renamed: `github.com/banakh/gocurl` → `github.com/drlinggg/gocurl`. Update all imports.
- `Output.Write` signature changed from `Write(*Response)` to `Write(*Request, *Response)`. Pretty-print and verbose flags are now read from the request, not the output constructor.
- `NewOutput` / `NewOutputWriter` no longer accept `pretty, verbose bool` parameters.
- `Engine.New` now requires `Input` and `Output` arguments. `Engine.Colors()` removed.

### Architecture

- Engine now owns the full request lifecycle: `Engine.Run()` reads from `Input`, executes, writes to `Output`.
- `core/io.Input` and `core/io.Output` interfaces are now used by Engine (not just defined). Compile-time assertions added in `cmd/gocurl/io`.
- `main.go` moved from `cmd/gocurl/main.go` to project root. Build: `go build .`

### Dependencies

- `go.mod`: `github.com/BurntSushi/toml` and `golang.org/x/term` correctly marked as direct (were `// indirect`).
- `golang.org/x/sys` remains indirect (transitive dep of `golang.org/x/term`).

### Fixes

- `embed/presents.toml` renamed to `embed/presets.toml` (typo).
- Removed dead code: `core/http/receive.go` (empty file), `FileHistory` stub in `core/storage/history.go`.
- Removed duplicate docs: `docs/docs.md` and `docs/presents.toml`.

## [0.0.1] — 2026-05-17

Initial release.
