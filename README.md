# gocurl v0.0.4

A curl replacement focused on REST APIs with named presets and a minimal syntax.

> **Platform support:** macOS + zsh only.

---

## Install

```zsh
make install
source ~/.zshrc

# or

go install github.com/drlinggg/gocurl@latest
```

`make install` runs tests, copies the binary to `$GOPATH/bin`, and adds `$GOPATH/bin` to `~/.zshrc` if it isn't there already.

---

## Development

| Command                      | Description                                  |
|------------------------------|----------------------------------------------|
| `make build`                 | Run tests, then compile `./gocurl`           |
| `make test`                  | Run all tests                                |
| `make run ARGS="<args>"`     | Run without installing (`go run .`)          |
| `make vet`                   | Run `go vet ./...`                           |
| `make fmt`                   | Run `go fmt ./...`                           |
| `make clean`                 | Remove the compiled `./gocurl` binary        |
| `make install`               | Build, install to `$GOPATH/bin`, patch PATH  |
| `make uninstall`             | Remove binary from `$GOPATH/bin`, undo PATH  |

```zsh
make run ARGS="POST httpbin.org/post name=alex"
make run ARGS="--pretty GET httpbin.org/json"
```

---

## Syntax

```
gocurl [METHOD] <url|preset[/path]> [fields...] [flags...]
gocurl set <preset> <field=value> [field=value ...]
gocurl --help
```

---

## URL

Scheme defaults to `https` if not specified. Three forms are accepted:

```zsh
gocurl example.com/api/users          # full host
gocurl dev /users                     # preset name + path (expands to base + /users)
gocurl dev                            # preset name alone  (expands to base)
```

The second and third forms require the preset to have a `base` set (see **Presets**).
If the URL starts with a preset's `base`, that preset is applied automatically even with a full URL.

---

## Method

Optional. Auto-detected when omitted:

- body fields present → `POST`
- no body fields → `GET`

Explicit values: `GET POST PUT PATCH DELETE HEAD OPTIONS`

---

## Fields

| Syntax          | Meaning                    | Example                    |
|-----------------|----------------------------|----------------------------|
| `key=value`     | JSON body field (string)   | `name=alex`                |
| `key:=value`    | JSON body field (raw JSON) | `age:=30`, `ids:=[1,2,3]`  |
| `@Header=Value` | Request header             | `@Authorization=Bearer x`  |
| `?key=value`    | Query parameter            | `?page=2`                  |

Body is sent as `application/json`. Raw (`:=`) values are embedded as-is in the JSON object.

---

## Flags

| Flag              | Description                                   |
|-------------------|-----------------------------------------------|
| `--pretty`        | Pretty-print JSON response with indentation   |
| `-v`              | Show response headers                         |
| `--http1`         | Force HTTP/1.1                                |
| `--http2`         | Force HTTP/2 (negotiated via TLS ALPN)        |
| `--http3`         | HTTP/3 (not yet implemented)                  |
| `--scheme http`   | Override URL scheme (`http` or `https`)       |
| `-t N`            | Timeout in seconds (overrides preset/default) |
| `--help`          | Print this help and exit                      |

---

## Output

Output order:

```
RESPONSE:
  <body>
STATUS: 200   TIME: 123ms
HEADERS:                    ← only with -v
  Key: Value
```

- **RESPONSE** — raw or pretty-printed JSON body
- **STATUS** — colored: green 2xx, yellow 4xx, red 5xx
- **TIME** — total request round-trip
- **HEADERS** — shown only with `-v`

Colors are applied via 24-bit ANSI codes and are disabled automatically when stdout is not a TTY (e.g. when piping).

---

## Presets

Presets live in `~/.config/gocurl/presets.toml` and are written with the `set` command.

### set command

```zsh
gocurl set <preset> <field> [field ...]
```

| Field syntax              | Effect                             |
|---------------------------|------------------------------------|
| `base=https://...`        | Base URL — enables short path form |
| `@Key=Value`              | Default request header             |
| `?key=value`              | Default query parameter            |
| `timeout=30`              | Request timeout in seconds         |
| `scheme=http`             | Default scheme (`http` or `https`) |
| `http=2`                  | Default HTTP version (1, 2, 3)     |
| `color.<field>=RRGGBB`    | Output color (hex, no `#`)         |

Color fields: `status_2xx`, `status_4xx`, `status_5xx`, `headers`, `body`, `elapsed`.

### Example

```zsh
gocurl set dev base=https://api.dev.example.com
gocurl set dev @Authorization=Bearer $TOKEN
gocurl set dev timeout=30

gocurl dev /users          # → GET https://api.dev.example.com/users
gocurl dev /users ?page=2  # → adds ?page=2
gocurl POST dev /users name=alex
```

### Preset resolution order (low → high priority)

```
[default] preset → matched preset → CLI flags
```

### TOML format

```toml
[default]
timeout      = 10
http_version = 1
scheme       = "https"

[default.colors]
status_2xx = "00c853"
status_4xx = "ffab00"
status_5xx = "ff1744"
headers    = "888888"
body       = "61dafb"
elapsed    = "555555"

[dev]
base = "https://api.dev.example.com"

[dev.headers]
Authorization = "Bearer $TOKEN"   # env vars expanded at request time
```

---

## .env

If a `.env` file exists in the working directory it is loaded automatically before any request. Variables already set in the shell are not overridden.

```dotenv
GITHUB_TOKEN=ghp_...
API_KEY=secret
```

Use them in preset headers:

```zsh
gocurl set github @Authorization="Bearer $GITHUB_TOKEN"
```

---

## History

Every request is appended to `~/.config/gocurl/history.jsonl`:

```jsonl
{"timestamp":"2026-05-17T03:27:39Z","method":"GET","url":"example.com","status":200,"elapsed":"12ms"}
```

Override the path with `GOCURL_DATA` env var. Override the config path with `GOCURL_CONFIG`.

---

## Examples

```zsh
# Simple GET
gocurl httpbin.org/get

# Pretty-print response
gocurl --pretty httpbin.org/json

# POST with body
gocurl httpbin.org/post name=alex age:=30

# With headers and query
gocurl @Authorization="Bearer token" ?page=2 api.example.com/users

# Show response headers
gocurl -v httpbin.org/get

# Configure a preset and use it
gocurl set dev base=https://httpbin.org
gocurl set dev @X-Token=secret
gocurl dev /get
gocurl --pretty dev /json

# Set default timeout
gocurl set default timeout=15

# Customize output colors (hex, no #)
gocurl set default color.body=eeeeee
gocurl set default color.status_2xx=00c853
gocurl set default color.elapsed=555555
```
