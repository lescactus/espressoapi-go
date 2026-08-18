package postgreserrors

import (
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
)

type Entity string

var (
	EntitySheet   Entity = "sheets"
	EntityRoaster Entity = "roasters"
	EntityBeans   Entity = "beans"
	EntityShot    Entity = "shots"

	checkConstraintErrors = map[string]error{
		"chk_beans_roast_level":                     domainerrors.ErrBeansRoastLevelOutOfRange,
		"chk_shots_comparison_with_previous_result": domainerrors.ErrShotComparisonWithPreviousResultOutOfRange,
	}
	foreignKeyReferenceErrors = map[string]error{
		"beans_roaster_id_fkey": domainerrors.ErrRoasterDoesNotExist,
		"shots_sheet_id_fkey":   domainerrors.ErrSheetDoesNotExist,
		"shots_beans_id_fkey":   domainerrors.ErrBeansDoesNotExist,
	}
	entityToErrAlreadyExists = map[Entity]error{
		EntitySheet:   domainerrors.ErrSheetAlreadyExists,
		EntityRoaster: domainerrors.ErrRoasterAlreadyExists,
		EntityBeans:   domainerrors.ErrBeansAlreadyExists,
		EntityShot:    domainerrors.ErrShotAlreadyExists,
	}
	entityToErrForeignKeyConstraint = map[Entity]error{
		EntityBeans: domainerrors.ErrBeansForeignKeyConstraint,
		EntityShot:  domainerrors.ErrShotForeignKeyConstraint,
	}
	entityToErrDoesNotExist = map[Entity]error{
		EntitySheet:   domainerrors.ErrSheetDoesNotExist,
		EntityRoaster: domainerrors.ErrRoasterDoesNotExist,
		EntityBeans:   domainerrors.ErrBeansDoesNotExist,
		EntityShot:    domainerrors.ErrShotDoesNotExist,
	}
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
		return mappedEntityError(entityToErrAlreadyExists, *entity, fallback)
	case "23503":
		if entity == nil {
			return mappedEntityError(entityToErrForeignKeyConstraint, Entity(postgresError.TableName), fallback)
		}
		if mappedError, ok := foreignKeyReferenceErrors[postgresError.ConstraintName]; ok {
			return mappedError
		}
		return mappedEntityError(entityToErrDoesNotExist, *entity, fallback)
	case "23514":
		if mappedError, ok := checkConstraintErrors[postgresError.ConstraintName]; ok {
			return mappedError
		}
	}

	return fallback
}

func mappedEntityError(entityErrors map[Entity]error, entity Entity, fallback error) error {
	if err, ok := entityErrors[entity]; ok {
		return err
	}

	return fallback
}
