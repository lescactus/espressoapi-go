package shot

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
)

func TestShotRepositoryPostgresGetShotsBySheetId(t *testing.T) {
	expectQuery := `
SELECT
	shots.id,
	shots.grind_setting,
	shots.quantity_in,
	shots.quantity_out,
	shots.shot_time,
	shots.water_temperature,
	shots.rating,
	shots.is_too_bitter,
	shots.is_too_sour,
	shots.comparison_with_previous_result,
	shots.additional_notes,
	shots.created_at,
	shots.updated_at,
	sheet.id as "sheet.id",
	sheet.name as "sheet.name",
	beans.id as "beans.id",
	beans.name as "beans.name",
	beans.roast_date as "beans.roast_date",
	beans.roast_level as "beans.roast_level",
	roaster.id AS "beans.roaster.id",
	roaster.name AS "beans.roaster.name",
	roaster.created_at AS "beans.roaster.created_at",
	roaster.updated_at AS "beans.roaster.updated_at"
FROM shots
INNER JOIN
	sheets sheet ON shots.sheet_id = sheet.id
INNER JOIN
	beans beans ON shots.beans_id = beans.id
INNER JOIN
	roasters roaster ON beans.roaster_id = roaster.id
WHERE shots.sheet_id = $1`

	tests := []struct {
		name string
		run  func(t *testing.T, repository *Shot, mock sqlmock.Sqlmock)
	}{
		{
			name: "uses the postgres $1 placeholder and returns an empty slice for a sheet without shots",
			run: func(t *testing.T, repository *Shot, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WithArgs(1).WillReturnRows(
					sqlmock.NewRows([]string{"id", "grind_setting", "quantity_in", "quantity_out", "shot_time", "water_temperature", "rating", "is_too_bitter", "is_too_sour", "comparison_with_previous_result", "additional_notes", "sheet.id", "sheet.name", "beans.id", "beans.name", "beans.roast_date", "beans.roast_level", "beans.roaster.id", "beans.roaster.name", "beans.roaster.created_at", "beans.roaster.updated_at"}),
				)

				got, err := repository.GetShotsBySheetId(context.Background(), 1)
				if err != nil {
					t.Fatalf("GetShotsBySheetId() error = %v", err)
				}
				if len(got) != 0 {
					t.Errorf("GetShotsBySheetId() = %v, want an empty slice", got)
				}
			},
		},
		{
			name: "returns the hydrated rows scoped to the given sheet",
			run: func(t *testing.T, repository *Shot, mock sqlmock.Sqlmock) {
				now := time.Now()
				mock.ExpectQuery(expectQuery).WithArgs(1).WillReturnRows(
					sqlmock.NewRows([]string{"id", "grind_setting", "quantity_in", "quantity_out", "shot_time", "water_temperature", "rating", "is_too_bitter", "is_too_sour", "comparison_with_previous_result", "additional_notes", "sheet.id", "sheet.name", "beans.id", "beans.name", "beans.roast_date", "beans.roast_level"}).
						AddRow(1, 11, 18.0, 36.0, int64(25000), 90.0, 4.5, false, true, sql.Better, "This is a test", 1, "sheet01", 1, "beans01", now, sql.RoastLevelLight),
				)

				want := []sql.Shot{
					{
						Id:                           1,
						GrindSetting:                 11,
						QuantityIn:                   18.0,
						QuantityOut:                  36.0,
						ShotTime:                     25 * time.Second,
						WaterTemperature:             90.0,
						Rating:                       4.5,
						IsTooBitter:                  false,
						IsTooSour:                    true,
						ComparisonWithPreviousResult: sql.Better,
						AdditionalNotes:              "This is a test",
						Sheet: &sql.Sheet{
							Id:   1,
							Name: "sheet01",
						},
						Beans: &sql.Beans{
							Id:         1,
							Name:       "beans01",
							RoastDate:  &now,
							RoastLevel: sql.RoastLevelLight,
						},
					},
				}

				got, err := repository.GetShotsBySheetId(context.Background(), 1)
				if err != nil {
					t.Fatalf("GetShotsBySheetId() error = %v", err)
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("GetShotsBySheetId() = %v, want %v", got, want)
				}
			},
		},
		{
			name: "propagates a database error",
			run: func(t *testing.T, repository *Shot, mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WithArgs(1).WillReturnError(errors.New("mock error"))

				if _, err := repository.GetShotsBySheetId(context.Background(), 1); err == nil {
					t.Fatal("GetShotsBySheetId() expected an error, got nil")
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
