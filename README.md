# espressoapi-go

Simple API to keep track of notes during espresso making written in Go.

## Datasource configuration

The API supports MySQL and PostgreSQL. MySQL is the default when no database
configuration is supplied.

| Database | `DATABASE_TYPE` | `DATABASE_DATASOURCE_NAME` |
| --- | --- | --- |
| MySQL | `mysql` | `root:root@tcp(127.0.0.1:3306)/espresso-api?parseTime=true` |
| PostgreSQL | `postgres` | `postgres://root:root@127.0.0.1:5432/espresso-api?sslmode=disable` |

Configuration is read from `config.json`, `config.yaml`, or `config.env` when
one of those files exists in the working directory. Otherwise, it is read from
environment variables. A configuration file is authoritative; environment
variables do not override values from that file.

The database type selects the matching `database/sql` driver, repository set,
and embedded migration directory. The HTTP API contract is the same for both
databases.

## Migrations

Run migrations with the same datasource environment used by the API:

```bash
go run main.go migrate up
go run main.go migrate down
go run main.go migrate redo
go run main.go migrate skip
```

`migrate up`, `down`, `redo`, and `skip` automatically select
`migrations/sql/mysql` or `migrations/sql/postgres` from `DATABASE_TYPE`.

## Local end-to-end testing

Start one database profile at a time. Each profile starts the matching API
container on `http://127.0.0.1:8080`.

```bash
docker compose --profile mysql up -d --build --wait
go run main.go migrate up
venom run -vv ./e2e/venom.e2e.beans.yaml
docker compose --profile mysql down -v
```

For PostgreSQL, pass the PostgreSQL datasource to migration and integration
commands running on the host:

```bash
docker compose --profile postgres up -d --build --wait
DATABASE_TYPE=postgres \
DATABASE_DATASOURCE_NAME='postgres://root:root@127.0.0.1:5432/espresso-api?sslmode=disable' \
go run main.go migrate up
DATABASE_TYPE=postgres \
E2E_DATABASE_DSN='postgres://root:root@127.0.0.1:5432/espresso-api?sslmode=disable' \
go test -tags=integration ./e2e
venom run -vv ./e2e/venom.e2e.beans.yaml
docker compose --profile postgres down -v
```

The GitHub Actions e2e workflow runs every Venom suite against both profiles.

## Kubernetes

The deployment manifest defaults to MySQL through inline environment values.
To use PostgreSQL, set `DATABASE_TYPE=postgres` and a PostgreSQL datasource DSN
in both the `migrate-up` init container and the API container. Both containers
must target the same database so that migrations run against the API's store.

## Web UI

Alongside the JSON REST API under `/rest/v1/...`, the same binary serves a
server-rendered web UI (templ + htmx + Pico CSS) on the same port:

| Route | Purpose |
| --- | --- |
| `/` | Home page |
| `/sheets`, `/sheets/add`, `/sheets/get/:id`, `/sheets/update/:id`, `/sheets/delete/:id` | Sheets list, add/edit (inline row), detail page (including its scoped shots section) |
| `/roasters`, `/roasters/add`, `/roasters/get/:id`, `/roasters/update/:id`, `/roasters/delete/:id` | Roasters list, add/edit (inline row) |
| `/beans`, `/beans/add`, `/beans/get/:id`, `/beans/update/:id`, `/beans/delete/:id` | Beans list, add/edit (dialog) |
| `/shots`, `/shots/add`, `/shots/get/:id`, `/shots/update/:id`, `/shots/delete/:id` | Shots list, add/edit (dialog); `/shots/add?sheet_id=N` locks the sheet, used from the sheet detail page |

**Direct navigation vs. htmx.** `GET` routes render either a full page (direct
browser navigation/refresh/deep link) or an htmx fragment, based on the
`HX-Request` header. All mutations (`POST`/`PUT`/`DELETE`) require an
`HX-Request: true` header; there is no progressive-enhancement fallback for
direct form posts. Successful mutations use an `HX-Trigger: dialog-close`
response header (dialog-based resources) or out-of-band (`hx-swap-oob`) row
and alert fragments to update the page without a full reload.

**Runtime frontend assets are CDN-hosted, not bundled.** The layout pins exact
versions with Subresource Integrity hashes:

- htmx `2.0.10` (`https://cdn.jsdelivr.net/npm/htmx.org@2.0.10/dist/htmx.min.js`)
- Pico CSS `2.1.1` (`https://cdn.jsdelivr.net/npm/@picocss/pico@2.1.1/css/pico.min.css`)

See `views/templates/shared/layout.templ` for the exact `<script>`/`<link>`
tags and `integrity` attributes. The scratch Docker image remains valid since
no frontend build step or bundler is involved.

**templ generation.** Templates live in `views/templates/**/*.templ`;
`*_templ.go` is generated and **committed** — it is a normal build input, not
a build-time artifact. Regenerate it after editing any `.templ` file with the
exact pinned CLI version (must match the `github.com/a-h/templ` version in
`go.mod`):

```bash
go install github.com/a-h/templ/cmd/templ@v0.3.1020
templ generate
```

CI (`templ-drift` job in `.github/workflows/go.yaml`) regenerates and runs
`git diff --exit-code`, failing the build if the committed `_templ.go` files
are out of date.

**Live reload.** `.air.toml` watches `.templ`/`.go`/`docs/swagger.json`, runs
`templ generate` before each rebuild, and excludes `_templ.go`/`_test.go` from
triggering additional rebuilds. Run `air` from the repository root for local
development.

**Swagger/OpenAPI regeneration.** `docs/swagger.json` documents `/rest/v1/...`
only (the web UI routes are not part of the JSON API contract and are not
documented there). It was generated with
[go-swagger](https://github.com/go-swagger/go-swagger) `v0.36.4`:

```bash
go install github.com/go-swagger/go-swagger/cmd/swagger@v0.36.4
swagger generate spec -o docs/swagger.json --scan-models
```

The generator does not preserve hand-curated enum descriptions/examples
already in `docs/swagger.json`; review the diff after regenerating and
reapply anything the generator dropped or overwrote before committing.

**No authentication, CSRF protection, or same-origin checks.** This is a
deliberate, accepted scope boundary (see `SPEC.md`), not an oversight: any
client that can reach the web UI's mutation routes with an `HX-Request: true`
header can create/update/delete data, including from a different origin. Do
not expose this service to untrusted networks without adding authentication
and CSRF protection first.

