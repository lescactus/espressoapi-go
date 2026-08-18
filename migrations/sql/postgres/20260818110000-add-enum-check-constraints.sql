-- +migrate Up
ALTER TABLE beans
    ADD CONSTRAINT chk_beans_roast_level
    CHECK (roast_level BETWEEN 0 AND 4);

ALTER TABLE shots
    ADD CONSTRAINT chk_shots_comparison_with_previous_result
    CHECK (comparison_with_previous_result BETWEEN 0 AND 3);

-- +migrate Down
ALTER TABLE shots
    DROP CONSTRAINT IF EXISTS chk_shots_comparison_with_previous_result;

ALTER TABLE beans
    DROP CONSTRAINT IF EXISTS chk_beans_roast_level;
