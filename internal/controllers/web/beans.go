package web

import (
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	viewbeans "github.com/lescactus/espressoapi-go/views/templates/beans"
	"github.com/lescactus/espressoapi-go/views/templates/shared"
)

var beanSortColumns = []string{"id", "name", "roast_date", "roast_level", "created_at", "updated_at"}

func sortBeans(beans []bean.Bean, col, order string) {
	col = normalizeSortColumn(col, beanSortColumns)
	less := func(i, j int) bool { return beanLess(beans[i], beans[j], col) }
	if normalizeSortOrder(order) == "desc" {
		less = func(i, j int) bool { return beanLess(beans[j], beans[i], col) }
	}
	sort.SliceStable(beans, less)
}

func beanLess(a, b bean.Bean, col string) bool {
	switch col {
	case "name":
		return a.Name < b.Name
	case "roast_date":
		return timeLess(a.RoastDate, b.RoastDate)
	case "roast_level":
		return a.RoastLevel < b.RoastLevel
	case "created_at":
		return timeLess(a.CreatedAt, b.CreatedAt)
	case "updated_at":
		return timeLess(a.UpdatedAt, b.UpdatedAt)
	default:
		return a.Id < b.Id
	}
}

const errInvalidBeanID = "The beans id must be a positive number."

// ListBeans handles GET /beans.
func (h *Handler) ListBeans(w http.ResponseWriter, r *http.Request) {
	beans, err := h.BeanService.GetAllBeans(r.Context())
	if err != nil {
		h.writeFullPageError(w, r, mapDomainError(err))
		return
	}
	sortCol := normalizeSortColumn(r.URL.Query().Get("sort"), beanSortColumns)
	order := normalizeSortOrder(r.URL.Query().Get("order"))
	sortBeans(beans, sortCol, order)

	writeHTMLStatus(w, http.StatusOK)
	if isHXRequest(r) {
		_ = viewbeans.Table(beans, sortCol, order).Render(r.Context(), w)
		return
	}
	_ = viewbeans.Page(beans, sortCol, order, nil).Render(r.Context(), w)
}

// beansListForPage fetches and default-sorts the full bean list, for the
// full-page fallback of a direct GET to an add/edit dialog route.
func (h *Handler) beansListForPage(r *http.Request) ([]bean.Bean, error) {
	beans, err := h.BeanService.GetAllBeans(r.Context())
	if err != nil {
		return nil, err
	}
	sortBeans(beans, "id", "asc")
	return beans, nil
}

// beanFormState builds a blank or pre-filled form state plus the roaster
// options needed to render the add/edit dialog.
func (h *Handler) beanRoasterOptions(r *http.Request) ([]roaster.Roaster, error) {
	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		return nil, err
	}
	sort.SliceStable(roasters, func(i, j int) bool { return roasters[i].Id < roasters[j].Id })
	return roasters, nil
}

// AddBeanForm handles GET /beans/add: the dialog form fragment for htmx, or
// the full beans list page with the dialog pre-opened for direct navigation.
func (h *Handler) AddBeanForm(w http.ResponseWriter, r *http.Request) {
	roasters, err := h.beanRoasterOptions(r)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}
	form := viewbeans.Form(viewbeans.FormState{}, roasters, true, "", "")

	if !isHXRequest(r) {
		beans, err := h.beansListForPage(r)
		if err != nil {
			h.writeFullPageError(w, r, mapDomainError(err))
			return
		}
		writeHTMLStatus(w, http.StatusOK)
		_ = viewbeans.Page(beans, "id", "asc", form).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = form.Render(r.Context(), w)
}

// parseBeanForm extracts and validates bean form fields, returning the raw
// FormState (for redisplay) and, on success, the parsed service model.
func parseBeanForm(r *http.Request, id int) (viewbeans.FormState, *bean.Bean, bool) {
	state := viewbeans.FormState{
		ID:         id,
		Name:       strings.TrimSpace(r.PostFormValue("name")),
		RoasterID:  strings.TrimSpace(r.PostFormValue("roaster_id")),
		RoastDate:  strings.TrimSpace(r.PostFormValue("roast_date")),
		RoastLevel: strings.TrimSpace(r.PostFormValue("roast_level")),
		Errors:     map[string]string{},
	}

	if state.Name == "" {
		state.Errors["name"] = "Beans name must not be empty."
	}

	roasterID, err := strconv.Atoi(state.RoasterID)
	if err != nil || roasterID <= 0 {
		state.Errors["roaster_id"] = "Select a roaster."
	}

	var roastDate *time.Time
	if state.RoastDate != "" {
		parsed, err := time.Parse("2006-01-02", state.RoastDate)
		if err != nil {
			state.Errors["roast_date"] = "Roast date must be a valid date."
		} else {
			roastDate = &parsed
		}
	}

	roastLevel := 0
	if state.RoastLevel == "" {
		state.Errors["roast_level"] = "Select a roast level."
	} else if n, err := strconv.Atoi(state.RoastLevel); err != nil || n < int(sql.RoastLevelLight) || n > int(sql.RoastLevelDark) {
		state.Errors["roast_level"] = "Invalid roast level."
	} else {
		roastLevel = n
	}

	if len(state.Errors) > 0 {
		return state, nil, false
	}

	return state, &bean.Bean{
		Id:         id,
		Name:       state.Name,
		Roaster:    &roaster.Roaster{Id: roasterID},
		RoastDate:  roastDate,
		RoastLevel: sql.RoastLevel(roastLevel),
	}, true
}

