package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

// fakeRoasterService is a hand-rolled fake with func fields, matching the
// pattern used by internal/controllers/rest.
type fakeRoasterService struct {
	t                   *testing.T
	createRoasterByName func(context.Context, string) (*roaster.Roaster, error)
	getRoasterByID      func(context.Context, int) (*roaster.Roaster, error)
	getAllRoasters      func(context.Context) ([]roaster.Roaster, error)
	updateRoasterByID   func(context.Context, int, *roaster.Roaster) (*roaster.Roaster, error)
	deleteRoasterByID   func(context.Context, int) error
}

var _ roaster.Service = (*fakeRoasterService)(nil)

func (f *fakeRoasterService) CreateRoasterByName(ctx context.Context, name string) (*roaster.Roaster, error) {
	if f.createRoasterByName == nil {
		f.t.Fatalf("unexpected CreateRoasterByName call")
	}
	return f.createRoasterByName(ctx, name)
}

func (f *fakeRoasterService) GetRoasterById(ctx context.Context, id int) (*roaster.Roaster, error) {
	if f.getRoasterByID == nil {
		f.t.Fatalf("unexpected GetRoasterById call")
	}
	return f.getRoasterByID(ctx, id)
}

func (f *fakeRoasterService) GetAllRoasters(ctx context.Context) ([]roaster.Roaster, error) {
	if f.getAllRoasters == nil {
		f.t.Fatalf("unexpected GetAllRoasters call")
	}
	return f.getAllRoasters(ctx)
}

func (f *fakeRoasterService) UpdateRoasterById(ctx context.Context, id int, value *roaster.Roaster) (*roaster.Roaster, error) {
	if f.updateRoasterByID == nil {
		f.t.Fatalf("unexpected UpdateRoasterById call")
	}
	return f.updateRoasterByID(ctx, id, value)
}

func (f *fakeRoasterService) DeleteRoasterById(ctx context.Context, id int) error {
	if f.deleteRoasterByID == nil {
		f.t.Fatalf("unexpected DeleteRoasterById call")
	}
	return f.deleteRoasterByID(ctx, id)
}

func (f *fakeRoasterService) Ping(context.Context) error { return nil }

// unusedSheetService satisfies Handler's sheet.Service dependency for tests
// that only exercise roaster routes.
type unusedSheetService struct{}

func (unusedSheetService) CreateSheetByName(context.Context, string) (*sheet.Sheet, error) {
	return nil, nil
}
func (unusedSheetService) GetSheetById(context.Context, int) (*sheet.Sheet, error) { return nil, nil }
func (unusedSheetService) GetAllSheets(context.Context) ([]sheet.Sheet, error)     { return nil, nil }
func (unusedSheetService) UpdateSheetById(context.Context, int, *sheet.Sheet) (*sheet.Sheet, error) {
	return nil, nil
}
func (unusedSheetService) DeleteSheetById(context.Context, int) error { return nil }
func (unusedSheetService) Ping(context.Context) error                 { return nil }

func newTestRoasterHandler(t *testing.T) (*Handler, *fakeRoasterService) {
	t.Helper()
	svc := &fakeRoasterService{t: t}
	return NewHandler(unusedSheetService{}, svc, unusedBeanService{}, unusedShotService{}), svc
}

func testRoaster(id int, name string) *roaster.Roaster {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	return &roaster.Roaster{Id: id, Name: name, CreatedAt: &created}
}

func TestListRoasters_FullPageVsFragment(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.getAllRoasters = func(context.Context) ([]roaster.Roaster, error) {
		return []roaster.Roaster{*testRoaster(1, "Blue Bottle")}, nil
	}

	fullPage := httptest.NewRecorder()
	h.ListRoasters(fullPage, newWebRequest(http.MethodGet, "/roasters", "", "", "", false))
	if !strings.Contains(fullPage.Body.String(), "<html") {
		t.Errorf("expected full HTML page without HX-Request, got: %s", fullPage.Body.String())
	}

	fragment := httptest.NewRecorder()
	h.ListRoasters(fragment, newWebRequest(http.MethodGet, "/roasters", "", "", "", true))
	if strings.Contains(fragment.Body.String(), "<html") || !strings.Contains(fragment.Body.String(), `id="roasters-table"`) {
		t.Errorf("expected a table fragment only with HX-Request, got: %s", fragment.Body.String())
	}
}

