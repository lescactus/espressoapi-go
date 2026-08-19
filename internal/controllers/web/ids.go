package web

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// parsePositiveID extracts and validates the ":id" URL parameter. Missing,
// non-integer, zero, and negative ids are rejected with ok=false, per the
// spec's web ID policy (400 for malformed, 404 for a valid-but-unknown id).
func parsePositiveID(r *http.Request) (int, bool) {
	raw := httprouter.ParamsFromContext(r.Context()).ByName("id")
	if raw == "" {
		return 0, false
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}
