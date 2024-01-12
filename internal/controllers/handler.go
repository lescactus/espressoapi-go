package controllers

import (
	"net/http"

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
