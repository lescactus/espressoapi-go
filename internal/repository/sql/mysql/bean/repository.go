package bean

import (
	"context"
	dbsql "database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/mysqlerrors"
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

func (db *Bean) CreateBeans(ctx context.Context, beans *sql.Beans) (int, error) {
	query := `INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)`
	res, err := db.db.ExecContext(ctx, query,
		beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel)
	if err != nil {
		return 0, mysqlerrors.ParseMySQLError(err, &mysqlerrors.EntityRoaster, fmt.Errorf("failed to insert record to the database: %w", err))
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to retrieve last inserted id: %w", err)
	}

	return int(id), nil
}

func (db *Bean) GetBeansById(ctx context.Context, id int) (*sql.Beans, error) {
	var b sql.Beans

	get := `
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
		beans.created_at,
		beans.updated_at,
		roaster.id AS "roaster.id",
		roaster.name AS "roaster.name",
		roaster.created_at AS "roaster.created_at",
		roaster.updated_at AS "roaster.updated_at"
	FROM beans
		INNER JOIN roasters roaster
			ON beans.roaster_id = roaster.id`

	if err := db.db.SelectContext(ctx, &beans, get); err != nil {
		return beans, fmt.Errorf("failed to read records for beans: %w", err)
	}

	return beans, nil
}

func (db *Bean) UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error) {
	_, err := db.db.ExecContext(ctx, `UPDATE beans SET name = ?, roaster_id = ?, roast_date = ?, roast_level = ? WHERE id = ?`,
		beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel, id)
	if err != nil {
		return nil, mysqlerrors.ParseMySQLError(err, &mysqlerrors.EntityRoaster, fmt.Errorf("failed to update record for beans id=%d: %w", id, err))
	}

	return beans, nil
}

func (db *Bean) DeleteBeansById(ctx context.Context, id int) error {
	res, err := db.db.ExecContext(ctx, `DELETE FROM beans WHERE id = ?`, id)
	if err != nil {
		return mysqlerrors.ParseMySQLError(err, nil, fmt.Errorf("failed to delete record for beans id=%d: %w", id, err))
	}

	if row, _ := res.RowsAffected(); row != 1 {
		return errors.ErrBeansDoesNotExist
	}

	return nil
}

func (db *Bean) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
