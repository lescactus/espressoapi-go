package sheet

import (
	"context"
	dbsql "database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
)

func TestSheetRepositoryPostgresBehavior(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, repository *Sheet, mock sqlmock.Sqlmock)
	}{
		{
			name: "create uses postgres placeholder",
			run: func(t *testing.T, repository *Sheet, mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sheets (name) VALUES ($1)").
					WithArgs("sheet").
					WillReturnResult(sqlmock.NewResult(1, 1))

				if err := repository.CreateSheet(context.Background(), &sql.Sheet{Name: "sheet"}); err != nil {
					t.Fatalf("CreateSheet() error = %v", err)
				}
			},
		},
		{
			name: "get missing sheet returns domain error",
			run: func(t *testing.T, repository *Sheet, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM sheets WHERE id = $1").
					WithArgs(42).
					WillReturnError(dbsql.ErrNoRows)

				_, err := repository.GetSheetById(context.Background(), 42)
				if !errors.Is(err, domainerrors.ErrSheetDoesNotExist) {
					t.Fatalf("GetSheetById() error = %v, want %v", err, domainerrors.ErrSheetDoesNotExist)
				}
			},
		},
		{
			name: "delete referenced sheet returns domain error",
			run: func(t *testing.T, repository *Sheet, mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM sheets WHERE id = $1").
					WithArgs(1).
					WillReturnError(&pgconn.PgError{Code: "23503", TableName: "shots"})

				err := repository.DeleteSheetById(context.Background(), 1)
				if !errors.Is(err, domainerrors.ErrShotForeignKeyConstraint) {
					t.Fatalf("DeleteSheetById() error = %v, want %v", err, domainerrors.ErrShotForeignKeyConstraint)
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
