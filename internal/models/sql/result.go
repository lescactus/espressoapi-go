package sql

type ComparaisonWithPreviousResult uint8

const (
	Worst ComparaisonWithPreviousResult = iota
	Same
	Better
	Unknown
)

type Result struct {
	Id                            int
	Rating                        float64
	IsTooBitter                   bool
	IsTooSour                     bool
	ComparaisonWithPreviousResult ComparaisonWithPreviousResult
	AdditionalNotes               string
}
