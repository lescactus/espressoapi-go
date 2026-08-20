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
	domainerrors.ErrShotTimeOutOfRange:                         {http.StatusBadRequest, "Shot time must be between 0 and 3600 seconds."},
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

// beanErrorField resolves a bean domain error to the form field it should
// be displayed under. Returns "" for anything not tied to a specific field
// (an unexpected/internal error), meaning the caller should show the
// message in the form's general, non-field-specific error slot instead.
func beanErrorField(err error) string {
	switch {
	case errors.Is(err, domainerrors.ErrRoasterDoesNotExist):
		return "roaster_id"
	case errors.Is(err, domainerrors.ErrBeansRoastLevelOutOfRange):
		return "roast_level"
	case errors.Is(err, domainerrors.ErrBeansAlreadyExists), errors.Is(err, domainerrors.ErrBeansNameIsEmpty):
		return "name"
	default:
		return ""
	}
}

// shotErrorField resolves a shot domain error to the form field it should
// be displayed under. Returns "" for anything not tied to a specific field
// (an unexpected/internal error), meaning the caller should show the
// message in the form's general, non-field-specific error slot instead.
func shotErrorField(err error) string {
	switch {
	case errors.Is(err, domainerrors.ErrSheetDoesNotExist):
		return "sheet_id"
	case errors.Is(err, domainerrors.ErrBeansDoesNotExist):
		return "beans_id"
	case errors.Is(err, domainerrors.ErrShotComparisonWithPreviousResultOutOfRange):
		return "comparison_with_previous_result"
	case errors.Is(err, domainerrors.ErrShotRatingOutOfRange):
		return "rating"
	case errors.Is(err, domainerrors.ErrShotTimeOutOfRange):
		return "shot_time"
	default:
		return ""
	}
}

// mapDeleteError resolves a delete-time domain error to a UI status/message
// pair. domainErrorMessages' entry for ErrShotForeignKeyConstraint hedges
// between "sheet or beans" since sheets and beans share that same sentinel
// error when referenced by shots; fkMessage substitutes the resource-
// specific wording for the caller (DeleteSheet/DeleteBean) instead.
func mapDeleteError(err error, fkMessage string) webError {
	if errors.Is(err, domainerrors.ErrShotForeignKeyConstraint) {
		return webError{Status: http.StatusConflict, Message: fkMessage}
	}
	return mapDomainError(err)
}
