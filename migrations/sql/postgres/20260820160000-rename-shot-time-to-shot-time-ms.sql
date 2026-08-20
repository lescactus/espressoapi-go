-- +migrate Up
ALTER TABLE shots RENAME COLUMN shot_time TO shot_time_ms;

-- +migrate Down
ALTER TABLE shots RENAME COLUMN shot_time_ms TO shot_time;
