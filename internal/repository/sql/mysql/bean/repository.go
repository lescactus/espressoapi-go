package bean

import (
	"context"
	dbsql "database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
)

var _ repository.BeansRepository = (*Bean)(nil)

type Bean struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Bean {
	return &Bean{
		db: db,
	}
}

func (db *Bean) CreateBeans(ctx context.Context, beans *sql.Beans) error {
	query := `INSERT INTO beans (roaster_name, beans_name, roast_date, roast_level) VALUES (?, ?, ?, ?)`
	_, err := db.db.ExecContext(ctx, query,
		beans.RoasterName, beans.BeansName, beans.RoastDate, beans.RoastLevel)
	if err != nil {
		return fmt.Errorf("failed to insert record to the database: %w", err)
	}

	return nil
}

func (db *Bean) GetBeansById(ctx context.Context, id int) (*sql.Beans, error) {
	var b sql.Beans

	err := db.db.QueryRowxContext(ctx, "SELECT * FROM beans WHERE id = ?", id).StructScan(&b)
	if err != nil {
		// No row found, return nil
		if err == dbsql.ErrNoRows {
			return nil, errors.ErrBeansDoesNotExist
		}
		return nil, fmt.Errorf("failed to read record for beans id=%d from the database: %w", id, err)
	}

	return &b, nil
}

func (db *Bean) GetAllBeans(ctx context.Context) ([]sql.Beans, error) {
	var beans = make([]sql.Beans, 0)
	err := db.db.SelectContext(ctx, &beans, "SELECT * FROM beans")
	if err != nil {
		return beans, fmt.Errorf("failed to read records for beans: %w", err)
	}

	return beans, nil
}

func (db *Bean) UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error) {
	beans.Id = id

	res, err := db.db.ExecContext(ctx, `UPDATE beans SET roaster_name = ?, beans_name = ?, roast_date = ?, roast_level = ? WHERE id = ?`,
		beans.RoasterName, beans.BeansName, beans.RoastDate, beans.RoastLevel, beans.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to update record for beans id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return nil, errors.ErrBeansDoesNotExist
	}

	return beans, nil
}

func (db *Bean) DeleteBeansById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, `DELETE FROM beans WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete record for beans id=%d: %w", id, err)
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return errors.ErrBeansDoesNotExist
	}

	return nil
}

func (db *Bean) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
