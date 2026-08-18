package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
)

var (
	ErrIDNotFound   = NewErrorResponse(http.StatusBadRequest, "id cannot be empty")
	ErrIDNotInteger = NewErrorResponse(http.StatusBadRequest, "id must be an integer")
)

// ErrorResponse represents the json response
// for http errors.
// It contains a message describing the error
//
// swagger:response ErrorResponse
type ErrorResponse struct {
	status int
	Msg    string `json:"msg"`
}

// Error method for ErrorResponse
func (e *ErrorResponse) Error() string {
	return e.Msg
}

// StatusCode method for ErrorResponse
func (e *ErrorResponse) StatusCode() int {
	return e.status
}

func NewErrorResponse(status int, msg string) *ErrorResponse {
	return &ErrorResponse{
		status: status,
		Msg:    msg,
	}
}

var domainErrorResponses = map[error]ErrorResponse{
	// Catch if the sheet does not exist
	domainerrors.ErrSheetDoesNotExist: {status: http.StatusNotFound, Msg: "no sheet found for given id"},
	// Catch if the sheet already exists
	domainerrors.ErrSheetAlreadyExists: {status: http.StatusConflict, Msg: "a sheet with the given name already exists"},
	// Catch if the sheet name is empty
	domainerrors.ErrSheetNameIsEmpty: {status: http.StatusBadRequest, Msg: "sheet name must not be empty"},
	// Catch if the roaster does not exist
	domainerrors.ErrRoasterDoesNotExist: {status: http.StatusNotFound, Msg: "no roaster found for given id"},
	// Catch if the roaster already exists
	domainerrors.ErrRoasterAlreadyExists: {status: http.StatusConflict, Msg: "a roaster with the given name already exists"},
	// Catch if the roaster name is empty
	domainerrors.ErrRoasterNameIsEmpty: {status: http.StatusBadRequest, Msg: "roaster name must not be empty"},
	// Catch if the beans does not exist
	domainerrors.ErrBeansDoesNotExist: {status: http.StatusNotFound, Msg: "no beans found for given id"},
	// Catch if beans already exist
	domainerrors.ErrBeansAlreadyExists: {status: http.StatusConflict, Msg: "beans already exist"},
	// Catch if the shot does not exist
	domainerrors.ErrShotDoesNotExist: {status: http.StatusNotFound, Msg: "no shot found for given id"},
	// Catch if a shot already exists
	domainerrors.ErrShotAlreadyExists: {status: http.StatusConflict, Msg: "shot already exists"},
	// Catch if the shot rating is out of range
	domainerrors.ErrShotRatingOutOfRange: {status: http.StatusBadRequest, Msg: "shot rating is out of range. Must be between 0.0 and 10.0"},
	// Catch if the shot comparison with previous result is out of range
	domainerrors.ErrShotComparisonWithPreviousResultOutOfRange: {status: http.StatusBadRequest, Msg: "shot comparison with previous result is out of range. Must be between 0 and 3"},
	// Catch if the beans roast level is out of range
	domainerrors.ErrBeansRoastLevelOutOfRange: {status: http.StatusBadRequest, Msg: "beans roast level is out of range. Must be between 0 and 4"},
	// Catch if the beans foreign key constraint failed
	domainerrors.ErrBeansForeignKeyConstraint: {status: http.StatusBadRequest, Msg: "cannot delete due to existing references: beans foreign key constraint failed"},
	// Catch if the shot foreign key constraint failed
	domainerrors.ErrShotForeignKeyConstraint: {status: http.StatusBadRequest, Msg: "cannot delete due to existing references: shot foreign key constraint failed"},
	// Catch if the beans name is empty
	domainerrors.ErrBeansNameIsEmpty: {status: http.StatusBadRequest, Msg: "beans name must not be empty"},
}

// SetErrorResponse will attempt to parse the given error
// and set the response status code and using the ResponseWriter
// according to the type of the error.
func (h *Handler) SetErrorResponse(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	w.Header().Set("Content-Type", ContentTypeApplicationJSON)

	var errResp *ErrorResponse

	if resp, ok := err.(*ErrorResponse); ok {
		errResp = resp
	} else {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var maxBytesError *http.MaxBytesError
		var timeParseError *time.ParseError

		for domainErr, response := range domainErrorResponses {
			if errors.Is(err, domainErr) {
				errResp = &response
				break
			}
		}

		if errResp == nil {
			switch {

			// Catch any syntax errors
			case errors.As(err, &syntaxError):
				msg := fmt.Sprintf("request body contains badly-formed json (at position %d)", syntaxError.Offset)
				errResp = &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

			// In some circumstances Decode() may also return an
			// io.ErrUnexpectedEOF error for syntax errors in the JSON
			case errors.Is(err, io.ErrUnexpectedEOF):
				msg := "request body contains badly-formed json"
				errResp = &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

			// Catch any type errors
			case errors.As(err, &unmarshalTypeError):
				msg := fmt.Sprintf("request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
				errResp = &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

			// Catch the error caused by extra unexpected fields in the request
			// body. We extract the field name from the error message and
			// interpolate it in our custom error message. There is an open
			// issue at https://github.com/golang/go/issues/29035 regarding
			// turning this into a sentinel error.
			case strings.HasPrefix(err.Error(), "json: unknown field "):
				fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
				msg := fmt.Sprintf("request body contains unknown field %s", fieldName)
				errResp = &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

			// An io.EOF error is returned by Decode() if the request body is
			// empty.
			case errors.Is(err, io.EOF):
				msg := "request body must not be empty"
				errResp = &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

			// Catch the error caused by the request body being too large
			case errors.As(err, &maxBytesError):
				msg := fmt.Sprintf("request body must not be larger than %d bytes", h.maxRequestSize)
				errResp = &ErrorResponse{status: http.StatusRequestEntityTooLarge, Msg: msg}

			// Catch if the error is due to a time parsing error
			case errors.As(err, &timeParseError):
				msg := fmt.Sprintf("invalid time format: %s", timeParseError)
				errResp = &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

			default:
				errResp = &ErrorResponse{status: http.StatusInternalServerError, Msg: "internal server error"}
			}
		}
	}

	w.WriteHeader(errResp.status)

	resp, _ := json.Marshal(errResp)
	w.Write(resp)
}
