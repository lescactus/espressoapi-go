package roaster

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

var _ repository.RoasterRepository = (*Roaster)(nil)

type Roaster struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Roaster {
	return &Roaster{
		db: db,
	}
}

func (db *Roaster) CreateRoaster(ctx context.Context, sheet *sql.Roaster) error {
	query := `INSERT INTO roasters (name) VALUES (?)`
	_, err := db.db.ExecContext(ctx, query, sheet.Name)
	if err != nil {
		return mysqlerrors.ParseMySQLError(err, &mysqlerrors.EntityRoaster, fmt.Errorf("failed to insert record to the database: %w", err))
	}

	return nil
}

func (db *Roaster) GetRoasterById(ctx context.Context, id int) (*sql.Roaster, error) {
	var r sql.Roaster

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM roasters WHERE id = ?", id).StructScan(&r)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrRoasterDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for roaster id=%d from the database: %w", id, err)
	}

	return &r, nil
}

func (db *Roaster) GetRoasterByName(ctx context.Context, name string) (*sql.Roaster, error) {
	var r sql.Roaster

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM roasters WHERE name = ?", name).StructScan(&r)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrRoasterDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for roaster name=\"%s\" from the database: %w", name, err)
	}

	return &r, nil
}

func (db *Roaster) GetAllRoasters(ctx context.Context) ([]sql.Roaster, error) {
	var roasters = make([]sql.Roaster, 0)
	err := db.db.SelectContext(ctx, &roasters, "SELECT * FROM roasters")
	if err != nil {
		return roasters, fmt.Errorf("failed to read records for roasters: %w", err)
	}

	return roasters, nil
}

func (db *Roaster) UpdateRoasterById(ctx context.Context, id int, roaster *sql.Roaster) (*sql.Roaster, error) {
	now := time.Now()
	roaster.Id = id
	roaster.UpdatedAt = &now

	// CreatedAt should be immutable
	res, err := db.db.ExecContext(ctx, `UPDATE roasters SET name = ?, updated_at = ? WHERE id = ?`, roaster.Name, roaster.UpdatedAt, roaster.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to update record for roaster id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return nil, errors.ErrRoasterDoesNotExist
	}

	return db.GetRoasterById(ctx, roaster.Id)
}

func (db *Roaster) DeleteRoasterById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, `DELETE FROM roasters WHERE id = ?`, id)
	if err != nil {
		return mysqlerrors.ParseMySQLError(err, nil, fmt.Errorf("failed to delete record for roaster id=%d: %w", id, err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return errors.ErrRoasterDoesNotExist
	}

	return nil
}

func (db *Roaster) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
