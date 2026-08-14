package controllers

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	modelsql "github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
)

func TestBeanHandlersHappyPaths(t *testing.T) {
	expectedRoastDate := time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC)
	created := testBean(1, "espresso blend")
	createdWithoutRoastDate := testBean(2, "espresso blend without roast date")
	createdWithoutRoastDate.RoastDate = nil
	found := testBean(7, "found beans")
	first := testBean(1, "first")
	second := testBean(2, "second")
	updated := testBean(9, "updated beans")
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		id        string
		status    int
		expected  any
		configure func(*testing.T, *fakeBeanService)
		handler   controllerHandler
	}{
		{
			name: "create", method: http.MethodPost, target: "/rest/v1/beans", body: `{"name":"espresso blend","roaster_id":6,"roast_date":"2026-02-18","roast_level":2}`,
			status: http.StatusCreated, expected: BeansResponse{*created}, handler: (*Handler).CreateBeans,
			configure: func(t *testing.T, service *fakeBeanService) {
				service.createBean = func(_ context.Context, value *bean.Bean) (*bean.Bean, error) {
					assertBeanRequest(t, value, 0, "espresso blend", 6, &expectedRoastDate, modelsql.RoastLevelMedium)
					return created, nil
				}
			},
		},
		{
			name: "create without roast date", method: http.MethodPost, target: "/rest/v1/beans", body: `{"name":"espresso blend without roast date","roaster_id":6,"roast_level":2}`,
			status: http.StatusCreated, expected: BeansResponse{*createdWithoutRoastDate}, handler: (*Handler).CreateBeans,
			configure: func(t *testing.T, service *fakeBeanService) {
				service.createBean = func(_ context.Context, value *bean.Bean) (*bean.Bean, error) {
					assertBeanRequest(t, value, 0, "espresso blend without roast date", 6, nil, modelsql.RoastLevelMedium)
					return createdWithoutRoastDate, nil
				}
			},
		},
		{
			name: "create with null roast date", method: http.MethodPost, target: "/rest/v1/beans", body: `{"name":"espresso blend without roast date","roaster_id":6,"roast_date":null,"roast_level":2}`,
			status: http.StatusCreated, expected: BeansResponse{*createdWithoutRoastDate}, handler: (*Handler).CreateBeans,
			configure: func(t *testing.T, service *fakeBeanService) {
				service.createBean = func(_ context.Context, value *bean.Bean) (*bean.Bean, error) {
					assertBeanRequest(t, value, 0, "espresso blend without roast date", 6, nil, modelsql.RoastLevelMedium)
					return createdWithoutRoastDate, nil
				}
			},
		},
		{
			name: "get by id", method: http.MethodGet, target: "/rest/v1/beans/7", id: "7",
			status: http.StatusOK, expected: BeansResponse{*found}, handler: (*Handler).GetBeansById,
			configure: func(t *testing.T, service *fakeBeanService) {
				service.getBeanByID = func(_ context.Context, id int) (*bean.Bean, error) {
					if id != found.Id {
						t.Errorf("id = %d, want %d", id, found.Id)
					}
					return found, nil
				}
			},
		},
		{
			name: "get all", method: http.MethodGet, target: "/rest/v1/beans",
			status: http.StatusOK, expected: []BeansResponse{{*first}, {*second}}, handler: (*Handler).GetAllBeans,
			configure: func(_ *testing.T, service *fakeBeanService) {
				service.getAllBeans = func(context.Context) ([]bean.Bean, error) {
					return []bean.Bean{*first, *second}, nil
				}
			},
		},
		{
			name: "update", method: http.MethodPut, target: "/rest/v1/beans/9", body: `{"name":"updated beans","roaster_id":6,"roast_date":"2026-02-18","roast_level":2}`, id: "9",
			status: http.StatusOK, expected: BeansResponse{*updated}, handler: (*Handler).UpdateBeanById,
			configure: func(t *testing.T, service *fakeBeanService) {
				service.updateBeanByID = func(_ context.Context, id int, value *bean.Bean) (*bean.Bean, error) {
					if id != 9 {
						t.Errorf("id = %d, want 9", id)
					}
					assertBeanRequest(t, value, 9, "updated beans", 6, &expectedRoastDate, modelsql.RoastLevelMedium)
					return updated, nil
				}
			},
		},
		{
			name: "delete", method: http.MethodDelete, target: "/rest/v1/beans/11", id: "11",
			status: http.StatusOK, expected: ItemDeletedResponse{Id: 11, Msg: "beans 11 deleted successfully"}, handler: (*Handler).DeleteBeansById,
			configure: func(t *testing.T, service *fakeBeanService) {
				service.deleteBeanByID = func(_ context.Context, id int) error {
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
			handler, _, _, service, _ := newTestHandler(t)
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

func TestBeanHandlersErrorPaths(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		id        string
		status    int
		message   string
		configure func(*fakeBeanService)
		handler   controllerHandler
	}{
		{
			name: "create with missing roaster", method: http.MethodPost, target: "/rest/v1/beans", body: `{"name":"beans","roaster_id":99,"roast_date":"2026-02-18","roast_level":2}`,
			status: http.StatusNotFound, message: "no roaster found for given id", handler: (*Handler).CreateBeans,
			configure: func(service *fakeBeanService) {
				service.createBean = func(_ context.Context, value *bean.Bean) (*bean.Bean, error) {
					if value.Roaster == nil || value.Roaster.Id != 99 {
						t.Errorf("bean roaster = %#v, want id 99", value.Roaster)
					}
					return nil, domainerrors.ErrRoasterDoesNotExist
				}
			},
		},
		{
			name: "get not found", method: http.MethodGet, target: "/rest/v1/beans/5", id: "5",
			status: http.StatusNotFound, message: "no beans found for given id", handler: (*Handler).GetBeansById,
			configure: func(service *fakeBeanService) {
				service.getBeanByID = func(context.Context, int) (*bean.Bean, error) { return nil, domainerrors.ErrBeansDoesNotExist }
			},
		},
		{
			name: "get all service failure", method: http.MethodGet, target: "/rest/v1/beans",
			status: http.StatusInternalServerError, message: "internal server error", handler: (*Handler).GetAllBeans,
			configure: func(service *fakeBeanService) {
				service.getAllBeans = func(context.Context) ([]bean.Bean, error) { return nil, errors.New("service failure") }
			},
		},
		{
			name: "update with empty name", method: http.MethodPut, target: "/rest/v1/beans/5", body: `{"name":"","roaster_id":1,"roast_date":"2026-02-18","roast_level":2}`, id: "5",
			status: http.StatusBadRequest, message: "beans name must not be empty", handler: (*Handler).UpdateBeanById,
			configure: func(service *fakeBeanService) {
				service.updateBeanByID = func(_ context.Context, _ int, value *bean.Bean) (*bean.Bean, error) {
					if value.Name != "" {
						t.Errorf("bean name = %q, want empty", value.Name)
					}
					return nil, domainerrors.ErrBeansNameIsEmpty
				}
			},
		},
		{
			name: "delete referenced beans", method: http.MethodDelete, target: "/rest/v1/beans/5", id: "5",
			status: http.StatusBadRequest, message: "cannot delete due to existing references: shot foreign key constraint failed", handler: (*Handler).DeleteBeansById,
			configure: func(service *fakeBeanService) {
				service.deleteBeanByID = func(context.Context, int) error { return domainerrors.ErrShotForeignKeyConstraint }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _, service, _ := newTestHandler(t)
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

func assertBeanRequest(t *testing.T, value *bean.Bean, id int, name string, roasterID int, roastDate *time.Time, roastLevel modelsql.RoastLevel) {
	t.Helper()
	if value.Id != id {
		t.Errorf("bean id = %d, want %d", value.Id, id)
	}
	if value.Name != name {
		t.Errorf("bean name = %q, want %q", value.Name, name)
	}
	if value.Roaster == nil || value.Roaster.Id != roasterID {
		t.Errorf("bean roaster = %#v, want id %d", value.Roaster, roasterID)
	}
	if (value.RoastDate == nil) != (roastDate == nil) || value.RoastDate != nil && !value.RoastDate.Equal(*roastDate) {
		t.Errorf("bean roast date = %v, want %v", value.RoastDate, roastDate)
	}
	if value.RoastLevel != roastLevel {
		t.Errorf("bean roast level = %v, want %v", value.RoastLevel, roastLevel)
	}
}
