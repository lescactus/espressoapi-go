package controllers

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	tests := []struct {
		name     string
		pingErr  error
		status   int
		expected any
	}{
		{
			name: "healthy", status: http.StatusOK,
			expected: PingResponse{Ping: "pong"},
		},
		{
			name: "unhealthy", pingErr: errors.New("database unavailable"), status: http.StatusInternalServerError,
			expected: ErrorResponse{Msg: "unhealthy database"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, service, _, _, _ := newTestHandler(t)
			service.ping = func(context.Context) error { return tt.pingErr }
			req := newControllerRequest(t, http.MethodGet, "/ping", "", "", "")

			recorder := executeHandler(handler.Ping, req)

			assertJSONResponse(t, recorder, tt.status, tt.expected)
		})
	}
}
