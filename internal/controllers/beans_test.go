package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func TestLogBeansFromRequestOptionalDates(t *testing.T) {
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	roastDate := time.Date(2025, time.December, 28, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name      string
		createdAt *time.Time
		roastDate *time.Time
	}{
		{name: "includes both dates", createdAt: &createdAt, roastDate: &roastDate},
		{name: "omits nil roast date", createdAt: &createdAt},
		{name: "omits nil creation timestamp", roastDate: &roastDate},
		{name: "omits both nil dates"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput bytes.Buffer
			logger := zerolog.New(&logOutput).Level(zerolog.DebugLevel)
			req := httptest.NewRequest(http.MethodGet, "/rest/v1/beans/1", nil)
			beans := testBean(1, "espresso blend")
			beans.CreatedAt = tt.createdAt
			beans.RoastDate = tt.roastDate

			hlog.NewHandler(logger)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				logBeansFromRequest(r, beans, "beans found by id")
			})).ServeHTTP(httptest.NewRecorder(), req)

			var logEntry map[string]any
			if err := json.Unmarshal(logOutput.Bytes(), &logEntry); err != nil {
				t.Fatalf("decode log output %q: %v", logOutput.String(), err)
			}
			if got := logEntry[zerolog.MessageFieldName]; got != "beans found by id" {
				t.Errorf("message = %v, want %q", got, "beans found by id")
			}

			loggedBeans, ok := logEntry["beans"].(map[string]any)
			if !ok {
				t.Fatalf("beans field = %#v, want object", logEntry["beans"])
			}

			for field, want := range map[string]*time.Time{
				"created_at": tt.createdAt,
				"roast_date": tt.roastDate,
			} {
				got, exists := loggedBeans[field]
				if exists != (want != nil) {
					t.Errorf("%s presence = %t, want %t", field, exists, want != nil)
					continue
				}
				if want != nil && got != want.Format(time.RFC3339) {
					t.Errorf("%s = %v, want %q", field, got, want.Format(time.RFC3339))
				}
			}
		})
	}
}
