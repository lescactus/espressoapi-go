package web

import (
	"errors"
	"net/http"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
)

// webError pairs an HTTP status with a human-facing UI message. This table
// mirrors rest.domainErrorResponses but with UI-friendly wording, matched
// with errors.Is since services wrap errors with fmt.Errorf("%w", ...).
type webError struct {
	Status  int
	Message string
}

var domainErrorMessages = map[error]webError{
	domainerrors.ErrSheetDoesNotExist:  {http.StatusNotFound, "No sheet found for the given id."},
	domainerrors.ErrSheetAlreadyExists: {http.StatusConflict, "A sheet with this name already exists."},
	domainerrors.ErrSheetNameIsEmpty:   {http.StatusBadRequest, "Sheet name must not be empty."},

	domainerrors.ErrRoasterDoesNotExist:  {http.StatusNotFound, "No roaster found for the given id."},
	domainerrors.ErrRoasterAlreadyExists: {http.StatusConflict, "A roaster with this name already exists."},
	domainerrors.ErrRoasterNameIsEmpty:   {http.StatusBadRequest, "Roaster name must not be empty."},

	domainerrors.ErrBeansDoesNotExist:         {http.StatusNotFound, "No beans found for the given id."},
	domainerrors.ErrBeansAlreadyExists:        {http.StatusConflict, "Beans with this name already exist."},
	domainerrors.ErrBeansNameIsEmpty:          {http.StatusBadRequest, "Beans name must not be empty."},
	domainerrors.ErrBeansRoastLevelOutOfRange: {http.StatusBadRequest, "Roast level must be between light and dark."},
	domainerrors.ErrBeansForeignKeyConstraint: {http.StatusConflict, "This roaster is still used by beans. Delete or reassign those beans first."},

	domainerrors.ErrShotDoesNotExist:                           {http.StatusNotFound, "No shot found for the given id."},
	domainerrors.ErrShotAlreadyExists:                          {http.StatusConflict, "Shot already exists."},
	domainerrors.ErrShotRatingOutOfRange:                       {http.StatusBadRequest, "Rating must be between 0 and 10."},
	domainerrors.ErrShotComparisonWithPreviousResultOutOfRange: {http.StatusBadRequest, "Invalid comparison value."},
	domainerrors.ErrShotForeignKeyConstraint:                   {http.StatusConflict, "This sheet or beans selection is still referenced by shots. Delete those shots first."},
}

// mapDomainError resolves a service error to a UI status/message pair,
// falling back to 500 for anything unrecognized.
func mapDomainError(err error) webError {
	for domainErr, msg := range domainErrorMessages {
		if errors.Is(err, domainErr) {
			return msg
		}
	}
	return webError{Status: http.StatusInternalServerError, Message: "Something went wrong. Please try again."}
}
