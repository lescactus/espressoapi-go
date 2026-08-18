package adapters

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	sqlerrors "github.com/lescactus/espressoapi-go/internal/repository/sql/errors"
	mysqlerrors "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/mysqlerrors"
	postgreserrors "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/postgreserrors"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/shared"
)

func MySQL() shared.Dialect {
	return shared.Dialect{
		Rebind:     func(query string) string { return query },
		ParseError: mysqlerrors.ParseMySQLError,
		InsertID: func(ctx context.Context, db *sqlx.DB, query string, entity *sqlerrors.Entity, args ...any) (int, error) {
			result, err := db.ExecContext(ctx, query, args...)
			if err != nil {
				return 0, mysqlerrors.ParseMySQLError(err, entity, fmt.Errorf("failed to insert record to the database: %w", err))
			}
			id, err := result.LastInsertId()
			if err != nil {
				return 0, fmt.Errorf("failed to retrieve last inserted id: %w", err)
			}
			return int(id), nil
		},
	}
}

func PostgreSQL() shared.Dialect {
	return shared.Dialect{
		Rebind:     func(query string) string { return sqlx.Rebind(sqlx.DOLLAR, query) },
		ParseError: postgreserrors.ParsePostgresError,
		InsertID: func(ctx context.Context, db *sqlx.DB, query string, entity *sqlerrors.Entity, args ...any) (int, error) {
			var id int
			if err := db.QueryRowxContext(ctx, query+" RETURNING id", args...).Scan(&id); err != nil {
				return 0, postgreserrors.ParsePostgresError(err, entity, fmt.Errorf("failed to insert record to the database: %w", err))
			}
			return id, nil
		},
	}
}
