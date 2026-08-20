package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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
	const maxRequestSize = int64(4)
	tests := []struct {
		name         string
		body         string
		wantStatus   int
		wantTooLarge bool
	}{
		{
			name: "below limit", body: "abc", wantStatus: http.StatusNoContent,
		},
		{
			name: "at limit", body: "abcd", wantStatus: http.StatusNoContent,
		},
		{
			name: "over limit", body: "abcde", wantStatus: http.StatusRequestEntityTooLarge, wantTooLarge: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(nil, nil, nil, nil, maxRequestSize)
			var body []byte
			var readErr error
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, readErr = io.ReadAll(r.Body)
				if readErr != nil {
					handler.SetErrorResponse(w, readErr)
					return
				}
				w.WriteHeader(http.StatusNoContent)
			})
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.body))
			recorder := httptest.NewRecorder()

			handler.MaxReqSize()(next).ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}

			var maxBytesError *http.MaxBytesError
			if errors.As(readErr, &maxBytesError) != tt.wantTooLarge {
				t.Fatalf("read error = %v, MaxBytesError = %t, want %t", readErr, maxBytesError != nil, tt.wantTooLarge)
			}
			if maxBytesError != nil {
				if maxBytesError.Limit != maxRequestSize {
					t.Errorf("MaxBytesError limit = %d, want %d", maxBytesError.Limit, maxRequestSize)
				}
				assertJSONResponse(t, recorder, http.StatusRequestEntityTooLarge, ErrorResponse{Msg: "request body must not be larger than 4 bytes"})
				return
			}

			if readErr != nil {
				t.Fatalf("read body: %v", readErr)
			}
			if string(body) != tt.body {
				t.Errorf("body = %q, want %q", body, tt.body)
			}
		})
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
