package web

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
)

// fakeShotServiceForWeb is a hand-rolled fake with func fields, matching the
// pattern used by internal/controllers/rest.
type fakeShotServiceForWeb struct {
	t                 *testing.T
	createShot        func(context.Context, *shot.Shot) (*shot.Shot, error)
	getShotByID       func(context.Context, int) (*shot.Shot, error)
	getAllShots       func(context.Context) ([]shot.Shot, error)
	getShotsBySheetID func(context.Context, int) ([]shot.Shot, error)
	updateShotByID    func(context.Context, int, *shot.Shot) (*shot.Shot, error)
	deleteShotByID    func(context.Context, int) error
}

var _ shot.Service = (*fakeShotServiceForWeb)(nil)

func (f *fakeShotServiceForWeb) CreateShot(ctx context.Context, value *shot.Shot) (*shot.Shot, error) {
	if f.createShot == nil {
		f.t.Fatalf("unexpected CreateShot call")
	}
	return f.createShot(ctx, value)
}

func (f *fakeShotServiceForWeb) GetShotById(ctx context.Context, id int) (*shot.Shot, error) {
	if f.getShotByID == nil {
		f.t.Fatalf("unexpected GetShotById call")
	}
	return f.getShotByID(ctx, id)
}

func (f *fakeShotServiceForWeb) GetAllShots(ctx context.Context) ([]shot.Shot, error) {
	if f.getAllShots == nil {
		f.t.Fatalf("unexpected GetAllShots call")
	}
	return f.getAllShots(ctx)
}

func (f *fakeShotServiceForWeb) GetShotsBySheetId(ctx context.Context, sheetId int) ([]shot.Shot, error) {
	if f.getShotsBySheetID == nil {
		f.t.Fatalf("unexpected GetShotsBySheetId call")
	}
	return f.getShotsBySheetID(ctx, sheetId)
}

func (f *fakeShotServiceForWeb) UpdateShotById(ctx context.Context, id int, value *shot.Shot) (*shot.Shot, error) {
	if f.updateShotByID == nil {
		f.t.Fatalf("unexpected UpdateShotById call")
	}
	return f.updateShotByID(ctx, id, value)
}

func (f *fakeShotServiceForWeb) DeleteShotById(ctx context.Context, id int) error {
	if f.deleteShotByID == nil {
		f.t.Fatalf("unexpected DeleteShotById call")
	}
	return f.deleteShotByID(ctx, id)
}

func (f *fakeShotServiceForWeb) Ping(context.Context) error { return nil }

// fakeSheetServiceForShots and fakeBeanServiceForShots return fixed,
// non-empty lists so shot forms can populate their <select> options.
type fakeSheetServiceForShots struct {
	unusedSheetService
	sheets []sheet.Sheet
}

func (f fakeSheetServiceForShots) GetAllSheets(context.Context) ([]sheet.Sheet, error) {
	return f.sheets, nil
}

func (f fakeSheetServiceForShots) GetSheetById(_ context.Context, id int) (*sheet.Sheet, error) {
	for _, s := range f.sheets {
		if s.Id == id {
			return &s, nil
		}
	}
	return nil, errors.ErrSheetDoesNotExist
}

type fakeBeanServiceForShots struct {
	unusedBeanService
	beans []bean.Bean
}

func (f fakeBeanServiceForShots) GetAllBeans(context.Context) ([]bean.Bean, error) {
	return f.beans, nil
}

func newTestShotHandler(t *testing.T, sheets []sheet.Sheet, beans []bean.Bean) (*Handler, *fakeShotServiceForWeb) {
	t.Helper()
	svc := &fakeShotServiceForWeb{t: t}
	h := NewHandler(fakeSheetServiceForShots{sheets: sheets}, unusedRoasterService{}, fakeBeanServiceForShots{beans: beans}, svc)
	return h, svc
}

