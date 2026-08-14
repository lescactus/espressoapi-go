package controllers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type failingJSONMarshaler struct{}

func (failingJSONMarshaler) MarshalJSON() ([]byte, error) {
	return nil, errors.New("marshal failed")
}

func TestWriteJSONResponse(t *testing.T) {
	tests := []struct {
		name           string
		status         int
		value          any
		expectedStatus int
		expected       any
	}{
		{
			name:           "writes marshaled response",
			status:         http.StatusCreated,
			value:          map[string]string{"status": "created"},
			expectedStatus: http.StatusCreated,
			expected:       map[string]string{"status": "created"},
		},
		{
			name:           "returns internal server error when marshaling fails",
			status:         http.StatusCreated,
			value:          failingJSONMarshaler{},
			expectedStatus: http.StatusInternalServerError,
			expected:       ErrorResponse{Msg: "internal server error"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := &Handler{}

			handler.writeJSONResponse(recorder, tt.status, tt.value)

			assertJSONResponse(t, recorder, tt.expectedStatus, tt.expected)
		})
	}
}
