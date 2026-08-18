package bean

import (
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/adapters"
	"github.com/lescactus/espressoapi-go/internal/repository/sql/shared"
)

var _ repository.BeansRepository = (*Bean)(nil)

type Bean struct {
	*shared.Bean
}

func New(db *sqlx.DB) *Bean {
	return &Bean{shared.NewBean(db, adapters.MySQL())}
}
