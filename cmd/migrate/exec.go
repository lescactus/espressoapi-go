package migrate

import (
	"github.com/lescactus/espressoapi-go/cmd/app"
	sqlmigrate "github.com/rubenv/sql-migrate"
)

func execMigrations(direction sqlmigrate.MigrationDirection) (int, error) {
	migrations, err := migrationSource(app.App.MigrationsFS, app.App.Cfg.DatabaseType)
	if err != nil {
		return 0, err
	}

	var n int
	if version >= 0 {
		n, err = sqlmigrate.ExecVersion(app.App.Db.DB, string(app.App.Cfg.DatabaseType), migrations, direction, version)
	} else {
		n, err = sqlmigrate.ExecMax(app.App.Db.DB, string(app.App.Cfg.DatabaseType), migrations, direction, limit)
	}
	if err != nil {
		return n, err
	}

	return n, nil
}
