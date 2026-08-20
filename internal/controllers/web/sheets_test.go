package web

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
)

// fakeSheetService is a hand-rolled fake with func fields, matching the
// pattern used by internal/controllers/rest.
type fakeSheetService struct {
	t                 *testing.T
	createSheetByName func(context.Context, string) (*sheet.Sheet, error)
	getSheetByID      func(context.Context, int) (*sheet.Sheet, error)
	getAllSheets      func(context.Context) ([]sheet.Sheet, error)
	updateSheetByID   func(context.Context, int, *sheet.Sheet) (*sheet.Sheet, error)
	deleteSheetByID   func(context.Context, int) error
}

var _ sheet.Service = (*fakeSheetService)(nil)

func (f *fakeSheetService) CreateSheetByName(ctx context.Context, name string) (*sheet.Sheet, error) {
	if f.createSheetByName == nil {
		f.t.Fatalf("unexpected CreateSheetByName call")
	}
	return f.createSheetByName(ctx, name)
}

func (f *fakeSheetService) GetSheetById(ctx context.Context, id int) (*sheet.Sheet, error) {
	if f.getSheetByID == nil {
		f.t.Fatalf("unexpected GetSheetById call")
	}
	return f.getSheetByID(ctx, id)
}

func (f *fakeSheetService) GetAllSheets(ctx context.Context) ([]sheet.Sheet, error) {
	if f.getAllSheets == nil {
		f.t.Fatalf("unexpected GetAllSheets call")
	}
	return f.getAllSheets(ctx)
}

func (f *fakeSheetService) UpdateSheetById(ctx context.Context, id int, value *sheet.Sheet) (*sheet.Sheet, error) {
	if f.updateSheetByID == nil {
		f.t.Fatalf("unexpected UpdateSheetById call")
	}
	return f.updateSheetByID(ctx, id, value)
}

func (f *fakeSheetService) DeleteSheetById(ctx context.Context, id int) error {
	if f.deleteSheetByID == nil {
		f.t.Fatalf("unexpected DeleteSheetById call")
	}
	return f.deleteSheetByID(ctx, id)
}

func (f *fakeSheetService) Ping(context.Context) error { return nil }

// unusedRoasterService/unusedBeanService/unusedShotService satisfy the
// remaining Handler dependencies for tests that only exercise sheet routes.
type unusedRoasterService struct{}

func (unusedRoasterService) CreateRoasterByName(context.Context, string) (*roaster.Roaster, error) {
	return nil, nil
}
func (unusedRoasterService) GetRoasterById(context.Context, int) (*roaster.Roaster, error) {
	return nil, nil
}
func (unusedRoasterService) GetAllRoasters(context.Context) ([]roaster.Roaster, error) {
	return nil, nil
}
func (unusedRoasterService) UpdateRoasterById(context.Context, int, *roaster.Roaster) (*roaster.Roaster, error) {
	return nil, nil
}
func (unusedRoasterService) DeleteRoasterById(context.Context, int) error { return nil }
func (unusedRoasterService) Ping(context.Context) error                   { return nil }

type unusedBeanService struct{}

func (unusedBeanService) CreateBean(context.Context, *bean.Bean) (*bean.Bean, error) { return nil, nil }
func (unusedBeanService) GetBeanById(context.Context, int) (*bean.Bean, error)       { return nil, nil }
func (unusedBeanService) GetAllBeans(context.Context) ([]bean.Bean, error)           { return nil, nil }
func (unusedBeanService) UpdateBeanById(context.Context, int, *bean.Bean) (*bean.Bean, error) {
	return nil, nil
}
func (unusedBeanService) DeleteBeanById(context.Context, int) error { return nil }
func (unusedBeanService) Ping(context.Context) error                { return nil }

type unusedShotService struct{}

func (unusedShotService) CreateShot(context.Context, *shot.Shot) (*shot.Shot, error) { return nil, nil }
func (unusedShotService) GetShotById(context.Context, int) (*shot.Shot, error)       { return nil, nil }
func (unusedShotService) GetAllShots(context.Context) ([]shot.Shot, error)           { return nil, nil }
func (unusedShotService) GetShotsBySheetId(context.Context, int) ([]shot.Shot, error) {
	return nil, nil
}
func (unusedShotService) UpdateShotById(context.Context, int, *shot.Shot) (*shot.Shot, error) {
	return nil, nil
}
func (unusedShotService) DeleteShotById(context.Context, int) error { return nil }
func (unusedShotService) Ping(context.Context) error                { return nil }

