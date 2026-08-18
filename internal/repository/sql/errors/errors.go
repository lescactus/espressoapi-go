// Package sqlerrors contains shared database error mappings.
package sqlerrors

import domainerrors "github.com/lescactus/espressoapi-go/internal/errors"

// Entity identifies a database table used by repository error parsing.
type Entity string

const (
	EntitySheet   Entity = "sheets"
	EntityRoaster Entity = "roasters"
	EntityBeans   Entity = "beans"
	EntityShot    Entity = "shots"
)

// EntityToErrAlreadyExists maps entities to duplicate-entry domain errors.
var EntityToErrAlreadyExists = map[Entity]error{
	EntitySheet:   domainerrors.ErrSheetAlreadyExists,
	EntityRoaster: domainerrors.ErrRoasterAlreadyExists,
	EntityBeans:   domainerrors.ErrBeansAlreadyExists,
	EntityShot:    domainerrors.ErrShotAlreadyExists,
}

// EntityToErrForeignKeyConstraint maps entities to delete constraint errors.
var EntityToErrForeignKeyConstraint = map[Entity]error{
	EntityBeans: domainerrors.ErrBeansForeignKeyConstraint,
	EntityShot:  domainerrors.ErrShotForeignKeyConstraint,
}

// EntityToErrDoesNotExist maps entities to missing-record domain errors.
var EntityToErrDoesNotExist = map[Entity]error{
	EntitySheet:   domainerrors.ErrSheetDoesNotExist,
	EntityRoaster: domainerrors.ErrRoasterDoesNotExist,
	EntityBeans:   domainerrors.ErrBeansDoesNotExist,
	EntityShot:    domainerrors.ErrShotDoesNotExist,
}

// MappedEntityError returns the mapped error for an entity or the fallback.
func MappedEntityError(entityErrors map[Entity]error, entity Entity, fallback error) error {
	if err, ok := entityErrors[entity]; ok {
		return err
	}

	return fallback
}
