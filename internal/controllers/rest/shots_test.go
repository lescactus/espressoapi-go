package rest

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

func TestLogShotFromRequestCreatedAt(t *testing.T) {
	createdAt := time.Date(2026, time.January, 5, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		name          string
		createdAt     *time.Time
		wantCreatedAt bool
	}{
		{name: "includes supplied timestamp", createdAt: &createdAt, wantCreatedAt: true},
		{name: "omits nil timestamp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput bytes.Buffer
			logger := zerolog.New(&logOutput).Level(zerolog.DebugLevel)
			req := httptest.NewRequest(http.MethodGet, "/rest/v1/shots/1", nil)
			value := testShot(1)
			value.CreatedAt = tt.createdAt

			hlog.NewHandler(logger)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				logShotFromRequest(r, value, "shot found by id")
			})).ServeHTTP(httptest.NewRecorder(), req)

			var logEntry map[string]any
			if err := json.Unmarshal(logOutput.Bytes(), &logEntry); err != nil {
				t.Fatalf("decode log output %q: %v", logOutput.String(), err)
			}
			if got := logEntry[zerolog.MessageFieldName]; got != "shot found by id" {
				t.Errorf("message = %v, want %q", got, "shot found by id")
			}

			loggedShot, ok := logEntry["shot"].(map[string]any)
			if !ok {
				t.Fatalf("shot field = %#v, want object", logEntry["shot"])
			}
			gotCreatedAt, hasCreatedAt := loggedShot["created_at"]
			if hasCreatedAt != tt.wantCreatedAt {
				t.Fatalf("created_at presence = %t, want %t", hasCreatedAt, tt.wantCreatedAt)
			}
			if tt.wantCreatedAt && gotCreatedAt != createdAt.Format(time.RFC3339) {
				t.Errorf("created_at = %v, want %q", gotCreatedAt, createdAt.Format(time.RFC3339))
			}
		})
	}
}
