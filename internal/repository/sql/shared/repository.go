package shared

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	sqlerrors "github.com/lescactus/espressoapi-go/internal/repository/sql/errors"
)

type Dialect struct {
	Rebind     func(string) string
	ParseError func(error, *sqlerrors.Entity, error) error
	InsertID   func(context.Context, *sqlx.DB, string, *sqlerrors.Entity, ...any) (int, error)
}

var (
	entityBeans   = sqlerrors.EntityBeans
	entityRoaster = sqlerrors.EntityRoaster
	entitySheet   = sqlerrors.EntitySheet
	entityShot    = sqlerrors.EntityShot
)

type Bean struct {
	db      *sqlx.DB
	dialect Dialect
}

func NewBean(db *sqlx.DB, dialect Dialect) *Bean { return &Bean{db: db, dialect: dialect} }

func (db *Bean) CreateBeans(ctx context.Context, beans *sql.Beans) (int, error) {
	query := db.dialect.Rebind("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)")
	return db.dialect.InsertID(ctx, db.db, query, &entityBeans, beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel)
}

func (db *Bean) GetBeansById(ctx context.Context, id int) (*sql.Beans, error) {
	var beans sql.Beans
	query := db.dialect.Rebind(`
SELECT
	beans.id,
	beans.name,
	beans.roast_date,
	beans.roast_level,
	beans.created_at,
	beans.updated_at,
	roaster.id AS "roaster.id",
	roaster.name AS "roaster.name",
	roaster.created_at AS "roaster.created_at",
	roaster.updated_at AS "roaster.updated_at"
FROM beans
	INNER JOIN roasters roaster
		ON beans.roaster_id = roaster.id
WHERE
	beans.id = ?`)
	if err := db.db.QueryRowxContext(ctx, query, id).StructScan(&beans); err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, domainerrors.ErrBeansDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for beans id=%d from the database: %w", id, err)
	}
	return &beans, nil
}

func (db *Bean) GetAllBeans(ctx context.Context) ([]sql.Beans, error) {
	beans := make([]sql.Beans, 0)
	query := db.dialect.Rebind(`
	SELECT
		beans.id,
		beans.name,
		beans.roast_date,
		beans.roast_level,
		beans.created_at,
		beans.updated_at,
		roaster.id AS "roaster.id",
		roaster.name AS "roaster.name",
		roaster.created_at AS "roaster.created_at",
		roaster.updated_at AS "roaster.updated_at"
	FROM beans
		INNER JOIN roasters roaster
			ON beans.roaster_id = roaster.id`)
	if err := db.db.SelectContext(ctx, &beans, query); err != nil {
		return beans, fmt.Errorf("failed to read records for beans: %w", err)
	}
	return beans, nil
}

func (db *Bean) UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error) {
	query := db.dialect.Rebind(`UPDATE beans SET name = ?, roaster_id = ?, roast_date = ?, roast_level = ? WHERE id = ?`)
	if _, err := db.db.ExecContext(ctx, query, beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel, id); err != nil {
		return nil, db.dialect.ParseError(err, &entityBeans, fmt.Errorf("failed to update record for beans id=%d: %w", id, err))
	}
	return beans, nil
}

func (db *Bean) DeleteBeansById(ctx context.Context, id int) error {
	query := db.dialect.Rebind(`DELETE FROM beans WHERE id = ?`)
	res, err := db.db.ExecContext(ctx, query, id)
	if err != nil {
		return db.dialect.ParseError(err, nil, fmt.Errorf("failed to delete record for beans id=%d: %w", id, err))
	}
	if row, _ := res.RowsAffected(); row != 1 {
		return domainerrors.ErrBeansDoesNotExist
	}
	return nil
}

func (db *Bean) Ping(ctx context.Context) error { return db.db.PingContext(ctx) }

type Roaster struct {
	db      *sqlx.DB
	dialect Dialect
}

func NewRoaster(db *sqlx.DB, dialect Dialect) *Roaster { return &Roaster{db: db, dialect: dialect} }

func (db *Roaster) CreateRoaster(ctx context.Context, roaster *sql.Roaster) error {
	query := db.dialect.Rebind(`INSERT INTO roasters (name) VALUES (?)`)
	_, err := db.db.ExecContext(ctx, query, roaster.Name)
	if err != nil {
		return db.dialect.ParseError(err, &entityRoaster, fmt.Errorf("failed to insert record to the database: %w", err))
	}
	return nil
}

