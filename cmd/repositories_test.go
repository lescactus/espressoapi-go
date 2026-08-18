package cmd

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/config"
	mysqlbean "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/bean"
	mysqlroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/roaster"
	mysqlsheet "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/sheet"
	mysqlshot "github.com/lescactus/espressoapi-go/internal/repository/sql/mysql/shot"
	postgresbean "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/bean"
	postgresroaster "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/roaster"
	postgressheet "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/sheet"
	postgresshot "github.com/lescactus/espressoapi-go/internal/repository/sql/postgresql/shot"
)

func TestNewRepositorySet(t *testing.T) {
	tests := []struct {
		name         string
		databaseType config.DatabaseType
		assert       func(t *testing.T, repositories repositorySet)
		wantErr      bool
	}{
		{
			name:         "mysql",
			databaseType: config.DatabaseTypeMySQL,
			assert: func(t *testing.T, repositories repositorySet) {
				if _, ok := repositories.sheet.(*mysqlsheet.Sheet); !ok {
					t.Errorf("sheet repository = %T, want *mysqlsheet.Sheet", repositories.sheet)
				}
				if _, ok := repositories.roaster.(*mysqlroaster.Roaster); !ok {
					t.Errorf("roaster repository = %T, want *mysqlroaster.Roaster", repositories.roaster)
				}
				if _, ok := repositories.beans.(*mysqlbean.Bean); !ok {
					t.Errorf("beans repository = %T, want *mysqlbean.Bean", repositories.beans)
				}
				if _, ok := repositories.shot.(*mysqlshot.Shot); !ok {
					t.Errorf("shot repository = %T, want *mysqlshot.Shot", repositories.shot)
				}
			},
		},
		{
			name:         "postgres",
			databaseType: config.DatabaseTypePostgres,
			assert: func(t *testing.T, repositories repositorySet) {
				if _, ok := repositories.sheet.(*postgressheet.Sheet); !ok {
					t.Errorf("sheet repository = %T, want *postgressheet.Sheet", repositories.sheet)
				}
				if _, ok := repositories.roaster.(*postgresroaster.Roaster); !ok {
					t.Errorf("roaster repository = %T, want *postgresroaster.Roaster", repositories.roaster)
				}
				if _, ok := repositories.beans.(*postgresbean.Bean); !ok {
					t.Errorf("beans repository = %T, want *postgresbean.Bean", repositories.beans)
				}
				if _, ok := repositories.shot.(*postgresshot.Shot); !ok {
					t.Errorf("shot repository = %T, want *postgresshot.Shot", repositories.shot)
				}
			},
		},
		{
			name:         "unsupported database type",
			databaseType: "sqlite",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repositories, err := newRepositorySet(tt.databaseType, &sqlx.DB{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("newRepositorySet() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				tt.assert(t, repositories)
			}
		})
	}
}
