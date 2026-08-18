package roaster

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

var _ repository.RoasterRepository = (*Roaster)(nil)

type Roaster struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Roaster {
	return &Roaster{db: db}
}

func (db *Roaster) CreateRoaster(ctx context.Context, roaster *sql.Roaster) error {
	_, err := db.db.ExecContext(ctx, "INSERT INTO roasters (name) VALUES ($1)", roaster.Name)
	if err != nil {
		return postgreserrors.ParsePostgresError(err, &postgreserrors.EntityRoaster, fmt.Errorf("failed to insert record to the database: %w", err))
	}

	return nil
}

func (db *Roaster) GetRoasterById(ctx context.Context, id int) (*sql.Roaster, error) {
	var roaster sql.Roaster
	if err := db.db.QueryRowxContext(ctx, "SELECT id, name, created_at, updated_at FROM roasters WHERE id = $1", id).StructScan(&roaster); err != nil {
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrRoasterDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for roaster id=%d from the database: %w", id, err)
	}

	return &roaster, nil
}

func (db *Roaster) GetRoasterByName(ctx context.Context, name string) (*sql.Roaster, error) {
	var roaster sql.Roaster
	if err := db.db.QueryRowxContext(ctx, "SELECT id, name, created_at, updated_at FROM roasters WHERE name = $1", name).StructScan(&roaster); err != nil {
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrRoasterDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for roaster name=\"%s\" from the database: %w", name, err)
	}

	return &roaster, nil
}

func (db *Roaster) GetAllRoasters(ctx context.Context) ([]sql.Roaster, error) {
	roasters := make([]sql.Roaster, 0)
	if err := db.db.SelectContext(ctx, &roasters, "SELECT id, name, created_at, updated_at FROM roasters"); err != nil {
		return roasters, fmt.Errorf("failed to read records for roasters: %w", err)
	}

	return roasters, nil
}

func (db *Roaster) UpdateRoasterById(ctx context.Context, id int, roaster *sql.Roaster) (*sql.Roaster, error) {
	roaster.Id = id
	result, err := db.db.ExecContext(ctx, "UPDATE roasters SET name = $1 WHERE id = $2", roaster.Name, roaster.Id)
	if err != nil {
		return nil, postgreserrors.ParsePostgresError(err, &postgreserrors.EntityRoaster, fmt.Errorf("failed to update record for roaster id=%d: %w", id, err))
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		if _, err := db.GetRoasterById(ctx, id); err != nil {
			return nil, err
		}
	}

	return roaster, nil
}

func (db *Roaster) DeleteRoasterById(ctx context.Context, id int) error {
	result, err := db.db.ExecContext(ctx, "DELETE FROM roasters WHERE id = $1", id)
	if err != nil {
		return postgreserrors.ParsePostgresError(err, nil, fmt.Errorf("failed to delete record for roaster id=%d: %w", id, err))
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected != 1 {
		return errors.ErrRoasterDoesNotExist
	}

	return nil
}

func (db *Roaster) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
