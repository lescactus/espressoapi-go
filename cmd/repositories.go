package cmd

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/config"
	"github.com/lescactus/espressoapi-go/internal/repository"
	mysqlbean "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/bean"
	mysqlroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/roaster"
	mysqlsheet "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/sheet"
	mysqlshot "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/shot"
	postgresbean "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/bean"
	postgresroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/roaster"
	postgressheet "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/sheet"
	postgresshot "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/shot"
)

type repositorySet struct {
	sheet   repository.SheetRepository
	roaster repository.RoasterRepository
	beans   repository.BeansRepository
	shot    repository.ShotRepository
}

func newRepositorySet(databaseType config.DatabaseType, db *sqlx.DB) (repositorySet, error) {
	switch databaseType {
	case config.DatabaseTypeMySQL:
		return repositorySet{
			sheet:   mysqlsheet.New(db),
			roaster: mysqlroaster.New(db),
			beans:   mysqlbean.New(db),
			shot:    mysqlshot.New(db),
		}, nil
	case config.DatabaseTypePostgres:
		return repositorySet{
			sheet:   postgressheet.New(db),
			roaster: postgresroaster.New(db),
			beans:   postgresbean.New(db),
			shot:    postgresshot.New(db),
		}, nil
	default:
		return repositorySet{}, fmt.Errorf("unsupported database type %q", databaseType)
	}
}
