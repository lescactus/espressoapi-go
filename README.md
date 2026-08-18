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
