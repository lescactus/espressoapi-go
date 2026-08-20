# `shot_time`: storage unit, Go type, and wire format

## Status

- **Implemented** (commit `db0b4e8`, branch `feat/web-frontend-htmx`): milliseconds-in-`INT`
  storage with conversion at the repository boundary. See "Background" below.
- **Implemented — "Option A"** (this commit, branch `feat/web-frontend-htmx`):
  human-readable **float seconds on the wire** (REST) and the SQL column renamed to
  `shot_time_ms`. See "[Option A — decision record](#option-a--decision-record)" below
  for what was built and the one deliberate deviation from the original spec.

## Unit map (current, Option A implemented)

| Layer | Representation | Unit |
|---|---|---|
| REST JSON (`shot_time`) | JSON number (float) | **seconds** (`28.5` = 28.5 s) |
| Web form input/display | `<input type=number step=0.1>` / `28.5 s` | **seconds** |
| Go in-memory (`shot.Shot.ShotTime`, `sql.Shot.ShotTime`) | `time.Duration` | **nanoseconds** (native) |
| SQL column (`shots.shot_time_ms`) | `INT` | **milliseconds** |

The only unit-crossing points are: (1) the REST/web parse+render boundary
(seconds ↔ `time.Duration`), and (2) the repository SQL read/write boundary
(`time.Duration` ↔ milliseconds). Nothing in between converts anything.

## Background (already implemented — for context only)

`sql.Shot.ShotTime` ([`internal/models/sql/shot.go`](../internal/models/sql/shot.go)) is
Go's `time.Duration` (an `int64` of nanoseconds). The `shots` shot-time column is a
32-bit SQL `INT` (max `2,147,483,647`), so raw nanoseconds overflow at ~2.15 s. This
was silently unreachable until the web form submitted a real value (28.5 s) and MySQL
rejected the INSERT (`Error 1264 ... Out of range value for column 'shot_time'`).

Fix chosen (commit `db0b4e8`): keep the `INT` column, store **milliseconds**, convert
only at the repository boundary in
[`internal/repository/sql/shared/repository.go`](../internal/repository/sql/shared/repository.go):

- Writes (`CreateShot`, `UpdateShotById`): bind `shot.ShotTime.Milliseconds()`.
- Reads (`GetShotById`, `GetAllShots`, `GetShotsBySheetId`): after sqlx struct-scan,
  correct with `shots[i].ShotTime *= time.Millisecond`.

Milliseconds give ~24.8 days of headroom and represent the UI's 0.1 s granularity
losslessly. Whole seconds (violates the 0.1 s requirement) and microseconds (needless
tight headroom) were rejected; widening to `BIGINT` nanoseconds was rejected as a
larger blast radius for no practical benefit. Test literals were updated accordingly
(`shot_time: 24000000` in venom, `int64(25000)` in sqlmock rows).

Before Option A, the REST JSON contract was the **raw nanosecond integer**
(`"shot_time": 28500000000` = 28.5 s).

---

## Option A — decision record

`shot_time` is now a **JSON number of seconds** on the REST API (`25` = 25 s,
`25.5` = 25.5 s), on every endpoint that carries a shot. Go code keeps
`time.Duration` everywhere internally. SQL keeps `INT` milliseconds, and the column
was renamed `shot_time` → `shot_time_ms` so the stored unit is self-documenting (the
original overflow bug existed precisely because the column's unit was invisible).

This was a **deliberate breaking change** to the REST contract — there were no
external consumers.

### What was built

1. Wire unit: seconds, JSON **number** only. A JSON string (`"25.5"`) is rejected.
2. Validation: value must be finite, `> 0`, and `<= 3600` (max plausible shot time,
   also a guard so a legacy nanosecond-denominated caller fails loudly with a 400
   instead of silently storing centuries). Validation lives entirely in
   `DurationSeconds.UnmarshalJSON` ([`internal/controllers/rest/types.go`](../internal/controllers/rest/types.go)) —
   it only fires when the `shot_time` JSON key is present, matching the pre-existing
   handler behavior of treating an omitted `shot_time` as "not provided" rather than
   inventing a new required-field check (several pre-existing tests, both unit and
   e2e, submit bodies that omit `shot_time` while testing unrelated fields).
3. Precision: parse → **round to the nearest millisecond** → stored as seconds
   rounded to that precision. Marshal from that same rounded value, never from raw
   nanoseconds — this guarantees `25.3` round-trips as exactly `25.3`, avoiding
   binary-float artifacts like `25.299999999`.
4. Storage: unchanged `INT` milliseconds; column renamed to `shot_time_ms`
   (`migrations/sql/{mysql,postgres}/20260820160000-rename-shot-time-to-shot-time-ms.sql`).
5. Service layer: no contract change. `shot.Shot.ShotTime` stays `time.Duration`.
6. NaN/Inf need no explicit JSON-side check — they are not valid JSON numbers and
   `encoding/json` rejects them.

### One deliberate deviation from the original spec

The original instructions suggested `type DurationSeconds time.Duration`. The
implementation instead uses **`type DurationSeconds float64`**
([`internal/controllers/rest/types.go`](../internal/controllers/rest/types.go)).
This is a pure internal-representation choice — it does not affect the wire format,
validation rules, or rounding behavior above — and was made because a `float64`
underlying type lets go-swagger's reflection infer `type: number, format: double`
directly, with zero custom swagger annotations, confirmed correct in the regenerated
`docs/swagger.json`.

### Web layer

Deliberately left unchanged. The web form already speaks seconds and already allows
`shot_time` to be optional (empty string → zero duration, no error) — a constraint
incompatible with the REST rule that a *present* `shot_time` must be `> 0`. Unifying
the two would have changed existing web behavior, so no shared helper was extracted;
REST and web validate independently, as before.

### Verification performed

- Full `go test ./...`, `go vet ./...`, `gofmt -l .`, `go mod tidy -diff` (clean).
- Migration up/down/up round-trip verified live on both MySQL 8 and PostgreSQL 16.
- Full venom e2e suites (shots, sheets, beans, roasters, web, swagger) passed against
  fresh MySQL; the shots suite (the one exercising `shot_time`) also passed against
  fresh PostgreSQL.
- Manual smoke test: created a shot via REST with `"shot_time": 28.5`; response
  echoed `"shot_time": 28.5` exactly; DB row held `28500` in `shot_time_ms`.

---

## Historical note: why not other designs

Recorded from the design discussion (2026-08-20) so the decision is traceable:

- **New field `shot_time_seconds` + deprecation** — rejected: compatibility tax with
  zero consumers, and the embedded-DTO pattern makes field suppression awkward.
- **Duration strings (`"25.5s"`)** — rejected: user wants plain floats; HTML number
  inputs can't produce strings; `Duration.String()` formats `90s` as `1m30s`.
- **Versioned `/rest/v2`** — rejected: route duplication overkill for a unit change
  with no consumers.
- **`BIGINT` nanoseconds storage** — rejected: the wire needs nothing finer than ms;
  a rename to a self-documenting `shot_time_ms` is more valuable than raw-ns fidelity.
- **`DECIMAL(10,3)` seconds storage** — rejected: requires a scan shim in `sql.Shot`,
  threatening the keep-`time.Duration` requirement at the model layer.
