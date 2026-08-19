// Package shots renders the shots list, dialog form, and reusable table
// templates (shared by /shots and the sheet detail page). It is imported
// into internal/controllers/web as viewshots to avoid clashing with the
// services/shot package.
package shots

import "strconv"

// FormState carries a shot add/edit form's submitted values and per-field
// validation errors so invalid input can be redisplayed after a 400/409
// response. When SheetLocked is true, the sheet is fixed (e.g. added from
// the sheet detail page) and submitted via a hidden field, not a <select>.
type FormState struct {
	ID                           int
	SheetID                      string
	SheetName                    string
	SheetLocked                  bool
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
