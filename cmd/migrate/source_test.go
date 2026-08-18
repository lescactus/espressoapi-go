package migrate

import (
	"embed"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/lescactus/espressoapi-go/internal/config"
)

func TestMigrationSource(t *testing.T) {
	tests := []struct {
		name         string
		databaseType config.DatabaseType
		wantRoot     string
		wantErr      bool
	}{
		{
			name:         "mysql",
			databaseType: config.DatabaseTypeMySQL,
			wantRoot:     "migrations/sql/mysql",
		},
		{
			name:         "postgres",
			databaseType: config.DatabaseTypePostgres,
			wantRoot:     "migrations/sql/postgres",
		},
		{
			name:         "unsupported database type",
			databaseType: "sqlite",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := migrationSource(&embed.FS{}, tt.databaseType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("migrationSource() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && source.Root != tt.wantRoot {
				t.Errorf("migrationSource().Root = %q, want %q", source.Root, tt.wantRoot)
			}
		})
	}
}

func TestMigrationVersionParity(t *testing.T) {
	mysqlVersions := migrationVersions(t, "mysql")
	postgresVersions := migrationVersions(t, "postgres")
	if !reflect.DeepEqual(mysqlVersions, postgresVersions) {
		t.Errorf("migration versions differ: mysql=%v postgres=%v", mysqlVersions, postgresVersions)
	}
}

func migrationVersions(t *testing.T, databaseType string) []string {
	t.Helper()

	entries, err := os.ReadDir(filepath.Join("..", "..", "migrations", "sql", databaseType))
	if err != nil {
		t.Fatalf("read %s migrations: %v", databaseType, err)
	}

	versions := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".sql" {
			continue
		}
		versions = append(versions, strings.SplitN(entry.Name(), "-", 2)[0])
	}

	return versions
}