func testShot(id int) *shot.Shot {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	return &shot.Shot{
		Id:               id,
		Sheet:            &sheet.Sheet{Id: 1, Name: "Morning"},
		Beans:            &bean.Bean{Id: 2, Name: "Ethiopia"},
		GrindSetting:     12,
		QuantityIn:       18,
		QuantityOut:      36,
		ShotTime:         28 * time.Second,
		WaterTemperature: 93,
		Rating:           8.5,
		CreatedAt:        &created,
	}
}

const validShotForm = "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&shot_time=28.5&rating=8.5&comparison_with_previous_result=0"

func TestListShots_FullPageVsFragment(t *testing.T) {
	h, svc := newTestShotHandler(t, nil, nil)
	svc.getAllShots = func(context.Context) ([]shot.Shot, error) { return []shot.Shot{*testShot(1)}, nil }

	fullPage := httptest.NewRecorder()
	h.ListShots(fullPage, newWebRequest(http.MethodGet, "/shots", "", "", "", false))
	if !strings.Contains(fullPage.Body.String(), "<html") || !strings.Contains(fullPage.Body.String(), `id="shot-dialog"`) {
		t.Errorf("expected the full page with the dialog target, got: %s", fullPage.Body.String())
	}

	fragment := httptest.NewRecorder()
	h.ListShots(fragment, newWebRequest(http.MethodGet, "/shots", "", "", "", true))
	if strings.Contains(fragment.Body.String(), "<html") || !strings.Contains(fragment.Body.String(), `id="shots-table"`) {
		t.Errorf("expected a table fragment only with HX-Request, got: %s", fragment.Body.String())
	}
}

func TestAddShotForm_LocksSheetWhenQueryParamGiven(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	rec := httptest.NewRecorder()
	h.AddShotForm(rec, newWebRequest(http.MethodGet, "/shots/add?sheet_id=1", "", "", "", true))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `type="hidden" name="sheet_id" value="1"`) {
		t.Errorf("expected the sheet to be locked via a hidden field, got: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Morning") {
		t.Errorf("expected the locked sheet name, got: %s", rec.Body.String())
	}
}

func TestAddShotForm_NoLockShowsSelect(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	rec := httptest.NewRecorder()
	h.AddShotForm(rec, newWebRequest(http.MethodGet, "/shots/add", "", "", "", true))

	if !strings.Contains(rec.Body.String(), `<select name="sheet_id"`) {
		t.Errorf("expected an unlocked sheet select, got: %s", rec.Body.String())
	}
}

