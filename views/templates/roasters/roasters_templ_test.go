package roasters

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/a-h/templ"
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

func testRoaster() roaster.Roaster {
	created := time.Date(2026, 1, 2, 3, 4, 0, 0, time.UTC)
	updated := time.Date(2026, 1, 5, 6, 7, 0, 0, time.UTC)
	return roaster.Roaster{Id: 7, Name: "Blue Bottle", CreatedAt: &created, UpdatedAt: &updated}
}

func TestRow_ShowsAllColumnsAndActions(t *testing.T) {
	html := render(t, Row(testRoaster()))

	for _, want := range []string{"7", "Blue Bottle", "2026-01-02 03:04", "2026-01-05 06:07", "Edit", "Delete"} {
		if !strings.Contains(html, want) {
			t.Errorf("expected row to contain %q, got: %s", want, html)
		}
	}
	if !strings.Contains(html, `hx-confirm="Are you sure you want to delete Blue Bottle?"`) {
		t.Errorf("expected delete confirm to include the roaster name, got: %s", html)
	}
}

func TestEditRow_MetadataIsReadOnly(t *testing.T) {
	html := render(t, EditRow(FormState{ID: 7, Name: "Blue Bottle"}, "2026-01-02 03:04", "2026-01-05 06:07"))

	if strings.Contains(html, `name="id"`) || strings.Contains(html, `name="created_at"`) || strings.Contains(html, `name="updated_at"`) {
		t.Errorf("expected id/created_at/updated_at to not be editable inputs, got: %s", html)
	}
	if !strings.Contains(html, "2026-01-02 03:04") || !strings.Contains(html, "2026-01-05 06:07") {
		t.Errorf("expected read-only timestamps to still be displayed, got: %s", html)
	}
}

func TestRowPage_WrapsRowInAOneRowTable(t *testing.T) {
	html := render(t, RowPage(testRoaster()))

	if !strings.Contains(html, "<html") || !strings.Contains(html, "<table") || !strings.Contains(html, "Blue Bottle") {
		t.Errorf("expected a full page with a one-row table, got: %s", html)
	}
}
