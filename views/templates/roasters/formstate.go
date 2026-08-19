// Package roasters renders the roasters list and inline row templates. It is
// imported into internal/controllers/web as viewroasters to avoid clashing
// with the services/roaster package.
package roasters

import "strconv"

// FormState carries a roaster add/edit form's submitted values and any
// validation error so invalid input can be redisplayed after a 400/409
// response.
type FormState struct {
	ID    int
	Name  string
	Error string
}

func rowElementID(id int) string { return "roaster-row-" + strconv.Itoa(id) }
func updatePath(id int) string   { return "/roasters/update/" + strconv.Itoa(id) }
func deletePath(id int) string   { return "/roasters/delete/" + strconv.Itoa(id) }
func getPath(id int) string      { return "/roasters/get/" + strconv.Itoa(id) }
