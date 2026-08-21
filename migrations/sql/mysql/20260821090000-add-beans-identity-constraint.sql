-- +migrate Up
-- NULL roast dates share the 1000-01-01 key sentinel; the service rejects real roast dates before 1900.
ALTER TABLE beans ADD UNIQUE INDEX uq_beans_identity (name, roaster_id, (COALESCE(roast_date, CAST('1000-01-01' AS DATE))));

-- +migrate Down
ALTER TABLE beans DROP INDEX uq_beans_identity;