func TestCreateRoaster_HappyPath(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.createRoasterByName = func(_ context.Context, name string) (*roaster.Roaster, error) {
		return testRoaster(3, name), nil
	}

	req := newWebRequest(http.MethodPost, "/roasters/add", "name=Blue+Bottle", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateRoaster(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Blue Bottle") || !strings.Contains(rec.Body.String(), `hx-swap-oob="beforeend"`) {
		t.Errorf("expected the new row and an OOB success alert, got: %s", rec.Body.String())
	}
}

func TestCreateRoaster_DuplicateNameReturns409(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.createRoasterByName = func(context.Context, string) (*roaster.Roaster, error) {
		return nil, errors.ErrRoasterAlreadyExists
	}

	req := newWebRequest(http.MethodPost, "/roasters/add", "name=Blue+Bottle", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateRoaster(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestCreateRoaster_EmptyNameReturns400(t *testing.T) {
	h, _ := newTestRoasterHandler(t)

	req := newWebRequest(http.MethodPost, "/roasters/add", "name=", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateRoaster(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateRoaster_MalformedFormBodyReturns400(t *testing.T) {
	h, _ := newTestRoasterHandler(t)

	req := newWebRequest(http.MethodPost, "/roasters/add", "name=%zz", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateRoaster(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a malformed body, got %d", rec.Code)
	}
}

func TestCreateRoaster_OversizedBodyReturns413(t *testing.T) {
	h, _ := newTestRoasterHandler(t)

	req := newOversizedWebRequest(http.MethodPost, "/roasters/add", formURLEncoded, "")
	rec := httptest.NewRecorder()
	h.CreateRoaster(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for an oversized body, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetRoaster_InvalidIDReturns400(t *testing.T) {
	h, _ := newTestRoasterHandler(t)
	rec := httptest.NewRecorder()
	h.GetRoaster(rec, newWebRequest(http.MethodGet, "/roasters/get/abc", "", "", "abc", false))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetRoaster_UnknownIDReturns404(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.getRoasterByID = func(context.Context, int) (*roaster.Roaster, error) {
		return nil, errors.ErrRoasterDoesNotExist
	}

	rec := httptest.NewRecorder()
	h.GetRoaster(rec, newWebRequest(http.MethodGet, "/roasters/get/99", "", "", "99", false))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetRoaster_FragmentVsOneRowTablePage(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.getRoasterByID = func(context.Context, int) (*roaster.Roaster, error) { return testRoaster(1, "Blue Bottle"), nil }

	fragment := httptest.NewRecorder()
	h.GetRoaster(fragment, newWebRequest(http.MethodGet, "/roasters/get/1", "", "", "1", true))
	if strings.Contains(fragment.Body.String(), "<html") || !strings.Contains(fragment.Body.String(), `id="roaster-row-1"`) {
		t.Errorf("expected a bare row fragment with HX-Request, got: %s", fragment.Body.String())
	}

	full := httptest.NewRecorder()
	h.GetRoaster(full, newWebRequest(http.MethodGet, "/roasters/get/1", "", "", "1", false))
	if !strings.Contains(full.Body.String(), "<html") || !strings.Contains(full.Body.String(), "<table") {
		t.Errorf("expected the full page with a one-row table, got: %s", full.Body.String())
	}
}

func TestUpdateRoaster_HappyPath(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.updateRoasterByID = func(_ context.Context, id int, r *roaster.Roaster) (*roaster.Roaster, error) {
		return testRoaster(id, r.Name), nil
	}

	req := newWebRequest(http.MethodPut, "/roasters/update/1", "name=Renamed", formURLEncoded, "1", true)
	rec := httptest.NewRecorder()
	h.UpdateRoaster(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Renamed") {
		t.Errorf("expected the updated row, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "successfully updated") {
		t.Errorf("expected an OOB success alert, got: %s", rec.Body.String())
	}
}

func TestUpdateRoaster_EmptyNameReturns400(t *testing.T) {
	h, _ := newTestRoasterHandler(t)

	req := newWebRequest(http.MethodPut, "/roasters/update/1", "name=", formURLEncoded, "1", true)
	rec := httptest.NewRecorder()
	h.UpdateRoaster(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteRoaster_HappyPath(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.deleteRoasterByID = func(context.Context, int) error { return nil }

	req := newWebRequest(http.MethodDelete, "/roasters/delete/1", "", "", "1", true)
	rec := httptest.NewRecorder()
	h.DeleteRoaster(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "successfully deleted") {
		t.Errorf("expected 200 with an OOB success alert, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteRoaster_ForeignKeyViolationReturns409WithReswapNone(t *testing.T) {
	h, svc := newTestRoasterHandler(t)
	svc.deleteRoasterByID = func(context.Context, int) error { return errors.ErrBeansForeignKeyConstraint }

	req := newWebRequest(http.MethodDelete, "/roasters/delete/1", "", "", "1", true)
	rec := httptest.NewRecorder()
	h.DeleteRoaster(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	if rec.Header().Get("HX-Reswap") != "none" {
		t.Errorf("expected HX-Reswap: none so the existing row remains, got %q", rec.Header().Get("HX-Reswap"))
	}
	if !strings.Contains(rec.Body.String(), "still used by beans") {
		t.Errorf("expected a human FK-violation message, got: %s", rec.Body.String())
	}
}
