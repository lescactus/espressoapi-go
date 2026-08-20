// Package beans renders the beans list and dialog add/edit form templates.
// It is imported into internal/controllers/web as viewbeans to avoid
// clashing with the services/bean package.
package beans

import "strconv"

// FormState carries a bean add/edit form's submitted values and per-field
// validation errors so invalid input can be redisplayed after a 400/409
// response. FormError holds an error not tied to any single field (a
// malformed or oversized request body, or an unexpected failure).
type FormState struct {
	ID         int
	Name       string
	RoasterID  string
	RoastDate  string
	RoastLevel string
	Errors     map[string]string
	FormError  string
}

func (s FormState) fieldError(field string) string {
	if s.Errors == nil {
		return ""
	}
	return s.Errors[field]
}

func rowElementID(id int) string { return "bean-row-" + strconv.Itoa(id) }
func updatePath(id int) string   { return "/beans/update/" + strconv.Itoa(id) }
func deletePath(id int) string   { return "/beans/delete/" + strconv.Itoa(id) }
