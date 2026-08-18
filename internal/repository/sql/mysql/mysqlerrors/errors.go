package mysqlerrors

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	sqlerrors "github.com/lescactus/espressoapi-go/internal/repository/sql/errors"
)

type Entity = sqlerrors.Entity

var (
	error1451TablePattern = regexp.MustCompile(`\x60([^\x60]+)\x60\.\x60([^\x60]+)\x60`)
	error1452TablePattern = regexp.MustCompile(`FOREIGN KEY \(\x60(.+?)\x60\) REFERENCES \x60(.+?)\x60 \(\x60id\x60`)
	checkConstraintErrors = map[string]error{
		"chk_beans_roast_level":                     domainerrors.ErrBeansRoastLevelOutOfRange,
		"chk_shots_comparison_with_previous_result": domainerrors.ErrShotComparisonWithPreviousResultOutOfRange,
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

// ParseMySQLError parses a MySQL error and returns a more specific error based on the error code.
// If the error is nil, it returns nil.
// If the error is a duplicate entry error (ERROR 1062), it returns the corresponding error for the entity.
// If the error is a foreign key constraint error (ERROR 1451), it returns the corresponding error for the entity or table.
// If the error is a foreign key constraint error indicating that the entity does not exist (ERROR 1452),
// it returns the corresponding error for the entity or table.
// If the error does not match any specific error code, it returns the fallback error.
func ParseMySQLError(err error, entity *Entity, fallback error) error {
	if err == nil {
		return nil
	}

	// Checking if the entry inserted is a duplicate:
	// ERROR 1062 (23000): Duplicate entry 'xxxxx' for key 'yyyy'
	var me *mysql.MySQLError
	if errors.As(err, &me) {
		// ERROR 3819 (HY000): Check constraint 'constraint_name' is violated.
		if me.Number == 3819 {
			for constraint, mappedErr := range checkConstraintErrors {
				if strings.Contains(me.Message, constraint) {
					return mappedErr
				}
			}
			return fallback
		}

		if me.Number == 1062 {
			if entity == nil {
				return fallback
			}
			return sqlerrors.MappedEntityError(entityToErrAlreadyExists, *entity, fallback)
		}

		// Checking if the error is due to a foreign key constraint
		// which will indicate the entity cannot be deleted due to existing references:
		// ERROR 1451 (23000): Cannot delete or update a parent row: a foreign key constraint fails
		if me.Number == 1451 {
			if entity == nil {
				table, err := ExtractTableNameFromError1451(*me)
				if err != nil {
					return fallback
				}
				return sqlerrors.MappedEntityError(entityToErrForeignKeyConstraint, table, fallback)
			}
			return sqlerrors.MappedEntityError(entityToErrForeignKeyConstraint, *entity, fallback)
		}

		// Checking if the error is due to a foreign key constraint
		// which will indicate the entity does not exists:
		// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails
		if me.Number == 1452 {
			table, err := ExtractTableNameFromError1452(*me)
			if err == nil {
				return sqlerrors.MappedEntityError(entityToErrDoesNotExist, table, fallback)
			}
			if entity == nil {
				return fallback
			}
			return sqlerrors.MappedEntityError(entityToErrDoesNotExist, *entity, fallback)
		}
	}

	return fallback
}

// ExtractTableNameFromError1451 extracts the table name from a MySQL error with code 1451.
// It uses a regular expression to find the table name in the error message.
// If a match is found, it returns the table name. Otherwise, it returns an error.
//
// Example error message:
// "Cannot delete or update a parent row: a foreign key constraint fails (`espresso-api`.`beans`, CONSTRAINT `beans_ibfk_1` FOREIGN KEY (`roaster_id`) REFERENCES `roasters` (`id`))"
func ExtractTableNameFromError1451(err mysql.MySQLError) (Entity, error) {
	if err.Number != 1451 {
		return "", fmt.Errorf("error is not mysql error 1451")
	}

	// Use the regular expression to find the table name in the error message
	matches := error1451TablePattern.FindStringSubmatch(err.Error())

	// Check if a match was found
	if len(matches) > 0 {
		// The second element in matches will be the table name
		return Entity(matches[2]), nil
	} else {
		return "", fmt.Errorf("failed to extract table name from error message")
	}
}

// ExtractTableNameFromError1452 extracts the table name from a MySQL error with code 1452.
// It uses a regular expression to find the table name in the error message.
// If a match is found, it returns the table name. Otherwise, it returns an error.
//
// Example error message:
// "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `sheets` (`id`))"
func ExtractTableNameFromError1452(err mysql.MySQLError) (Entity, error) {
	if err.Number != 1452 {
		return "", fmt.Errorf("error is not mysql error 1452")
	}

	// Use the regular expression to find the table name in the error message
	matches := error1452TablePattern.FindStringSubmatch(err.Error())

	// Check if a match was found
	if len(matches) > 0 {
		// The second element in matches will be the table name
		return Entity(matches[2]), nil
	} else {
		return "", fmt.Errorf("failed to extract table name from error message")
	}
}
