package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
)

// fakeBeanService is a hand-rolled fake with func fields, matching the
// pattern used by internal/controllers/rest.
type fakeBeanService struct {
	t              *testing.T
	createBean     func(context.Context, *bean.Bean) (*bean.Bean, error)
	getBeanByID    func(context.Context, int) (*bean.Bean, error)
	getAllBeans    func(context.Context) ([]bean.Bean, error)
	updateBeanByID func(context.Context, int, *bean.Bean) (*bean.Bean, error)
	deleteBeanByID func(context.Context, int) error
}

var _ bean.Service = (*fakeBeanService)(nil)

func (f *fakeBeanService) CreateBean(ctx context.Context, value *bean.Bean) (*bean.Bean, error) {
	if f.createBean == nil {
		f.t.Fatalf("unexpected CreateBean call")
	}
	return f.createBean(ctx, value)
}

func (f *fakeBeanService) GetBeanById(ctx context.Context, id int) (*bean.Bean, error) {
	if f.getBeanByID == nil {
		f.t.Fatalf("unexpected GetBeanById call")
	}
	return f.getBeanByID(ctx, id)
}

func (f *fakeBeanService) GetAllBeans(ctx context.Context) ([]bean.Bean, error) {
	if f.getAllBeans == nil {
		f.t.Fatalf("unexpected GetAllBeans call")
	}
	return f.getAllBeans(ctx)
}

func (f *fakeBeanService) UpdateBeanById(ctx context.Context, id int, value *bean.Bean) (*bean.Bean, error) {
	if f.updateBeanByID == nil {
		f.t.Fatalf("unexpected UpdateBeanById call")
	}
	return f.updateBeanByID(ctx, id, value)
}

func (f *fakeBeanService) DeleteBeanById(ctx context.Context, id int) error {
	if f.deleteBeanByID == nil {
		f.t.Fatalf("unexpected DeleteBeanById call")
	}
	return f.deleteBeanByID(ctx, id)
}

func (f *fakeBeanService) Ping(context.Context) error { return nil }

// fakeRoasterServiceForBeans returns a fixed, non-empty roaster list so bean
// forms can populate the roaster <select>.
type fakeRoasterServiceForBeans struct {
	unusedRoasterService
	roasters []roaster.Roaster
}

func (f fakeRoasterServiceForBeans) GetAllRoasters(context.Context) ([]roaster.Roaster, error) {
	return f.roasters, nil
}

func newTestBeanHandler(t *testing.T, roasters []roaster.Roaster) (*Handler, *fakeBeanService) {
	t.Helper()
	svc := &fakeBeanService{t: t}
	h := NewHandler(unusedSheetService{}, fakeRoasterServiceForBeans{roasters: roasters}, svc, unusedShotService{})
	return h, svc
}

func testBean(id int, name string) *bean.Bean {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	return &bean.Bean{Id: id, Name: name, Roaster: &roaster.Roaster{Id: 1, Name: "Roaster"}, CreatedAt: &created}
}

func TestListBeans_FullPageVsFragment(t *testing.T) {
	h, svc := newTestBeanHandler(t, nil)
	svc.getAllBeans = func(context.Context) ([]bean.Bean, error) { return []bean.Bean{*testBean(1, "Ethiopia")}, nil }

	fullPage := httptest.NewRecorder()
	h.ListBeans(fullPage, newWebRequest(http.MethodGet, "/beans", "", "", "", false))
	if !strings.Contains(fullPage.Body.String(), "<html") || !strings.Contains(fullPage.Body.String(), `id="bean-dialog"`) {
		t.Errorf("expected the full page with the dialog target, got: %s", fullPage.Body.String())
	}

	fragment := httptest.NewRecorder()
	h.ListBeans(fragment, newWebRequest(http.MethodGet, "/beans", "", "", "", true))
	if strings.Contains(fragment.Body.String(), "<html") || !strings.Contains(fragment.Body.String(), `id="beans-table"`) {
		t.Errorf("expected a table fragment only with HX-Request, got: %s", fragment.Body.String())
	}
}

