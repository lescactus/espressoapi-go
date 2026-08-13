package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type controllerHandler func(*Handler, http.ResponseWriter, *http.Request)

func executeControllerHandler(handler *Handler, endpoint controllerHandler, req *http.Request) *httptest.ResponseRecorder {
	return executeHandler(func(w http.ResponseWriter, r *http.Request) {
		endpoint(handler, w, r)
	}, req)
}

func TestCreateAndUpdateHandlersRejectInvalidRequests(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		target    string
		id        string
		handler   controllerHandler
		validBody string
	}{
		{name: "create sheet", method: http.MethodPost, target: "/rest/v1/sheets", handler: (*Handler).CreateSheet, validBody: `{"name":"sheet"}`},
		{name: "update sheet", method: http.MethodPut, target: "/rest/v1/sheets/1", id: "1", handler: (*Handler).UpdateSheetById, validBody: `{"name":"sheet"}`},
		{name: "create roaster", method: http.MethodPost, target: "/rest/v1/roasters", handler: (*Handler).CreateRoaster, validBody: `{"name":"roaster"}`},
		{name: "update roaster", method: http.MethodPut, target: "/rest/v1/roasters/1", id: "1", handler: (*Handler).UpdateRoasterById, validBody: `{"name":"roaster"}`},
		{name: "create beans", method: http.MethodPost, target: "/rest/v1/beans", handler: (*Handler).CreateBeans, validBody: `{"name":"beans","roaster_id":1,"roast_date":"2026-01-01","roast_level":2}`},
		{name: "update beans", method: http.MethodPut, target: "/rest/v1/beans/1", id: "1", handler: (*Handler).UpdateBeanById, validBody: `{"name":"beans","roaster_id":1,"roast_date":"2026-01-01","roast_level":2}`},
		{name: "create shot", method: http.MethodPost, target: "/rest/v1/shots", handler: (*Handler).CreateShot, validBody: `{}`},
		{name: "update shot", method: http.MethodPut, target: "/rest/v1/shots/1", id: "1", handler: (*Handler).UpdateShotById, validBody: `{}`},
	}

	invalidRequests := []struct {
		name        string
		body        func(string) string
		contentType string
		status      int
		message     string
	}{
		{
			name:    "missing content type",
			body:    func(validBody string) string { return validBody },
			status:  http.StatusUnsupportedMediaType,
			message: "Content-Type header is not application/json",
		},
		{
			name:        "malformed JSON",
			body:        func(string) string { return "{" },
			contentType: ContentTypeApplicationJSON,
			status:      http.StatusBadRequest,
			message:     "request body contains badly-formed json",
		},
	}

	for _, tt := range tests {
		for _, invalid := range invalidRequests {
			t.Run(tt.name+"/"+invalid.name, func(t *testing.T) {
				handler, _, _, _, _ := newTestHandler(t)
				req := newControllerRequest(t, tt.method, tt.target, invalid.body(tt.validBody), invalid.contentType, tt.id)

				recorder := executeControllerHandler(handler, tt.handler, req)

				assertJSONResponse(t, recorder, invalid.status, ErrorResponse{Msg: invalid.message})
			})
		}
	}
}

func TestCreateSheetRejectsInvalidBodies(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		contentType string
		message     string
	}{
		{name: "wrong content type", body: `{"name":"sheet"}`, contentType: "text/plain", message: "Content-Type header is not application/json"},
		{name: "empty body", contentType: ContentTypeApplicationJSON, message: "request body must not be empty"},
		{name: "invalid field type", body: `{"name": 1}`, contentType: ContentTypeApplicationJSON, message: `request body contains an invalid value for the "name" field (at position 10)`},
		{name: "unknown field", body: `{"unknown":"value"}`, contentType: ContentTypeApplicationJSON, message: `request body contains unknown field "unknown"`},
		{name: "multiple objects", body: `{"name":"sheet"} {"name":"second"}`, contentType: ContentTypeApplicationJSON, message: "request body must only contain a single json object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _, _, _ := newTestHandler(t)
			req := newControllerRequest(t, http.MethodPost, "/rest/v1/sheets", tt.body, tt.contentType, "")

			recorder := executeHandler(handler.CreateSheet, req)

			status := http.StatusBadRequest
			if tt.contentType == "text/plain" {
				status = http.StatusUnsupportedMediaType
			}
			assertJSONResponse(t, recorder, status, ErrorResponse{Msg: tt.message})
		})
	}
}