func TestAddShotForm_FullPageFallbackForDirectNavigation(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.getAllShots = func(context.Context) ([]shot.Shot, error) { return []shot.Shot{*testShot(1)}, nil }

	rec := httptest.NewRecorder()
	h.AddShotForm(rec, newWebRequest(http.MethodGet, "/shots/add", "", "", "", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "<html") || !strings.Contains(body, `id="shot-dialog"`) {
		t.Fatalf("expected the full shots page, got %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, `<select name="sheet_id"`) {
		t.Errorf("expected the add form to be rendered inside the dialog, got: %s", body)
	}
}

func TestAddShotForm_SheetLockedFullPageFallbackOmitsSheetShotsViewContext(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.getAllShots = func(context.Context) ([]shot.Shot, error) { return []shot.Shot{*testShot(1)}, nil }

	rec := httptest.NewRecorder()
	h.AddShotForm(rec, newWebRequest(http.MethodGet, "/shots/add?sheet_id=1", "", "", "", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `id="shot-dialog"`) {
		t.Fatalf("expected the full shots page, got %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, `type="hidden" name="sheet_id" value="1"`) {
		t.Errorf("expected the sheet lock to still apply, got: %s", body)
	}
	if strings.Contains(body, `name="view_context"`) {
		t.Errorf("expected no view_context field on the standalone-page fallback (would misalign the OOB row against its 16-column table), got: %s", body)
	}
}

func TestCreateShot_HappyPath(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(_ context.Context, s *shot.Shot) (*shot.Shot, error) {
		if s.ShotTime != 28500*time.Millisecond {
			t.Errorf("expected 28.5s converted to a duration, got %v", s.ShotTime)
		}
		return testShot(5), nil
	}

	req := newWebRequest(http.MethodPost, "/shots/add", validShotForm, formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("HX-Reswap") != "none" || rec.Header().Get("HX-Trigger") != "dialog-close" {
		t.Errorf("expected HX-Reswap: none and HX-Trigger: dialog-close, got %v", rec.Header())
	}
	if !strings.Contains(rec.Body.String(), `hx-swap-oob="beforeend:#shots-tbody"`) {
		t.Errorf("expected the new row inserted OOB, got: %s", rec.Body.String())
	}
}

func TestCreateShot_SheetShotsContextOmitsSheetColumnInOOBRow(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) { return testShot(5), nil }

	req := newWebRequest(http.MethodPost, "/shots/add", validShotForm+"&view_context=sheet-shots", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "Morning") {
		t.Errorf("expected the OOB row to omit the sheet column on the sheet detail page, got: %s", rec.Body.String())
	}
}

func TestCreateShot_DefaultContextIncludesSheetColumnInOOBRow(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) { return testShot(5), nil }

	req := newWebRequest(http.MethodPost, "/shots/add", validShotForm, formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if !strings.Contains(rec.Body.String(), "Morning") {
		t.Errorf("expected the OOB row to include the sheet column on the standalone /shots list, got: %s", rec.Body.String())
	}
}

func TestCreateShot_MissingSheetReturns400(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&rating=8.5", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "Select a sheet") {
		t.Errorf("expected 400 with a sheet field error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateShot_NegativeQuantityReturns400(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=-1&quantity_out=36&rating=8.5", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "non-negative") {
		t.Errorf("expected 400 with a quantity field error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateShot_NonFiniteNumericFieldsReturn400(t *testing.T) {
	cases := []struct {
		name  string
		form  string
		field string
	}{
		{"quantity_in NaN", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=NaN&quantity_out=36&rating=8.5&comparison_with_previous_result=0", "quantity_in"},
		{"quantity_in Inf", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=Inf&quantity_out=36&rating=8.5&comparison_with_previous_result=0", "quantity_in"},
		{"quantity_out NaN", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=NaN&rating=8.5&comparison_with_previous_result=0", "quantity_out"},
		{"shot_time NaN", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&shot_time=NaN&rating=8.5&comparison_with_previous_result=0", "shot_time"},
		{"water_temperature NaN", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&water_temperature=NaN&rating=8.5&comparison_with_previous_result=0", "water_temperature"},
		{"water_temperature Inf", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&water_temperature=Inf&rating=8.5&comparison_with_previous_result=0", "water_temperature"},
		{"rating NaN", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&rating=NaN&comparison_with_previous_result=0", "rating"},
		{"rating Inf", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&rating=Inf&comparison_with_previous_result=0", "rating"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
			svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) {
				t.Fatalf("service must not be called for a non-finite %s", tc.field)
				return nil, nil
			}

			req := newWebRequest(http.MethodPost, "/shots/add", tc.form, formURLEncoded, "", true)
			rec := httptest.NewRecorder()
			h.CreateShot(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreateShot_ShotTimeOutOfRangeMapsToShotTimeField(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) {
		return nil, errors.ErrShotTimeOutOfRange
	}

	// 999999 seconds parses fine (it's a valid number); range enforcement is
	// deferred to the service, matching rating/comparison.
	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&shot_time=999999&rating=8.5&comparison_with_previous_result=0", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	body := rec.Body.String()
	shotTimeIdx := strings.Index(body, `name="shot_time"`)
	msgIdx := strings.Index(body, "Shot time must be between 0 and 3600 seconds.")
	if rec.Code != http.StatusBadRequest || shotTimeIdx < 0 || msgIdx < 0 || !(shotTimeIdx < msgIdx) {
		t.Errorf("expected the shot-time-out-of-range error rendered under the shot_time field, got %d: %s", rec.Code, body)
	}
}

func TestCreateShot_MissingComparisonReturns400WithFieldError(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&rating=8.5", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "Select a comparison value") {
		t.Errorf("expected 400 with a comparison field error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateShot_OutOfRangeComparisonReturns400InsteadOfWrapping(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) {
		t.Fatalf("service must not be called for an out-of-range comparison value")
		return nil, nil
	}

	// 256 wraps to 0 (uint8) if converted without a range check.
	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&rating=8.5&comparison_with_previous_result=256", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "Invalid comparison value") {
		t.Errorf("expected 400 with an invalid comparison error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateShot_SheetDoesNotExistDomainErrorMapsToSheetField(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) {
		return nil, errors.ErrSheetDoesNotExist
	}

	req := newWebRequest(http.MethodPost, "/shots/add", validShotForm, formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	body := rec.Body.String()
	sheetIdx := strings.Index(body, `name="sheet_id"`)
	msgIdx := strings.Index(body, "No sheet found for the given id.")
	if sheetIdx < 0 || msgIdx < 0 || !(sheetIdx < msgIdx) {
		t.Errorf("expected the sheet-not-found error rendered under the sheet field, got: %s", body)
	}
}

func TestCreateShot_BeansDoesNotExistDomainErrorMapsToBeansField(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) {
		return nil, errors.ErrBeansDoesNotExist
	}

	req := newWebRequest(http.MethodPost, "/shots/add", validShotForm, formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	body := rec.Body.String()
	beansIdx := strings.Index(body, `name="beans_id"`)
	msgIdx := strings.Index(body, "No beans found for the given id.")
	if beansIdx < 0 || msgIdx < 0 || !(beansIdx < msgIdx) {
		t.Errorf("expected the beans-not-found error rendered under the beans field, got: %s", body)
	}
}

func TestCreateShot_ComparisonOutOfRangeDomainErrorMapsToComparisonField(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) {
		return nil, errors.ErrShotComparisonWithPreviousResultOutOfRange
	}

	req := newWebRequest(http.MethodPost, "/shots/add", validShotForm, formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	body := rec.Body.String()
	comparisonIdx := strings.Index(body, `name="comparison_with_previous_result"`)
	msgIdx := strings.Index(body, "Invalid comparison value.")
	if comparisonIdx < 0 || msgIdx < 0 || !(comparisonIdx < msgIdx) {
		t.Errorf("expected the comparison-out-of-range error rendered under the comparison field, got: %s", body)
	}
}

func TestCreateShot_RatingOutOfRangeDeferredToService(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) {
		return nil, errors.ErrShotRatingOutOfRange
	}

	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=1&beans_id=2&grind_setting=12&quantity_in=18&quantity_out=36&rating=11&comparison_with_previous_result=0", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 (mapped from the service's domain error), got %d", rec.Code)
	}
}

func TestCreateShot_WrongContentTypeReturns415(t *testing.T) {
	h, _ := newTestShotHandler(t, nil, nil)

	req := newWebRequest(http.MethodPost, "/shots/add", `{"rating":8}`, "application/json", "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Fatalf("expected 415, got %d", rec.Code)
	}
}

func TestCreateShot_MalformedFormBodyReturns400(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=%zz", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a malformed body, got %d", rec.Code)
	}
}

func TestCreateShot_MalformedFormBodyShowsGeneralFormErrorNotOnRatingField(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	req := newWebRequest(http.MethodPost, "/shots/add", "sheet_id=%zz", formURLEncoded, "", true)
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	body := rec.Body.String()
	if !strings.Contains(body, `<p role="alert">Invalid form submission.</p>`) {
		t.Errorf("expected the malformed-body error in the general form error slot, got: %s", body)
	}
	if strings.Contains(body, "aria-invalid") {
		t.Errorf("expected no field to be marked invalid for a body-level error, got: %s", body)
	}
}

func TestCreateShot_OversizedBodyReturns413(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	req := newOversizedWebRequest(http.MethodPost, "/shots/add", formURLEncoded, "")
	rec := httptest.NewRecorder()
	h.CreateShot(rec, req)

	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 for an oversized body, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateShot_UnexpectedDomainErrorShowsGeneralFormError(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.createShot = func(context.Context, *shot.Shot) (*shot.Shot, error) { return nil, stderrors.New("boom") }

	rec := httptest.NewRecorder()
	h.CreateShot(rec, newWebRequest(http.MethodPost, "/shots/add", validShotForm, formURLEncoded, "", true))

	body := rec.Body.String()
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, `<p role="alert">Something went wrong. Please try again.</p>`) {
		t.Errorf("expected the unexpected error in the general form error slot, got: %s", body)
	}
}

func TestGetShot_InvalidIDReturns400(t *testing.T) {
	h, _ := newTestShotHandler(t, nil, nil)
	rec := httptest.NewRecorder()
	h.GetShot(rec, newWebRequest(http.MethodGet, "/shots/get/abc", "", "", "abc", false))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestGetShot_UnknownIDReturns404(t *testing.T) {
	h, svc := newTestShotHandler(t, nil, nil)
	svc.getShotByID = func(context.Context, int) (*shot.Shot, error) { return nil, errors.ErrShotDoesNotExist }

	rec := httptest.NewRecorder()
	h.GetShot(rec, newWebRequest(http.MethodGet, "/shots/get/9", "", "", "9", true))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestGetShot_FullPageFallbackForDirectNavigation(t *testing.T) {
	h, svc := newTestShotHandler(t, nil, nil)
	svc.getShotByID = func(context.Context, int) (*shot.Shot, error) { return testShot(5), nil }

	rec := httptest.NewRecorder()
	h.GetShot(rec, newWebRequest(http.MethodGet, "/shots/get/5", "", "", "5", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "<html") || !strings.Contains(body, "Ethiopia") {
		t.Fatalf("expected a full page with the shot row, got %d: %s", rec.Code, body)
	}
}

func TestEditShotForm_PrefillsSecondsFromDuration(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.getShotByID = func(context.Context, int) (*shot.Shot, error) { return testShot(5), nil }

	rec := httptest.NewRecorder()
	h.EditShotForm(rec, newWebRequest(http.MethodGet, "/shots/update/5", "", "", "5", true))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `value="28.0"`) {
		t.Errorf("expected shot_time prefilled as seconds, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEditShotForm_ReadsViewContextFromQueryParam(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.getShotByID = func(context.Context, int) (*shot.Shot, error) { return testShot(5), nil }

	rec := httptest.NewRecorder()
	h.EditShotForm(rec, newWebRequest(http.MethodGet, "/shots/update/5?view_context=sheet-shots", "", "", "5", true))

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `type="hidden" name="view_context" value="sheet-shots"`) {
		t.Errorf("expected the view_context to round-trip as a hidden field, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEditShotForm_FullPageFallbackForDirectNavigation(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.getShotByID = func(context.Context, int) (*shot.Shot, error) { return testShot(5), nil }
	svc.getAllShots = func(context.Context) ([]shot.Shot, error) { return []shot.Shot{*testShot(5)}, nil }

	rec := httptest.NewRecorder()
	h.EditShotForm(rec, newWebRequest(http.MethodGet, "/shots/update/5", "", "", "5", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, "<html") || !strings.Contains(body, `id="shot-dialog"`) {
		t.Fatalf("expected the full shots page, got %d: %s", rec.Code, body)
	}
	if !strings.Contains(body, `value="28.0"`) {
		t.Errorf("expected the edit form to be pre-filled inside the dialog, got: %s", body)
	}
}

func TestEditShotForm_FullPageFallbackOmitsSheetShotsViewContextEvenIfRequested(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.getShotByID = func(context.Context, int) (*shot.Shot, error) { return testShot(5), nil }
	svc.getAllShots = func(context.Context) ([]shot.Shot, error) { return []shot.Shot{*testShot(5)}, nil }

	rec := httptest.NewRecorder()
	h.EditShotForm(rec, newWebRequest(http.MethodGet, "/shots/update/5?view_context=sheet-shots", "", "", "5", false))

	body := rec.Body.String()
	if rec.Code != http.StatusOK || !strings.Contains(body, `id="shot-dialog"`) {
		t.Fatalf("expected the full shots page, got %d: %s", rec.Code, body)
	}
	if strings.Contains(body, `name="view_context"`) {
		t.Errorf("expected no view_context field on the standalone-page fallback (would misalign the OOB row against its 16-column table), got: %s", body)
	}
}

func TestUpdateShot_HappyPath(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.updateShotByID = func(_ context.Context, id int, s *shot.Shot) (*shot.Shot, error) {
		return testShot(id), nil
	}

	req := newWebRequest(http.MethodPut, "/shots/update/5", validShotForm, formURLEncoded, "5", true)
	rec := httptest.NewRecorder()
	h.UpdateShot(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `hx-swap-oob="true"`) {
		t.Errorf("expected the updated row to replace by id OOB, got %d: %s", rec.Code, rec.Body.String())
	}
	if rec.Header().Get("HX-Trigger") != "dialog-close" {
		t.Errorf("expected HX-Trigger: dialog-close, got %v", rec.Header())
	}
}

func TestUpdateShot_SheetShotsContextOmitsSheetColumnInOOBRow(t *testing.T) {
	h, svc := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})
	svc.updateShotByID = func(_ context.Context, id int, s *shot.Shot) (*shot.Shot, error) {
		return testShot(id), nil
	}

	req := newWebRequest(http.MethodPut, "/shots/update/5", validShotForm+"&view_context=sheet-shots", formURLEncoded, "5", true)
	rec := httptest.NewRecorder()
	h.UpdateShot(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "Morning") {
		t.Errorf("expected the OOB row to omit the sheet column on the sheet detail page, got: %s", rec.Body.String())
	}
}

func TestUpdateShot_InvalidGrindSettingReturns400(t *testing.T) {
	h, _ := newTestShotHandler(t, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}})

	req := newWebRequest(http.MethodPut, "/shots/update/5", "sheet_id=1&beans_id=2&grind_setting=abc&quantity_in=18&quantity_out=36&rating=8.5", formURLEncoded, "5", true)
	rec := httptest.NewRecorder()
	h.UpdateShot(rec, req)

	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "whole number") {
		t.Errorf("expected 400 with a grind_setting field error, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteShot_HappyPath(t *testing.T) {
	h, svc := newTestShotHandler(t, nil, nil)
	svc.deleteShotByID = func(context.Context, int) error { return nil }

	req := newWebRequest(http.MethodDelete, "/shots/delete/5", "", "", "5", true)
	rec := httptest.NewRecorder()
	h.DeleteShot(rec, req)

	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "successfully deleted") {
		t.Errorf("expected 200 with an OOB success alert, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestDeleteShot_InvalidIDReturns400WithReswapNone(t *testing.T) {
	h, _ := newTestShotHandler(t, nil, nil)

	req := newWebRequest(http.MethodDelete, "/shots/delete/abc", "", "", "abc", true)
	rec := httptest.NewRecorder()
	h.DeleteShot(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
	if rec.Header().Get("HX-Reswap") != "none" {
		t.Errorf("expected HX-Reswap: none, got %q", rec.Header().Get("HX-Reswap"))
	}
}
