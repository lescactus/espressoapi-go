package rest

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	modelsql "github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
	"github.com/rs/zerolog"
)

type fakeSheetService struct {
	t                 *testing.T
	createSheetByName func(context.Context, string) (*sheet.Sheet, error)
	getSheetByID      func(context.Context, int) (*sheet.Sheet, error)
	getAllSheets      func(context.Context) ([]sheet.Sheet, error)
	updateSheetByID   func(context.Context, int, *sheet.Sheet) (*sheet.Sheet, error)
	deleteSheetByID   func(context.Context, int) error
	ping              func(context.Context) error
}

var _ sheet.Service = (*fakeSheetService)(nil)

func (f *fakeSheetService) CreateSheetByName(ctx context.Context, name string) (*sheet.Sheet, error) {
	if f.createSheetByName == nil {
		f.t.Fatalf("unexpected CreateSheetByName call")
		return nil, nil
	}
	return f.createSheetByName(ctx, name)
}

func (f *fakeSheetService) GetSheetById(ctx context.Context, id int) (*sheet.Sheet, error) {
	if f.getSheetByID == nil {
		f.t.Fatalf("unexpected GetSheetById call")
		return nil, nil
	}
	return f.getSheetByID(ctx, id)
}

func (f *fakeSheetService) GetAllSheets(ctx context.Context) ([]sheet.Sheet, error) {
	if f.getAllSheets == nil {
		f.t.Fatalf("unexpected GetAllSheets call")
		return nil, nil
	}
	return f.getAllSheets(ctx)
}

func (f *fakeSheetService) UpdateSheetById(ctx context.Context, id int, value *sheet.Sheet) (*sheet.Sheet, error) {
	if f.updateSheetByID == nil {
		f.t.Fatalf("unexpected UpdateSheetById call")
		return nil, nil
	}
	return f.updateSheetByID(ctx, id, value)
}

func (f *fakeSheetService) DeleteSheetById(ctx context.Context, id int) error {
	if f.deleteSheetByID == nil {
		f.t.Fatalf("unexpected DeleteSheetById call")
		return nil
	}
	return f.deleteSheetByID(ctx, id)
}

func (f *fakeSheetService) Ping(ctx context.Context) error {
	if f.ping == nil {
		f.t.Fatalf("unexpected sheet Ping call")
		return nil
	}
	return f.ping(ctx)
}

type fakeRoasterService struct {
	t                   *testing.T
	createRoasterByName func(context.Context, string) (*roaster.Roaster, error)
	getRoasterByID      func(context.Context, int) (*roaster.Roaster, error)
	getAllRoasters      func(context.Context) ([]roaster.Roaster, error)
	updateRoasterByID   func(context.Context, int, *roaster.Roaster) (*roaster.Roaster, error)
	deleteRoasterByID   func(context.Context, int) error
	ping                func(context.Context) error
}

var _ roaster.Service = (*fakeRoasterService)(nil)

func (f *fakeRoasterService) CreateRoasterByName(ctx context.Context, name string) (*roaster.Roaster, error) {
	if f.createRoasterByName == nil {
		f.t.Fatalf("unexpected CreateRoasterByName call")
		return nil, nil
	}
	return f.createRoasterByName(ctx, name)
}

func (f *fakeRoasterService) GetRoasterById(ctx context.Context, id int) (*roaster.Roaster, error) {
	if f.getRoasterByID == nil {
		f.t.Fatalf("unexpected GetRoasterById call")
		return nil, nil
	}
	return f.getRoasterByID(ctx, id)
}

func (f *fakeRoasterService) GetAllRoasters(ctx context.Context) ([]roaster.Roaster, error) {
	if f.getAllRoasters == nil {
		f.t.Fatalf("unexpected GetAllRoasters call")
		return nil, nil
	}
	return f.getAllRoasters(ctx)
}

func (f *fakeRoasterService) UpdateRoasterById(ctx context.Context, id int, value *roaster.Roaster) (*roaster.Roaster, error) {
	if f.updateRoasterByID == nil {
		f.t.Fatalf("unexpected UpdateRoasterById call")
		return nil, nil
	}
	return f.updateRoasterByID(ctx, id, value)
}

func (f *fakeRoasterService) DeleteRoasterById(ctx context.Context, id int) error {
	if f.deleteRoasterByID == nil {
		f.t.Fatalf("unexpected DeleteRoasterById call")
		return nil
	}
	return f.deleteRoasterByID(ctx, id)
}