func TestAddBeanForm_EmptyRoastersDisablesSubmit(t *testing.T) {
	h, _ := newTestBeanHandler(t, nil)

	rec := httptest.NewRecorder()
	h.AddBeanForm(rec, newWebRequest(http.MethodGet, "/beans/add", "", "", "", true))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Create one first") {
		t.Errorf("expected a hint to create a roaster first, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAddBeanForm_FullPageFallbackForDirectNavigation(t *testing.T) {
	h, svc := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})
	svc.getAllBeans = func(context.Context) ([]bean.Bean, error) { return []bean.Bean{*testBean(1, "Ethiopia")}, nil }

	rec := httptest.NewRecorder()
	h.AddBeanForm(rec, newWebRequest(http.MethodGet, "/beans/add", "", "", "", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "<html") || !strings.Contains(body, `id="bean-dialog"`) {
		t.Fatalf("expected the full beans page, got %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, "Ethiopia") {
		t.Errorf("expected the beans list to be rendered behind the dialog, got: %s", body)
	}
	if !strings.Contains(body, `name="name"`) {
		t.Errorf("expected the add form to be pre-populated inside the dialog, got: %s", body)
	}
}

func TestCreateBean_HappyPath(t *testing.T) {
	h, svc := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})
	svc.createBean = func(_ context.Context, b *bean.Bean) (*bean.Bean, error) {
		if b.Roaster == nil || b.Roaster.Id != 1 {
			t.Errorf("expected roaster id 1, got %#v", b.Roaster)
		}
		return testBean(5, b.Name), nil
	}

	req := newWebRequest(http.MethodPost, "/beans/add", "name=Ethiopia&roaster_id=1&roast_level=2", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateBean(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("HX-Reswap") != "none" || rec.Header().Get("HX-Trigger") != "dialog-close" {
		t.Errorf("expected HX-Reswap: none and HX-Trigger: dialog-close, got %v", rec.Header())
	}
	if !strings.Contains(rec.Body.String(), "Ethiopia") || !strings.Contains(rec.Body.String(), `hx-swap-oob="beforeend:#beans-tbody"`) {
		t.Errorf("expected the new row inserted OOB, got: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "successfully created") {
		t.Errorf("expected an OOB success alert, got: %s", rec.Body.String())
	}
}

func TestCreateBean_MissingRoasterReturns400WithFieldError(t *testing.T) {
	h, _ := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})

	req := newWebRequest(http.MethodPost, "/beans/add", "name=Ethiopia&roaster_id=", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateBean(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Select a roaster") {
		t.Errorf("expected inline roaster field error, got: %s", rec.Body.String())
	}
}

func TestCreateBean_InvalidRoastDateReturns400(t *testing.T) {
	h, _ := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})

	req := newWebRequest(http.MethodPost, "/beans/add", "name=Ethiopia&roaster_id=1&roast_date=not-a-date", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateBean(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "valid date") {
		t.Errorf("expected inline roast date error, got: %s", rec.Body.String())
	}
}

