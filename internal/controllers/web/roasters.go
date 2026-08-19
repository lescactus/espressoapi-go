package web

import (
	"net/http"
	"sort"
	"strings"

	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	viewroasters "github.com/lescactus/espressoapi-go/views/templates/roasters"
	"github.com/lescactus/espressoapi-go/views/templates/shared"
)

var roasterSortColumns = []string{"id", "name", "created_at", "updated_at"}

// sortRoasters sorts in place by col/order, falling back to id ascending for
// an unknown column. Nil timestamps sort after non-nil ones ascending.
func sortRoasters(roasters []roaster.Roaster, col, order string) {
	col = normalizeSortColumn(col, roasterSortColumns)
	less := func(i, j int) bool { return roasterLess(roasters[i], roasters[j], col) }
	if normalizeSortOrder(order) == "desc" {
		less = func(i, j int) bool { return roasterLess(roasters[j], roasters[i], col) }
	}
	sort.SliceStable(roasters, less)
}

func roasterLess(a, b roaster.Roaster, col string) bool {
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

const errInvalidRoasterID = "The roaster id must be a positive number."

// ListRoasters handles GET /roasters.
func (h *Handler) ListRoasters(w http.ResponseWriter, r *http.Request) {
	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		h.writeFullPageError(w, r, mapDomainError(err))
		return
	}
	sortCol := normalizeSortColumn(r.URL.Query().Get("sort"), roasterSortColumns)
	order := normalizeSortOrder(r.URL.Query().Get("order"))
	sortRoasters(roasters, sortCol, order)

	writeHTMLStatus(w, http.StatusOK)
	if isHXRequest(r) {
		_ = viewroasters.Table(roasters, sortCol, order, false).Render(r.Context(), w)
		return
	}
	_ = viewroasters.Page(roasters, sortCol, order, false).Render(r.Context(), w)
}

// AddRoasterForm handles GET /roasters/add.
func (h *Handler) AddRoasterForm(w http.ResponseWriter, r *http.Request) {
	if isHXRequest(r) {
		writeHTMLStatus(w, http.StatusOK)
		_ = viewroasters.AddRow(viewroasters.FormState{}).Render(r.Context(), w)
		return
	}
	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		h.writeFullPageError(w, r, mapDomainError(err))
		return
	}
	sortRoasters(roasters, "id", "asc")
	writeHTMLStatus(w, http.StatusOK)
	_ = viewroasters.Page(roasters, "id", "asc", true).Render(r.Context(), w)
}

// CreateRoaster handles POST /roasters/add.
func (h *Handler) CreateRoaster(w http.ResponseWriter, r *http.Request) {
	if !isFormURLEncoded(r) {
		writeHTMLStatus(w, http.StatusUnsupportedMediaType)
		_ = viewroasters.AddRow(viewroasters.FormState{Error: "Invalid form. Please try again."}).Render(r.Context(), w)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		writeHTMLStatus(w, status)
		_ = viewroasters.AddRow(viewroasters.FormState{Error: message}).Render(r.Context(), w)
		return
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		writeHTMLStatus(w, http.StatusBadRequest)
		_ = viewroasters.AddRow(viewroasters.FormState{Name: name, Error: "Roaster name must not be empty."}).Render(r.Context(), w)
		return
	}

	created, err := h.RoasterService.CreateRoasterByName(r.Context(), name)
	if err != nil {
		we := mapDomainError(err)
		writeHTMLStatus(w, we.Status)
		_ = viewroasters.AddRow(viewroasters.FormState{Name: name, Error: we.Message}).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = viewroasters.Row(*created).Render(r.Context(), w)
	_ = shared.SuccessAlertOOB("Roaster successfully created.").Render(r.Context(), w)
}

// GetRoaster handles GET /roasters/get/:id: a single row fragment in view
// mode for htmx, or the full page with a one-row table for direct navigation.
func (h *Handler) GetRoaster(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidRoasterID})
		return
	}
	roasterVal, err := h.RoasterService.GetRoasterById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	if !isHXRequest(r) {
		_ = viewroasters.RowPage(*roasterVal).Render(r.Context(), w)
		return
	}
	_ = viewroasters.Row(*roasterVal).Render(r.Context(), w)
}

// EditRoasterForm handles GET /roasters/update/:id.
func (h *Handler) EditRoasterForm(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidRoasterID})
		return
	}
	roasterVal, err := h.RoasterService.GetRoasterById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	state := viewroasters.FormState{ID: roasterVal.Id, Name: roasterVal.Name}
	createdAt := shared.FormatTimestamp(roasterVal.CreatedAt)
	updatedAt := shared.FormatTimestamp(roasterVal.UpdatedAt)

	if !isHXRequest(r) {
		roasters, err := h.RoasterService.GetAllRoasters(r.Context())
		if err != nil {
			h.writeFullPageError(w, r, mapDomainError(err))
			return
		}
		sortRoasters(roasters, "id", "asc")
		writeHTMLStatus(w, http.StatusOK)
		_ = viewroasters.EditRowPage(roasters, state, createdAt, updatedAt).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = viewroasters.EditRow(state, createdAt, updatedAt).Render(r.Context(), w)
}

// UpdateRoaster handles PUT /roasters/update/:id.
func (h *Handler) UpdateRoaster(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		writeHTMLStatus(w, http.StatusBadRequest)
		w.Header().Set("HX-Reswap", "none")
		_ = shared.ErrorAlertOOB(errInvalidRoasterID).Render(r.Context(), w)
		return
	}

	if !isFormURLEncoded(r) {
		writeHTMLStatus(w, http.StatusUnsupportedMediaType)
		_ = viewroasters.EditRow(viewroasters.FormState{ID: id, Error: "Invalid form. Please try again."}, "", "").Render(r.Context(), w)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		writeHTMLStatus(w, status)
		_ = viewroasters.EditRow(viewroasters.FormState{ID: id, Error: message}, "", "").Render(r.Context(), w)
		return
	}

	name := strings.TrimSpace(r.PostFormValue("name"))
	if name == "" {
		writeHTMLStatus(w, http.StatusBadRequest)
		_ = viewroasters.EditRow(viewroasters.FormState{ID: id, Name: name, Error: "Roaster name must not be empty."}, "", "").Render(r.Context(), w)
		return
	}

	updated, err := h.RoasterService.UpdateRoasterById(r.Context(), id, &roaster.Roaster{Id: id, Name: name})
	if err != nil {
		we := mapDomainError(err)
		writeHTMLStatus(w, we.Status)
		_ = viewroasters.EditRow(viewroasters.FormState{ID: id, Name: name, Error: we.Message}, "", "").Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = viewroasters.Row(*updated).Render(r.Context(), w)
	_ = shared.SuccessAlertOOB("Roaster successfully updated.").Render(r.Context(), w)
}

// DeleteRoaster handles DELETE /roasters/delete/:id.
func (h *Handler) DeleteRoaster(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, http.StatusBadRequest)
		_ = shared.ErrorAlertOOB(errInvalidRoasterID).Render(r.Context(), w)
		return
	}

	if err := h.RoasterService.DeleteRoasterById(r.Context(), id); err != nil {
		we := mapDomainError(err)
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, we.Status)
		_ = shared.ErrorAlertOOB(we.Message).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = shared.SuccessAlertOOB("Roaster successfully deleted.").Render(r.Context(), w)
}
