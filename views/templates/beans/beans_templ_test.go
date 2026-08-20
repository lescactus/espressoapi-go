package beans

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), &b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func testBean() bean.Bean {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	roastDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	return bean.Bean{
		Id:         9,
		Name:       "Ethiopia Yirgacheffe",
		Roaster:    &roaster.Roaster{Id: 3, Name: "Blue Bottle"},
		RoastDate:  &roastDate,
		RoastLevel: sql.RoastLevelMedium,
		CreatedAt:  &created,
	}
}

func TestRow_ShowsAllColumnsAndHumanRoastLevel(t *testing.T) {
	html := render(t, Row(testBean(), ""))

	for _, want := range []string{"9", "Ethiopia Yirgacheffe", "Blue Bottle", "2026-01-01", "Medium", "Edit", "Delete"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected row to contain %q, got: %s", want, html)
		}
	}
}

func TestRow_RoasterCellShowsIDAndName(t *testing.T) {
	html := render(t, Row(testBean(), ""))

	if !strings.Contains(html, "#3 Blue Bottle") {
		t.Errorf("expected the roaster cell to show its id and name, got: %s", html)
	}
}

func TestRow_OOBModes(t *testing.T) {
	insert := render(t, Row(testBean(), "insert"))
	if !strings.Contains(insert, `hx-swap-oob="beforeend:#beans-tbody"`) {
		t.Errorf("expected insert oob attribute, got: %s", insert)
	}

	replace := render(t, Row(testBean(), "replace"))
	if !strings.Contains(replace, `hx-swap-oob="true"`) {
		t.Errorf("expected replace oob attribute, got: %s", replace)
	}

	plain := render(t, Row(testBean(), ""))
	if strings.Contains(plain, "hx-swap-oob") {
		t.Errorf("expected no oob attribute for a normal row, got: %s", plain)
	}
}

func TestForm_EmptyRoastersDisablesSubmitAndShowsHint(t *testing.T) {
	html := render(t, Form(FormState{}, nil, true, "", ""))

	if !strings.Contains(html, "Create one first") {
		t.Errorf("expected a hint to create a roaster first, got: %s", html)
	}
	if !strings.Contains(html, "disabled") {
		t.Errorf("expected submit to be disabled, got: %s", html)
	}
}

func TestForm_EditModeShowsReadOnlyMetadata(t *testing.T) {
	state := FormState{ID: 9, Name: "Ethiopia", RoasterID: "3", RoastLevel: "2"}
	html := render(t, Form(state, []roaster.Roaster{{Id: 3, Name: "Blue Bottle"}}, false, "2026-01-02 03:04", "2026-01-05 06:07"))

	if strings.Contains(html, `name="id"`) || strings.Contains(html, `name="created_at"`) {
		t.Errorf("expected id/created_at to not be editable inputs, got: %s", html)
	}
	if !strings.Contains(html, "2026-01-02 03:04") || !strings.Contains(html, "2026-01-05 06:07") {
		t.Errorf("expected read-only metadata to be displayed, got: %s", html)
	}
	if !strings.Contains(html, `selected`) {
		t.Errorf("expected the current roaster/roast level to be pre-selected, got: %s", html)
	}
}

func TestForm_ShowsInlineFieldErrors(t *testing.T) {
	state := FormState{Name: "", Errors: map[string]string{"name": "Beans name must not be empty."}}
	html := render(t, Form(state, []roaster.Roaster{{Id: 1, Name: "Roaster"}}, true, "", ""))

	if !strings.Contains(html, "Beans name must not be empty.") {
		t.Errorf("expected inline error message, got: %s", html)
	}
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid on the errored field, got: %s", html)
	}
}
