package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	domerrors "github.com/lescactus/espressoapi-go/internal/errors"
)

var (
	ErrIDNotFound   = NewErrorResponse(http.StatusBadRequest, "id cannot be empty")
	ErrIDNotInteger = NewErrorResponse(http.StatusBadRequest, "id must be an integer")
)

// ErrorResponse represents the json response
// for http errors
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

		switch {
		// Catch if the sheet does not exist
		case errors.Is(err, domerrors.ErrSheetDoesNotExist):
			errResp = &ErrorResponse{status: http.StatusNotFound, Msg: "no sheet found for given id"}

		// Catch if the sheet already exists
		case errors.Is(err, domerrors.ErrSheetAlreadyExists):
			errResp = &ErrorResponse{status: http.StatusConflict, Msg: "a sheet with the given name already exists"}

		// Catch if the roaster does not exist
		case errors.Is(err, domerrors.ErrRoasterDoesNotExist):
			errResp = &ErrorResponse{status: http.StatusNotFound, Msg: "no roaster found for given id"}

		// Catch if the roaster already exists
		case errors.Is(err, domerrors.ErrRoasterAlreadyExists):
			errResp = &ErrorResponse{status: http.StatusConflict, Msg: "a roaster with the given name already exists"}

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
		case err.Error() == "http: request body too large":
			msg := fmt.Sprintf("request body must not be larger than %d bytes", h.maxRequestSize)
			errResp = &ErrorResponse{status: http.StatusRequestEntityTooLarge, Msg: msg}

		default:
			errResp = &ErrorResponse{status: http.StatusInternalServerError, Msg: "internal server error"}
		}
	}

	w.WriteHeader(errResp.status)

	resp, _ := json.Marshal(errResp)
	w.Write(resp)
}
