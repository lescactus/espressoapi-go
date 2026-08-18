package bean

import (
	"context"
	dbsql "database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/postgreserrors"
)

var _ repository.BeansRepository = (*Bean)(nil)

type Bean struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Bean {
	return &Bean{db: db}
}

func (db *Bean) CreateBeans(ctx context.Context, beans *sql.Beans) (int, error) {
	var id int
	err := db.db.QueryRowxContext(ctx, "INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES ($1, $2, $3, $4) RETURNING id",
		beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel).Scan(&id)
	if err != nil {
		return 0, postgreserrors.ParsePostgresError(err, &postgreserrors.EntityBeans, fmt.Errorf("failed to insert record to the database: %w", err))
	}

	return id, nil
}

func (db *Bean) GetBeansById(ctx context.Context, id int) (*sql.Beans, error) {
	var beans sql.Beans
	query := `
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
	beans.id = $1`

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
	query := `
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

	if err := db.db.SelectContext(ctx, &beans, query); err != nil {
		return beans, fmt.Errorf("failed to read records for beans: %w", err)
	}

	return beans, nil
}

func (db *Bean) UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error) {
	_, err := db.db.ExecContext(ctx, "UPDATE beans SET name = $1, roaster_id = $2, roast_date = $3, roast_level = $4 WHERE id = $5",
		beans.Name, beans.Roaster.Id, beans.RoastDate, beans.RoastLevel, id)
	if err != nil {
		return nil, postgreserrors.ParsePostgresError(err, &postgreserrors.EntityBeans, fmt.Errorf("failed to update record for beans id=%d: %w", id, err))
	}

	return beans, nil
}

func (db *Bean) DeleteBeansById(ctx context.Context, id int) error {
	result, err := db.db.ExecContext(ctx, "DELETE FROM beans WHERE id = $1", id)
	if err != nil {
		return postgreserrors.ParsePostgresError(err, nil, fmt.Errorf("failed to delete record for beans id=%d: %w", id, err))
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected != 1 {
		return domainerrors.ErrBeansDoesNotExist
	}

	return nil
}

func (db *Bean) Ping(ctx context.Context) error {
	return db.db.PingContext(ctx)
}
