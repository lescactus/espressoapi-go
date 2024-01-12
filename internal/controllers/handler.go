package controllers

import (
	"net/http"
	"strings"

	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

const (
	// ContentTypeApplicationJSON represent the applcation/json Content-Type value
	ContentTypeApplicationJSON = "application/json"
)

type Handler struct {
	SheetService   sheet.Service
	maxRequestSize int64
}

func NewHandler(sheetService sheet.Service, serverMaxRequestSize int64) *Handler {
	return &Handler{
		SheetService:   sheetService,
		maxRequestSize: serverMaxRequestSize,
	}
}

// MaxReqSize is a HTTP middleware limiting the size of the request
// by using http.MaxBytesReader() on the request body.
func (h *Handler) MaxReqSize() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, h.maxRequestSize)
			next.ServeHTTP(w, r)
		})
	}
}

// parseContentType checks if the "Content-Type" header of the HTTP request is "application/json".
// It returns an error if the header is missing or has a different value.
func (h *Handler) parseContentType(r *http.Request) error {
	// If the "Content-Type" header is present, check that it has the value "application/json".
	// Parse and normalize the header to remove any additional parameters by stripping
	// whitespace and converting to lowercase before we checking the value
	ct := r.Header.Get("Content-Type")
	if ct != "" {
		mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
		if mediaType != "application/json" {
			return &ErrorResponse{status: http.StatusUnsupportedMediaType, Msg: "Content-Type header is not application/json"}
		}
	} else {
		return &ErrorResponse{status: http.StatusUnsupportedMediaType, Msg: "Content-Type header is not application/json"}
	}

	return nil
}
