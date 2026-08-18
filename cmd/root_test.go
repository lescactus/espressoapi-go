package cmd

import (
	"testing"

	"github.com/lescactus/espressoapi-go/internal/config"
)

func TestDatabaseDriverName(t *testing.T) {
	tests := []struct {
		name         string
		databaseType config.DatabaseType
		want         string
		wantErr      bool
	}{
		{
			name:         "mysql",
			databaseType: config.DatabaseTypeMySQL,
			want:         "mysql",
		},
		{
			name:         "postgres",
			databaseType: config.DatabaseTypePostgres,
			want:         "pgx",
		},
		{
			name:         "unsupported database type",
			databaseType: "sqlite",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := databaseDriverName(tt.databaseType)
			if (err != nil) != tt.wantErr {
				t.Fatalf("databaseDriverName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("databaseDriverName() = %q, want %q", got, tt.want)
			}
		})
	}
}
