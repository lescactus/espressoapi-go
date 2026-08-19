# `shot_time` storage precision: INT vs BIGINT, nanoseconds vs milliseconds

## Status

Implemented (commit `db0b4e8`, branch `feat/web-frontend-htmx`). Documenting here so
a future change can be made deliberately instead of re-discovering the constraint.

## The problem

`sql.Shot.ShotTime` (see [`internal/models/sql/shot.go`](../internal/models/sql/shot.go))
is typed as Go's `time.Duration`, i.e. an `int64` count of **nanoseconds**. Before this
fix, [`internal/repository/sql/shared/repository.go`](../internal/repository/sql/shared/repository.go)
passed that raw nanosecond value straight through to the SQL driver on
`CreateShot`/`UpdateShotById`, and scanned the column straight back into a
`time.Duration` on every read (`GetShotById`, `GetAllShots`, `GetShotsBySheetId`).

The `shots.shot_time` column, however, is a 32-bit SQL `INT`
(see `migrations/sql/mysql/20240104154454-.sql` and the Postgres equivalent), whose
max value is `2,147,483,647`. In nanoseconds that's about **2.15 seconds**. Any
shot duration longer than ~2.15s overflows the column outright.

This bug pre-dates the web UI work; it was silently unreachable for a long time
because no existing REST caller (fixtures, e2e tests) ever submitted a realistic
`shot_time` value — all the venom e2e literals used tiny placeholder values like
`shot_time: 24` (i.e. 24 **nanoseconds**), which round-trips fine but is not
representative of an actual espresso shot (typically 20-40 real seconds). It was
only caught when the new web form (`/shots/add`) submitted a real value (28.5s) and
MySQL rejected the INSERT with:

```
Error 1264 (22003): Out of range value for column 'shot_time' at row 1
```

## Options considered

1. **Widen the column to `BIGINT` (schema migration), keep storing raw nanoseconds.**
   - Pro: no precision loss whatsoever, no unit-conversion code at the repository
     boundary.
   - Con: requires a new migration file for both MySQL and Postgres, plus a
     production migration rollout. Higher blast radius for what is otherwise a
     boundary-only bug.
2. **Keep the `INT` column, but change the *stored unit* from nanoseconds to a
   coarser unit, converting only at the SQL read/write boundary in the
   repository layer.** Go's `time.Duration` type and the REST/JSON contract stay
   exactly as they are; only the bytes that hit the database change meaning.
   - Sub-choice: which coarser unit?
     - **Whole seconds**: max ~2.147 billion seconds (~68 years) of headroom, but
       truncates any sub-second precision. This conflicts with `SPEC.md`'s
       explicit requirement that the web UI accept shot time at **0.1-second**
       granularity — storing whole seconds would silently round `28.5s` to `28s`
       or `29s` server-side.
     - **Milliseconds** *(chosen)*: max ~2.147 billion ms (~24.8 days) of
       headroom — comfortably more than any real shot duration — while
       preserving the UI's 0.1s (100ms) granularity **losslessly** (100ms is an
       exact multiple of 1ms, so nothing is rounded away).
     - **Microseconds**: max ~2.147 billion µs (~35.8 minutes) of headroom. This
       is still more than any real shot duration (typically well under a minute),
       but the safety margin is far smaller than milliseconds (35.8 minutes vs.
       24.8 days) for no practical precision benefit, since the web UI only needs
       0.1s (100ms) granularity. Not chosen for that reason.
3. **Do nothing / defer, and document the limitation** (e.g. reject or clamp
   shot times above ~2.1s at the web/service layer). Rejected: this would make
   the shots feature nearly useless, since virtually every real shot exceeds 2.1s.

## Decision: option 2, milliseconds

No schema migration. The persistence boundary in
[`internal/repository/sql/shared/repository.go`](../internal/repository/sql/shared/repository.go)
now does the conversion **only** at the four call sites that touch the
`shot_time` column:

- `CreateShot` / `UpdateShotById` (write): `shot.ShotTime.Milliseconds()` — converts
  the in-memory `time.Duration` (nanoseconds) to an `int64` count of milliseconds
  before binding it as a query argument.
- `GetShotById` / `GetAllShots` / `GetShotsBySheetId` (read): after the `sqlx`
  struct-scan populates `ShotTime` with the raw column value (which sqlx treats as
  a `time.Duration`-shaped `int64`, i.e. the number is currently *mislabelled* as
  nanoseconds), each result's `ShotTime` field is corrected with
  `shot.ShotTime *= time.Millisecond` (equivalently, for slices,
  `shots[i].ShotTime *= time.Millisecond`), which multiplies the raw millisecond
  count by `1,000,000` to yield the correct nanosecond-denominated
  `time.Duration`.

Everything **outside** the repository package is unaffected:

- `sql.Shot.ShotTime` is still `time.Duration` (nanoseconds) everywhere in Go code
  (services, REST controllers, web controllers, templates).
- The REST JSON contract for `shot_time` is unchanged — it is still the raw
  nanosecond integer, exactly as before this fix. A client sending
  `"shot_time": 28500000000` (28.5s) gets back `28500000000` again, byte-for-byte,
  because the round trip through milliseconds is lossless for any value that is a
  whole multiple of 1ms.
- Only values with **sub-millisecond** precision (fewer than 6 significant
  nanosecond digits, e.g. `"shot_time": 24` = 24ns) are truncated to the nearest
  millisecond on write — in practice `0`. This is a real, disclosed behavior
  change for that edge case; see "Known consequence" below.

## Known consequence: e2e/test literals updated

The pre-existing venom e2e suite ([`e2e/venom.e2e.shots.yaml`](../e2e/venom.e2e.shots.yaml))
and the MySQL shot repository sqlmock tests
([`internal/repository/sql/mysql/shot/repository_test.go`](../internal/repository/sql/mysql/shot/repository_test.go))
used placeholder values like `shot_time: 24` (24 nanoseconds) purely as "any
non-zero integer" filler — not as a realistic value. Under millisecond-precision
storage those tiny literals would round-trip to `0`, so they were updated to
`24000000` (24,000,000 ns = 24 ms) in the e2e YAML, and to `int64(25000)` (25,000 ms
= the millisecond-column value corresponding to the previously-asserted
`25 * time.Second` Go struct field) in the repository test mocks. No other
REST/domain behavior was changed.

## If a future change is needed

If nanosecond (or even microsecond) precision genuinely becomes a requirement:

1. Widen `shots.shot_time` to `BIGINT` via a new migration in both
   `migrations/sql/mysql/` and `migrations/sql/postgres/`.
2. Remove the `.Milliseconds()` conversion on write and the `* time.Millisecond`
   conversion on read in `internal/repository/sql/shared/repository.go`, going
   back to storing/reading the raw `time.Duration` value directly.
3. Revert the `shot_time` literals in `e2e/venom.e2e.shots.yaml` and the mocked
   row values in the MySQL/Postgres shot repository tests back to their original
   raw-nanosecond form if desired (not required — the current literals are still
   valid nanosecond values, just chosen to also survive millisecond rounding).
4. Re-run `go test ./... `, the full venom e2e suite (per-suite against a fresh
   database — the suites are not designed to be chained on one database), and a
   live Docker Compose smoke test before merging.