func (f *fakeRoasterService) Ping(ctx context.Context) error {
	if f.ping == nil {
		f.t.Fatalf("unexpected roaster Ping call")
		return nil
	}
	return f.ping(ctx)
}

type fakeBeanService struct {
	t              *testing.T
	createBean     func(context.Context, *bean.Bean) (*bean.Bean, error)
	getBeanByID    func(context.Context, int) (*bean.Bean, error)
	getAllBeans    func(context.Context) ([]bean.Bean, error)
	updateBeanByID func(context.Context, int, *bean.Bean) (*bean.Bean, error)
	deleteBeanByID func(context.Context, int) error
	ping           func(context.Context) error
}

var _ bean.Service = (*fakeBeanService)(nil)

func (f *fakeBeanService) CreateBean(ctx context.Context, value *bean.Bean) (*bean.Bean, error) {
	if f.createBean == nil {
		f.t.Fatalf("unexpected CreateBean call")
		return nil, nil
	}
	return f.createBean(ctx, value)
}

func (f *fakeBeanService) GetBeanById(ctx context.Context, id int) (*bean.Bean, error) {
	if f.getBeanByID == nil {
		f.t.Fatalf("unexpected GetBeanById call")
		return nil, nil
	}
	return f.getBeanByID(ctx, id)
}

func (f *fakeBeanService) GetAllBeans(ctx context.Context) ([]bean.Bean, error) {
	if f.getAllBeans == nil {
		f.t.Fatalf("unexpected GetAllBeans call")
		return nil, nil
	}
	return f.getAllBeans(ctx)
}

func (f *fakeBeanService) UpdateBeanById(ctx context.Context, id int, value *bean.Bean) (*bean.Bean, error) {
	if f.updateBeanByID == nil {
		f.t.Fatalf("unexpected UpdateBeanById call")
		return nil, nil
	}
	return f.updateBeanByID(ctx, id, value)
}

func (f *fakeBeanService) DeleteBeanById(ctx context.Context, id int) error {
	if f.deleteBeanByID == nil {
		f.t.Fatalf("unexpected DeleteBeanById call")
		return nil
	}
	return f.deleteBeanByID(ctx, id)
}

func (f *fakeBeanService) Ping(ctx context.Context) error {
	if f.ping == nil {
		f.t.Fatalf("unexpected bean Ping call")
		return nil
	}
	return f.ping(ctx)
}

type fakeShotService struct {
	t                 *testing.T
	createShot        func(context.Context, *shot.Shot) (*shot.Shot, error)
	getShotByID       func(context.Context, int) (*shot.Shot, error)
	getAllShots       func(context.Context) ([]shot.Shot, error)
	getShotsBySheetID func(context.Context, int) ([]shot.Shot, error)
	updateShotByID    func(context.Context, int, *shot.Shot) (*shot.Shot, error)
	deleteShotByID    func(context.Context, int) error
	ping              func(context.Context) error
}

var _ shot.Service = (*fakeShotService)(nil)

func (f *fakeShotService) CreateShot(ctx context.Context, value *shot.Shot) (*shot.Shot, error) {
	if f.createShot == nil {
		f.t.Fatalf("unexpected CreateShot call")
		return nil, nil
	}
	return f.createShot(ctx, value)
}

func (f *fakeShotService) GetShotById(ctx context.Context, id int) (*shot.Shot, error) {
	if f.getShotByID == nil {
		f.t.Fatalf("unexpected GetShotById call")
		return nil, nil
	}
	return f.getShotByID(ctx, id)
}

func (f *fakeShotService) GetAllShots(ctx context.Context) ([]shot.Shot, error) {
	if f.getAllShots == nil {
		f.t.Fatalf("unexpected GetAllShots call")
		return nil, nil
	}
	return f.getAllShots(ctx)
}

func (f *fakeShotService) GetShotsBySheetId(ctx context.Context, sheetId int) ([]shot.Shot, error) {
	if f.getShotsBySheetID == nil {
		f.t.Fatalf("unexpected GetShotsBySheetId call")
		return nil, nil
	}
	return f.getShotsBySheetID(ctx, sheetId)
}

