package web

import (
	"net/http"
	"strings"
)

const contentTypeHTML = "text/html; charset=utf-8"

// isHXRequest reports whether the request was issued by htmx (as opposed to
// a direct browser navigation / refresh / deep link).
func isHXRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

// writeHTMLStatus sets the HTML content type, writes status, then renders.
func writeHTMLStatus(w http.ResponseWriter, status int) {
	w.Header().Set("Content-Type", contentTypeHTML)
	w.WriteHeader(status)
}

// isFormURLEncoded reports whether the request's Content-Type header is
// application/x-www-form-urlencoded (ignoring parameters and case).
func isFormURLEncoded(r *http.Request) bool {
	ct := r.Header.Get("Content-Type")
	if ct == "" {
		return false
	}
	mediaType := strings.ToLower(strings.TrimSpace(strings.Split(ct, ";")[0]))
	return mediaType == "application/x-www-form-urlencoded"
}

// closed view_context values. Anything else falls back to the resource's
// default context; the value is never used to build a template name or
// selector directly.
const (
	viewContextList   = "sheet-list"
	viewContextDetail = "sheet-detail"
)

// viewContext resolves the request's view_context query parameter to one of
// the closed values above, defaulting to viewContextList (the common case:
// most resources only ever have a list row, not a detail page).
func viewContext(r *http.Request) string {
	if r.URL.Query().Get("view_context") == viewContextDetail {
		return viewContextDetail
	}
	return viewContextList
}
