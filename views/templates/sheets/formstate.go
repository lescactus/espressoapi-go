// Package sheets renders the sheets list, detail, and inline row templates.
// It is imported into internal/controllers/web as viewsheets to avoid
// clashing with the services/sheet package.
package sheets

import "strconv"

// FormState carries a sheet add/edit form's submitted values and any
// validation error so invalid input can be redisplayed after a 400/409
// response.
type FormState struct {
	ID    int
	Name  string
	Error string
}

func rowElementID(id int) string { return "sheet-row-" + strconv.Itoa(id) }
func updatePath(id int) string   { return "/sheets/update/" + strconv.Itoa(id) }
func deletePath(id int) string   { return "/sheets/delete/" + strconv.Itoa(id) }
func getPath(id int) string      { return "/sheets/get/" + strconv.Itoa(id) }
