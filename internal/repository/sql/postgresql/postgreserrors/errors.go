package postgreserrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	sqlerrors "github.com/lescactus/espressoapi-go/internal/repository/sql/errors"
)

type Entity = sqlerrors.Entity

var (
	checkConstraintErrors = map[string]error{
		"chk_beans_roast_level":                     domainerrors.ErrBeansRoastLevelOutOfRange,
		"chk_shots_comparison_with_previous_result": domainerrors.ErrShotComparisonWithPreviousResultOutOfRange,
	}
	foreignKeyReferenceErrors = map[string]error{
		"beans_roaster_id_fkey": domainerrors.ErrRoasterDoesNotExist,
		"shots_sheet_id_fkey":   domainerrors.ErrSheetDoesNotExist,
		"shots_beans_id_fkey":   domainerrors.ErrBeansDoesNotExist,
	}
)

var (
	EntitySheet   = sqlerrors.EntitySheet
	EntityRoaster = sqlerrors.EntityRoaster
	EntityBeans   = sqlerrors.EntityBeans
	EntityShot    = sqlerrors.EntityShot
)

var (
	entityToErrAlreadyExists        = sqlerrors.EntityToErrAlreadyExists
	entityToErrForeignKeyConstraint = sqlerrors.EntityToErrForeignKeyConstraint
	entityToErrDoesNotExist         = sqlerrors.EntityToErrDoesNotExist
)

func ParsePostgresError(err error, entity *Entity, fallback error) error {
	if err == nil {
		return nil
	}

	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return fallback
	}

	switch postgresError.Code {
	case "23505":
		if entity == nil {
			return fallback
		}
		return sqlerrors.MappedEntityError(entityToErrAlreadyExists, *entity, fallback)
	case "23503":
		if entity == nil {
			return sqlerrors.MappedEntityError(entityToErrForeignKeyConstraint, Entity(postgresError.TableName), fallback)
		}
		if mappedError, ok := foreignKeyReferenceErrors[postgresError.ConstraintName]; ok {
			return mappedError
		}
		return sqlerrors.MappedEntityError(entityToErrDoesNotExist, *entity, fallback)
	case "23514":
		if mappedError, ok := checkConstraintErrors[postgresError.ConstraintName]; ok {
			return mappedError
		}
	}

	return fallback
}
