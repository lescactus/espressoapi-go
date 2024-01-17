package controllers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/rs/zerolog"
)

const (
	// ContentTypeApplicationJSON represent the applcation/json Content-Type value
	ContentTypeApplicationJSON = "application/json"
)

type Handler struct {
	SheetService   sheet.Service
	RoasterService roaster.Service
	maxRequestSize int64
}

func NewHandler(sheetService sheet.Service, roasterService roaster.Service, serverMaxRequestSize int64) *Handler {
	return &Handler{
		SheetService:   sheetService,
		RoasterService: roasterService,
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

// IdParameterLoggerHandler returns a middleware function that logs the value of the specified
// field key from the request URL parameters.
func (h *Handler) IdParameterLoggerHandler(fieldKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return h.idParameterLoggerHttpHandler(next, fieldKey)
	}
}

// idParameterLoggerHttpHandler extracts the ID from the request parameters and adds it to
// the logger context.
// If the ID cannot be fetched from the request params, it is not added to the logger context.
func (hr *Handler) idParameterLoggerHttpHandler(h http.Handler, fieldKey string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := hr.getIdFromParams(r.Context())
		// In case the id is fetched from the request params, we add it to the logger context
		if err == nil {
			zerolog.Ctx(r.Context()).UpdateContext(func(c zerolog.Context) zerolog.Context {
				return c.Int(fieldKey, id)
			})
		}
		h.ServeHTTP(w, r)
	})
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

// getIdFromParams extracts the ID parameter from the context and converts it to an integer.
// It returns the extracted ID and any error encountered during the process.
func (h *Handler) getIdFromParams(ctx context.Context) (int, error) {
	params := httprouter.ParamsFromContext(ctx)
	idParam := params.ByName("id")
	if idParam == "" {
		return 0, ErrIDNotFound
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return 0, ErrIDNotInteger
	}

	return id, nil
}
