package controllers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func jsonDecodeBody(w http.ResponseWriter, r *http.Request, dst interface{}) error {
	// If the "Content-Type" header is present, check that it has the value "application/json".
	// Parse and normalize the header to remove any additional parameters by stripping
	// whitespace and converting to lowercase before we checking the value
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			return &ErrorResponse{status: http.StatusUnsupportedMediaType, Msg: "Content-Type header is not application/json"}
		}
	} else {
		return &ErrorResponse{status: http.StatusUnsupportedMediaType, Msg: "Content-Type header is not application/json"}
	}

	// Use http.MaxBytesReader to enforce a maximum read of 1MB from the
	// response body. A request body larger than that will now result in
	// Decode() returning a "http: request body too large" error
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	// Setup the decoder and call the DisallowUnknownFields() method on it.
	// This will cause Decode() to return a "json: unknown field ..." error
	// if it encounters any extra unexpected fields in the JSON. Strictly
	// speaking, it returns an error for "keys which do not match any
	// non-ignored, exported fields in the destination"
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(&dst)
	if err != nil {
		return err
		// var syntaxError *json.SyntaxError
		// var unmarshalTypeError *json.UnmarshalTypeError

		// switch {
		// // Catch any syntax errors
		// case errors.As(err, &syntaxError):
		// 	msg := fmt.Sprintf("request body contains badly-formed json (at position %d)", syntaxError.Offset)
		// 	return &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

		// // In some circumstances Decode() may also return an
		// // io.ErrUnexpectedEOF error for syntax errors in the JSON
		// case errors.Is(err, io.ErrUnexpectedEOF):
		// 	msg := "request body contains badly-formed JSON"
		// 	return &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

		// // Catch any type errors
		// case errors.As(err, &unmarshalTypeError):
		// 	msg := fmt.Sprintf("request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset)
		// 	return &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

		// // Catch the error caused by extra unexpected fields in the request
		// // body. We extract the field name from the error message and
		// // interpolate it in our custom error message. There is an open
		// // issue at https://github.com/golang/go/issues/29035 regarding
		// // turning this into a sentinel error.
		// case strings.HasPrefix(err.Error(), "json: unknown field "):
		// 	fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
		// 	msg := fmt.Sprintf("request body contains unknown field %s", fieldName)
		// 	return &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

		// // An io.EOF error is returned by Decode() if the request body is
		// // empty.
		// case errors.Is(err, io.EOF):
		// 	msg := "request body must not be empty"
		// 	return &ErrorResponse{status: http.StatusBadRequest, Msg: msg}

		// // Catch the error caused by the request body being too large
		// case err.Error() == "http: request body too large":
		// 	msg := "request body must not be larger than 1MB"
		// 	return &ErrorResponse{status: http.StatusRequestEntityTooLarge, Msg: msg}

		// default:
		// 	return &ErrorResponse{status: http.StatusInternalServerError, Msg: "internal server error: could not decode json input"}
		// }
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	err = dec.Decode(&struct{}{})
	if !errors.Is(err, io.EOF) {
		msg := "request body must only contain a single JSON object"
		return &ErrorResponse{status: http.StatusBadRequest, Msg: msg}
	}

	return nil
}
