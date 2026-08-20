package rest

import (
	"context"
	"errors"
	"net/http"
	"testing"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

func TestSheetHandlersHappyPaths(t *testing.T) {
	created := testSheet(1, "morning shots")
	found := testSheet(7, "dial in")
	first := testSheet(1, "first")
	second := testSheet(2, "second")
	updated := testSheet(9, "updated")
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		id        string
		status    int
		expected  any
		configure func(*testing.T, *fakeSheetService)
		handler   controllerHandler
	}{
		{
			name: "create", method: http.MethodPost, target: "/rest/v1/sheets", body: `{"name":"morning shots"}`,
			status: http.StatusCreated, expected: SheetResponse{*created}, handler: (*Handler).CreateSheet,
			configure: func(t *testing.T, service *fakeSheetService) {
				service.createSheetByName = func(_ context.Context, name string) (*sheet.Sheet, error) {
					if name != created.Name {
						t.Errorf("name = %q, want %q", name, created.Name)
					}
					return created, nil
				}
			},
		},
		{
			name: "get by id", method: http.MethodGet, target: "/rest/v1/sheets/7", id: "7",
			status: http.StatusOK, expected: SheetResponse{*found}, handler: (*Handler).GetSheetById,
			configure: func(t *testing.T, service *fakeSheetService) {
				service.getSheetByID = func(_ context.Context, id int) (*sheet.Sheet, error) {
					if id != found.Id {
						t.Errorf("id = %d, want %d", id, found.Id)
					}
					return found, nil
				}
			},
		},
		{
			name: "get all", method: http.MethodGet, target: "/rest/v1/sheets",
			status: http.StatusOK, expected: []SheetResponse{{*first}, {*second}}, handler: (*Handler).GetAllSheets,
			configure: func(_ *testing.T, service *fakeSheetService) {
				service.getAllSheets = func(context.Context) ([]sheet.Sheet, error) {
					return []sheet.Sheet{*first, *second}, nil
				}
			},
		},
		{
			name: "get all empty", method: http.MethodGet, target: "/rest/v1/sheets",
			status: http.StatusOK, expected: []SheetResponse{}, handler: (*Handler).GetAllSheets,
			configure: func(_ *testing.T, service *fakeSheetService) {
				service.getAllSheets = func(context.Context) ([]sheet.Sheet, error) {
					return []sheet.Sheet{}, nil
				}
			},
		},
		{
			name: "update", method: http.MethodPut, target: "/rest/v1/sheets/9", body: `{"name":"updated"}`, id: "9",
			status: http.StatusOK, expected: SheetResponse{*updated}, handler: (*Handler).UpdateSheetById,
			configure: func(t *testing.T, service *fakeSheetService) {
				service.updateSheetByID = func(_ context.Context, id int, value *sheet.Sheet) (*sheet.Sheet, error) {
					if id != updated.Id {
						t.Errorf("id = %d, want %d", id, updated.Id)
					}
					if value.Id != updated.Id || value.Name != updated.Name {
						t.Errorf("sheet = %#v, want id %d and name %q", value, updated.Id, updated.Name)
					}
					return updated, nil
				}
			},
		},
		{
			name: "delete", method: http.MethodDelete, target: "/rest/v1/sheets/11", id: "11",
			status: http.StatusOK, expected: ItemDeletedResponse{Id: 11, Msg: "sheet 11 deleted successfully"}, handler: (*Handler).DeleteSheetById,
			configure: func(t *testing.T, service *fakeSheetService) {
				service.deleteSheetByID = func(_ context.Context, id int) error {
					if id != 11 {
						t.Errorf("id = %d, want 11", id)
					}
					return nil
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, service, _, _, _ := newTestHandler(t)
			tt.configure(t, service)
			contentType := ""
			if tt.body != "" {
				contentType = ContentTypeApplicationJSON
			}
			req := newControllerRequest(t, tt.method, tt.target, tt.body, contentType, tt.id)

			recorder := executeControllerHandler(handler, tt.handler, req)

			assertJSONResponse(t, recorder, tt.status, tt.expected)
		})
	}
}

func TestSheetHandlersErrorPaths(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		id        string
		status    int
		message   string
		configure func(*fakeSheetService)
		handler   func(*Handler, http.ResponseWriter, *http.Request)
	}{
		{
			name: "create duplicate", method: http.MethodPost, target: "/rest/v1/sheets", body: `{"name":"duplicate"}`,
			status: http.StatusConflict, message: "a sheet with the given name already exists", handler: (*Handler).CreateSheet,
			configure: func(service *fakeSheetService) {
				service.createSheetByName = func(context.Context, string) (*sheet.Sheet, error) { return nil, domainerrors.ErrSheetAlreadyExists }
			},
		},
		{
			name: "get not found", method: http.MethodGet, target: "/rest/v1/sheets/5", id: "5",
			status: http.StatusNotFound, message: "no sheet found for given id", handler: (*Handler).GetSheetById,
			configure: func(service *fakeSheetService) {
				service.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) { return nil, domainerrors.ErrSheetDoesNotExist }
			},
		},
		{
			name: "get all service failure", method: http.MethodGet, target: "/rest/v1/sheets",
			status: http.StatusInternalServerError, message: "internal server error", handler: (*Handler).GetAllSheets,
			configure: func(service *fakeSheetService) {
				service.getAllSheets = func(context.Context) ([]sheet.Sheet, error) { return nil, errors.New("service failure") }
			},
		},
		{
			name: "update duplicate", method: http.MethodPut, target: "/rest/v1/sheets/5", body: `{"name":"duplicate"}`, id: "5",
			status: http.StatusConflict, message: "a sheet with the given name already exists", handler: (*Handler).UpdateSheetById,
			configure: func(service *fakeSheetService) {
				service.updateSheetByID = func(context.Context, int, *sheet.Sheet) (*sheet.Sheet, error) {
					return nil, domainerrors.ErrSheetAlreadyExists
				}
			},
		},
		{
			name: "delete referenced sheet", method: http.MethodDelete, target: "/rest/v1/sheets/5", id: "5",
			status: http.StatusBadRequest, message: "cannot delete due to existing references: shot foreign key constraint failed", handler: (*Handler).DeleteSheetById,
			configure: func(service *fakeSheetService) {
				service.deleteSheetByID = func(context.Context, int) error { return domainerrors.ErrShotForeignKeyConstraint }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, service, _, _, _ := newTestHandler(t)
			tt.configure(service)
			contentType := ""
			if tt.body != "" {
				contentType = ContentTypeApplicationJSON
			}
			req := newControllerRequest(t, tt.method, tt.target, tt.body, contentType, tt.id)

			recorder := executeControllerHandler(handler, tt.handler, req)

			assertJSONResponse(t, recorder, tt.status, ErrorResponse{Msg: tt.message})
		})
	}
}
