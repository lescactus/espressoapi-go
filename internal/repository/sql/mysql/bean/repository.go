package bean

import (
	"context"
	dbsql "database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"

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
	query := `INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)`
	_, err := db.db.ExecContext(ctx, query,
		beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel)
	if err != nil {
		// Checking if the error is due to a foreign key constraint
		// which will indicate the roaste does not exists:
		// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails
		if me, ok := err.(*mysql.MySQLError); ok && me.Number == 1452 {
			return errors.ErrRoasterDoesNotExist
		}

		return fmt.Errorf("failed to insert record to the database: %w", err)
	}

	return nil
}

func (db *Bean) GetBeansById(ctx context.Context, id int) (*sql.Beans, error) {
	var b sql.Beans

	get := `
SELECT
	beans.id,
	beans.name,
	beans.roast_date,
	beans.roast_level,
	roaster.id AS "roaster.id",
	roaster.name AS "roaster.name"
FROM beans
	INNER JOIN roasters roaster
		ON beans.roaster_id = roaster.id
WHERE
	beans.id = ?`

	if err := db.db.QueryRowxContext(ctx, get, id).StructScan(&b); err != nil {
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

	get := `
	SELECT
		beans.id,
		beans.name,
		beans.roast_date,
		beans.roast_level,
		roaster.id AS "roaster.id",
		roaster.name AS "roaster.name"
	FROM beans
		INNER JOIN roasters roaster
			ON beans.roaster_id = roaster.id`

	if err := db.db.SelectContext(ctx, &beans, get); err != nil {
		return beans, fmt.Errorf("failed to read records for beans: %w", err)
	}

	return beans, nil
}

func (db *Bean) UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error) {
	beans.Id = id

	res, err := db.db.ExecContext(ctx, `UPDATE beans SET name = ?, roaster_id = ? roast_date = ?, roast_level = ? WHERE id = ?`,
		beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel, beans.Id)
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
