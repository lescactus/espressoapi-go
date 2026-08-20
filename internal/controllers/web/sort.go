package web

import "slices"

// normalizeSortColumn returns col if it is in the resource's whitelist,
// otherwise "id" (the default, always-present sort column).
func normalizeSortColumn(col string, whitelist []string) string {
	if slices.Contains(whitelist, col) {
		return col
	}
	return "id"
}

// normalizeSortOrder returns "desc" only for an explicit "desc" value;
// everything else (missing, "asc", or unknown) defaults to "asc".
func normalizeSortOrder(order string) string {
	if order == "desc" {
		return "desc"
	}
	return "asc"
}
