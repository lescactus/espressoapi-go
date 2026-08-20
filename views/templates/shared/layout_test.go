package shared

import (
	"context"
	"strings"
	"testing"

	"github.com/a-h/templ"
)

func render(t *testing.T, c templ.Component) string {
	t.Helper()
	var b strings.Builder
	if err := c.Render(context.Background(), &b); err != nil {
		t.Fatalf("render: %v", err)
	}
	return b.String()
}

func TestLayout_RendersNavAndContent(t *testing.T) {
	html := render(t, Layout("Sheets", "sheets"))

	if !strings.Contains(html, "<title>Sheets - espressoapi-go</title>") {
		t.Errorf("expected title in rendered HTML, got: %s", html)
	}
	if !strings.Contains(html, `aria-current="page"`) {
		t.Errorf("expected active nav link to have aria-current, got: %s", html)
	}
	if !strings.Contains(html, `id="alerts"`) {
		t.Errorf("expected alerts region, got: %s", html)
	}
}

func TestNav_MarksOnlyActiveLinkCurrent(t *testing.T) {
	html := render(t, Nav("roasters"))

	if strings.Count(html, `aria-current="page"`) != 1 {
		t.Errorf("expected exactly one aria-current=page link, got: %s", html)
	}
	if !strings.Contains(html, `href="/roasters" aria-current="page"`) {
		t.Errorf("expected roasters link to be marked current, got: %s", html)
	}
}

func TestNotFoundPage_Renders404(t *testing.T) {
	html := render(t, NotFoundPage())

	if !strings.Contains(html, "404") {
		t.Errorf("expected 404 page content, got: %s", html)
	}
}
