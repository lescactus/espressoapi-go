package sheet

import (
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/adapters"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/shared"
)

var _ repository.SheetRepository = (*Sheet)(nil)

type Sheet struct {
	*shared.Sheet
}

func New(db *sqlx.DB) *Sheet {
	return &Sheet{shared.NewSheet(db, adapters.MySQL())}
}
