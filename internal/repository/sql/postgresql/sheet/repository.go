package sheet

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

var _ repository.SheetRepository = (*Sheet)(nil)

type Sheet struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Sheet {
	return &Sheet{db: db}
}

func (db *Sheet) CreateSheet(ctx context.Context, sheet *sql.Sheet) error {
	_, err := db.db.ExecContext(ctx, "INSERT INTO sheets (name) VALUES ($1)", sheet.Name)
	if err != nil {
		return postgreserrors.ParsePostgresError(err, &postgreserrors.EntitySheet, fmt.Errorf("failed to insert record to the database: %w", err))
	}

	return nil
}

func (db *Sheet) GetSheetById(ctx context.Context, id int) (*sql.Sheet, error) {
	var sheet sql.Sheet
	if err := db.db.QueryRowxContext(ctx, "SELECT * FROM sheets WHERE id = $1", id).StructScan(&sheet); err != nil {
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet id=%d from the database: %w", id, err)
	}

	return &sheet, nil
}

func (db *Sheet) GetSheetByName(ctx context.Context, name string) (*sql.Sheet, error) {
	var sheet sql.Sheet
	if err := db.db.QueryRowxContext(ctx, "SELECT * FROM sheets WHERE name = $1", name).StructScan(&sheet); err != nil {
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet name=\"%s\" from the database: %w", name, err)
	}

	return &sheet, nil
}

func (db *Sheet) GetAllSheets(ctx context.Context) ([]sql.Sheet, error) {
	sheets := make([]sql.Sheet, 0)
	if err := db.db.SelectContext(ctx, &sheets, "SELECT * FROM sheets"); err != nil {
		return sheets, fmt.Errorf("failed to read records for sheets: %w", err)
	}

	return sheets, nil
}

func (db *Sheet) UpdateSheetById(ctx context.Context, id int, sheet *sql.Sheet) (*sql.Sheet, error) {
	sheet.Id = id
	result, err := db.db.ExecContext(ctx, "UPDATE sheets SET name = $1 WHERE id = $2", sheet.Name, sheet.Id)
	if err != nil {
		return nil, postgreserrors.ParsePostgresError(err, &postgreserrors.EntitySheet, fmt.Errorf("failed to update record for sheet id=%d: %w", id, err))
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		if _, err := db.GetSheetById(ctx, id); err != nil {
			return nil, err
		}
	}

	return sheet, nil
}

func (db *Sheet) DeleteSheetById(ctx context.Context, id int) error {
	result, err := db.db.ExecContext(ctx, "DELETE FROM sheets WHERE id = $1", id)
	if err != nil {
		return postgreserrors.ParsePostgresError(err, nil, fmt.Errorf("failed to delete record for sheet id=%d: %w", id, err))
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected != 1 {
		return errors.ErrSheetDoesNotExist
	}

	return nil
}

func (db *Sheet) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
