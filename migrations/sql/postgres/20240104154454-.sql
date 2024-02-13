-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
--
-- Auto update the field updated_at
CREATE FUNCTION update_updated_at() RETURNS TRIGGER AS $$ BEGIN NEW.updated_at = now(); RETURN NEW;END;$$ language 'plpgsql';
CREATE TABLE IF NOT EXISTS "sheets" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(255) UNIQUE,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE -- updated by trigger
);
-- -- Trigger the "update_updated_at" function
CREATE TRIGGER update_updated_at_sheets BEFORE
UPDATE ON sheets FOR EACH ROW EXECUTE PROCEDURE update_updated_at();
CREATE TABLE IF NOT EXISTS "roasters" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(255) UNIQUE,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE -- updated by trigger
);
-- Trigger the "update_updated_at" function function
CREATE TRIGGER update_updated_at_roasters BEFORE
UPDATE ON roasters FOR EACH ROW EXECUTE PROCEDURE update_updated_at();
CREATE TABLE IF NOT EXISTS "beans" (
    "id" SERIAL PRIMARY KEY,
    "name" VARCHAR(255) NOT NULL,
    "roast_date" DATE NULL,
    "roast_level" SMALLINT NOT NULL,
    "roaster_id" INT NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE, -- updated by trigger
    FOREIGN KEY (roaster_id) REFERENCES roasters(id)
);
-- -- Trigger the "update_updated_at" function
CREATE TRIGGER update_updated_at_beans BEFORE
UPDATE ON beans FOR EACH ROW EXECUTE PROCEDURE update_updated_at();
CREATE TABLE IF NOT EXISTS "shots" (
    "id" SERIAL PRIMARY KEY,
    "grind_setting" INT NOT NULL,
    "quantity_in" DECIMAL NOT NULL,
    "quantity_out" DECIMAL NOT NULL,
    "shot_time" INT NOT NULL,
    "water_temperature" DECIMAL NOT NULL DEFAULT(93.0),
    "rating" DECIMAL NOT NULL CHECK ("rating" >= 0.0 AND "rating" <= 10.0),
    "is_too_bitter" BOOLEAN NOT NULL,
    "is_too_sour" BOOLEAN NOT NULL,
    "comparaison_with_previous_result" SMALLINT NOT NULL,
    "additional_notes" VARCHAR(511),
    "sheet_id" INT NOT NULL,
    "beans_id" INT NOT NULL,
    "created_at" TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP NOT NULL,
    "updated_at" TIMESTAMP WITH TIME ZONE, -- updated by trigger
    FOREIGN KEY (sheet_id) REFERENCES sheets(id),
    FOREIGN KEY (beans_id) REFERENCES beans(id)
);
-- -- Trigger the "update_updated_at" function
CREATE TRIGGER update_updated_at_shots BEFORE
UPDATE ON shots FOR EACH ROW EXECUTE PROCEDURE update_updated_at();
-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE IF EXISTS beans DROP CONSTRAINT beans_roaster_id_fkey;
ALTER TABLE IF EXISTS shots DROP CONSTRAINT shots_sheet_id_fkey;
ALTER TABLE IF EXISTS shots DROP CONSTRAINT shots_beans_id_fkey;
DROP TRIGGER IF EXISTS update_updated_at_sheets ON sheets;
DROP TRIGGER IF EXISTS update_updated_at_roasters ON roasters;
DROP TRIGGER IF EXISTS update_updated_at_beans ON beans;
DROP TRIGGER IF EXISTS update_updated_at_shots ON shots;
DROP FUNCTION IF EXISTS update_updated_at;
DROP TABLE IF EXISTS beans;
DROP TABLE IF EXISTS roasters;
DROP TABLE IF EXISTS sheets;
DROP TABLE IF EXISTS shots;