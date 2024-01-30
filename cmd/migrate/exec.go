package migrate

import (
	"embed"

	"github.com/lescactus/espressoapi-go/cmd/app"
	sqlmigrate "github.com/rubenv/sql-migrate"
)

func execMigrations(direction sqlmigrate.MigrationDirection) (int, error) {
	migrations := sqlmigrate.EmbedFileSystemMigrationSource{
		FileSystem: (*app.App.MigrationsFS).(embed.FS),
		Root:       "migrations/sql/mysql",
	}

	var n int
	var err error
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
