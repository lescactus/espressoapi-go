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
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
)

const validShotRequestBody = `{
	"sheet_id":3,
	"beans_id":4,
	"grind_setting":12,
	"quantity_in":18.5,
	"quantity_out":37,
	"shot_time":28,
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
			status: http.StatusCreated, expected: newShotResponse(*created), handler: (*Handler).CreateShot,
			configure: func(t *testing.T, service *fakeShotService) {
				service.createShot = func(_ context.Context, value *shot.Shot) (*shot.Shot, error) {
					assertShotRequest(t, value, 0)
					return created, nil
				}
			},
		},
		{
			name: "get by id", method: http.MethodGet, target: "/rest/v1/shots/7", id: "7",
			status: http.StatusOK, expected: newShotResponse(*found), handler: (*Handler).GetShotById,
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
			status: http.StatusOK, expected: []ShotResponse{newShotResponse(*first), newShotResponse(*second)}, handler: (*Handler).GetAllShots,
			configure: func(_ *testing.T, service *fakeShotService) {
				service.getAllShots = func(context.Context) ([]shot.Shot, error) {
					return []shot.Shot{*first, *second}, nil
				}
			},
		},
		{
			name: "update", method: http.MethodPut, target: "/rest/v1/shots/9", body: validShotRequestBody, id: "9",
			status: http.StatusOK, expected: newShotResponse(*updated), handler: (*Handler).UpdateShotById,
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

func TestCreateShot_ShotTimeValidation(t *testing.T) {
	// A non-numeric JSON value fails during decode, before the request ever
	// reaches the service.
	t.Run("string value rejected at decode time", func(t *testing.T) {
		// service.createShot is left nil: the fake fails the test if the
		// service is reached despite the invalid shot_time.
		handler, _, _, _, _ := newTestHandler(t)
		body := `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":"25.5","rating":8}`
		req := newControllerRequest(t, http.MethodPost, "/rest/v1/shots", body, ContentTypeApplicationJSON, "")

		recorder := executeControllerHandler(handler, (*Handler).CreateShot, req)

		assertJSONResponse(t, recorder, http.StatusBadRequest, ErrorResponse{Msg: `request body contains an invalid value for the "shot_time" field (at position 6)`})
	})
}

// TestCreateShot_ShotTimeOutOfRangeDeferredToService covers values that
// decode successfully (they are valid JSON numbers) but fall outside the
// service's [0, 3600s] range: negative, just over the max, and a legacy
// nanosecond-denominated value. Range enforcement lives solely in
// shot.Service (see internal/services/shot/service.go), matching the
// rating/comparison pattern, so these reach the service and come back as
// domainerrors.ErrShotTimeOutOfRange.
func TestCreateShot_ShotTimeOutOfRangeDeferredToService(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "negative", body: `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":-1,"rating":8}`},
		{name: "just over max", body: `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":3600.1,"rating":8}`},
		{name: "legacy nanoseconds value", body: `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":28500000000,"rating":8}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _, _, service := newTestHandler(t)
			service.createShot = func(_ context.Context, _ *shot.Shot) (*shot.Shot, error) {
				return nil, domainerrors.ErrShotTimeOutOfRange
			}
			req := newControllerRequest(t, http.MethodPost, "/rest/v1/shots", tt.body, ContentTypeApplicationJSON, "")

			recorder := executeControllerHandler(handler, (*Handler).CreateShot, req)

			assertJSONResponse(t, recorder, http.StatusBadRequest, ErrorResponse{Msg: "shot time is out of range. Must be between 0 and 3600 seconds"})
		})
	}
}

func TestCreateShot_ShotTimeSecondsRoundTrip(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantEcho string
	}{
		{
			name:     "zero is accepted as not recorded",
			body:     `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":0,"rating":8}`,
			wantEcho: `"shot_time":0`,
		},
		{
			name:     "null is a no-op, same as an omitted key",
			body:     `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":null,"rating":8}`,
			wantEcho: `"shot_time":0`,
		},
		{
			name:     "boundary max 3600 accepted",
			body:     `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":3600,"rating":8}`,
			wantEcho: `"shot_time":3600`,
		},
		{
			name:     "fractional value echoes exactly",
			body:     `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":25.3,"rating":8}`,
			wantEcho: `"shot_time":25.3`,
		},
		{
			name:     "sub-millisecond value rounds to the nearest ms",
			body:     `{"sheet_id":1,"beans_id":1,"grind_setting":12,"quantity_in":18,"quantity_out":36,"shot_time":25.0004,"rating":8}`,
			wantEcho: `"shot_time":25`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler, _, _, _, service := newTestHandler(t)
			service.createShot = func(_ context.Context, value *shot.Shot) (*shot.Shot, error) {
				s := testShot(1)
				s.ShotTime = value.ShotTime
				return s, nil
			}
			req := newControllerRequest(t, http.MethodPost, "/rest/v1/shots", tt.body, ContentTypeApplicationJSON, "")

			recorder := executeControllerHandler(handler, (*Handler).CreateShot, req)

			if recorder.Code != http.StatusCreated {
				t.Fatalf("status = %d, want 201: %s", recorder.Code, recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), tt.wantEcho) {
				t.Errorf("expected shot_time echoed as %q, got: %s", tt.wantEcho, recorder.Body.String())
			}
		})
	}
}

