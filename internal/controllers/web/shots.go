package web

import (
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
	"github.com/lescactus/espressoapi-go/views/templates/shared"
	viewshots "github.com/lescactus/espressoapi-go/views/templates/shots"
)

var shotSortColumns = []string{
	"id", "grind_setting", "quantity_in", "quantity_out", "shot_time",
	"water_temperature", "rating", "created_at", "updated_at",
}

func sortShots(shots []shot.Shot, col, order string) {
	col = normalizeSortColumn(col, shotSortColumns)
	less := func(i, j int) bool { return shotLess(shots[i], shots[j], col) }
	if normalizeSortOrder(order) == "desc" {
		less = func(i, j int) bool { return shotLess(shots[j], shots[i], col) }
	}
	sort.SliceStable(shots, less)
}

func shotLess(a, b shot.Shot, col string) bool {
	switch col {
	case "grind_setting":
		return a.GrindSetting < b.GrindSetting
	case "quantity_in":
		return a.QuantityIn < b.QuantityIn
	case "quantity_out":
		return a.QuantityOut < b.QuantityOut
	case "shot_time":
		return a.ShotTime < b.ShotTime
	case "water_temperature":
		return a.WaterTemperature < b.WaterTemperature
	case "rating":
		return a.Rating < b.Rating
	case "created_at":
		return timeLess(a.CreatedAt, b.CreatedAt)
	case "updated_at":
		return timeLess(a.UpdatedAt, b.UpdatedAt)
	default:
		return a.Id < b.Id
	}
}

const errInvalidShotID = "The shot id must be a positive number."

func (h *Handler) shotFormOptions(r *http.Request) ([]sheet.Sheet, []bean.Bean, error) {
	sheets, err := h.SheetService.GetAllSheets(r.Context())
	if err != nil {
		return nil, nil, err
	}
	sort.SliceStable(sheets, func(i, j int) bool { return sheets[i].Id < sheets[j].Id })

	beans, err := h.BeanService.GetAllBeans(r.Context())
	if err != nil {
		return nil, nil, err
	}
	sort.SliceStable(beans, func(i, j int) bool { return beans[i].Id < beans[j].Id })

	return sheets, beans, nil
}

// ListShots handles GET /shots.
func (h *Handler) ListShots(w http.ResponseWriter, r *http.Request) {
	shots, err := h.ShotService.GetAllShots(r.Context())
	if err != nil {
		h.writeFullPageError(w, r, mapDomainError(err))
		return
	}
	sortCol := normalizeSortColumn(r.URL.Query().Get("sort"), shotSortColumns)
	order := normalizeSortOrder(r.URL.Query().Get("order"))
	sortShots(shots, sortCol, order)

	writeHTMLStatus(w, http.StatusOK)
	if isHXRequest(r) {
		_ = viewshots.Table(shots, sortCol, order, true, true).Render(r.Context(), w)
		return
	}
	_ = viewshots.Page(shots, sortCol, order, nil).Render(r.Context(), w)
}

// shotsListForPage fetches and default-sorts the full shot list, for the
// full-page fallback of a direct GET to an add/edit dialog route.
func (h *Handler) shotsListForPage(r *http.Request) ([]shot.Shot, error) {
	shots, err := h.ShotService.GetAllShots(r.Context())
	if err != nil {
		return nil, err
	}
	sortShots(shots, "id", "asc")
	return shots, nil
}

