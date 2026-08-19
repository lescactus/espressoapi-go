package rest

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	modelsql "github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
)

const validShotRequestBody = `{
	"sheet_id":3,
	"beans_id":4,
	"grind_setting":12,
	"quantity_in":18.5,
	"quantity_out":37,
	"shot_time":28000000000,
	"water_temperature":93.5,
	"rating":8.5,
	"is_too_bitter":false,
	"is_too_sour":true,
	"comparison_with_previous_result":2,
	"additional_notes":"test notes"
}`

func TestShotHandlersHappyPaths(t *testing.T) {
	created := testShot(1)
	found := testShot(7)
	first := testShot(1)
	second := testShot(2)
	updated := testShot(9)
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		id        string
		status    int
		expected  any
		configure func(*testing.T, *fakeShotService)
		handler   controllerHandler
	}{
		{
			name: "create", method: http.MethodPost, target: "/rest/v1/shots", body: validShotRequestBody,
			status: http.StatusCreated, expected: ShotResponse{*created}, handler: (*Handler).CreateShot,
			configure: func(t *testing.T, service *fakeShotService) {
				service.createShot = func(_ context.Context, value *shot.Shot) (*shot.Shot, error) {
					assertShotRequest(t, value, 0)
					return created, nil
				}
			},
		},
		{
			name: "get by id", method: http.MethodGet, target: "/rest/v1/shots/7", id: "7",
			status: http.StatusOK, expected: ShotResponse{*found}, handler: (*Handler).GetShotById,
			configure: func(t *testing.T, service *fakeShotService) {
				service.getShotByID = func(_ context.Context, id int) (*shot.Shot, error) {
					if id != found.Id {
						t.Errorf("id = %d, want %d", id, found.Id)
					}
					return found, nil
				}
			},
		},
		{
			name: "get all", method: http.MethodGet, target: "/rest/v1/shots",
			status: http.StatusOK, expected: []ShotResponse{{*first}, {*second}}, handler: (*Handler).GetAllShots,
			configure: func(_ *testing.T, service *fakeShotService) {
				service.getAllShots = func(context.Context) ([]shot.Shot, error) {
					return []shot.Shot{*first, *second}, nil
				}
			},
		},
		{
			name: "update", method: http.MethodPut, target: "/rest/v1/shots/9", body: validShotRequestBody, id: "9",
			status: http.StatusOK, expected: ShotResponse{*updated}, handler: (*Handler).UpdateShotById,
			configure: func(t *testing.T, service *fakeShotService) {
				service.updateShotByID = func(_ context.Context, id int, value *shot.Shot) (*shot.Shot, error) {
					if id != 9 {
						t.Errorf("id = %d, want 9", id)
					}
					assertShotRequest(t, value, 9)
					return updated, nil
				}
			},
		},
		{
			name: "delete", method: http.MethodDelete, target: "/rest/v1/shots/11", id: "11",
			status: http.StatusOK, expected: ItemDeletedResponse{Id: 11, Msg: "shot 11 deleted successfully"}, handler: (*Handler).DeleteShotById,
			configure: func(t *testing.T, service *fakeShotService) {
				service.deleteShotByID = func(_ context.Context, id int) error {
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
			handler, _, _, _, service := newTestHandler(t)
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

func TestShotHandlersErrorPaths(t *testing.T) {
	invalidRatingBody := strings.Replace(validShotRequestBody, `"rating":8.5`, `"rating":11`, 1)
	tests := []struct {
		name      string
		method    string
		target    string
		body      string
		id        string
		status    int
		message   string
		configure func(*fakeShotService)
		handler   controllerHandler
	}{
		{
			name: "create invalid rating", method: http.MethodPost, target: "/rest/v1/shots", body: invalidRatingBody,
			status: http.StatusBadRequest, message: "shot rating is out of range. Must be between 0.0 and 10.0", handler: (*Handler).CreateShot,
			configure: func(service *fakeShotService) {
				service.createShot = func(_ context.Context, value *shot.Shot) (*shot.Shot, error) {
					if value.Rating != 11 {
						t.Errorf("shot rating = %v, want 11", value.Rating)
					}
					return nil, domainerrors.ErrShotRatingOutOfRange
				}
			},
		},
		{
			name: "get not found", method: http.MethodGet, target: "/rest/v1/shots/5", id: "5",
			status: http.StatusNotFound, message: "no shot found for given id", handler: (*Handler).GetShotById,
			configure: func(service *fakeShotService) {
				service.getShotByID = func(context.Context, int) (*shot.Shot, error) { return nil, domainerrors.ErrShotDoesNotExist }
			},
		},
		{
			name: "get all service failure", method: http.MethodGet, target: "/rest/v1/shots",
			status: http.StatusInternalServerError, message: "internal server error", handler: (*Handler).GetAllShots,
			configure: func(service *fakeShotService) {
				service.getAllShots = func(context.Context) ([]shot.Shot, error) { return nil, errors.New("service failure") }
			},
		},
		{
			name: "update invalid rating", method: http.MethodPut, target: "/rest/v1/shots/5", body: invalidRatingBody, id: "5",
			status: http.StatusBadRequest, message: "shot rating is out of range. Must be between 0.0 and 10.0", handler: (*Handler).UpdateShotById,
			configure: func(service *fakeShotService) {
				service.updateShotByID = func(_ context.Context, _ int, value *shot.Shot) (*shot.Shot, error) {
					if value.Rating != 11 {
						t.Errorf("shot rating = %v, want 11", value.Rating)
					}
					return nil, domainerrors.ErrShotRatingOutOfRange
				}
			},
		},
		{
			name: "delete not found", method: http.MethodDelete, target: "/rest/v1/shots/5", id: "5",
			status: http.StatusNotFound, message: "no shot found for given id", handler: (*Handler).DeleteShotById,
			configure: func(service *fakeShotService) {
				service.deleteShotByID = func(context.Context, int) error { return domainerrors.ErrShotDoesNotExist }
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _, _, service := newTestHandler(t)
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

func assertShotRequest(t *testing.T, value *shot.Shot, id int) {
	t.Helper()
	if value.Id != id {
		t.Errorf("shot id = %d, want %d", value.Id, id)
	}
	if value.Sheet == nil || value.Sheet.Id != 3 {
		t.Errorf("shot sheet = %#v, want id 3", value.Sheet)
	}
	if value.Beans == nil || value.Beans.Id != 4 {
		t.Errorf("shot beans = %#v, want id 4", value.Beans)
	}
	if value.GrindSetting != 12 || value.QuantityIn != 18.5 || value.QuantityOut != 37 {
		t.Errorf("shot grind/quantities = %d/%v/%v, want 12/18.5/37", value.GrindSetting, value.QuantityIn, value.QuantityOut)
	}
	if value.ShotTime != 28*time.Second || value.WaterTemperature != 93.5 || value.Rating != 8.5 {
		t.Errorf("shot time/temperature/rating = %v/%v/%v, want 28s/93.5/8.5", value.ShotTime, value.WaterTemperature, value.Rating)
	}
	if value.IsTooBitter || !value.IsTooSour {
		t.Errorf("shot taste flags = bitter:%t sour:%t, want false/true", value.IsTooBitter, value.IsTooSour)
	}
	if value.ComparisonWithPreviousResult != modelsql.Better {
		t.Errorf("shot comparison = %v, want %v", value.ComparisonWithPreviousResult, modelsql.Better)
	}
	if value.AdditionalNotes != "test notes" {
		t.Errorf("shot notes = %q, want %q", value.AdditionalNotes, "test notes")
	}
}
