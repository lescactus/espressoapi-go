package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
	"time"

	"github.com/justinas/alice"
	"github.com/lescactus/espressoapi-go/cmd/app"
	"github.com/lescactus/espressoapi-go/internal/controllers/rest"
	"github.com/lescactus/espressoapi-go/internal/controllers/web"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
)

// stubNow backs every stubbed CreatedAt/UpdatedAt so handler logging that
// dereferences timestamps does not panic on nil.
var stubNow = time.Now()

// TestMain initializes the app.App global with a stub Swagger filesystem so
// the /swagger.json route can dispatch without the real embedded doc.
func TestMain(m *testing.M) {
	app.App = &app.Application{
		SwaggerFS: fstest.MapFS{
			"docs/swagger.json": &fstest.MapFile{Data: []byte(`{}`)},
		},
	}
	m.Run()
}

// stubSheetService is a minimal no-op sheet.Service used to exercise routing only.
type stubSheetService struct{}

func (stubSheetService) CreateSheetByName(context.Context, string) (*sheet.Sheet, error) {
	return &sheet.Sheet{Id: 1, Name: "stub", CreatedAt: &stubNow, UpdatedAt: &stubNow}, nil
}
func (stubSheetService) GetSheetById(context.Context, int) (*sheet.Sheet, error) {
	return &sheet.Sheet{Id: 1, Name: "stub", CreatedAt: &stubNow, UpdatedAt: &stubNow}, nil
}
func (stubSheetService) GetAllSheets(context.Context) ([]sheet.Sheet, error) { return nil, nil }
func (stubSheetService) UpdateSheetById(context.Context, int, *sheet.Sheet) (*sheet.Sheet, error) {
	return &sheet.Sheet{Id: 1, Name: "stub", CreatedAt: &stubNow, UpdatedAt: &stubNow}, nil
}
func (stubSheetService) DeleteSheetById(context.Context, int) error { return nil }
func (stubSheetService) Ping(context.Context) error                 { return nil }

// stubRoasterService is a minimal no-op roaster.Service used to exercise routing only.
type stubRoasterService struct{}

func (stubRoasterService) CreateRoasterByName(context.Context, string) (*roaster.Roaster, error) {
	return &roaster.Roaster{Id: 1, Name: "stub", CreatedAt: &stubNow, UpdatedAt: &stubNow}, nil
}
func (stubRoasterService) GetRoasterById(context.Context, int) (*roaster.Roaster, error) {
	return &roaster.Roaster{Id: 1, Name: "stub", CreatedAt: &stubNow, UpdatedAt: &stubNow}, nil
}
func (stubRoasterService) GetAllRoasters(context.Context) ([]roaster.Roaster, error) {
	return nil, nil
}
func (stubRoasterService) UpdateRoasterById(context.Context, int, *roaster.Roaster) (*roaster.Roaster, error) {
	return &roaster.Roaster{Id: 1, Name: "stub", CreatedAt: &stubNow, UpdatedAt: &stubNow}, nil
}
func (stubRoasterService) DeleteRoasterById(context.Context, int) error { return nil }
func (stubRoasterService) Ping(context.Context) error                   { return nil }

// stubBeanService is a minimal no-op bean.Service used to exercise routing only.
type stubBeanService struct{}

func stubBean() *bean.Bean {
	return &bean.Bean{
		Id:        1,
		Name:      "stub",
		Roaster:   &roaster.Roaster{Id: 1, Name: "stub roaster", CreatedAt: &stubNow, UpdatedAt: &stubNow},
		RoastDate: &stubNow,
		CreatedAt: &stubNow,
		UpdatedAt: &stubNow,
	}
}

func (stubBeanService) CreateBean(context.Context, *bean.Bean) (*bean.Bean, error) {
	return stubBean(), nil
}
func (stubBeanService) GetBeanById(context.Context, int) (*bean.Bean, error) {
	return stubBean(), nil
}
func (stubBeanService) GetAllBeans(context.Context) ([]bean.Bean, error) { return nil, nil }
func (stubBeanService) UpdateBeanById(context.Context, int, *bean.Bean) (*bean.Bean, error) {
	return stubBean(), nil
}
func (stubBeanService) DeleteBeanById(context.Context, int) error { return nil }
func (stubBeanService) Ping(context.Context) error                { return nil }

// stubShotService is a minimal no-op shot.Service used to exercise routing only.
type stubShotService struct{}

func stubShot() *shot.Shot {
	return &shot.Shot{
		Id:        1,
		Sheet:     &sheet.Sheet{Id: 1, Name: "stub sheet", CreatedAt: &stubNow, UpdatedAt: &stubNow},
		Beans:     stubBean(),
		CreatedAt: &stubNow,
		UpdatedAt: &stubNow,
	}
}

func (stubShotService) CreateShot(context.Context, *shot.Shot) (*shot.Shot, error) {
	return stubShot(), nil
}
func (stubShotService) GetShotById(context.Context, int) (*shot.Shot, error) {
	return stubShot(), nil
}
func (stubShotService) GetAllShots(context.Context) ([]shot.Shot, error) { return nil, nil }
func (stubShotService) GetShotsBySheetId(context.Context, int) ([]shot.Shot, error) {
	return nil, nil
}
func (stubShotService) UpdateShotById(context.Context, int, *shot.Shot) (*shot.Shot, error) {
	return stubShot(), nil
}
func (stubShotService) DeleteShotById(context.Context, int) error { return nil }
func (stubShotService) Ping(context.Context) error                { return nil }