func TestGetShotsBySheetId(t *testing.T) {
	t.Run("populated sheet", func(t *testing.T) {
		handler, sheetSvc, _, _, shotSvc := newTestHandler(t)
		sheetSvc.getSheetByID = func(_ context.Context, id int) (*sheet.Sheet, error) {
			if id != 3 {
				t.Errorf("id = %d, want 3", id)
			}
			return &sheet.Sheet{Id: 3, Name: "sheet03"}, nil
		}
		shotSvc.getShotsBySheetID = func(_ context.Context, sheetId int) ([]shot.Shot, error) {
			if sheetId != 3 {
				t.Errorf("sheetId = %d, want 3", sheetId)
			}
			return []shot.Shot{*testShot(1)}, nil
		}

		req := newControllerRequest(t, http.MethodGet, "/rest/v1/sheets/3/shots", "", "", "3")
		recorder := executeControllerHandler(handler, (*Handler).GetShotsBySheetId, req)

		assertJSONResponse(t, recorder, http.StatusOK, &[]ShotResponse{newShotResponse(*testShot(1))})
	})

	t.Run("existing sheet without shots returns an empty array", func(t *testing.T) {
		handler, sheetSvc, _, _, shotSvc := newTestHandler(t)
		sheetSvc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) {
			return &sheet.Sheet{Id: 3, Name: "sheet03"}, nil
		}
		shotSvc.getShotsBySheetID = func(context.Context, int) ([]shot.Shot, error) {
			return []shot.Shot{}, nil
		}

		req := newControllerRequest(t, http.MethodGet, "/rest/v1/sheets/3/shots", "", "", "3")
		recorder := executeControllerHandler(handler, (*Handler).GetShotsBySheetId, req)

		assertJSONResponse(t, recorder, http.StatusOK, &[]ShotResponse{})
	})

	t.Run("missing sheet returns 404 before querying shots", func(t *testing.T) {
		handler, sheetSvc, _, _, shotSvc := newTestHandler(t)
		sheetSvc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) {
			return nil, domainerrors.ErrSheetDoesNotExist
		}
		shotSvc.getShotsBySheetID = func(context.Context, int) ([]shot.Shot, error) {
			t.Fatal("GetShotsBySheetId should not be called when the sheet does not exist")
			return nil, nil
		}

		req := newControllerRequest(t, http.MethodGet, "/rest/v1/sheets/99/shots", "", "", "99")
		recorder := executeControllerHandler(handler, (*Handler).GetShotsBySheetId, req)

		assertJSONResponse(t, recorder, http.StatusNotFound, ErrorResponse{Msg: "no sheet found for given id"})
	})

	t.Run("invalid id returns 400", func(t *testing.T) {
		handler, _, _, _, _ := newTestHandler(t)

		req := newControllerRequest(t, http.MethodGet, "/rest/v1/sheets/abc/shots", "", "", "abc")
		recorder := executeControllerHandler(handler, (*Handler).GetShotsBySheetId, req)

		assertJSONResponse(t, recorder, http.StatusBadRequest, ErrorResponse{Msg: "id must be an integer"})
	})
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