func (db *Roaster) GetRoasterById(ctx context.Context, id int) (*sql.Roaster, error) {
	var roaster sql.Roaster
	query := db.dialect.Rebind("SELECT id, name, created_at, updated_at FROM roasters WHERE id = ?")
	if err := db.db.QueryRowxContext(ctx, query, id).StructScan(&roaster); err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, domainerrors.ErrRoasterDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for roaster id=%d from the database: %w", id, err)
	}
	return &roaster, nil
}

func (db *Roaster) GetRoasterByName(ctx context.Context, name string) (*sql.Roaster, error) {
	var roaster sql.Roaster
	query := db.dialect.Rebind("SELECT id, name, created_at, updated_at FROM roasters WHERE name = ?")
	if err := db.db.QueryRowxContext(ctx, query, name).StructScan(&roaster); err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, domainerrors.ErrRoasterDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for roaster name=\"%s\" from the database: %w", name, err)
	}
	return &roaster, nil
}

func (db *Roaster) GetAllRoasters(ctx context.Context) ([]sql.Roaster, error) {
	roasters := make([]sql.Roaster, 0)
	query := db.dialect.Rebind("SELECT id, name, created_at, updated_at FROM roasters")
	if err := db.db.SelectContext(ctx, &roasters, query); err != nil {
		return roasters, fmt.Errorf("failed to read records for roasters: %w", err)
	}
	return roasters, nil
}

func (db *Roaster) UpdateRoasterById(ctx context.Context, id int, roaster *sql.Roaster) (*sql.Roaster, error) {
	roaster.Id = id
	query := db.dialect.Rebind(`UPDATE roasters SET name = ? WHERE id = ?`)
	res, err := db.db.ExecContext(ctx, query, roaster.Name, roaster.Id)
	if err != nil {
		return nil, db.dialect.ParseError(err, &entityRoaster, fmt.Errorf("failed to update record for roaster id=%d: %w", id, err))
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		if _, err := db.GetRoasterById(ctx, id); err != nil {
			return nil, err
		}
	}
	return roaster, nil
}

func (db *Roaster) DeleteRoasterById(ctx context.Context, id int) error {
	query := db.dialect.Rebind(`DELETE FROM roasters WHERE id = ?`)
	res, err := db.db.ExecContext(ctx, query, id)
	if err != nil {
		return db.dialect.ParseError(err, nil, fmt.Errorf("failed to delete record for roaster id=%d: %w", id, err))
	}
	if row, _ := res.RowsAffected(); row != 1 {
		return domainerrors.ErrRoasterDoesNotExist
	}
	return nil
}

func (db *Roaster) Ping(ctx context.Context) error { return db.db.PingContext(ctx) }

type Sheet struct {
	db      *sqlx.DB
	dialect Dialect
}

func NewSheet(db *sqlx.DB, dialect Dialect) *Sheet { return &Sheet{db: db, dialect: dialect} }

func (db *Sheet) CreateSheet(ctx context.Context, sheet *sql.Sheet) error {
	query := db.dialect.Rebind(`INSERT INTO sheets (name) VALUES (?)`)
	_, err := db.db.ExecContext(ctx, query, sheet.Name)
	if err != nil {
		return db.dialect.ParseError(err, &entitySheet, fmt.Errorf("failed to insert record to the database: %w", err))
	}
	return nil
}

func (db *Sheet) GetSheetById(ctx context.Context, id int) (*sql.Sheet, error) {
	var sheet sql.Sheet
	query := db.dialect.Rebind("SELECT id, name, created_at, updated_at FROM sheets WHERE id = ?")
	if err := db.db.QueryRowxContext(ctx, query, id).StructScan(&sheet); err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, domainerrors.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet id=%d from the database: %w", id, err)
	}
	return &sheet, nil
}

func (db *Sheet) GetSheetByName(ctx context.Context, name string) (*sql.Sheet, error) {
	var sheet sql.Sheet
	query := db.dialect.Rebind("SELECT id, name, created_at, updated_at FROM sheets WHERE name = ?")
	if err := db.db.QueryRowxContext(ctx, query, name).StructScan(&sheet); err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, domainerrors.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet name=\"%s\" from the database: %w", name, err)
	}
	return &sheet, nil
}

func (db *Sheet) GetAllSheets(ctx context.Context) ([]sql.Sheet, error) {
	sheets := make([]sql.Sheet, 0)
	query := db.dialect.Rebind("SELECT id, name, created_at, updated_at FROM sheets")
	if err := db.db.SelectContext(ctx, &sheets, query); err != nil {
		return sheets, fmt.Errorf("failed to read records for sheets: %w", err)
	}
	return sheets, nil
}

