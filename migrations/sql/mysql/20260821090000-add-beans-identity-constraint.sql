-- +migrate Up
ALTER TABLE beans ADD UNIQUE INDEX uq_beans_identity (name, roaster_id, (COALESCE(roast_date, CAST('1000-01-01' AS DATE))));

-- +migrate Down
ALTER TABLE beans DROP INDEX uq_beans_identity;