// AddShotForm handles GET /shots/add. An optional ?sheet_id= query param
// locks the sheet (used from the sheet detail page's "Add shot" button).
// htmx requests get the dialog form fragment; direct navigation gets the
// full shots list page with the dialog pre-opened.
func (h *Handler) AddShotForm(w http.ResponseWriter, r *http.Request) {
	sheets, beans, err := h.shotFormOptions(r)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	state := viewshots.FormState{}
	if lockedID := r.URL.Query().Get("sheet_id"); lockedID != "" {
		if id, err := strconv.Atoi(lockedID); err == nil && id > 0 {
			if s, err := h.SheetService.GetSheetById(r.Context(), id); err == nil {
				state.SheetLocked = true
				state.SheetID = strconv.Itoa(s.Id)
				state.SheetName = s.Name
				state.ViewContext = viewshots.ViewContextSheetShots
			}
		}
	}
	form := viewshots.Form(state, sheets, beans, true, "", "")

	if !isHXRequest(r) {
		allShots, err := h.shotsListForPage(r)
		if err != nil {
			h.writeFullPageError(w, r, mapDomainError(err))
			return
		}
		// The full-page fallback always renders the standalone (17-column,
		// Sheet column included) shots page, even for a sheet-locked add, so
		// clear ViewContext here: a submission from this page must render its
		// OOB row with the Sheet column, not assume the sheet-detail page's
		// 16-column shape the hidden view_context field would otherwise carry.
		fallbackState := state
		fallbackState.ViewContext = ""
		fallbackForm := viewshots.Form(fallbackState, sheets, beans, true, "", "")
		writeHTMLStatus(w, http.StatusOK)
		_ = viewshots.Page(allShots, "id", "asc", fallbackForm).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = form.Render(r.Context(), w)
}

// parseShotForm extracts and validates shot form fields, returning the raw
// FormState (for redisplay) and, on success, the parsed service model.
// Rating, comparison, and shot_time range checks are intentionally left to
// the service (it already validates and returns the same domain errors as
// the REST API).
func parseShotForm(r *http.Request, id int) (viewshots.FormState, *shot.Shot, bool) {
	state := viewshots.FormState{
		ID:                           id,
		ViewContext:                  strings.TrimSpace(r.PostFormValue("view_context")),
		SheetID:                      strings.TrimSpace(r.PostFormValue("sheet_id")),
		BeansID:                      strings.TrimSpace(r.PostFormValue("beans_id")),
		GrindSetting:                 strings.TrimSpace(r.PostFormValue("grind_setting")),
		QuantityIn:                   strings.TrimSpace(r.PostFormValue("quantity_in")),
		QuantityOut:                  strings.TrimSpace(r.PostFormValue("quantity_out")),
		ShotTimeSeconds:              strings.TrimSpace(r.PostFormValue("shot_time")),
		WaterTemperature:             strings.TrimSpace(r.PostFormValue("water_temperature")),
		Rating:                       strings.TrimSpace(r.PostFormValue("rating")),
		IsTooBitter:                  r.PostFormValue("is_too_bitter") != "",
		IsTooSour:                    r.PostFormValue("is_too_sour") != "",
		ComparisonWithPreviousResult: strings.TrimSpace(r.PostFormValue("comparison_with_previous_result")),
		AdditionalNotes:              r.PostFormValue("additional_notes"),
		Errors:                       map[string]string{},
	}

	sheetID, err := strconv.Atoi(state.SheetID)
	if err != nil || sheetID <= 0 {
		state.Errors["sheet_id"] = "Select a sheet."
	}

	beansID, err := strconv.Atoi(state.BeansID)
	if err != nil || beansID <= 0 {
		state.Errors["beans_id"] = "Select beans."
	}

	grindSetting, err := strconv.Atoi(state.GrindSetting)
	if err != nil {
		state.Errors["grind_setting"] = "Grind setting must be a whole number."
	}

	quantityIn, err := strconv.ParseFloat(state.QuantityIn, 64)
	if err != nil || math.IsNaN(quantityIn) || math.IsInf(quantityIn, 0) || quantityIn < 0 {
		state.Errors["quantity_in"] = "Quantity in must be a non-negative number."
	}

	quantityOut, err := strconv.ParseFloat(state.QuantityOut, 64)
	if err != nil || math.IsNaN(quantityOut) || math.IsInf(quantityOut, 0) || quantityOut < 0 {
		state.Errors["quantity_out"] = "Quantity out must be a non-negative number."
	}

	var shotTime time.Duration
	if state.ShotTimeSeconds != "" {
		seconds, err := strconv.ParseFloat(state.ShotTimeSeconds, 64)
		if err != nil || math.IsNaN(seconds) || math.IsInf(seconds, 0) {
			state.Errors["shot_time"] = "Shot time must be a number."
		} else {
			shotTime = shot.SecondsToDuration(seconds)
		}
	}

	var waterTemperature float64
	if state.WaterTemperature != "" {
		waterTemperature, err = strconv.ParseFloat(state.WaterTemperature, 64)
		if err != nil || math.IsNaN(waterTemperature) || math.IsInf(waterTemperature, 0) {
			state.Errors["water_temperature"] = "Water temperature must be a number."
		}
	}

	rating, err := strconv.ParseFloat(state.Rating, 64)
	if err != nil || math.IsNaN(rating) || math.IsInf(rating, 0) {
		state.Errors["rating"] = "Rating must be a number."
	}

	comparison := 0
	if state.ComparisonWithPreviousResult == "" {
		state.Errors["comparison_with_previous_result"] = "Select a comparison value."
	} else if n, err := strconv.Atoi(state.ComparisonWithPreviousResult); err != nil || n < int(sql.Worst) || n > int(sql.Unknown) {
		state.Errors["comparison_with_previous_result"] = "Invalid comparison value."
	} else {
		comparison = n
	}

	if len(state.AdditionalNotes) > 511 {
		state.Errors["additional_notes"] = "Additional notes must be 511 characters or fewer."
	}

	if len(state.Errors) > 0 {
		return state, nil, false
	}

	return state, &shot.Shot{
		Id:                           id,
		Sheet:                        &sheet.Sheet{Id: sheetID},
		Beans:                        &bean.Bean{Id: beansID},
		GrindSetting:                 grindSetting,
		QuantityIn:                   quantityIn,
		QuantityOut:                  quantityOut,
		ShotTime:                     shotTime,
		WaterTemperature:             waterTemperature,
		Rating:                       rating,
		IsTooBitter:                  state.IsTooBitter,
		IsTooSour:                    state.IsTooSour,
		ComparisonWithPreviousResult: sql.ComparisonWithPreviousResult(comparison),
		AdditionalNotes:              state.AdditionalNotes,
	}, true
}

// CreateShot handles POST /shots/add.
func (h *Handler) CreateShot(w http.ResponseWriter, r *http.Request) {
	if !isFormURLEncoded(r) {
		h.renderShotFormError(w, r, viewshots.FormState{}, true, http.StatusUnsupportedMediaType)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		state := viewshots.FormState{FormError: message}
		h.renderShotFormError(w, r, state, true, status)
		return
	}

	state, model, ok := parseShotForm(r, 0)
	if !ok {
		h.renderShotFormError(w, r, state, true, http.StatusBadRequest)
		return
	}

	created, err := h.ShotService.CreateShot(r.Context(), model)
	if err != nil {
		we := mapDomainError(err)
		if field := shotErrorField(err); field != "" {
			state.Errors[field] = we.Message
		} else {
			state.FormError = we.Message
		}
		h.renderShotFormError(w, r, state, true, we.Status)
		return
	}

	w.Header().Set("HX-Reswap", "none")
	w.Header().Set("HX-Trigger", "dialog-close")
	writeHTMLStatus(w, http.StatusOK)
	showSheetColumn := state.ViewContext != viewshots.ViewContextSheetShots
	_ = viewshots.Row(*created, showSheetColumn, "insert").Render(r.Context(), w)
	_ = shared.SuccessAlertOOB("Shot successfully created.").Render(r.Context(), w)
}

// GetShot handles GET /shots/get/:id.
func (h *Handler) GetShot(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidShotID})
		return
	}
	s, err := h.ShotService.GetShotById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	if !isHXRequest(r) {
		_ = viewshots.RowPage(*s).Render(r.Context(), w)
		return
	}
	_ = viewshots.Row(*s, true, "").Render(r.Context(), w)
}

