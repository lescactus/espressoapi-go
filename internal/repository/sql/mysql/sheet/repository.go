package sheet

import (
	"context"
	"fmt"
	"time"

	dbsql "database/sql"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/mysqlerrors"
)

var _ repository.SheetRepository = (*Sheet)(nil)

type Sheet struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Sheet {
	return &Sheet{
		db: db,
	}
}

func (db *Sheet) CreateSheet(ctx context.Context, sheet *sql.Sheet) error {
	query := `INSERT INTO sheets (name) VALUES (?)`
	_, err := db.db.ExecContext(ctx, query, sheet.Name)
	if err != nil {
		return mysqlerrors.ParseMySQLError(err, &mysqlerrors.EntitySheet, fmt.Errorf("failed to insert record to the database: %w", err))
	}
	return nil
}

func (db *Sheet) GetSheetById(ctx context.Context, id int) (*sql.Sheet, error) {
	var s sql.Sheet

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM sheets WHERE id = ?", id).StructScan(&s)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet id=%d from the database: %w", id, err)
	}

	return &s, nil
}

func (db *Sheet) GetSheetByName(ctx context.Context, name string) (*sql.Sheet, error) {
	var s sql.Sheet

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM sheets WHERE name = ?", name).StructScan(&s)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrSheetDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for sheet name=\"%s\" from the database: %w", name, err)
	}

	return &s, nil
}

func (db *Sheet) GetAllSheets(ctx context.Context) ([]sql.Sheet, error) {
	var sheets = make([]sql.Sheet, 0)
	err := db.db.SelectContext(ctx, &sheets, "SELECT * FROM sheets")
	if err != nil {
		return sheets, fmt.Errorf("failed to read records for sheets: %w", err)
	}

	return sheets, nil
}

func (db *Sheet) UpdateSheetById(ctx context.Context, id int, sheet *sql.Sheet) (*sql.Sheet, error) {
	now := time.Now()
	sheet.Id = id
	sheet.UpdatedAt = &now

	// CreatedAt should be immutable
	res, err := db.db.ExecContext(ctx, `UPDATE sheets SET name = ?, updated_at = ? WHERE id = ?`, sheet.Name, sheet.UpdatedAt, sheet.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to update record for sheet id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return nil, errors.ErrSheetDoesNotExist
	}

	return sheet, nil
}

func (db *Sheet) DeleteSheetById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, `DELETE FROM sheets WHERE id = ?`, id)
	if err != nil {
		return mysqlerrors.ParseMySQLError(err, nil, fmt.Errorf("failed to delete record for sheet id=%d: %w", id, err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return errors.ErrSheetDoesNotExist
	}

	return nil
}

func (db *Sheet) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
