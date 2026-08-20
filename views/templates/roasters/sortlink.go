package roasters

func nextSortOrder(currentSort, currentOrder, col string) string {
	if currentSort == col && currentOrder == "asc" {
		return "desc"
	}
	return "asc"
}
