package shots

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), &b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func testShot() shot.Shot {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	return shot.Shot{
		Id:                           5,
		Sheet:                        &sheet.Sheet{Id: 1, Name: "Morning"},
		Beans:                        &bean.Bean{Id: 2, Name: "Ethiopia", Roaster: &roaster.Roaster{Id: 3, Name: "Blue Bottle"}},
		GrindSetting:                 12,
		QuantityIn:                   18,
		QuantityOut:                  36,
		ShotTime:                     28500 * time.Millisecond,
		WaterTemperature:             93,
		Rating:                       8.5,
		IsTooBitter:                  false,
		IsTooSour:                    true,
		ComparisonWithPreviousResult: sql.Better,
		AdditionalNotes:              "great shot",
		CreatedAt:                    &created,
	}
}

func TestRow_ShowsEveryPersistedField(t *testing.T) {
	html := render(t, Row(testShot(), true, ""))

	for _, want := range []string{
		"5", "Morning", "Ethiopia", "12", "18.0", "36.0", "28.5 s", "93.0", "8.5",
		"No", "Yes", "Better", "great shot", "2026-01-02 03:04",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("expected row to contain %q, got: %s", want, html)
		}
	}
}

func TestRow_HidesSheetColumnWhenRequested(t *testing.T) {
	html := render(t, Row(testShot(), false, ""))
	if strings.Contains(html, "Morning") {
		t.Errorf("expected the sheet column to be hidden, got: %s", html)
	}
}

func TestRow_SheetAndBeansCellsShowIDAndName(t *testing.T) {
	html := render(t, Row(testShot(), true, ""))

	if !strings.Contains(html, "#1 Morning") {
		t.Errorf("expected the sheet cell to show its id and name, got: %s", html)
	}
	if !strings.Contains(html, `<a href="/sheets/get/1">#1 Morning</a>`) {
		t.Errorf("expected the sheet cell to link to its detail page, got: %s", html)
	}
	if !strings.Contains(html, "#2 Ethiopia") {
		t.Errorf("expected the beans cell to show its id and name, got: %s", html)
	}
}

func TestRowPage_IncludesDialogTargetForEditLink(t *testing.T) {
	html := render(t, RowPage(testShot()))

	if !strings.Contains(html, "<html") || !strings.Contains(html, "<table") {
		t.Errorf("expected a full page with a one-row table, got: %s", html)
	}
	if !strings.Contains(html, `id="shot-dialog"`) {
		t.Errorf("expected the row's Edit link (hx-target=\"#shot-dialog\") to have a matching dialog target, got: %s", html)
	}
}

func TestRow_EditLinkCarriesViewContextOnlyWhenSheetColumnHidden(t *testing.T) {
	withSheetColumn := render(t, Row(testShot(), true, ""))
	if !strings.Contains(withSheetColumn, `hx-get="/shots/update/5"`) {
		t.Errorf("expected a plain edit link when the sheet column is shown, got: %s", withSheetColumn)
	}
	if strings.Contains(withSheetColumn, "view_context") {
		t.Errorf("expected no view_context on the standalone list's edit link, got: %s", withSheetColumn)
	}

	withoutSheetColumn := render(t, Row(testShot(), false, ""))
	if !strings.Contains(withoutSheetColumn, `hx-get="/shots/update/5?view_context=sheet-shots"`) {
		t.Errorf("expected the edit link to carry view_context=sheet-shots, got: %s", withoutSheetColumn)
	}
}

func TestRow_DeleteConfirmUsesShotID(t *testing.T) {
	html := render(t, Row(testShot(), true, ""))
	if !strings.Contains(html, `hx-confirm="Are you sure you want to delete shot #5?"`) {
		t.Errorf("expected the delete confirm to reference the shot id, got: %s", html)
	}
}

func TestForm_SheetLockedRendersHiddenFieldAndReadOnlyName(t *testing.T) {
	state := FormState{SheetLocked: true, SheetID: "1", SheetName: "Morning"}
	html := render(t, Form(state, nil, []bean.Bean{{Id: 2, Name: "Ethiopia"}}, true, "", ""))

	if !strings.Contains(html, `type="hidden" name="sheet_id" value="1"`) {
		t.Errorf("expected a hidden sheet_id field, got: %s", html)
	}
	if strings.Contains(html, `<select name="sheet_id"`) {
		t.Errorf("expected no sheet select when locked, got: %s", html)
	}
	if !strings.Contains(html, "Morning") {
		t.Errorf("expected the locked sheet name to be displayed, got: %s", html)
	}
}

func TestForm_UnlockedShowsSheetSelectWithHintWhenEmpty(t *testing.T) {
	html := render(t, Form(FormState{}, nil, nil, true, "", ""))
	if !strings.Contains(html, "Create one first") {
		t.Errorf("expected hints for empty sheets/beans, got: %s", html)
	}
	if !strings.Contains(html, "disabled") {
		t.Errorf("expected submit to be disabled, got: %s", html)
	}
}

func TestForm_EditModeShowsReadOnlyMetadata(t *testing.T) {
	state := FormState{ID: 5, SheetID: "1", BeansID: "2"}
	html := render(t, Form(state, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}}, false, "2026-01-02 03:04", "2026-01-05 06:07"))

	if !strings.Contains(html, "2026-01-02 03:04") || !strings.Contains(html, "2026-01-05 06:07") {
		t.Errorf("expected read-only metadata, got: %s", html)
	}
}

func TestForm_ShowsInlineFieldErrors(t *testing.T) {
	state := FormState{Errors: map[string]string{"rating": "Rating must be between 0 and 10."}}
	html := render(t, Form(state, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}}, true, "", ""))

	if !strings.Contains(html, "Rating must be between 0 and 10.") {
		t.Errorf("expected the inline rating error, got: %s", html)
	}
}

func TestForm_ViewContextRendersHiddenFieldWhenSet(t *testing.T) {
	state := FormState{ID: 5, ViewContext: ViewContextSheetShots}
	html := render(t, Form(state, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}}, false, "", ""))

	if !strings.Contains(html, `type="hidden" name="view_context" value="sheet-shots"`) {
		t.Errorf("expected a hidden view_context field, got: %s", html)
	}
}

func TestForm_ViewContextOmittedWhenUnset(t *testing.T) {
	html := render(t, Form(FormState{ID: 5}, []sheet.Sheet{{Id: 1, Name: "Morning"}}, []bean.Bean{{Id: 2, Name: "Ethiopia"}}, false, "", ""))
	if strings.Contains(html, "view_context") {
		t.Errorf("expected no view_context field when unset, got: %s", html)
	}
}
