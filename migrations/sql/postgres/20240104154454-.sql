-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS "sheets" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(255) UNIQUE,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE
);
CREATE FUNCTION update_updated_at_sheets() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = now(); RETURN NEW; END; $$ language 'plpgsql';
-- -- Trigger the above function
CREATE TRIGGER update_updated_at_sheets BEFORE
UPDATE ON sheets FOR EACH ROW EXECUTE PROCEDURE update_updated_at_sheets();
CREATE TABLE IF NOT EXISTS "beans" (
    "id" SERIAL PRIMARY KEY,
    "roaster_name" VARCHAR(255) NOT NULL,
    "beans_name" VARCHAR(255) NOT NULL,
    "roast_date" DATE NULL,
    "roast_level" SMALLINT NOT NULL
);
CREATE TABLE IF NOT EXISTS "shots" (
    "id" SERIAL PRIMARY KEY,
    "grind_setting" INT NOT NULL,
    "quantity_in" DECIMAL NOT NULL,
    "quantity_out" DECIMAL NOT NULL,
    "shot_time" INT NOT NULL,
    "water_temperature" INT NOT NULL DEFAULT(93),
    "sheet_id" INT NOT NULL,
    "beans_id" INT NOT NULL,
    FOREIGN KEY (sheet_id) REFERENCES sheets(id),
    FOREIGN KEY (beans_id) REFERENCES beans(id)
);
CREATE TABLE IF NOT EXISTS "results" (
    "id" SERIAL PRIMARY KEY,
    "rating" DECIMAL(2, 1) NOT NULL,
    "is_too_bitter" BOOLEAN NOT NULL,
    "is_too_sour" BOOLEAN NOT NULL,
    "comparaison_with_previous_result" SMALLINT NOT NULL,
    "additional_notes" VARCHAR(511),
    "shot_id" INT NOT NULL,
    FOREIGN KEY (shot_id) REFERENCES shots(id)
);
-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE IF EXISTS shots DROP CONSTRAINT shots_sheet_id_fkey;
ALTER TABLE IF EXISTS shots DROP CONSTRAINT shots_beans_id_fkey;
ALTER TABLE IF EXISTS results DROP CONSTRAINT results_shot_id_fkey;
DROP TRIGGER IF EXISTS update_updated_at_sheets ON sheets;
DROP FUNCTION IF EXISTS update_updated_at_sheets;
DROP TABLE IF EXISTS beans;
DROP TABLE IF EXISTS results;
DROP TABLE IF EXISTS sheets;
DROP TABLE IF EXISTS shots;