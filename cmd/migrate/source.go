package migrate

import (
	"embed"
	"fmt"

	"github.com/lescactus/espressoapi-go/internal/config"
	sqlmigrate "github.com/rubenv/sql-migrate"
)

func migrationSource(migrationFS *embed.FS, databaseType config.DatabaseType) (sqlmigrate.EmbedFileSystemMigrationSource, error) {
	root, err := migrationRoot(databaseType)
	if err != nil {
		return sqlmigrate.EmbedFileSystemMigrationSource{}, err
	}

	return sqlmigrate.EmbedFileSystemMigrationSource{
		FileSystem: *migrationFS,
		Root:       root,
	}, nil
}

func migrationRoot(databaseType config.DatabaseType) (string, error) {
	switch databaseType {
	case config.DatabaseTypeMySQL:
		return "migrations/sql/mysql", nil
	case config.DatabaseTypePostgres:
		return "migrations/sql/postgres", nil
	default:
		return "", fmt.Errorf("unsupported database type %q", databaseType)
	}
}
