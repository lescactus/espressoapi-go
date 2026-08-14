package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		sheetService         sheet.Service
		roasterService       roaster.Service
		beanService          bean.Service
		shotService          shot.Service
		serverMaxRequestSize int64
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "nil args",
			args: args{nil, nil, nil, nil, 0},
			want: &Handler{nil, nil, nil, nil, 0},
		},
		{
			name: "non nil args",
			args: args{sheet.New(nil), roaster.New(nil), bean.New(nil), shot.New(nil), 10},
			want: &Handler{sheet.New(nil), roaster.New(nil), bean.New(nil), shot.New(nil), 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.sheetService, tt.args.roasterService, tt.args.beanService, tt.args.shotService, tt.args.serverMaxRequestSize); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxReqSizeMiddleware(t *testing.T) {
	// Create a test handler to wrap with the MaxReqSize middleware
	testHandlerOK := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	testHandlerTooLarge := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusRequestEntityTooLarge)
	})

	// Create an instance of the Handler with a test SheetService and maxRequestSize
	handler := NewHandler(nil, nil, nil, nil, 1024)

	// Create a request with a body larger than maxRequestSize
	requestBody := "a" + strings.Repeat("b", 1024)
	req, err := http.NewRequest("POST", "/test", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Set the "Content-Type" header to "application/json"
	req.Header.Set("Content-Type", "application/json")

	// Create a recorder to capture the response
	rr := httptest.NewRecorder()

	// Wrap the test handler with the MaxReqSize middleware
	maxReqSizeMiddleware := handler.MaxReqSize()
	handlerWithMiddleware := maxReqSizeMiddleware(testHandlerOK)

	// Serve the request
	handlerWithMiddleware.ServeHTTP(rr, req)

	// Check the response status code
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, rr.Code)
	}

	// Create a request with a body larger than maxRequestSize
	requestBody = strings.Repeat("a", 2048)
	req, err = http.NewRequest("POST", "/test", strings.NewReader(requestBody))
	if err != nil {
		t.Fatal(err)
	}

	// Set the "Content-Type" header to "application/json"
	req.Header.Set("Content-Type", "application/json")

	// Reset the recorder for the second request
	rr = httptest.NewRecorder()

	// Serve the request
	// Wrap the test handler with the MaxReqSize middleware
	handlerWithMiddleware = maxReqSizeMiddleware(testHandlerTooLarge)

	// Serve the request
	handlerWithMiddleware.ServeHTTP(rr, req)

	// Check the response status code for the second request
	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("Expected status code %d, got %d", http.StatusRequestEntityTooLarge, rr.Code)
	}
}

func TestHandlerParseContentType(t *testing.T) {
	type fields struct{}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "valid content type",
			fields: fields{},
			args: args{
				r: &http.Request{
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "invalid content type",
			fields: fields{},
			args: args{
				r: &http.Request{
					Header: http.Header{
						"Content-Type": []string{"text/plain"},
					},
				},
			},
			wantErr: true,
		},
		{
			name:   "missing content type",
			fields: fields{},
			args: args{
				r: &http.Request{
					Header: http.Header{},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{}
			if err := h.parseContentType(tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("Handler.parseContentType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandlerIdParameterLoggerHandler(t *testing.T) {
	tests := []struct {
		name         string
		fieldKey     string
		idParam      string
		wantID       int
		wantIDInLogs bool
	}{
		{
			name: "valid id", fieldKey: "id", idParam: "123",
			wantID: 123, wantIDInLogs: true,
		},
		{
			name: "custom field key", fieldKey: "sheet_id", idParam: "456",
			wantID: 456, wantIDInLogs: true,
		},
		{
			name: "non-integer id", fieldKey: "id", idParam: "invalid",
		},
		{
			name: "missing id", fieldKey: "id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput bytes.Buffer
			logger := zerolog.New(&logOutput)
			handler := NewHandler(nil, nil, nil, nil, 1024)
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				hlog.FromRequest(r).Info().Msg("handled")
				w.WriteHeader(http.StatusNoContent)
			})
			middleware := hlog.NewHandler(logger)(handler.IdParameterLoggerHandler(tt.fieldKey)(next))

			req := httptest.NewRequest(http.MethodGet, "/rest/v1/sheets/"+tt.idParam, nil)
			if tt.idParam != "" {
				params := httprouter.Params{{Key: "id", Value: tt.idParam}}
				ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
				req = req.WithContext(ctx)
			}
			recorder := httptest.NewRecorder()

			middleware.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusNoContent {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
			}

			var logEvent map[string]json.RawMessage
			if err := json.Unmarshal(logOutput.Bytes(), &logEvent); err != nil {
				t.Fatalf("decode log event: %v", err)
			}

			var message string
			if err := json.Unmarshal(logEvent[zerolog.MessageFieldName], &message); err != nil {
				t.Fatalf("decode log message: %v", err)
			}
			if message != "handled" {
				t.Errorf("message = %q, want %q", message, "handled")
			}

			idValue, idInLogs := logEvent[tt.fieldKey]
			if idInLogs != tt.wantIDInLogs {
				t.Fatalf("log field %q present = %t, want %t", tt.fieldKey, idInLogs, tt.wantIDInLogs)
			}
			if idInLogs {
				var id int
				if err := json.Unmarshal(idValue, &id); err != nil {
					t.Fatalf("decode log field %q: %v", tt.fieldKey, err)
				}
				if id != tt.wantID {
					t.Errorf("log field %q = %d, want %d", tt.fieldKey, id, tt.wantID)
				}
			}
		})
	}
}
