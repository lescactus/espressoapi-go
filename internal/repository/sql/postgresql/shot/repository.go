package shot

import (
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/adapters"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/shared"
)

var _ repository.ShotRepository = (*Shot)(nil)

type Shot struct {
	*shared.Shot
}

func New(db *sqlx.DB) *Shot {
	return &Shot{shared.NewShot(db, adapters.PostgreSQL())}
}
