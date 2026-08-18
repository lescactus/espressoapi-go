package roaster

import (
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/adapters"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/shared"
)

var _ repository.RoasterRepository = (*Roaster)(nil)

type Roaster struct {
	*shared.Roaster
}

func New(db *sqlx.DB) *Roaster {
	return &Roaster{shared.NewRoaster(db, adapters.PostgreSQL())}
}
