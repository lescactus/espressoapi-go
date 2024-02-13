package shot

import (
	"context"
	dbsql "database/sql"
	"fmt"
	"regexp"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
)

var _ repository.ShotRepository = (*Shot)(nil)

type Shot struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Shot {
	return &Shot{
		db: db,
	}
}

func (db *Shot) CreateShot(ctx context.Context, shot *sql.Shot) (int, error) {
	query := `INSERT INTO 
	shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`
	res, err := db.db.ExecContext(ctx, query,
		shot.Sheet.Id, shot.Beans.Id, shot.GrindSetting, shot.QuantityIn, shot.QuantityOut, shot.ShotTime, shot.WaterTemperature, shot.Rating, shot.IsTooBitter, shot.IsTooSour, shot.ComparaisonWithPreviousResult, shot.AdditionalNotes)
	if err != nil {
		// Checking if the error is due to a foreign key constraint
		// which will indicate the sheet or beans does not exists:
		// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1452 {
			table, err := extractTableNameFromError1452(*me)
			if err != nil {
				// Generic error
				return 0, fmt.Errorf("failed to insert record to the database: %w", err)
			}
			switch table {
			case "sheets":
				return 0, errors.ErrSheetDoesNotExist
			case "beans":
				return 0, errors.ErrBeansDoesNotExist
			}
		}
		return 0, fmt.Errorf("failed to insert record to the database: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last inserted id: %w", err)
	}

	return int(id), nil
}

func (db *Shot) GetShotById(ctx context.Context, id int) (*sql.Shot, error) {
	var s sql.Shot

	get := `
SELECT
	shots.id,
	shots.grind_setting,
	shots.quantity_in,
	shots.quantity_out,
	shots.shot_time,
	shots.water_temperature,
	shots.rating,
	shots.is_too_bitter,
	shots.is_too_sour,
	shots.comparaison_with_previous_result,
	shots.additional_notes,
	shots.created_at,
	shots.updated_at,
	sheet.id as "sheet.id",
	sheet.name as "sheet.name",
	beans.id as "beans.id",
	beans.name as "beans.name",
	beans.roast_date as "beans.roast_date",
	beans.roast_level as "beans.roast_level",
	roaster.id AS "beans.roaster.id",
	roaster.name AS "beans.roaster.name",
	roaster.created_at AS "beans.roaster.created_at",
	roaster.updated_at AS "beans.roaster.updated_at"
FROM shots
INNER JOIN
	sheets sheet ON shots.sheet_id = sheet.id
INNER JOIN
	beans beans ON shots.beans_id = beans.id
INNER JOIN
	roasters roaster ON beans.roaster_id = roaster.id
WHERE shots.id = ?`

	if err := db.db.QueryRowxContext(ctx, get, id).StructScan(&s); err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrShotDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for shot id=%d from the database: %w", id, err)
	}

	return &s, nil
}
func (db *Shot) GetAllShots(ctx context.Context) ([]sql.Shot, error) {
	var shots = make([]sql.Shot, 0)
	get := `
SELECT
	shots.id,
	shots.grind_setting,
	shots.quantity_in,
	shots.quantity_out,
	shots.shot_time,
	shots.water_temperature,
	shots.rating,
	shots.is_too_bitter,
	shots.is_too_sour,
	shots.comparaison_with_previous_result,
	shots.additional_notes,
	shots.created_at,
	shots.updated_at,
	sheet.id as "sheet.id",
	sheet.name as "sheet.name",
	beans.id as "beans.id",
	beans.name as "beans.name",
	beans.roast_date as "beans.roast_date",
	beans.roast_level as "beans.roast_level",
	roaster.id AS "beans.roaster.id",
	roaster.name AS "beans.roaster.name",
	roaster.created_at AS "beans.roaster.created_at",
	roaster.updated_at AS "beans.roaster.updated_at"
FROM shots
INNER JOIN
	sheets sheet ON shots.sheet_id = sheet.id
INNER JOIN
	beans beans ON shots.beans_id = beans.id
INNER JOIN
	roasters roaster ON beans.roaster_id = roaster.id`

	if err := db.db.SelectContext(ctx, &shots, get); err != nil {
		return shots, fmt.Errorf("failed to read records for shots: %w", err)
	}

	return shots, nil
}
func (db *Shot) UpdateShotById(ctx context.Context, id int, shot *sql.Shot) (*sql.Shot, error) {
	_, err := db.db.ExecContext(ctx, `UPDATE shots SET
	sheet_id = ?,
	beans_id = ?,
	grind_setting = ?,
	quantity_in = ?,
	quantity_out = ?,
	shot_time = ?,
	water_temperature = ?,
	rating = ?,
	is_too_bitter = ?,
	is_too_sour = ?,
	comparaison_with_previous_result = ?,
	additional_notes = ?
	WHERE id = ?`,
		shot.Sheet.Id,
		shot.Beans.Id,
		shot.GrindSetting,
		shot.QuantityIn,
		shot.QuantityOut,
		shot.ShotTime,
		shot.WaterTemperature,
		shot.Rating,
		shot.IsTooBitter,
		shot.IsTooSour,
		shot.ComparaisonWithPreviousResult,
		shot.AdditionalNotes,
		id)
	if err != nil {
		// Checking if the error is due to a foreign key constraint
		// which will indicate the sheet or beans does not exists:
		// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1452 {
			table, err := extractTableNameFromError1452(*me)
			if err != nil {
				// Generic error
				return nil, fmt.Errorf("failed to insert record to the database: %w", err)
			}
			switch table {
			case "sheets":
				return nil, errors.ErrSheetDoesNotExist
			case "beans":
				return nil, errors.ErrBeansDoesNotExist
			}
		}

		return nil, fmt.Errorf("failed to update record for shots id=%d from the database: %w", id, err)
	}

	return shot, nil
}
func (db *Shot) DeleteShotById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, `DELETE FROM shots WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete record for shots id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return errors.ErrShotDoesNotExist
	}

	return nil
}
func (db *Shot) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}

// extractTableNameFromError1452 extracts the table name from a MySQL error with code 1452.
// It uses a regular expression to find the table name in the error message.
// If a match is found, it returns the table name. Otherwise, it returns an error.
//
// Example error message:
// "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `sheets` (`id`))"
func extractTableNameFromError1452(err mysql.MySQLError) (string, error) {
	if err.Number != 1452 {
		return "", fmt.Errorf("error is not mysql error 1452")
	}

	// Define the regular expression
	// x60 is the backtick character (`)
	re := regexp.MustCompile(`FOREIGN KEY \(\x60(.+?)\x60\) REFERENCES \x60(.+?)\x60 \(\x60id\x60`)

	// Use the regular expression to find the table name in the error message
	matches := re.FindStringSubmatch(err.Error())

	// Check if a match was found
	if len(matches) > 0 {
		// The second element in matches will be the table name
		return matches[2], nil
	} else {
		return "", fmt.Errorf("failed to extract table name from error message")
	}
}