func (f *fakeShotService) UpdateShotById(ctx context.Context, id int, value *shot.Shot) (*shot.Shot, error) {
	if f.updateShotByID == nil {
		f.t.Fatalf("unexpected UpdateShotById call")
		return nil, nil
	}
	return f.updateShotByID(ctx, id, value)
}

func (f *fakeShotService) DeleteShotById(ctx context.Context, id int) error {
	if f.deleteShotByID == nil {
		f.t.Fatalf("unexpected DeleteShotById call")
		return nil
	}
	return f.deleteShotByID(ctx, id)
}

func (f *fakeShotService) Ping(ctx context.Context) error {
	if f.ping == nil {
		f.t.Fatalf("unexpected shot Ping call")
		return nil
	}
	return f.ping(ctx)
}

func newTestHandler(t *testing.T) (*Handler, *fakeSheetService, *fakeRoasterService, *fakeBeanService, *fakeShotService) {
	t.Helper()

	sheetService := &fakeSheetService{t: t}
	roasterService := &fakeRoasterService{t: t}
	beanService := &fakeBeanService{t: t}
	shotService := &fakeShotService{t: t}

	return NewHandler(sheetService, roasterService, beanService, shotService, 64), sheetService, roasterService, beanService, shotService
}

func newControllerRequest(t *testing.T, method, target, body, contentType, id string) *http.Request {
	t.Helper()

	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	ctx := req.Context()
	if id != "" {
		params := httprouter.Params{{Key: "id", Value: id}}
		ctx = context.WithValue(ctx, httprouter.ParamsKey, params)
	}
	logger := zerolog.New(io.Discard)
	ctx = logger.WithContext(ctx)

	return req.WithContext(ctx)
}

func executeHandler(handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func assertJSONResponse(t *testing.T, recorder *httptest.ResponseRecorder, status int, expected any) {
	t.Helper()

	if recorder.Code != status {
		t.Errorf("status = %d, want %d", recorder.Code, status)
	}
	if contentType := recorder.Header().Get("Content-Type"); contentType != ContentTypeApplicationJSON {
		t.Errorf("Content-Type = %q, want %q", contentType, ContentTypeApplicationJSON)
	}

	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("marshal expected response: %v", err)
	}

	var gotValue any
	if err := json.Unmarshal(recorder.Body.Bytes(), &gotValue); err != nil {
		t.Fatalf("decode response body %q: %v", recorder.Body.String(), err)
	}
	var expectedValue any
	if err := json.Unmarshal(expectedJSON, &expectedValue); err != nil {
		t.Fatalf("decode expected response: %v", err)
	}

	if !reflect.DeepEqual(gotValue, expectedValue) {
		t.Errorf("response = %#v, want %#v", gotValue, expectedValue)
	}
}

func testSheet(id int, name string) *sheet.Sheet {
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	return &sheet.Sheet{Id: id, Name: name, CreatedAt: &createdAt, UpdatedAt: &updatedAt}
}

func testRoaster(id int, name string) *roaster.Roaster {
	createdAt := time.Date(2026, time.January, 3, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	return &roaster.Roaster{Id: id, Name: name, CreatedAt: &createdAt, UpdatedAt: &updatedAt}
}

func testBean(id int, name string) *bean.Bean {
	createdAt := time.Date(2026, time.January, 4, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	roastDate := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	return &bean.Bean{
		Id:         id,
		Roaster:    testRoaster(2, "test roaster"),
		Name:       name,
		RoastDate:  &roastDate,
		RoastLevel: modelsql.RoastLevelMedium,
		CreatedAt:  &createdAt,
		UpdatedAt:  &updatedAt,
	}
}

func testShot(id int) *shot.Shot {
	createdAt := time.Date(2026, time.January, 5, 3, 4, 5, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	return &shot.Shot{
		Id:                           id,
		Sheet:                        testSheet(3, "test sheet"),
		Beans:                        testBean(4, "test beans"),
		GrindSetting:                 12,
		QuantityIn:                   18.5,
		QuantityOut:                  37.0,
		ShotTime:                     28 * time.Second,
		WaterTemperature:             93.5,
		Rating:                       8.5,
		IsTooBitter:                  false,
		IsTooSour:                    true,
		ComparisonWithPreviousResult: modelsql.Better,
		AdditionalNotes:              "test notes",
		CreatedAt:                    &createdAt,
		UpdatedAt:                    &updatedAt,
	}
}
