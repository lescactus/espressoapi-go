package roaster

import (
	"context"
	dbsql "database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
)

func TestRoasterRepositoryPostgresBehavior(t *testing.T) {
	tests := []struct {
		name string
		run  func(t *testing.T, repository *Roaster, mock sqlmock.Sqlmock)
	}{
		{
			name: "create uses postgres placeholder",
			run: func(t *testing.T, repository *Roaster, mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO roasters (name) VALUES ($1)").
					WithArgs("roaster").
					WillReturnResult(sqlmock.NewResult(1, 1))

				if err := repository.CreateRoaster(context.Background(), &sql.Roaster{Name: "roaster"}); err != nil {
					t.Fatalf("CreateRoaster() error = %v", err)
				}
			},
		},
		{
			name: "get missing roaster returns domain error",
			run: func(t *testing.T, repository *Roaster, mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, name, created_at, updated_at FROM roasters WHERE id = $1").
					WithArgs(42).
					WillReturnError(dbsql.ErrNoRows)

				_, err := repository.GetRoasterById(context.Background(), 42)
				if !errors.Is(err, domainerrors.ErrRoasterDoesNotExist) {
					t.Fatalf("GetRoasterById() error = %v, want %v", err, domainerrors.ErrRoasterDoesNotExist)
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
