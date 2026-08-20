package web

import (
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/views/templates/shared"
	viewsheets "github.com/lescactus/espressoapi-go/views/templates/sheets"
)

var sheetSortColumns = []string{"id", "name", "created_at", "updated_at"}

// sortSheets sorts in place by col/order, falling back to id ascending for
// an unknown column. Nil timestamps sort after non-nil ones ascending (and
// therefore before them descending, since descending simply reverses order).
func sortSheets(sheets []sheet.Sheet, col, order string) {
	col = normalizeSortColumn(col, sheetSortColumns)
	less := func(i, j int) bool { return sheetLess(sheets[i], sheets[j], col) }
	if normalizeSortOrder(order) == "desc" {
		less = func(i, j int) bool { return sheetLess(sheets[j], sheets[i], col) }
	}
	sort.SliceStable(sheets, less)
}

func sheetLess(a, b sheet.Sheet, col string) bool {
	switch col {
	case "name":
		return a.Name < b.Name
	case "created_at":
		return timeLess(a.CreatedAt, b.CreatedAt)
	case "updated_at":
		return timeLess(a.UpdatedAt, b.UpdatedAt)
	default:
		return a.Id < b.Id
	}
}

func timeLess(a, b *time.Time) bool {
	if a == nil {
		return false
	}
	if b == nil {
		return true
	}
	return a.Before(*b)
}

// parseFormError classifies an r.ParseForm() error into a status/message
// pair, distinguishing an oversized body (413) from a malformed one (400).
func parseFormError(err error) (status int, message string) {
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return http.StatusRequestEntityTooLarge, "Request body is too large."
	}
	return http.StatusBadRequest, "Invalid form submission."
}

// writeGetError renders an error for a GET request: a small OOB alert for an
// htmx-driven request (main content stays put), or a full styled error page
// for direct browser navigation.
func (h *Handler) writeGetError(w http.ResponseWriter, r *http.Request, we webError) {
	if isHXRequest(r) {
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, we.Status)
		_ = shared.ErrorAlertOOB(we.Message).Render(r.Context(), w)
		return
	}
	h.writeFullPageError(w, r, we)
}

func (h *Handler) writeFullPageError(w http.ResponseWriter, r *http.Request, we webError) {
	writeHTMLStatus(w, we.Status)
	switch we.Status {
	case http.StatusNotFound:
		_ = shared.NotFoundPage().Render(r.Context(), w)
	case http.StatusBadRequest:
		_ = shared.BadRequestPage(we.Message).Render(r.Context(), w)
	default:
		_ = shared.InternalErrorPage().Render(r.Context(), w)
	}
}

const errInvalidSheetID = "The sheet id must be a positive number."

// Home renders the "/" landing page.
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	sheets, err := h.SheetService.GetAllSheets(r.Context())
	if err != nil {
		h.writeFullPageError(w, r, mapDomainError(err))
		return
	}
	sortSheets(sheets, "id", "asc")
	writeHTMLStatus(w, http.StatusOK)
	_ = viewsheets.Home(sheets).Render(r.Context(), w)
}

// ListSheets renders GET /sheets: the full page, or just the sortable table
// fragment for an htmx sort-refresh.
func (h *Handler) ListSheets(w http.ResponseWriter, r *http.Request) {
	sheets, err := h.SheetService.GetAllSheets(r.Context())
	if err != nil {
		h.writeFullPageError(w, r, mapDomainError(err))
		return
	}
	sortCol := normalizeSortColumn(r.URL.Query().Get("sort"), sheetSortColumns)
	order := normalizeSortOrder(r.URL.Query().Get("order"))
	sortSheets(sheets, sortCol, order)

	writeHTMLStatus(w, http.StatusOK)
	if isHXRequest(r) {
		_ = viewsheets.Table(sheets, sortCol, order, false).Render(r.Context(), w)
		return
	}
	_ = viewsheets.Page(sheets, sortCol, order, false).Render(r.Context(), w)
}

// AddSheetForm renders GET /sheets/add: a blank inline row fragment, or the
// full list page with that row already open for direct navigation.
func (h *Handler) AddSheetForm(w http.ResponseWriter, r *http.Request) {
	if isHXRequest(r) {
		writeHTMLStatus(w, http.StatusOK)
		_ = viewsheets.AddRow(viewsheets.FormState{}).Render(r.Context(), w)
		return
	}
	sheets, err := h.SheetService.GetAllSheets(r.Context())
	if err != nil {
		h.writeFullPageError(w, r, mapDomainError(err))
		return
	}
	sortSheets(sheets, "id", "asc")
	writeHTMLStatus(w, http.StatusOK)
	_ = viewsheets.Page(sheets, "id", "asc", true).Render(r.Context(), w)
}

