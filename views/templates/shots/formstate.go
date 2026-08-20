// Package shots renders the shots list, dialog form, and reusable table
// templates (shared by /shots and the sheet detail page). It is imported
// into internal/controllers/web as viewshots to avoid clashing with the
// services/shot package.
package shots

import "strconv"

// ViewContextSheetShots is the closed view_context value for a shot
// add/edit form (and the resulting OOB row) opened from the sheet detail
// page's scoped shots table, which has no Sheet column. Any other value
// (including a tampered one) falls back to the standalone /shots table.
const ViewContextSheetShots = "sheet-shots"

// FormState carries a shot add/edit form's submitted values and per-field
// validation errors so invalid input can be redisplayed after a 400/409
// response. When SheetLocked is true, the sheet is fixed (e.g. added from
// the sheet detail page) and submitted via a hidden field, not a <select>.
// ViewContext round-trips which shots table (standalone list or a sheet's
// scoped table) the form was opened from, so the created/updated row is
// rendered with the matching column set.
type FormState struct {
	ID                           int
	SheetID                      string
	SheetName                    string
	SheetLocked                  bool
	ViewContext                  string
	BeansID                      string
	GrindSetting                 string
	QuantityIn                   string
	QuantityOut                  string
	ShotTimeSeconds              string
	WaterTemperature             string
	Rating                       string
	IsTooBitter                  bool
	IsTooSour                    bool
	ComparisonWithPreviousResult string
	AdditionalNotes              string
	Errors                       map[string]string
}

func (s FormState) fieldError(field string) string {
	if s.Errors == nil {
		return ""
	}
	return s.Errors[field]
}

func rowElementID(id int) string { return "shot-row-" + strconv.Itoa(id) }
func updatePath(id int) string   { return "/shots/update/" + strconv.Itoa(id) }
func deletePath(id int) string   { return "/shots/delete/" + strconv.Itoa(id) }

// editPath is the Edit link's target: it carries view_context=sheet-shots
// when the row is rendered without the Sheet column, so the edit dialog
// (and its resulting OOB row) knows to keep matching that column set.
func editPath(id int, showSheetColumn bool) string {
	if showSheetColumn {
		return updatePath(id)
	}
	return updatePath(id) + "?view_context=" + ViewContextSheetShots
}
