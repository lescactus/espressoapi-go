-- +migrate Up
ALTER TABLE beans ADD CONSTRAINT uq_beans_identity UNIQUE NULLS NOT DISTINCT (name, roaster_id, roast_date);

-- +migrate Down
ALTER TABLE beans DROP CONSTRAINT uq_beans_identity;