func TestCreateSheetRejectsOversizedBody(t *testing.T) {
	handler, _, _, _, _ := newTestHandler(t)
	body := `{"name":"` + strings.Repeat("a", 64) + `"}`
	req := newControllerRequest(t, http.MethodPost, "/rest/v1/sheets", body, ContentTypeApplicationJSON, "")

	recorder := executeHandler(handler.MaxReqSize()(http.HandlerFunc(handler.CreateSheet)).ServeHTTP, req)

	assertJSONResponse(t, recorder, http.StatusRequestEntityTooLarge, ErrorResponse{Msg: "request body must not be larger than 64 bytes"})
}

func TestBeanHandlersRejectInvalidRoastDate(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		target  string
		id      string
		handler controllerHandler
	}{
		{name: "create", method: http.MethodPost, target: "/rest/v1/beans", handler: (*Handler).CreateBeans},
		{name: "update", method: http.MethodPut, target: "/rest/v1/beans/1", id: "1", handler: (*Handler).UpdateBeanById},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _, _, _ := newTestHandler(t)
			body := `{"name":"beans","roaster_id":1,"roast_date":"not-a-date","roast_level":2}`
			req := newControllerRequest(t, tt.method, tt.target, body, ContentTypeApplicationJSON, tt.id)

			recorder := executeControllerHandler(handler, tt.handler, req)

			assertJSONResponse(t, recorder, http.StatusBadRequest, ErrorResponse{Msg: `invalid time format: parsing time "not-a-date" as "2006-01-02": cannot parse "not-a-date" as "2006"`})
		})
	}
}

func TestIDHandlersRejectInvalidIDs(t *testing.T) {
	tests := []struct {
		name    string
		method  string
		target  string
		body    string
		handler controllerHandler
	}{
		{name: "get sheet", method: http.MethodGet, target: "/rest/v1/sheets/:id", handler: (*Handler).GetSheetById},
		{name: "update sheet", method: http.MethodPut, target: "/rest/v1/sheets/:id", body: `{"name":"sheet"}`, handler: (*Handler).UpdateSheetById},
		{name: "delete sheet", method: http.MethodDelete, target: "/rest/v1/sheets/:id", handler: (*Handler).DeleteSheetById},
		{name: "get roaster", method: http.MethodGet, target: "/rest/v1/roasters/:id", handler: (*Handler).GetRoasterById},
		{name: "update roaster", method: http.MethodPut, target: "/rest/v1/roasters/:id", body: `{"name":"roaster"}`, handler: (*Handler).UpdateRoasterById},
		{name: "delete roaster", method: http.MethodDelete, target: "/rest/v1/roasters/:id", handler: (*Handler).DeleteRoasterById},
		{name: "get beans", method: http.MethodGet, target: "/rest/v1/beans/:id", handler: (*Handler).GetBeansById},
		{name: "update beans", method: http.MethodPut, target: "/rest/v1/beans/:id", body: `{"name":"beans","roaster_id":1,"roast_date":"2026-01-01","roast_level":2}`, handler: (*Handler).UpdateBeanById},
		{name: "delete beans", method: http.MethodDelete, target: "/rest/v1/beans/:id", handler: (*Handler).DeleteBeansById},
		{name: "get shot", method: http.MethodGet, target: "/rest/v1/shots/:id", handler: (*Handler).GetShotById},
		{name: "update shot", method: http.MethodPut, target: "/rest/v1/shots/:id", body: `{}`, handler: (*Handler).UpdateShotById},
		{name: "delete shot", method: http.MethodDelete, target: "/rest/v1/shots/:id", handler: (*Handler).DeleteShotById},
	}

	invalidIDs := []struct {
		name    string
		id      string
		message string
	}{
		{name: "missing", message: "id cannot be empty"},
		{name: "not an integer", id: "abc", message: "id must be an integer"},
	}

	for _, tt := range tests {
		for _, invalid := range invalidIDs {
			t.Run(tt.name+"/"+invalid.name, func(t *testing.T) {
				handler, _, _, _, _ := newTestHandler(t)
				contentType := ""
				if tt.body != "" {
					contentType = ContentTypeApplicationJSON
				}
				req := newControllerRequest(t, tt.method, tt.target, tt.body, contentType, invalid.id)

				recorder := executeControllerHandler(handler, tt.handler, req)

				assertJSONResponse(t, recorder, http.StatusBadRequest, ErrorResponse{Msg: invalid.message})
			})
		}
	}
}
