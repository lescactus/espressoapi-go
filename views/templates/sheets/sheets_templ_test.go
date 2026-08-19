package sheets

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), &b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func testSheet() sheet.Sheet {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	updated := time.Date(2026, 1, 5, 6, 7, 0, 0, time.UTC)
	return sheet.Sheet{Id: 42, Name: "Double shot", CreatedAt: &created, UpdatedAt: &updated}
}

func TestRow_ShowsAllColumnsAndActions(t *testing.T) {
	html := render(t, Row(testSheet()))

	for _, want := range []string{"42", "Double shot", "2026-01-02 03:04", "2026-01-05 06:07", "Edit", "Delete"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected row to contain %q, got: %s", want, html)
		}
	}
	if !strings.Contains(html, `hx-confirm="Are you sure you want to delete Double shot?"`) {
		t.Errorf("expected delete confirm to include the sheet name, got: %s", html)
	}
}

func TestEditRow_MetadataIsReadOnly(t *testing.T) {
	html := render(t, EditRow(FormState{ID: 42, Name: "Double shot"}, "2026-01-02 03:04", "2026-01-05 06:07"))

	if strings.Contains(html, `name="id"`) || strings.Contains(html, `name="created_at"`) || strings.Contains(html, `name="updated_at"`) {
		t.Errorf("expected id/created_at/updated_at to not be editable inputs, got: %s", html)
	}
	if !strings.Contains(html, `name="name"`) {
		t.Errorf("expected an editable name input, got: %s", html)
	}
	if !strings.Contains(html, "2026-01-02 03:04") || !strings.Contains(html, "2026-01-05 06:07") {
		t.Errorf("expected read-only timestamps to still be displayed, got: %s", html)
	}
}

func TestEditRow_ShowsInlineError(t *testing.T) {
	html := render(t, EditRow(FormState{ID: 1, Name: "", Error: "Sheet name must not be empty."}, "", ""))

	if !strings.Contains(html, "Sheet name must not be empty.") {
		t.Errorf("expected inline error message, got: %s", html)
	}
	if !strings.Contains(html, `aria-invalid="true"`) {
		t.Errorf("expected aria-invalid on the errored field, got: %s", html)
	}
}

func TestDetailHeader_DeleteConfirmUsesName(t *testing.T) {
	html := render(t, DetailHeader(testSheet()))

	if !strings.Contains(html, `hx-confirm="Are you sure you want to delete Double shot?"`) {
		t.Errorf("expected delete confirm to include the sheet name, got: %s", html)
	}
	if !strings.Contains(html, "view_context=sheet-detail") {
		t.Errorf("expected detail actions to carry the sheet-detail view_context, got: %s", html)
	}
}

func TestTable_ShowsSortIndicatorOnActiveColumn(t *testing.T) {
	html := render(t, Table([]sheet.Sheet{testSheet()}, "name", "asc", false))

	if !strings.Contains(html, "Name") || !strings.Contains(html, "&#9650;") {
		t.Errorf("expected an ascending indicator on the active Name column, got: %s", html)
	}
}

func TestHome_EmptyStateShowsCallToAction(t *testing.T) {
	html := render(t, Home(nil))

	if !strings.Contains(html, "Add your first sheet") {
		t.Errorf("expected empty-state call to action, got: %s", html)
	}
}
