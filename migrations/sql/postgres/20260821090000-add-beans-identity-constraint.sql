-- +migrate Up
-- Identity comparison is case-insensitive but accent-sensitive: the key is LOWER(name).
-- NULLS NOT DISTINCT makes two NULL roast dates equal (requires PostgreSQL 15+).
CREATE UNIQUE INDEX uq_beans_identity ON beans (LOWER(name), roaster_id, roast_date) NULLS NOT DISTINCT;

-- +migrate Down
DROP INDEX uq_beans_identity;