package controllers

import (
	"context"
	"errors"
	"net/http"
	"testing"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
)

func TestRoasterHandlersHappyPaths(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		handler, _, service, _, _ := newTestHandler(t)
		created := testRoaster(1, "test roaster")
		service.createRoasterByName = func(_ context.Context, name string) (*roaster.Roaster, error) {
			if name != created.Name {
				t.Errorf("name = %q, want %q", name, created.Name)
			}
			return created, nil
		}
		req := newControllerRequest(t, http.MethodPost, "/rest/v1/roasters", `{"name":"test roaster"}`, ContentTypeApplicationJSON, "")

		recorder := executeHandler(handler.CreateRoaster, req)

		assertJSONResponse(t, recorder, http.StatusCreated, RoasterResponse{*created})
	})

	t.Run("get by id", func(t *testing.T) {
		handler, _, service, _, _ := newTestHandler(t)
		found := testRoaster(7, "found")
		service.getRoasterByID = func(_ context.Context, id int) (*roaster.Roaster, error) {
			if id != found.Id {
				t.Errorf("id = %d, want %d", id, found.Id)
			}
			return found, nil
		}
		req := newControllerRequest(t, http.MethodGet, "/rest/v1/roasters/7", "", "", "7")

		recorder := executeHandler(handler.GetRoasterById, req)

		assertJSONResponse(t, recorder, http.StatusOK, RoasterResponse{*found})
	})

	t.Run("get all", func(t *testing.T) {
		handler, _, service, _, _ := newTestHandler(t)
		first := testRoaster(1, "first")
		second := testRoaster(2, "second")
		service.getAllRoasters = func(context.Context) ([]roaster.Roaster, error) {
			return []roaster.Roaster{*first, *second}, nil
		}
		req := newControllerRequest(t, http.MethodGet, "/rest/v1/roasters", "", "", "")

		recorder := executeHandler(handler.GetAllRoasters, req)

		assertJSONResponse(t, recorder, http.StatusOK, []RoasterResponse{{*first}, {*second}})
	})

	t.Run("update", func(t *testing.T) {
		handler, _, service, _, _ := newTestHandler(t)
		updated := testRoaster(9, "updated")
		service.updateRoasterByID = func(_ context.Context, id int, value *roaster.Roaster) (*roaster.Roaster, error) {
			if id != updated.Id {
				t.Errorf("id = %d, want %d", id, updated.Id)
			}
			if value.Id != updated.Id || value.Name != updated.Name {
				t.Errorf("roaster = %#v, want id %d and name %q", value, updated.Id, updated.Name)
			}
			return updated, nil
		}
		req := newControllerRequest(t, http.MethodPut, "/rest/v1/roasters/9", `{"name":"updated"}`, ContentTypeApplicationJSON, "9")

		recorder := executeHandler(handler.UpdateRoasterById, req)

		assertJSONResponse(t, recorder, http.StatusOK, RoasterResponse{*updated})
	})

	t.Run("delete", func(t *testing.T) {
		handler, _, service, _, _ := newTestHandler(t)
		service.deleteRoasterByID = func(_ context.Context, id int) error {
			if id != 11 {
				t.Errorf("id = %d, want 11", id)
			}
			return nil
		}
		req := newControllerRequest(t, http.MethodDelete, "/rest/v1/roasters/11", "", "", "11")

		recorder := executeHandler(handler.DeleteRoasterById, req)

		assertJSONResponse(t, recorder, http.StatusOK, ItemDeletedResponse{Id: 11, Msg: "roaster 11 deleted successfully"})
	})
}

func TestRoasterHandlersErrorPaths(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		id        string
		status    int
		message   string
		configure func(*fakeRoasterService)
		handler   controllerHandler
	}{
		{
			name: "create duplicate", method: http.MethodPost, target: "/rest/v1/roasters", body: `{"name":"duplicate"}`,
			status: http.StatusConflict, message: "a roaster with the given name already exists", handler: (*Handler).CreateRoaster,
			configure: func(service *fakeRoasterService) {
				service.createRoasterByName = func(context.Context, string) (*roaster.Roaster, error) {
					return nil, domainerrors.ErrRoasterAlreadyExists
				}
			},
		},
		{
			name: "get not found", method: http.MethodGet, target: "/rest/v1/roasters/5", id: "5",
			status: http.StatusNotFound, message: "no roaster found for given id", handler: (*Handler).GetRoasterById,
			configure: func(service *fakeRoasterService) {
				service.getRoasterByID = func(context.Context, int) (*roaster.Roaster, error) { return nil, domainerrors.ErrRoasterDoesNotExist }
			},
		},
		{
			name: "get all service failure", method: http.MethodGet, target: "/rest/v1/roasters",
			status: http.StatusInternalServerError, message: "internal server error", handler: (*Handler).GetAllRoasters,
			configure: func(service *fakeRoasterService) {
				service.getAllRoasters = func(context.Context) ([]roaster.Roaster, error) { return nil, errors.New("service failure") }
			},
		},
		{
			name: "update duplicate", method: http.MethodPut, target: "/rest/v1/roasters/5", body: `{"name":"duplicate"}`, id: "5",
			status: http.StatusConflict, message: "a roaster with the given name already exists", handler: (*Handler).UpdateRoasterById,
			configure: func(service *fakeRoasterService) {
				service.updateRoasterByID = func(context.Context, int, *roaster.Roaster) (*roaster.Roaster, error) {
					return nil, domainerrors.ErrRoasterAlreadyExists
				}
			},
		},
		{
			name: "delete referenced roaster", method: http.MethodDelete, target: "/rest/v1/roasters/5", id: "5",
			status: http.StatusBadRequest, message: "cannot delete due to existing references: beans foreign key constraint failed", handler: (*Handler).DeleteRoasterById,
			configure: func(service *fakeRoasterService) {
				service.deleteRoasterByID = func(context.Context, int) error { return domainerrors.ErrBeansForeignKeyConstraint }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, service, _, _ := newTestHandler(t)
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
