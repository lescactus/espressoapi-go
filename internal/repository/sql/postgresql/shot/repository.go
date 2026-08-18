package shot

import (
	"context"
	dbsql "database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/postgreserrors"
)

var _ repository.ShotRepository = (*Shot)(nil)

type Shot struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Shot {
	return &Shot{db: db}
}

func (db *Shot) CreateShot(ctx context.Context, shot *sql.Shot) (int, error) {
	var id int
	err := db.db.QueryRowxContext(ctx, "INSERT INTO shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparison_with_previous_result, additional_notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id",
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
		shot.ComparisonWithPreviousResult,
		shot.AdditionalNotes).Scan(&id)
	if err != nil {
		return 0, postgreserrors.ParsePostgresError(err, &postgreserrors.EntityShot, fmt.Errorf("failed to insert record to the database: %w", err))
	}

	return id, nil
}

func (db *Shot) GetShotById(ctx context.Context, id int) (*sql.Shot, error) {
	var shot sql.Shot
	query := `
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
	shots.comparison_with_previous_result,
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
WHERE shots.id = $1`

	if err := db.db.QueryRowxContext(ctx, query, id).StructScan(&shot); err != nil {
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrShotDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for shot id=%d from the database: %w", id, err)
	}

	return &shot, nil
}

func (db *Shot) GetAllShots(ctx context.Context) ([]sql.Shot, error) {
	shots := make([]sql.Shot, 0)
	query := `
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
	shots.comparison_with_previous_result,
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

	if err := db.db.SelectContext(ctx, &shots, query); err != nil {
		return shots, fmt.Errorf("failed to read records for shots: %w", err)
	}

	return shots, nil
}

func (db *Shot) UpdateShotById(ctx context.Context, id int, shot *sql.Shot) (*sql.Shot, error) {
	_, err := db.db.ExecContext(ctx, `UPDATE shots SET
	sheet_id = $1,
	beans_id = $2,
	grind_setting = $3,
	quantity_in = $4,
	quantity_out = $5,
	shot_time = $6,
	water_temperature = $7,
	rating = $8,
	is_too_bitter = $9,
	is_too_sour = $10,
	comparison_with_previous_result = $11,
	additional_notes = $12
	WHERE id = $13`,
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
		shot.ComparisonWithPreviousResult,
		shot.AdditionalNotes,
		id)
	if err != nil {
		return nil, postgreserrors.ParsePostgresError(err, &postgreserrors.EntityShot, fmt.Errorf("failed to update record in the database: %w", err))
	}

	return shot, nil
}

func (db *Shot) DeleteShotById(ctx context.Context, id int) error {
	result, err := db.db.ExecContext(ctx, "DELETE FROM shots WHERE id = $1", id)
	if err != nil {
		return postgreserrors.ParsePostgresError(err, nil, fmt.Errorf("failed to delete record for shots id=%d: %w", id, err))
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected != 1 {
		return errors.ErrShotDoesNotExist
	}

	return nil
}

func (db *Shot) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