// CreateSheet handles POST /sheets/add.
func (h *Handler) CreateSheet(w http.ResponseWriter, r *http.Request) {
	if !isFormURLEncoded(r) {
		writeHTMLStatus(w, http.StatusUnsupportedMediaType)
		_ = viewsheets.AddRow(viewsheets.FormState{Error: "Invalid form. Please try again."}).Render(r.Context(), w)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		writeHTMLStatus(w, status)
		_ = viewsheets.AddRow(viewsheets.FormState{Error: message}).Render(r.Context(), w)
		return
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		writeHTMLStatus(w, http.StatusBadRequest)
		_ = viewsheets.AddRow(viewsheets.FormState{Name: name, Error: "Sheet name must not be empty."}).Render(r.Context(), w)
		return
	}

	created, err := h.SheetService.CreateSheetByName(r.Context(), name)
	if err != nil {
		we := mapDomainError(err)
		writeHTMLStatus(w, we.Status)
		_ = viewsheets.AddRow(viewsheets.FormState{Name: name, Error: we.Message}).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = viewsheets.Row(*created).Render(r.Context(), w)
	_ = shared.SuccessAlertOOB("Sheet successfully created.").Render(r.Context(), w)
}

// GetSheet handles GET /sheets/get/:id: the full detail page for direct
// navigation, or the view_context-selected fragment for an htmx cancel.
func (h *Handler) GetSheet(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidSheetID})
		return
	}
	s, err := h.SheetService.GetSheetById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	if !isHXRequest(r) {
		shots, err := h.ShotService.GetShotsBySheetId(r.Context(), id)
		if err != nil {
			h.writeFullPageError(w, r, mapDomainError(err))
			return
		}
		sortShots(shots, "id", "asc")
		writeHTMLStatus(w, http.StatusOK)
		_ = viewsheets.Detail(*s, shots).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	if viewContext(r) == viewContextList {
		_ = viewsheets.Row(*s).Render(r.Context(), w)
		return
	}
	_ = viewsheets.DetailHeader(*s).Render(r.Context(), w)
}

// EditSheetForm handles GET /sheets/update/:id.
func (h *Handler) EditSheetForm(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidSheetID})
		return
	}
	s, err := h.SheetService.GetSheetById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	state := viewsheets.FormState{ID: s.Id, Name: s.Name}
	createdAt := shared.FormatTimestamp(s.CreatedAt)
	updatedAt := shared.FormatTimestamp(s.UpdatedAt)
	vc := viewContext(r)

	if !isHXRequest(r) {
		if vc == viewContextDetail {
			shots, err := h.ShotService.GetShotsBySheetId(r.Context(), id)
			if err != nil {
				h.writeFullPageError(w, r, mapDomainError(err))
				return
			}
			sortShots(shots, "id", "asc")
			writeHTMLStatus(w, http.StatusOK)
			_ = viewsheets.DetailEditing(state, createdAt, updatedAt, shots, s.Id).Render(r.Context(), w)
			return
		}
		sheets, err := h.SheetService.GetAllSheets(r.Context())
		if err != nil {
			h.writeFullPageError(w, r, mapDomainError(err))
			return
		}
		sortSheets(sheets, "id", "asc")
		writeHTMLStatus(w, http.StatusOK)
		_ = viewsheets.EditRowPage(sheets, state, createdAt, updatedAt).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	if vc == viewContextDetail {
		_ = viewsheets.DetailHeaderEdit(state, createdAt, updatedAt).Render(r.Context(), w)
		return
	}
	_ = viewsheets.EditRow(state, createdAt, updatedAt).Render(r.Context(), w)
}

// UpdateSheet handles PUT /sheets/update/:id.
func (h *Handler) UpdateSheet(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		writeHTMLStatus(w, http.StatusBadRequest)
		w.Header().Set("HX-Reswap", "none")
		_ = shared.ErrorAlertOOB(errInvalidSheetID).Render(r.Context(), w)
		return
	}
	vc := viewContext(r)

	if !isFormURLEncoded(r) {
		h.renderSheetFormError(w, r, viewsheets.FormState{ID: id, Error: "Invalid form. Please try again."}, vc, http.StatusUnsupportedMediaType)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		h.renderSheetFormError(w, r, viewsheets.FormState{ID: id, Error: message}, vc, status)
		return
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		h.renderSheetFormError(w, r, viewsheets.FormState{ID: id, Name: name, Error: "Sheet name must not be empty."}, vc, http.StatusBadRequest)
		return
	}

	updated, err := h.SheetService.UpdateSheetById(r.Context(), id, &sheet.Sheet{Id: id, Name: name})
	if err != nil {
		we := mapDomainError(err)
		h.renderSheetFormError(w, r, viewsheets.FormState{ID: id, Name: name, Error: we.Message}, vc, we.Status)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	if vc == viewContextDetail {
		_ = viewsheets.DetailHeader(*updated).Render(r.Context(), w)
	} else {
		_ = viewsheets.Row(*updated).Render(r.Context(), w)
	}
	_ = shared.SuccessAlertOOB("Sheet successfully updated.").Render(r.Context(), w)
}

// renderSheetFormError re-renders the edit fragment matching vc with the
// submitted (possibly invalid) state and an inline error message.
func (h *Handler) renderSheetFormError(w http.ResponseWriter, r *http.Request, state viewsheets.FormState, vc string, status int) {
	writeHTMLStatus(w, status)
	if vc == viewContextDetail {
		_ = viewsheets.DetailHeaderEdit(state, "", "").Render(r.Context(), w)
		return
	}
	_ = viewsheets.EditRow(state, "", "").Render(r.Context(), w)
}

// DeleteSheet handles DELETE /sheets/delete/:id.
func (h *Handler) DeleteSheet(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, http.StatusBadRequest)
		_ = shared.ErrorAlertOOB(errInvalidSheetID).Render(r.Context(), w)
		return
	}

	if err := h.SheetService.DeleteSheetById(r.Context(), id); err != nil {
		we := mapDeleteError(err, "This sheet is still used by shots. Delete those shots first.")
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, we.Status)
		_ = shared.ErrorAlertOOB(we.Message).Render(r.Context(), w)
		return
	}

	if viewContext(r) == viewContextDetail {
		w.Header().Set("HX-Redirect", "/sheets")
		w.WriteHeader(http.StatusOK)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = shared.SuccessAlertOOB("Sheet successfully deleted.").Render(r.Context(), w)
}