// EditShotForm handles GET /shots/update/:id.
func (h *Handler) EditShotForm(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		h.writeGetError(w, r, webError{Status: http.StatusBadRequest, Message: errInvalidShotID})
		return
	}
	s, err := h.ShotService.GetShotById(r.Context(), id)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}
	sheets, beans, err := h.shotFormOptions(r)
	if err != nil {
		h.writeGetError(w, r, mapDomainError(err))
		return
	}

	state := viewshots.FormState{
		ID:                           s.Id,
		GrindSetting:                 strconv.Itoa(s.GrindSetting),
		QuantityIn:                   strconv.FormatFloat(s.QuantityIn, 'f', 1, 64),
		QuantityOut:                  strconv.FormatFloat(s.QuantityOut, 'f', 1, 64),
		ShotTimeSeconds:              strconv.FormatFloat(s.ShotTime.Seconds(), 'f', 1, 64),
		WaterTemperature:             strconv.FormatFloat(s.WaterTemperature, 'f', 1, 64),
		Rating:                       strconv.FormatFloat(s.Rating, 'f', 1, 64),
		IsTooBitter:                  s.IsTooBitter,
		IsTooSour:                    s.IsTooSour,
		ComparisonWithPreviousResult: strconv.Itoa(int(s.ComparisonWithPreviousResult)),
		AdditionalNotes:              s.AdditionalNotes,
	}
	if s.Sheet != nil {
		state.SheetID = strconv.Itoa(s.Sheet.Id)
	}
	if s.Beans != nil {
		state.BeansID = strconv.Itoa(s.Beans.Id)
	}
	if r.URL.Query().Get("view_context") == viewshots.ViewContextSheetShots {
		state.ViewContext = viewshots.ViewContextSheetShots
	}
	form := viewshots.Form(state, sheets, beans, false, shared.FormatTimestamp(s.CreatedAt), shared.FormatTimestamp(s.UpdatedAt))

	if !isHXRequest(r) {
		allShots, err := h.shotsListForPage(r)
		if err != nil {
			h.writeFullPageError(w, r, mapDomainError(err))
			return
		}
		// See AddShotForm: the full-page fallback always renders the
		// standalone (17-column) shots page, so clear ViewContext for the
		// form rendered on it.
		fallbackState := state
		fallbackState.ViewContext = ""
		fallbackForm := viewshots.Form(fallbackState, sheets, beans, false, shared.FormatTimestamp(s.CreatedAt), shared.FormatTimestamp(s.UpdatedAt))
		writeHTMLStatus(w, http.StatusOK)
		_ = viewshots.Page(allShots, "id", "asc", fallbackForm).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = form.Render(r.Context(), w)
}

