package bean

import (
	"context"
	dbsql "database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
)

func TestBeanRepositoryPostgresBehavior(t *testing.T) {
	roastDate := time.Date(2026, time.August, 18, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		run  func(t *testing.T, repository *Bean, mock sqlmock.Sqlmock)
	}{
		{
			name: "create returns postgres generated id",
			run: func(t *testing.T, repository *Bean, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES ($1, $2, $3, $4) RETURNING id").
					WithArgs("beans", 1, roastDate, sql.RoastLevelMedium).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(7))

				id, err := repository.CreateBeans(context.Background(), &sql.Beans{
					Name:       "beans",
					Roaster:    &sql.Roaster{Id: 1},
					RoastDate:  &roastDate,
					RoastLevel: sql.RoastLevelMedium,
				})
				if err != nil {
					t.Fatalf("CreateBeans() error = %v", err)
				}
				if id != 7 {
					t.Errorf("CreateBeans() id = %d, want 7", id)
				}
			},
		},
		{
			name: "create duplicate identity returns domain error",
			run: func(t *testing.T, repository *Bean, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES ($1, $2, $3, $4) RETURNING id").
					WithArgs("beans", 1, roastDate, sql.RoastLevelMedium).
					WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "uq_beans_identity"})

				_, err := repository.CreateBeans(context.Background(), &sql.Beans{
					Name:       "beans",
					Roaster:    &sql.Roaster{Id: 1},
					RoastDate:  &roastDate,
					RoastLevel: sql.RoastLevelMedium,
				})
				if !errors.Is(err, domainerrors.ErrBeansAlreadyExists) {
					t.Fatalf("CreateBeans() error = %v, want %v", err, domainerrors.ErrBeansAlreadyExists)
				}
			},
		},
		{
			name: "create with missing roaster returns domain error",
			run: func(t *testing.T, repository *Bean, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES ($1, $2, $3, $4) RETURNING id").
					WithArgs("beans", 2, roastDate, sql.RoastLevelMedium).
					WillReturnError(&pgconn.PgError{Code: "23503", ConstraintName: "beans_roaster_id_fkey"})

				_, err := repository.CreateBeans(context.Background(), &sql.Beans{
					Name:       "beans",
					Roaster:    &sql.Roaster{Id: 2},
					RoastDate:  &roastDate,
					RoastLevel: sql.RoastLevelMedium,
				})
				if !errors.Is(err, domainerrors.ErrRoasterDoesNotExist) {
					t.Fatalf("CreateBeans() error = %v, want %v", err, domainerrors.ErrRoasterDoesNotExist)
				}
			},
		},
		{
			name: "update to duplicate identity returns domain error",
			run: func(t *testing.T, repository *Bean, mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET name = $1, roaster_id = $2, roast_date = $3, roast_level = $4 WHERE id = $5").
					WithArgs("beans", 1, roastDate, sql.RoastLevelMedium, 7).
					WillReturnError(&pgconn.PgError{Code: "23505", ConstraintName: "uq_beans_identity"})

				_, err := repository.UpdateBeansById(context.Background(), 7, &sql.Beans{
					Name:       "beans",
					Roaster:    &sql.Roaster{Id: 1},
					RoastDate:  &roastDate,
					RoastLevel: sql.RoastLevelMedium,
				})
				if !errors.Is(err, domainerrors.ErrBeansAlreadyExists) {
					t.Fatalf("UpdateBeansById() error = %v, want %v", err, domainerrors.ErrBeansAlreadyExists)
				}
			},
		},
		{
			name: "get missing beans returns domain error",
			run: func(t *testing.T, repository *Bean, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("\nSELECT\n\tbeans.id,\n\tbeans.name,\n\tbeans.roast_date,\n\tbeans.roast_level,\n\tbeans.created_at,\n\tbeans.updated_at,\n\troaster.id AS \"roaster.id\",\n\troaster.name AS \"roaster.name\",\n\troaster.created_at AS \"roaster.created_at\",\n\troaster.updated_at AS \"roaster.updated_at\"\nFROM beans\n\tINNER JOIN roasters roaster\n\t\tON beans.roaster_id = roaster.id\nWHERE\n\tbeans.id = $1").
					WithArgs(42).
					WillReturnError(dbsql.ErrNoRows)

				_, err := repository.GetBeansById(context.Background(), 42)
				if !errors.Is(err, domainerrors.ErrBeansDoesNotExist) {
					t.Fatalf("GetBeansById() error = %v, want %v", err, domainerrors.ErrBeansDoesNotExist)
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