func newTestRouter() http.Handler {
	h := rest.NewHandler(stubSheetService{}, stubRoasterService{}, stubBeanService{}, stubShotService{}, 1<<20)
	web := web.NewHandler(stubSheetService{}, stubRoasterService{}, stubBeanService{}, stubShotService{})
	return newRouter(h, web, alice.New())
}

func TestNewRouter_RegistersAllExistingRoutes(t *testing.T) {
	r := newTestRouter()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"ping", http.MethodGet, "/ping"},
		{"create sheet", http.MethodPost, "/rest/v1/sheets"},
		{"get sheet by id", http.MethodGet, "/rest/v1/sheets/1"},
		{"get all sheets", http.MethodGet, "/rest/v1/sheets"},
		{"update sheet by id", http.MethodPut, "/rest/v1/sheets/1"},
		{"delete sheet by id", http.MethodDelete, "/rest/v1/sheets/1"},
		{"create roaster", http.MethodPost, "/rest/v1/roasters"},
		{"get roaster by id", http.MethodGet, "/rest/v1/roasters/1"},
		{"get all roasters", http.MethodGet, "/rest/v1/roasters"},
		{"update roaster by id", http.MethodPut, "/rest/v1/roasters/1"},
		{"delete roaster by id", http.MethodDelete, "/rest/v1/roasters/1"},
		{"create beans", http.MethodPost, "/rest/v1/beans"},
		{"get beans by id", http.MethodGet, "/rest/v1/beans/1"},
		{"get all beans", http.MethodGet, "/rest/v1/beans"},
		{"update beans by id", http.MethodPut, "/rest/v1/beans/1"},
		{"delete beans by id", http.MethodDelete, "/rest/v1/beans/1"},
		{"create shot", http.MethodPost, "/rest/v1/shots"},
		{"get shot by id", http.MethodGet, "/rest/v1/shots/1"},
		{"get all shots", http.MethodGet, "/rest/v1/shots"},
		{"update shot by id", http.MethodPut, "/rest/v1/shots/1"},
		{"delete shot by id", http.MethodDelete, "/rest/v1/shots/1"},
		{"get shots by sheet id", http.MethodGet, "/rest/v1/sheets/1/shots"},
		{"redoc", http.MethodGet, "/redoc"},
		{"swagger ui", http.MethodGet, "/swagger"},
		{"swagger json", http.MethodGet, "/swagger.json"},
		{"web home", http.MethodGet, "/"},
		{"web list sheets", http.MethodGet, "/sheets"},
		{"web add sheet form", http.MethodGet, "/sheets/add"},
		{"web create sheet", http.MethodPost, "/sheets/add"},
		{"web get sheet", http.MethodGet, "/sheets/get/1"},
		{"web edit sheet form", http.MethodGet, "/sheets/update/1"},
		{"web update sheet", http.MethodPut, "/sheets/update/1"},
		{"web delete sheet", http.MethodDelete, "/sheets/delete/1"},
		{"web list roasters", http.MethodGet, "/roasters"},
		{"web add roaster form", http.MethodGet, "/roasters/add"},
		{"web create roaster", http.MethodPost, "/roasters/add"},
		{"web get roaster", http.MethodGet, "/roasters/get/1"},
		{"web edit roaster form", http.MethodGet, "/roasters/update/1"},
		{"web update roaster", http.MethodPut, "/roasters/update/1"},
		{"web delete roaster", http.MethodDelete, "/roasters/delete/1"},
		{"web list beans", http.MethodGet, "/beans"},
		{"web add bean form", http.MethodGet, "/beans/add"},
		{"web create bean", http.MethodPost, "/beans/add"},
		{"web get bean", http.MethodGet, "/beans/get/1"},
		{"web edit bean form", http.MethodGet, "/beans/update/1"},
		{"web update bean", http.MethodPut, "/beans/update/1"},
		{"web delete bean", http.MethodDelete, "/beans/delete/1"},
		{"web list shots", http.MethodGet, "/shots"},
		{"web add shot form", http.MethodGet, "/shots/add"},
		{"web create shot", http.MethodPost, "/shots/add"},
		{"web get shot", http.MethodGet, "/shots/get/1"},
		{"web edit shot form", http.MethodGet, "/shots/update/1"},
		{"web update shot", http.MethodPut, "/shots/update/1"},
		{"web delete shot", http.MethodDelete, "/shots/delete/1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			r.ServeHTTP(rec, req)

			if rec.Code == http.StatusNotFound {
				t.Fatalf("%s %s: route not registered, got 404", tt.method, tt.path)
			}
		})
	}
}

func TestNewRouter_UnknownPathReturns404(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/does-not-exist", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown path, got %d", rec.Code)
	}
}

func TestNewRouter_UnsupportedMethodOnKnownPathReturns405(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodPatch, "/rest/v1/sheets/1", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for unsupported method on known path, got %d", rec.Code)
	}
}

func TestNewRouter_TrailingSlashDoesNotMatch(t *testing.T) {
	r := newTestRouter()

	req := httptest.NewRequest(http.MethodGet, "/rest/v1/sheets/", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// httprouter redirects "/rest/v1/sheets/" to "/rest/v1/sheets" (no trailing slash) by default.
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("expected 301 redirect for trailing slash, got %d", rec.Code)
	}
}