func newTestSheetHandler(t *testing.T) (*Handler, *fakeSheetService) {
	t.Helper()
	svc := &fakeSheetService{t: t}
	return NewHandler(svc, unusedRoasterService{}, unusedBeanService{}, unusedShotService{}), svc
}

// shotsBySheetIDStub is a minimal shot.Service exposing only a configurable
// GetShotsBySheetId, used to test the sheet detail page's shots summary.
type shotsBySheetIDStub struct {
	unusedShotService
	getShotsBySheetID func(context.Context, int) ([]shot.Shot, error)
}

func (s shotsBySheetIDStub) GetShotsBySheetId(ctx context.Context, sheetId int) ([]shot.Shot, error) {
	return s.getShotsBySheetID(ctx, sheetId)
}

// newWebRequest builds a request with an optional :id URL param and optional
// htmx/HX-Request header, mirroring the rest package's test helper.
func newWebRequest(method, target, body, contentType, id string, hx bool) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if hx {
		req.Header.Set("HX-Request", "true")
	}
	if id != "" {
		params := httprouter.Params{{Key: "id", Value: id}}
		req = req.WithContext(context.WithValue(req.Context(), httprouter.ParamsKey, params))
	}
	return req
}

func testSheet(id int, name string) *sheet.Sheet {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	return &sheet.Sheet{Id: id, Name: name, CreatedAt: &created}
}

const formURLEncoded = "application/x-www-form-urlencoded"

func TestHome_RendersSheetCards(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.getAllSheets = func(context.Context) ([]sheet.Sheet, error) {
		return []sheet.Sheet{*testSheet(1, "Double shot")}, nil
	}

	rec := httptest.NewRecorder()
	h.Home(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Double shot") {
		t.Errorf("expected sheet name in home page, got: %s", rec.Body.String())
	}
}

func TestListSheets_FullPageVsFragment(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.getAllSheets = func(context.Context) ([]sheet.Sheet, error) {
		return []sheet.Sheet{*testSheet(1, "Double shot")}, nil
	}

	fullPage := httptest.NewRecorder()
	h.ListSheets(fullPage, newWebRequest(http.MethodGet, "/sheets", "", "", "", false))
	if !strings.Contains(fullPage.Body.String(), "<html") {
		t.Errorf("expected full HTML page without HX-Request, got: %s", fullPage.Body.String())
	}

	fragment := httptest.NewRecorder()
	h.ListSheets(fragment, newWebRequest(http.MethodGet, "/sheets", "", "", "", true))
	if strings.Contains(fragment.Body.String(), "<html") {
		t.Errorf("expected a table fragment only with HX-Request, got: %s", fragment.Body.String())
	}
	if !strings.Contains(fragment.Body.String(), `id="sheets-table"`) {
		t.Errorf("expected the sheets table fragment, got: %s", fragment.Body.String())
	}
}

func TestListSheets_SortsByRequestedColumnDefaultingToID(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.getAllSheets = func(context.Context) ([]sheet.Sheet, error) {
		return []sheet.Sheet{*testSheet(2, "Beta"), *testSheet(1, "Alpha")}, nil
	}

	req := newWebRequest(http.MethodGet, "/sheets?sort=name&order=asc", "", "", "", true)
	rec := httptest.NewRecorder()
	h.ListSheets(rec, req)

	body := rec.Body.String()
	if strings.Index(body, "Alpha") > strings.Index(body, "Beta") {
		t.Errorf("expected Alpha before Beta when sorted by name asc, got: %s", body)
	}

	req = newWebRequest(http.MethodGet, "/sheets?sort=bogus", "", "", "", true)
	rec = httptest.NewRecorder()
	h.ListSheets(rec, req)
	body = rec.Body.String()
	if strings.Index(body, "sheet-row-1") > strings.Index(body, "sheet-row-2") {
		t.Errorf("expected default id-ascending order for an unknown sort column, got: %s", body)
	}
}

