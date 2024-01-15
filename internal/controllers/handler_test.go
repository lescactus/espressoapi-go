package controllers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/rs/zerolog"
)

func TestNewHandler(t *testing.T) {
	type args struct {
		sheetService         sheet.Service
		serverMaxRequestSize int64
	}
	tests := []struct {
		name string
		args args
		want *Handler
	}{
		{
			name: "nil args",
			args: args{nil, 0},
			want: &Handler{nil, 0},
		},
		{
			name: "non nil args",
			args: args{sheet.New(nil), 10},
			want: &Handler{sheet.New(nil), 10},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHandler(tt.args.sheetService, tt.args.serverMaxRequestSize); !reflect.DeepEqual(got, tt.want) {
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
	handler := NewHandler(nil, 1024)

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
	handler := NewHandler(nil, 1024)

	// Create a chain with the IdParameterLoggerHandler
	c := alice.New().Append(handler.IdParameterLoggerHandler("id"))

	// Create a test server with the handler chain
	ts := httptest.NewServer(c.ThenFunc(func(w http.ResponseWriter, r *http.Request) {

	}))

	defer ts.Close()

	// Make a request to the test server with a sample ID
	req, err := http.NewRequest("GET", ts.URL+"/rest/v1/sheets/123", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a test context with httprouter parameters
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "123"}}
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)

	// Set the context with parameters to the request
	req = req.WithContext(ctx)

	// Perform the request
	http.DefaultClient.Do(req)
}

func TestHandlerIDParameterLoggerHTTPHandler(t *testing.T) {
	// Create a mock handler
	mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL)

		_ = zerolog.Ctx(r.Context())
	})

	// Create a mock request with a context containing the ID parameter
	req := httptest.NewRequest(http.MethodGet, "/test/123", nil)
	params := httprouter.Params{httprouter.Param{Key: "id", Value: "123"}}

	logger := zerolog.New(io.Discard)
	ctx := context.WithValue(req.Context(), httprouter.ParamsKey, params)
	ctx = logger.WithContext(ctx)
	req = req.WithContext(ctx)

	// Create a mock response recorder
	rr := httptest.NewRecorder()

	// Create a mock Handler instance
	handler := NewHandler(nil, 1024)

	// Call the idParameterLoggerHttpHandler function
	idParamLoggerHandler := handler.idParameterLoggerHttpHandler(mockHandler, "id")

	// Serve the request using the idParameterLoggerHttpHandler
	idParamLoggerHandler.ServeHTTP(rr, req)
}