// UpdateShot handles PUT /shots/update/:id.
func (h *Handler) UpdateShot(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		writeHTMLStatus(w, http.StatusBadRequest)
		w.Header().Set("HX-Reswap", "none")
		_ = shared.ErrorAlertOOB(errInvalidShotID).Render(r.Context(), w)
		return
	}

	if !isFormURLEncoded(r) {
		h.renderShotFormError(w, r, viewshots.FormState{ID: id}, false, http.StatusUnsupportedMediaType)
		return
	}
	if err := r.ParseForm(); err != nil {
		status, message := parseFormError(err)
		state := viewshots.FormState{ID: id, FormError: message}
		h.renderShotFormError(w, r, state, false, status)
		return
	}

	state, model, ok := parseShotForm(r, id)
	if !ok {
		h.renderShotFormError(w, r, state, false, http.StatusBadRequest)
		return
	}

	updated, err := h.ShotService.UpdateShotById(r.Context(), id, model)
	if err != nil {
		we := mapDomainError(err)
		if field := shotErrorField(err); field != "" {
			state.Errors[field] = we.Message
		} else {
			state.FormError = we.Message
		}
		h.renderShotFormError(w, r, state, false, we.Status)
		return
	}

	w.Header().Set("HX-Reswap", "none")
	w.Header().Set("HX-Trigger", "dialog-close")
	writeHTMLStatus(w, http.StatusOK)
	showSheetColumn := state.ViewContext != viewshots.ViewContextSheetShots
	_ = viewshots.Row(*updated, showSheetColumn, "replace").Render(r.Context(), w)
	_ = shared.SuccessAlertOOB("Shot successfully updated.").Render(r.Context(), w)
}

func (h *Handler) renderShotFormError(w http.ResponseWriter, r *http.Request, state viewshots.FormState, isAdd bool, status int) {
	sheets, beans, err := h.shotFormOptions(r)
	if err != nil {
		sheets, beans = nil, nil
	}
	writeHTMLStatus(w, status)
	_ = viewshots.Form(state, sheets, beans, isAdd, "", "").Render(r.Context(), w)
}

// DeleteShot handles DELETE /shots/delete/:id.
func (h *Handler) DeleteShot(w http.ResponseWriter, r *http.Request) {
	id, ok := parsePositiveID(r)
	if !ok {
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, http.StatusBadRequest)
		_ = shared.ErrorAlertOOB(errInvalidShotID).Render(r.Context(), w)
		return
	}

	if err := h.ShotService.DeleteShotById(r.Context(), id); err != nil {
		we := mapDomainError(err)
		w.Header().Set("HX-Reswap", "none")
		writeHTMLStatus(w, we.Status)
		_ = shared.ErrorAlertOOB(we.Message).Render(r.Context(), w)
		return
	}

	writeHTMLStatus(w, http.StatusOK)
	_ = shared.SuccessAlertOOB("Shot successfully deleted.").Render(r.Context(), w)
}
