package controllers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func TestLogBeansFromRequestWithoutRoastDate(t *testing.T) {
	var logOutput bytes.Buffer
	logger := zerolog.New(&logOutput).Level(zerolog.DebugLevel)
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	req := httptest.NewRequest(http.MethodGet, "/rest/v1/beans/1", nil)
	beans := &bean.Bean{
		Id:        1,
		Name:      "espresso blend",
		Roaster:   &roaster.Roaster{Id: 2, Name: "roaster"},
		CreatedAt: &createdAt,
	}

	hlog.NewHandler(logger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logBeansFromRequest(r, beans, "beans found by id")
	})).ServeHTTP(httptest.NewRecorder(), req)

	if !strings.Contains(logOutput.String(), `"beans found by id"`) {
		t.Errorf("log output = %s, want message", logOutput.String())
	}

	if strings.Contains(logOutput.String(), `"roast_date"`) {
		t.Errorf("log output = %s, should omit roast_date when it is nil", logOutput.String())
	}
}