func (db *Sheet) UpdateSheetById(ctx context.Context, id int, sheet *sql.Sheet) (*sql.Sheet, error) {
	sheet.Id = id
	query := db.dialect.Rebind(`UPDATE sheets SET name = ? WHERE id = ?`)
	res, err := db.db.ExecContext(ctx, query, sheet.Name, sheet.Id)
	if err != nil {
		return nil, db.dialect.ParseError(err, &entitySheet, fmt.Errorf("failed to update record for sheet id=%d: %w", id, err))
	}
	if rowsAffected, _ := res.RowsAffected(); rowsAffected == 0 {
		if _, err := db.GetSheetById(ctx, id); err != nil {
			return nil, err
		}
	}
	return sheet, nil
}

func (db *Sheet) DeleteSheetById(ctx context.Context, id int) error {
	query := db.dialect.Rebind(`DELETE FROM sheets WHERE id = ?`)
	res, err := db.db.ExecContext(ctx, query, id)
	if err != nil {
		return db.dialect.ParseError(err, nil, fmt.Errorf("failed to delete record for sheet id=%d: %w", id, err))
	}
	if row, _ := res.RowsAffected(); row != 1 {
		return domainerrors.ErrSheetDoesNotExist
	}
	return nil
}

func (db *Sheet) Ping(ctx context.Context) error { return db.db.PingContext(ctx) }

type Shot struct {
	db      *sqlx.DB
	dialect Dialect
}

func NewShot(db *sqlx.DB, dialect Dialect) *Shot { return &Shot{db: db, dialect: dialect} }

func (db *Shot) CreateShot(ctx context.Context, shot *sql.Shot) (int, error) {
	query := db.dialect.Rebind(`INSERT INTO
	shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparison_with_previous_result, additional_notes)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	return db.dialect.InsertID(ctx, db.db, query, &entityShot, shot.Sheet.Id, shot.Beans.Id, shot.GrindSetting, shot.QuantityIn, shot.QuantityOut, shot.ShotTime, shot.WaterTemperature, shot.Rating, shot.IsTooBitter, shot.IsTooSour, shot.ComparisonWithPreviousResult, shot.AdditionalNotes)
}

func (db *Shot) GetShotById(ctx context.Context, id int) (*sql.Shot, error) {
	var shot sql.Shot
	query := db.dialect.Rebind(shotQuery + "\nWHERE shots.id = ?")
	if err := db.db.QueryRowxContext(ctx, query, id).StructScan(&shot); err != nil {
		if errors.Is(err, dbsql.ErrNoRows) {
			return nil, domainerrors.ErrShotDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for shot id=%d from the database: %w", id, err)
	}
	return &shot, nil
}

func (db *Shot) GetAllShots(ctx context.Context) ([]sql.Shot, error) {
	shots := make([]sql.Shot, 0)
	if err := db.db.SelectContext(ctx, &shots, db.dialect.Rebind(shotQuery)); err != nil {
		return shots, fmt.Errorf("failed to read records for shots: %w", err)
	}
	return shots, nil
}

func (db *Shot) UpdateShotById(ctx context.Context, id int, shot *sql.Shot) (*sql.Shot, error) {
	query := db.dialect.Rebind(`UPDATE shots SET
	sheet_id = ?, beans_id = ?, grind_setting = ?, quantity_in = ?, quantity_out = ?, shot_time = ?, water_temperature = ?, rating = ?, is_too_bitter = ?, is_too_sour = ?, comparison_with_previous_result = ?, additional_notes = ?
	WHERE id = ?`)
	if _, err := db.db.ExecContext(ctx, query, shot.Sheet.Id, shot.Beans.Id, shot.GrindSetting, shot.QuantityIn, shot.QuantityOut, shot.ShotTime, shot.WaterTemperature, shot.Rating, shot.IsTooBitter, shot.IsTooSour, shot.ComparisonWithPreviousResult, shot.AdditionalNotes, id); err != nil {
		return nil, db.dialect.ParseError(err, &entityShot, fmt.Errorf("failed to update record in the database: %w", err))
	}
	return shot, nil
}

func (db *Shot) DeleteShotById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, db.dialect.Rebind(`DELETE FROM shots WHERE id = ?`), id)
	if err != nil {
		return db.dialect.ParseError(err, nil, fmt.Errorf("failed to delete record for shots id=%d: %w", id, err))
	}
	if row, _ := res.RowsAffected(); row != 1 {
		return domainerrors.ErrShotDoesNotExist
	}
	return nil
}

func (db *Shot) Ping(ctx context.Context) error { return db.db.PingContext(ctx) }

const shotQuery = `
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
