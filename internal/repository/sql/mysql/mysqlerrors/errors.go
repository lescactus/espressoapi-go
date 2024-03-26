package mysqlerrors

import (
	"fmt"
	"regexp"

	"github.com/go-sql-driver/mysql"
	"github.com/lescactus/espressoapi-go/internal/errors"
)

type Entity string

var (
	EntitySheet   Entity = "sheets"
	EntityRoaster Entity = "roaster"
	EntityBeans   Entity = "beans"
	EntityShot    Entity = "shots"

	entityToErrAlreadyExists = map[Entity]error{
		EntitySheet:   errors.ErrSheetAlreadyExists,
		EntityRoaster: errors.ErrRoasterAlreadyExists,
	}

	entityToErrForeignKeyConstraint = map[Entity]error{
		EntityBeans: errors.ErrBeansForeignKeyConstraint,
		EntityShot:  errors.ErrShotForeignKeyConstraint,
	}

	entityToErrDoesNotExist = map[Entity]error{
		EntitySheet:   errors.ErrSheetDoesNotExist,
		EntityRoaster: errors.ErrRoasterDoesNotExist,
		EntityBeans:   errors.ErrBeansDoesNotExist,
		EntityShot:    errors.ErrShotDoesNotExist,
	}
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
	if me, ok := err.(*mysql.MySQLError); ok {
		if me.Number == 1062 {
			if entity == nil {
				return fallback
			}
			return entityToErrAlreadyExists[*entity]
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
				return entityToErrForeignKeyConstraint[table]
			}
			return entityToErrForeignKeyConstraint[*entity]
		}

		// Checking if the error is due to a foreign key constraint
		// which will indicate the entity does not exists:
		// ERROR 1452 (23000): Cannot add or update a child row: a foreign key constraint fails
		if me.Number == 1452 {
			if entity == nil {
				table, err := ExtractTableNameFromError1452(*me)
				if err != nil {
					return fallback
				}
				return entityToErrDoesNotExist[table]
			}
			return entityToErrDoesNotExist[*entity]
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

	// Define the regular expression
	// x60 is the backtick character (`)
	re := regexp.MustCompile(`\x60([^\x60]+)\x60\.\x60([^\x60]+)\x60`)

	// Use the regular expression to find the table name in the error message
	matches := re.FindStringSubmatch(err.Error())

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

	// Define the regular expression
	// x60 is the backtick character (`)
	re := regexp.MustCompile(`FOREIGN KEY \(\x60(.+?)\x60\) REFERENCES \x60(.+?)\x60 \(\x60id\x60`)

	// Use the regular expression to find the table name in the error message
	matches := re.FindStringSubmatch(err.Error())

	// Check if a match was found
	if len(matches) > 0 {
		// The second element in matches will be the table name
		return Entity(matches[2]), nil
	} else {
		return "", fmt.Errorf("failed to extract table name from error message")
	}
}
