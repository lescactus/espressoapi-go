package controllers

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestPing(t *testing.T) {
	t.Run("healthy", func(t *testing.T) {
		handler, service, _, _, _ := newTestHandler(t)
		service.ping = func(context.Context) error { return nil }
		req := newControllerRequest(t, http.MethodGet, "/ping", "", "", "")

		recorder := executeHandler(handler.Ping, req)

		assertJSONResponse(t, recorder, http.StatusOK, PingResponse{Ping: "pong"})
	})

	t.Run("unhealthy", func(t *testing.T) {
		handler, service, _, _, _ := newTestHandler(t)
		service.ping = func(context.Context) error { return errors.New("database unavailable") }
		req := newControllerRequest(t, http.MethodGet, "/ping", "", "", "")

		recorder := executeHandler(handler.Ping, req)

		assertJSONResponse(t, recorder, http.StatusInternalServerError, ErrorResponse{Msg: "unhealthy database"})
	})
}
