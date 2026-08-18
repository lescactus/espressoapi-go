package postgreserrors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
)

func TestParsePostgresError(t *testing.T) {
	fallback := errors.New("fallback")

	tests := []struct {
		name     string
		err      error
		entity   *Entity
		fallback error
		want     error
	}{
		{
			name: "nil error",
		},
		{
			name:     "non postgres error",
			err:      errors.New("other"),
			fallback: fallback,
			want:     fallback,
		},
		{
			name:     "duplicate beans",
			err:      fmt.Errorf("insert: %w", &pgconn.PgError{Code: "23505"}),
			entity:   &EntityBeans,
			fallback: fallback,
			want:     domainerrors.ErrBeansAlreadyExists,
		},
		{
			name:     "missing roaster",
			err:      &pgconn.PgError{Code: "23503", ConstraintName: "beans_roaster_id_fkey"},
			entity:   &EntityBeans,
			fallback: fallback,
			want:     domainerrors.ErrRoasterDoesNotExist,
		},
		{
			name:     "missing sheet",
			err:      &pgconn.PgError{Code: "23503", ConstraintName: "shots_sheet_id_fkey"},
			entity:   &EntityShot,
			fallback: fallback,
			want:     domainerrors.ErrSheetDoesNotExist,
		},
		{
			name:     "missing beans",
			err:      &pgconn.PgError{Code: "23503", ConstraintName: "shots_beans_id_fkey"},
			entity:   &EntityShot,
			fallback: fallback,
			want:     domainerrors.ErrBeansDoesNotExist,
		},
		{
			name:     "referenced beans prevent delete",
			err:      &pgconn.PgError{Code: "23503", TableName: "beans"},
			fallback: fallback,
			want:     domainerrors.ErrBeansForeignKeyConstraint,
		},
		{
			name:     "referenced shots prevent delete",
			err:      &pgconn.PgError{Code: "23503", TableName: "shots"},
			fallback: fallback,
			want:     domainerrors.ErrShotForeignKeyConstraint,
		},
		{
			name:     "beans roast level check",
			err:      &pgconn.PgError{Code: "23514", ConstraintName: "chk_beans_roast_level"},
			fallback: fallback,
			want:     domainerrors.ErrBeansRoastLevelOutOfRange,
		},
		{
			name:     "shot comparison check",
			err:      &pgconn.PgError{Code: "23514", ConstraintName: "chk_shots_comparison_with_previous_result"},
			fallback: fallback,
			want:     domainerrors.ErrShotComparisonWithPreviousResultOutOfRange,
		},
		{
			name:     "unknown constraint",
			err:      &pgconn.PgError{Code: "23514", ConstraintName: "other"},
			fallback: fallback,
			want:     fallback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePostgresError(tt.err, tt.entity, tt.fallback)
			if !errors.Is(got, tt.want) {
				t.Errorf("ParsePostgresError() = %v, want %v", got, tt.want)
			}
		})
	}
}