func TestAddSheetForm_FragmentVsFullPage(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.getAllSheets = func(context.Context) ([]sheet.Sheet, error) { return nil, nil }

	fragment := httptest.NewRecorder()
	h.AddSheetForm(fragment, newWebRequest(http.MethodGet, "/sheets/add", "", "", "", true))
	if strings.Contains(fragment.Body.String(), "<html") {
		t.Errorf("expected a bare row fragment with HX-Request, got: %s", fragment.Body.String())
	}

	fullPage := httptest.NewRecorder()
	h.AddSheetForm(fullPage, newWebRequest(http.MethodGet, "/sheets/add", "", "", "", false))
	if !strings.Contains(fullPage.Body.String(), "<html") || !strings.Contains(fullPage.Body.String(), `id="sheet-row-add"`) {
		t.Errorf("expected the full list page with the add row open, got: %s", fullPage.Body.String())
	}
}

func TestCreateSheet_HappyPath(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.createSheetByName = func(_ context.Context, name string) (*sheet.Sheet, error) {
		return testSheet(5, name), nil
	}

	req := newWebRequest(http.MethodPost, "/sheets/add", "name=Cortado", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateSheet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Cortado") {
		t.Errorf("expected the new row in the response, got: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `hx-swap-oob="beforeend"`) {
		t.Errorf("expected an OOB success alert, got: %s", rec.Body.String())
	}
}

func TestCreateSheet_EmptyNameReturns400WithInlineError(t *testing.T) {
	h, _ := newTestSheetHandler(t)

	req := newWebRequest(http.MethodPost, "/sheets/add", "name=", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateSheet(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "must not be empty") {
		t.Errorf("expected inline validation error, got: %s", rec.Body.String())
	}
}

func TestCreateSheet_DuplicateNameReturns409(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.createSheetByName = func(context.Context, string) (*sheet.Sheet, error) {
		return nil, errors.ErrSheetAlreadyExists
	}

	req := newWebRequest(http.MethodPost, "/sheets/add", "name=Cortado", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateSheet(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "already exists") {
		t.Errorf("expected duplicate-name message, got: %s", rec.Body.String())
	}
}

func TestCreateSheet_WrongContentTypeReturns415(t *testing.T) {
	h, _ := newTestSheetHandler(t)

	req := newWebRequest(http.MethodPost, "/sheets/add", `{"name":"Cortado"}`, "application/json", "", true)
	rec := httptest.NewRecorder()
	h.CreateSheet(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestCreateSheet_MalformedFormBodyReturns400(t *testing.T) {
	h, _ := newTestSheetHandler(t)

	req := newWebRequest(http.MethodPost, "/sheets/add", "name=%zz", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateSheet(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a malformed body, got %d", rec.Code)
	}
}

func TestGetSheet_InvalidIDReturns400(t *testing.T) {
	for _, id := range []string{"", "abc", "0", "-1"} {
		t.Run(id, func(t *testing.T) {
			h, _ := newTestSheetHandler(t)
			rec := httptest.NewRecorder()
			h.GetSheet(rec, newWebRequest(http.MethodGet, "/sheets/get/"+id, "", "", id, false))
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 for id %q, got %d", id, rec.Code)
			}
		})
	}
}

func TestGetSheet_UnknownIDReturns404(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) {
		return nil, errors.ErrSheetDoesNotExist
	}

	rec := httptest.NewRecorder()
	h.GetSheet(rec, newWebRequest(http.MethodGet, "/sheets/get/99", "", "", "99", false))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "404") {
		t.Errorf("expected a styled 404 page for direct navigation, got: %s", rec.Body.String())
	}
}

func TestGetSheet_FullDetailPageVsViewContextFragments(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) { return testSheet(1, "Double shot"), nil }

	full := httptest.NewRecorder()
	h.GetSheet(full, newWebRequest(http.MethodGet, "/sheets/get/1", "", "", "1", false))
	if !strings.Contains(full.Body.String(), "<html") {
		t.Errorf("expected the full detail page for direct navigation, got: %s", full.Body.String())
	}

	detailFragment := httptest.NewRecorder()
	h.GetSheet(detailFragment, newWebRequest(http.MethodGet, "/sheets/get/1?view_context=sheet-detail", "", "", "1", true))
	if strings.Contains(detailFragment.Body.String(), "<html") || !strings.Contains(detailFragment.Body.String(), `id="sheet-detail-header"`) {
		t.Errorf("expected the detail header fragment for view_context=sheet-detail, got: %s", detailFragment.Body.String())
	}

	listFragment := httptest.NewRecorder()
	h.GetSheet(listFragment, newWebRequest(http.MethodGet, "/sheets/get/1", "", "", "1", true))
	if !strings.Contains(listFragment.Body.String(), `id="sheet-row-1"`) {
		t.Errorf("expected the list row fragment by default, got: %s", listFragment.Body.String())
	}
}

func TestGetSheet_DetailPageIncludesShots(t *testing.T) {
	sheetSvc := &fakeSheetService{t: t}
	sheetSvc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) { return testSheet(1, "Double shot"), nil }
	shotSvc := shotsBySheetIDStub{getShotsBySheetID: func(_ context.Context, sheetId int) ([]shot.Shot, error) {
		if sheetId != 1 {
			t.Errorf("sheetId = %d, want 1", sheetId)
		}
		return []shot.Shot{{Id: 9, Beans: &bean.Bean{Name: "Ethiopia"}}}, nil
	}}
	h := NewHandler(sheetSvc, unusedRoasterService{}, unusedBeanService{}, shotSvc)

	rec := httptest.NewRecorder()
	h.GetSheet(rec, newWebRequest(http.MethodGet, "/sheets/get/1", "", "", "1", false))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Ethiopia") {
		t.Errorf("expected the detail page to include the sheet's shots, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetSheet_ShotsFetchErrorReturnsErrorStatusNotOK(t *testing.T) {
	sheetSvc := &fakeSheetService{t: t}
	sheetSvc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) { return testSheet(1, "Double shot"), nil }
	shotSvc := shotsBySheetIDStub{getShotsBySheetID: func(context.Context, int) ([]shot.Shot, error) {
		return nil, stderrors.New("boom")
	}}
	h := NewHandler(sheetSvc, unusedRoasterService{}, unusedBeanService{}, shotSvc)

	rec := httptest.NewRecorder()
	h.GetSheet(rec, newWebRequest(http.MethodGet, "/sheets/get/1", "", "", "1", false))

	if rec.Code == http.StatusOK {
		t.Errorf("expected a non-200 status when fetching the sheet's shots fails, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetSheet_DetailPageSortsShotsByIDAscending(t *testing.T) {
	sheetSvc := &fakeSheetService{t: t}
	sheetSvc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) { return testSheet(1, "Double shot"), nil }
	shotSvc := shotsBySheetIDStub{getShotsBySheetID: func(context.Context, int) ([]shot.Shot, error) {
		return []shot.Shot{{Id: 9}, {Id: 3}, {Id: 5}}, nil
	}}
	h := NewHandler(sheetSvc, unusedRoasterService{}, unusedBeanService{}, shotSvc)

	rec := httptest.NewRecorder()
	h.GetSheet(rec, newWebRequest(http.MethodGet, "/sheets/get/1", "", "", "1", false))

	body := rec.Body.String()
	i3, i5, i9 := strings.Index(body, `id="shot-row-3"`), strings.Index(body, `id="shot-row-5"`), strings.Index(body, `id="shot-row-9"`)
	if i3 < 0 || i5 < 0 || i9 < 0 {
		t.Fatalf("expected all three shot rows to be rendered, got: %s", body)
	}
	if !(i3 < i5 && i5 < i9) {
		t.Errorf("expected shots sorted by id ascending (3, 5, 9), got order in: %s", body)
	}
}

func TestEditSheetForm_DetailContextSortsShotsByIDAscending(t *testing.T) {
	sheetSvc := &fakeSheetService{t: t}
	sheetSvc.getSheetByID = func(context.Context, int) (*sheet.Sheet, error) { return testSheet(1, "Double shot"), nil }
	shotSvc := shotsBySheetIDStub{getShotsBySheetID: func(context.Context, int) ([]shot.Shot, error) {
		return []shot.Shot{{Id: 9}, {Id: 3}, {Id: 5}}, nil
	}}
	h := NewHandler(sheetSvc, unusedRoasterService{}, unusedBeanService{}, shotSvc)

	rec := httptest.NewRecorder()
	h.EditSheetForm(rec, newWebRequest(http.MethodGet, "/sheets/update/1?view_context=sheet-detail", "", "", "1", false))

	body := rec.Body.String()
	i3, i5, i9 := strings.Index(body, `id="shot-row-3"`), strings.Index(body, `id="shot-row-5"`), strings.Index(body, `id="shot-row-9"`)
	if i3 < 0 || i5 < 0 || i9 < 0 {
		t.Fatalf("expected all three shot rows to be rendered, got: %s", body)
	}
	if !(i3 < i5 && i5 < i9) {
		t.Errorf("expected shots sorted by id ascending (3, 5, 9), got order in: %s", body)
	}
}

func TestUpdateSheet_HappyPathBothViewContexts(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.updateSheetByID = func(_ context.Context, id int, s *sheet.Sheet) (*sheet.Sheet, error) {
		return testSheet(id, s.Name), nil
	}

	listReq := newWebRequest(http.MethodPut, "/sheets/update/1", "name=Renamed", formURLEncoded, "1", true)
	listRec := httptest.NewRecorder()
	h.UpdateSheet(listRec, listReq)
	if listRec.Code != http.StatusOK || !strings.Contains(listRec.Body.String(), `id="sheet-row-1"`) {
		t.Errorf("expected the updated row for sheet-list context, got %d: %s", listRec.Code, listRec.Body.String())
	}

	detailReq := newWebRequest(http.MethodPut, "/sheets/update/1?view_context=sheet-detail", "name=Renamed", formURLEncoded, "1", true)
	detailRec := httptest.NewRecorder()
	h.UpdateSheet(detailRec, detailReq)
	if detailRec.Code != http.StatusOK || !strings.Contains(detailRec.Body.String(), `id="sheet-detail-header"`) {
		t.Errorf("expected the updated header for sheet-detail context, got %d: %s", detailRec.Code, detailRec.Body.String())
	}
	if !strings.Contains(detailRec.Body.String(), "successfully updated") {
		t.Errorf("expected an OOB success alert, got: %s", detailRec.Body.String())
	}
}

func TestUpdateSheet_EmptyNamePreservesSubmittedValueAndError(t *testing.T) {
	h, _ := newTestSheetHandler(t)

	req := newWebRequest(http.MethodPut, "/sheets/update/1", "name=", formURLEncoded, "1", true)
	rec := httptest.NewRecorder()
	h.UpdateSheet(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "must not be empty") {
		t.Errorf("expected inline validation error, got: %s", rec.Body.String())
	}
}

func TestDeleteSheet_ListContextReturnsOOBSuccess(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.deleteSheetByID = func(context.Context, int) error { return nil }

	req := newWebRequest(http.MethodDelete, "/sheets/delete/1", "", "", "1", true)
	rec := httptest.NewRecorder()
	h.DeleteSheet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("HX-Redirect") != "" {
		t.Errorf("did not expect a redirect for the list context, got: %s", rec.Header().Get("HX-Redirect"))
	}
	if !strings.Contains(rec.Body.String(), "successfully deleted") {
		t.Errorf("expected an OOB success alert, got: %s", rec.Body.String())
	}
}

func TestDeleteSheet_DetailContextRedirectsToSheetsList(t *testing.T) {
	h, svc := newTestSheetHandler(t)
	svc.deleteSheetByID = func(context.Context, int) error { return nil }

	req := newWebRequest(http.MethodDelete, "/sheets/delete/1?view_context=sheet-detail", "", "", "1", true)
	rec := httptest.NewRecorder()
	h.DeleteSheet(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("HX-Redirect") != "/sheets" {
		t.Errorf("expected HX-Redirect to /sheets, got %q", rec.Header().Get("HX-Redirect"))
	}
}

func TestDeleteSheet_InvalidIDReturns400WithReswapNone(t *testing.T) {
	h, _ := newTestSheetHandler(t)

	req := newWebRequest(http.MethodDelete, "/sheets/delete/abc", "", "", "abc", true)
	rec := httptest.NewRecorder()
	h.DeleteSheet(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if rec.Header().Get("HX-Reswap") != "none" {
		t.Errorf("expected HX-Reswap: none so the existing row remains, got %q", rec.Header().Get("HX-Reswap"))
	}
}
