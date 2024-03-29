package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	domerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

func TestErrorResponse(t *testing.T) {
	err := ErrorResponse{
		status: http.StatusBadRequest,
		Msg:    "id cannot be empty",
	}

	msg := "id cannot be empty"
	if got := err.Error(); !reflect.DeepEqual(got, msg) {
		t.Errorf("ErrorResponse.Error() = %v, want %v", got, msg)
	}

	status := http.StatusBadRequest
	if got := err.StatusCode(); !reflect.DeepEqual(got, status) {
		t.Errorf("ErrorResponse.StatusCode() = %v, want %v", got, status)
	}
}

func TestSetErrorResponse(t *testing.T) {
	h := Handler{
		SheetService:   sheet.New(nil),
		maxRequestSize: 10,
	}

	type args struct {
		w   http.ResponseWriter
		err error
	}
	tests := []struct {
		name           string
		args           args
		want           *ErrorResponse
		wantStatusCode int
	}{
		{
			name:           "nil error",
			args:           args{w: httptest.NewRecorder(), err: nil},
			want:           nil,
			wantStatusCode: 200,
		},
		{
			name:           "ErrIDNotFound error",
			args:           args{w: httptest.NewRecorder(), err: ErrIDNotFound},
			want:           ErrIDNotFound,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "ErrIDNotInteger error",
			args:           args{w: httptest.NewRecorder(), err: ErrIDNotInteger},
			want:           ErrIDNotInteger,
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "errors.ErrSheetDoesNotExist error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrSheetDoesNotExist},
			want:           &ErrorResponse{status: http.StatusNotFound, Msg: "no sheet found for given id"},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "errors.ErrSheetAlreadyExists error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrSheetAlreadyExists},
			want:           &ErrorResponse{status: http.StatusConflict, Msg: "a sheet with the given name already exists"},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "errors.ErrSheetNameIsEmpty error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrSheetNameIsEmpty},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "sheet name must not be empty"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "errors.ErrRoasterDoesNotExist error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrRoasterDoesNotExist},
			want:           &ErrorResponse{status: http.StatusNotFound, Msg: "no roaster found for given id"},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "errors.ErrRoasterAlreadyExists error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrRoasterAlreadyExists},
			want:           &ErrorResponse{status: http.StatusConflict, Msg: "a roaster with the given name already exists"},
			wantStatusCode: http.StatusConflict,
		},
		{
			name:           "errors.ErrRoasterNameIsEmpty error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrRoasterNameIsEmpty},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "roaster name must not be empty"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "errors.ErrBeansDoesNotExist error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrBeansDoesNotExist},
			want:           &ErrorResponse{status: http.StatusNotFound, Msg: "no beans found for given id"},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "errors.ErrShotDoesNotExist error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrShotDoesNotExist},
			want:           &ErrorResponse{status: http.StatusNotFound, Msg: "no shot found for given id"},
			wantStatusCode: http.StatusNotFound,
		},
		{
			name:           "errors.ErrShotRatingOutOfRange error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrShotRatingOutOfRange},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "shot rating is out of range. Must be between 0.0 and 10.0"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "errors.ErrBeansForeignKeyConstraint error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrBeansForeignKeyConstraint},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: fmt.Sprintf("cannot delete due to existing references: %s", domerrors.ErrBeansForeignKeyConstraint)},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "errors.ErrShotForeignKeyConstraint error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrShotForeignKeyConstraint},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: fmt.Sprintf("cannot delete due to existing references: %s", domerrors.ErrShotForeignKeyConstraint)},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "errors.ErrBeansNameIsEmpty error",
			args:           args{w: httptest.NewRecorder(), err: domerrors.ErrBeansNameIsEmpty},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "beans name must not be empty"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "json.SyntaxError error",
			args:           args{w: httptest.NewRecorder(), err: &json.SyntaxError{}},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "request body contains badly-formed json (at position 0)"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "io.ErrUnexpectedEOF error",
			args:           args{w: httptest.NewRecorder(), err: io.ErrUnexpectedEOF},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "request body contains badly-formed json"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "json.UnmarshalTypeError error",
			args:           args{w: httptest.NewRecorder(), err: &json.UnmarshalTypeError{}},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "request body contains an invalid value for the \"\" field (at position 0)"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "has prefix 'json: unknown field ' error",
			args:           args{w: httptest.NewRecorder(), err: errors.New("json: unknown field 'unknownfield'")},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "request body contains unknown field 'unknownfield'"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "io.EOF error",
			args:           args{w: httptest.NewRecorder(), err: io.EOF},
			want:           &ErrorResponse{status: http.StatusBadRequest, Msg: "request body must not be empty"},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "http: request body too large error",
			args:           args{w: httptest.NewRecorder(), err: errors.New("http: request body too large")},
			want:           &ErrorResponse{status: http.StatusRequestEntityTooLarge, Msg: fmt.Sprintf("request body must not be larger than %d bytes", h.maxRequestSize)},
			wantStatusCode: http.StatusRequestEntityTooLarge,
		},
		{
			name: "invalid time format",
			args: args{w: httptest.NewRecorder(), err: &time.ParseError{
				Layout: "2024-01-26T14:05:54Z",
			}},
			want:           &ErrorResponse{status: http.StatusRequestEntityTooLarge, Msg: "invalid time format: parsing time \"\" as \"2024-01-26T14:05:54Z\": cannot parse \"\" as \"\""},
			wantStatusCode: http.StatusBadRequest,
		},
		{
			name:           "default error",
			args:           args{w: httptest.NewRecorder(), err: errors.New("")},
			want:           &ErrorResponse{status: http.StatusInternalServerError, Msg: "internal server error"},
			wantStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.SetErrorResponse(tt.args.w, tt.args.err)

			resp := tt.args.w.(*httptest.ResponseRecorder)
			resp.Result()

			if resp.Code != tt.wantStatusCode {
				t.Errorf("SetErrorResponse() status code = %v, want %v", resp.Code, tt.wantStatusCode)
			}

			if tt.want != nil {
				if resp.Header().Get("Content-Type") != ContentTypeApplicationJSON {
					t.Errorf("SetErrorResponse() response header = %v, want %v", resp.Header().Get("Content-Type"), ContentTypeApplicationJSON)
				}

				body, err := io.ReadAll(resp.Body)
				if err != nil {
					t.Errorf("SetErrorResponse() error reading response body: %v", err)
				}

				errResp := &ErrorResponse{}
				if err := json.Unmarshal(body, errResp); err != nil {
					t.Errorf("SetErrorResponse() error unmarshalling response body: %v", err)
				}

				if errResp.Error() != tt.want.Error() {
					t.Errorf("SetErrorResponse() error message = %v, want %v", errResp.Error(), tt.want.Error())
				}
			}
		})
	}
}
