-- +migrate Up
-- SQL in section 'Up' is executed when this migration is applied
CREATE TABLE IF NOT EXISTS `sheets` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) UNIQUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
);
CREATE TABLE IF NOT EXISTS `roasters` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) UNIQUE,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`)
);
CREATE TABLE IF NOT EXISTS `beans` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(255) NOT NULL,
    `roast_date` DATE NULL,
    `roast_level` TINYINT NOT NULL,
    `roaster_id` INT NOT NULL,
    `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated_at` TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (`id`),
    FOREIGN KEY (roaster_id) REFERENCES roasters(id)
);
CREATE TABLE IF NOT EXISTS `shots` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `grind_setting` INT NOT NULL,
    `quantity_in` DOUBLE NOT NULL,
    `quantity_out` DOUBLE NOT NULL,
    `shot_time` INT NOT NULL,
    `water_temperature` DOUBLE NOT NULL DEFAULT(93.0),
    `sheet_id` INT NOT NULL,
    `beans_id` INT NOT NULL,
    PRIMARY KEY (`id`),
    FOREIGN KEY (sheet_id) REFERENCES sheets(id),
    FOREIGN KEY (beans_id) REFERENCES beans(id)
);
CREATE TABLE IF NOT EXISTS `results` (
    `id` INT NOT NULL AUTO_INCREMENT,
    `rating` DOUBLE(2, 1) NOT NULL,
    `is_too_bitter` BOOL NOT NULL,
    `is_too_sour` BOOL NOT NULL,
    `comparaison_with_previous_result` TINYINT NOT NULL,
    `additional_notes` VARCHAR(511),
    `shot_id` INT NOT NULL,
    PRIMARY KEY (`id`),
    FOREIGN KEY (shot_id) REFERENCES shots(id)
);
-- +migrate Down
-- SQL section 'Down' is executed when this migration is rolled back
ALTER TABLE beans DROP FOREIGN KEY beans_ibfk_1;
ALTER TABLE shots DROP FOREIGN KEY shots_ibfk_1;
ALTER TABLE shots DROP FOREIGN KEY shots_ibfk_2;
ALTER TABLE results DROP FOREIGN KEY results_ibfk_1;
DROP TABLE beans;
DROP TABLE results;
DROP TABLE sheets;
DROP TABLE roasters;
DROP TABLE shots;