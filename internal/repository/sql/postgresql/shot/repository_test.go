package shot

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
)

func TestShotRepositoryPostgresCreate(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, repository *Shot, mock sqlmock.Sqlmock)
	}{
		{
			name: "create returns postgres generated id",
			run: func(t *testing.T, repository *Shot, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparison_with_previous_result, additional_notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id").
					WithArgs(1, 2, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "notes").
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(8))

				id, err := repository.CreateShot(context.Background(), &sql.Shot{
					Sheet:           &sql.Sheet{Id: 1},
					Beans:           &sql.Beans{Id: 2},
					AdditionalNotes: "notes",
				})
				if err != nil {
					t.Fatalf("CreateShot() error = %v", err)
				}
				if id != 8 {
					t.Errorf("CreateShot() id = %d, want 8", id)
				}
			},
		},
		{
			name: "create with missing sheet returns domain error",
			run: func(t *testing.T, repository *Shot, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparison_with_previous_result, additional_notes) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id").
					WithArgs(1, 2, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "notes").
					WillReturnError(&pgconn.PgError{Code: "23503", ConstraintName: "shots_sheet_id_fkey"})

				_, err := repository.CreateShot(context.Background(), &sql.Shot{
					Sheet:           &sql.Sheet{Id: 1},
					Beans:           &sql.Beans{Id: 2},
					AdditionalNotes: "notes",
				})
				if !errors.Is(err, domainerrors.ErrSheetDoesNotExist) {
					t.Fatalf("CreateShot() error = %v, want %v", err, domainerrors.ErrSheetDoesNotExist)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("sqlmock.New() error = %v", err)
			}
			defer db.Close()

			tt.run(t, New(sqlx.NewDb(db, "sqlmock")), mock)
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