func TestCreateBean_DuplicateReturns409(t *testing.T) {
	h, svc := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})
	svc.createBean = func(context.Context, *bean.Bean) (*bean.Bean, error) { return nil, errors.ErrBeansAlreadyExists }

	req := newWebRequest(http.MethodPost, "/beans/add", "name=Ethiopia&roaster_id=1", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateBean(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
}

func TestCreateBean_WrongContentTypeReturns415(t *testing.T) {
	h, _ := newTestBeanHandler(t, nil)

	req := newWebRequest(http.MethodPost, "/beans/add", `{"name":"Ethiopia"}`, "application/json", "", true)
	rec := httptest.NewRecorder()
	h.CreateBean(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestGetBean_InvalidIDReturns400(t *testing.T) {
	h, _ := newTestBeanHandler(t, nil)
	rec := httptest.NewRecorder()
	h.GetBean(rec, newWebRequest(http.MethodGet, "/beans/get/abc", "", "", "abc", false))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetBean_UnknownIDReturns404(t *testing.T) {
	h, svc := newTestBeanHandler(t, nil)
	svc.getBeanByID = func(context.Context, int) (*bean.Bean, error) { return nil, errors.ErrBeansDoesNotExist }

	rec := httptest.NewRecorder()
	h.GetBean(rec, newWebRequest(http.MethodGet, "/beans/get/9", "", "", "9", true))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetBean_FullPageFallbackForDirectNavigation(t *testing.T) {
	h, svc := newTestBeanHandler(t, nil)
	svc.getBeanByID = func(context.Context, int) (*bean.Bean, error) { return testBean(9, "Ethiopia"), nil }

	rec := httptest.NewRecorder()
	h.GetBean(rec, newWebRequest(http.MethodGet, "/beans/get/9", "", "", "9", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "<html") || !strings.Contains(body, "Ethiopia") {
		t.Fatalf("expected a full page with the bean row, got %d: %s", rec.Code, body)
	}
}

func TestEditBeanForm_PrefillsExistingValues(t *testing.T) {
	h, svc := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})
	svc.getBeanByID = func(context.Context, int) (*bean.Bean, error) { return testBean(9, "Ethiopia"), nil }

	rec := httptest.NewRecorder()
	h.EditBeanForm(rec, newWebRequest(http.MethodGet, "/beans/update/9", "", "", "9", true))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Ethiopia") {
		t.Errorf("expected the prefilled form, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEditBeanForm_FullPageFallbackForDirectNavigation(t *testing.T) {
	h, svc := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})
	svc.getBeanByID = func(context.Context, int) (*bean.Bean, error) { return testBean(9, "Ethiopia"), nil }
	svc.getAllBeans = func(context.Context) ([]bean.Bean, error) { return []bean.Bean{*testBean(9, "Ethiopia")}, nil }

	rec := httptest.NewRecorder()
	h.EditBeanForm(rec, newWebRequest(http.MethodGet, "/beans/update/9", "", "", "9", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "<html") || !strings.Contains(body, `id="bean-dialog"`) {
		t.Fatalf("expected the full beans page, got %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, `value="Ethiopia"`) {
		t.Errorf("expected the edit form to be pre-filled inside the dialog, got: %s", body)
	}
}

func TestUpdateBean_HappyPath(t *testing.T) {
	h, svc := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})
	svc.updateBeanByID = func(_ context.Context, id int, b *bean.Bean) (*bean.Bean, error) {
		return testBean(id, b.Name), nil
	}

	req := newWebRequest(http.MethodPut, "/beans/update/9", "name=Renamed&roaster_id=1", formURLEncoded, "9", true)
	rec := httptest.NewRecorder()
	h.UpdateBean(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Renamed") {
		t.Errorf("expected the updated row, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `hx-swap-oob="true"`) {
		t.Errorf("expected the row to replace by id OOB, got: %s", rec.Body.String())
	}
	if rec.Header().Get("HX-Trigger") != "dialog-close" {
		t.Errorf("expected HX-Trigger: dialog-close, got %v", rec.Header())
	}
}

func TestUpdateBean_EmptyNameReturns400(t *testing.T) {
	h, _ := newTestBeanHandler(t, []roaster.Roaster{{Id: 1, Name: "Roaster"}})

	req := newWebRequest(http.MethodPut, "/beans/update/9", "name=&roaster_id=1", formURLEncoded, "9", true)
	rec := httptest.NewRecorder()
	h.UpdateBean(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestDeleteBean_HappyPath(t *testing.T) {
	h, svc := newTestBeanHandler(t, nil)
	svc.deleteBeanByID = func(context.Context, int) error { return nil }

	req := newWebRequest(http.MethodDelete, "/beans/delete/9", "", "", "9", true)
	rec := httptest.NewRecorder()
	h.DeleteBean(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "successfully deleted") {
		t.Errorf("expected 200 with an OOB success alert, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteBean_ForeignKeyViolationReturns409(t *testing.T) {
	h, svc := newTestBeanHandler(t, nil)
	svc.deleteBeanByID = func(context.Context, int) error { return errors.ErrShotForeignKeyConstraint }

	req := newWebRequest(http.MethodDelete, "/beans/delete/9", "", "", "9", true)
	rec := httptest.NewRecorder()
	h.DeleteBean(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d", rec.Code)
	}
	if rec.Header().Get("HX-Reswap") != "none" {
		t.Errorf("expected HX-Reswap: none, got %q", rec.Header().Get("HX-Reswap"))
	}
}