// CreateBean handles POST /beans/add.
func (h *Handler) CreateBean(w http.ResponseWriter, r *http.Request) {
	if !isFormURLEncoded(r) {
		h.renderBeanFormError(w, r, viewbeans.FormState{}, true, http.StatusUnsupportedMediaType)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		state := viewbeans.FormState{FormError: message}
		h.renderBeanFormError(w, r, state, true, status)
		return
	}

	state, model, ok := parseBeanForm(r, 0)
	if !ok {
		h.renderBeanFormError(w, r, state, true, http.StatusBadRequest)
		return
	}

	created, err := h.BeanService.CreateBean(r.Context(), model)
	if err != nil {
		we := mapDomainError(err)
		if field := beanErrorField(err); field != "" {
			state.Errors[field] = we.Message
		} else {
			state.FormError = we.Message
		}
		h.renderBeanFormError(w, r, state, true, we.Status)
		return
	}

	w.Header().Set("HX-Reswap", "none")
	w.Header().Set("HX-Trigger", "dialog-close")
	writeHTMLStatus(w, http.StatusOK)
	_ = viewbeans.Row(*created, "insert").Render(r.Context(), w)
	_ = shared.SuccessAlertOOB("Beans successfully created.").Render(r.Context(), w)
}

// GetBean handles GET /beans/get/:id: a single row fragment in view mode for
// htmx, or the full page with a one-row table for direct navigation.
func (h *Handler) GetBean(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidBeanID})
		return
	}
	b, err := h.BeanService.GetBeanById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	if !isHXRequest(r) {
		_ = viewbeans.RowPage(*b).Render(r.Context(), w)
		return
	}
	_ = viewbeans.Row(*b, "").Render(r.Context(), w)
}

// EditBeanForm handles GET /beans/update/:id: the dialog form fragment,
// pre-filled.
func (h *Handler) EditBeanForm(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidBeanID})
		return
	}
	b, err := h.BeanService.GetBeanById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}
	roasters, err := h.beanRoasterOptions(r)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	state := viewbeans.FormState{ID: b.Id, Name: b.Name, RoastLevel: strconv.Itoa(int(b.RoastLevel))}
	if b.Roaster != nil {
		state.RoasterID = strconv.Itoa(b.Roaster.Id)
	}
	if b.RoastDate != nil {
		state.RoastDate = b.RoastDate.UTC().Format("2006-01-02")
	}
	form := viewbeans.Form(state, roasters, false, shared.FormatTimestamp(b.CreatedAt), shared.FormatTimestamp(b.UpdatedAt))

	if !isHXRequest(r) {
		beans, err := h.beansListForPage(r)
		if err != nil {
			h.writeFullPageError(w, r, mapDomainError(err))
			return
		}
		writeHTMLStatus(w, http.StatusOK)
		_ = viewbeans.Page(beans, "id", "asc", form).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = form.Render(r.Context(), w)
}

// UpdateBean handles PUT /beans/update/:id.
func (h *Handler) UpdateBean(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		writeHTMLStatus(w, http.StatusBadRequest)
		w.Header().Set("HX-Reswap", "none")
		_ = shared.ErrorAlertOOB(errInvalidBeanID).Render(r.Context(), w)
		return
	}

	if !isFormURLEncoded(r) {
		h.renderBeanFormError(w, r, viewbeans.FormState{ID: id}, false, http.StatusUnsupportedMediaType)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		state := viewbeans.FormState{ID: id, FormError: message}
		h.renderBeanFormError(w, r, state, false, status)
		return
	}

	state, model, ok := parseBeanForm(r, id)
	if !ok {
		h.renderBeanFormError(w, r, state, false, http.StatusBadRequest)
		return
	}

	updated, err := h.BeanService.UpdateBeanById(r.Context(), id, model)
	if err != nil {
		we := mapDomainError(err)
		if field := beanErrorField(err); field != "" {
			state.Errors[field] = we.Message
		} else {
			state.FormError = we.Message
		}
		h.renderBeanFormError(w, r, state, false, we.Status)
		return
	}

	w.Header().Set("HX-Reswap", "none")
	w.Header().Set("HX-Trigger", "dialog-close")
	writeHTMLStatus(w, http.StatusOK)
	_ = viewbeans.Row(*updated, "replace").Render(r.Context(), w)
	_ = shared.SuccessAlertOOB("Beans successfully updated.").Render(r.Context(), w)
}

func (h *Handler) renderBeanFormError(w http.ResponseWriter, r *http.Request, state viewbeans.FormState, isAdd bool, status int) {
	roasters, err := h.beanRoasterOptions(r)
	if err != nil {
		roasters = nil
	}
	writeHTMLStatus(w, status)
	_ = viewbeans.Form(state, roasters, isAdd, "", "").Render(r.Context(), w)
}

// DeleteBean handles DELETE /beans/delete/:id.
func (h *Handler) DeleteBean(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, http.StatusBadRequest)
		_ = shared.ErrorAlertOOB(errInvalidBeanID).Render(r.Context(), w)
		return
	}

	if err := h.BeanService.DeleteBeanById(r.Context(), id); err != nil {
		we := mapDeleteError(err, "These beans are still used by shots. Delete those shots first.")
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, we.Status)
		_ = shared.ErrorAlertOOB(we.Message).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = shared.SuccessAlertOOB("Beans successfully deleted.").Render(r.Context(), w)
}
