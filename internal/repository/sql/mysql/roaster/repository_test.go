package roaster

import (
	"context"
	dbsql "database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
)

// ref: https://github.com/DATA-DOG/go-sqlmock#matching-arguments-like-timetime
type AnyTime struct{}

// Match satisfies sqlmock.Argument interface
func (a AnyTime) Match(v driver.Value) bool {
	_, ok := v.(time.Time)
	return ok
}

func TestNew(t *testing.T) {
	sqlxdb := &sqlx.DB{}
	type args struct {
		db *sqlx.DB
	}
	tests := []struct {
		name string
		args args
		want *Roaster
	}{
		{"Nil args", args{nil}, &Roaster{nil}},
		{"Non nil args", args{sqlxdb}, &Roaster{sqlxdb}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.db)
			if got.db != tt.want.db {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDBCreateRoaster(t *testing.T) {
	type args struct {
		ctx     context.Context
		roaster *sql.Roaster
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "Unique roaster - no error",
			args: args{ctx: context.TODO(), roaster: &sql.Roaster{Name: "roaster01"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO roasters (name) VALUES (?)").WithArgs("roaster01").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Duplicate roaster - no error",
			args: args{ctx: context.TODO(), roaster: &sql.Roaster{Name: "roasteralreadyexists"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO roasters (name) VALUES (?)").WithArgs("roasteralreadyexists").WillReturnError(&mysql.MySQLError{
					Number: 1062, // Error 1062 is "Duplicate entry"
				})
			},
			wantErr: true,
		},
		{
			name: "Unique roaster - error",
			args: args{ctx: context.TODO(), roaster: &sql.Roaster{Name: "roaster02"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO roasters (name) VALUES (?)").WithArgs("roaster02").WillReturnError(fmt.Errorf("mock error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DB and mock
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Roaster{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			err = mdb.CreateRoaster(tt.args.ctx, tt.args.roaster)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.CreateRoaster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRoasterGetRoasterById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Roaster
		wantErr     bool
	}{
		{
			name: "Roaster exists",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM roasters WHERE id = \\?$").WithArgs(1).WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "roaster01"),
				)
			},
			want:    &sql.Roaster{Id: 1, Name: "roaster01"},
			wantErr: false,
		},
		{
			name: "Roaster does not exists",
			args: args{ctx: context.TODO(), id: 2},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM roasters WHERE id = \\?$").WithArgs(2).WillReturnError(dbsql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error",
			args: args{ctx: context.TODO(), id: 3},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM roasters WHERE id = \\?$").WithArgs(3).WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DB and mock
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Roaster{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetRoasterById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.GetRoasterById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.GetRoasterById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRoasterGetRoasterByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Roaster
		wantErr     bool
	}{
		{
			name: "Roaster exists",
			args: args{ctx: context.TODO(), name: "roaster01"},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM roasters WHERE name = \\?$").WithArgs("roaster01").WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "roaster01"),
				)
			},
			want:    &sql.Roaster{Id: 1, Name: "roaster01"},
			wantErr: false,
		},
		{
			name: "Roaster does not exists",
			args: args{ctx: context.TODO(), name: "roaster02"},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM roasters WHERE name = \\?$").WithArgs("roaster02").WillReturnError(dbsql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error",
			args: args{ctx: context.TODO(), name: "roaster03"},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM roasters WHERE name = \\?$").WithArgs("roaster03").WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DB and mock
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Roaster{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetRoasterByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.GetRoasterByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.GetRoasterByName() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRoasterGetAllRoasters(t *testing.T) {
	now := time.Now()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        []sql.Roaster
		wantErr     bool
	}{
		{
			name: "Empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM roasters").WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}),
				)
			},
			want:    []sql.Roaster{},
			wantErr: false,
		},
		{
			name: "Non empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM roasters").WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
						AddRow(1, "roaster01", now, nil).
						AddRow(2, "roaster02", now, now).
						AddRow(3, "roaster03", now, now),
				)
			},
			want: []sql.Roaster{
				{Id: 1, Name: "roaster01", CreatedAt: &now, UpdatedAt: nil},
				{Id: 2, Name: "roaster02", CreatedAt: &now, UpdatedAt: &now},
				{Id: 3, Name: "roaster03", CreatedAt: &now, UpdatedAt: &now},
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM roasters").WillReturnError(fmt.Errorf("mock error"))
			},
			want:    []sql.Roaster{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DB and mock
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Roaster{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetAllRoasters(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.GetAllRoasters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.GetAllRoasters() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRoasterUpdateRoasterById(t *testing.T) {
	type args struct {
		ctx     context.Context
		id      int
		roaster *sql.Roaster
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Roaster
		wantErr     bool
	}{
		{
			name: "Roaster.Id matching id - No error",
			args: args{ctx: context.TODO(), id: 1, roaster: &sql.Roaster{Id: 1, Name: "roasternewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE roasters SET name = ?, updated_at = ? WHERE id = ?").WithArgs("roasternewname", AnyTime{}, 1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Roaster{Id: 1, Name: "roasternewname"},
			wantErr: false,
		},
		{
			name: "Roaster.Id matching id - Error",
			args: args{ctx: context.TODO(), id: 1, roaster: &sql.Roaster{Id: 1, Name: "roasternewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE roasters SET name = ?, updated_at = ? WHERE id = ?").WithArgs("roasternewname", AnyTime{}, 1).WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Roaster.Id not matching id - No error",
			args: args{ctx: context.TODO(), id: 1, roaster: &sql.Roaster{Id: 2, Name: "roasternewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE roasters SET name = ?, updated_at = ? WHERE id = ?").WithArgs("roasternewname", AnyTime{}, 1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Roaster{Id: 1, Name: "roasternewname"},
			wantErr: false,
		},
		{
			name: "Roaster.Id not matching id - Error",
			args: args{ctx: context.TODO(), id: 1, roaster: &sql.Roaster{Id: 2, Name: "roasternewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE roasters SET name = ?, updated_at = ? WHERE id = ?").WithArgs("roasternewname", AnyTime{}, 1).WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Roaster does not exist",
			args: args{ctx: context.TODO(), id: 2, roaster: &sql.Roaster{Id: 2, Name: "roasternewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE roasters SET name = ?, updated_at = ? WHERE id = ?").WithArgs("roasternewname", AnyTime{}, 2).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DB and mock
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Roaster{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.UpdateRoasterById(tt.args.ctx, tt.args.id, tt.args.roaster)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.Roaster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isEqual := func(a, b *sql.Roaster) bool {
				return a == b || (a != nil && b != nil && a.Id == b.Id && a.Name == b.Name)
			}

			if !isEqual(got, tt.want) {
				t.Errorf("Roaster.Roaster() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestRoasterDeleteRoasterById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "Roaster found - no error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM roasters WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "Roaster found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from roasters where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
			},
			wantErr: true,
		},
		{
			name: "Roaster not found - No error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM roasters WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "Roaster not found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from roasters where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DB and mock
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Roaster{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.DeleteRoasterById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Roaster.DeleteRoasterById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoasterPing(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "No error",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing()
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectPing().WillReturnError(fmt.Errorf("mock error"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// DB and mock
			db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual), sqlmock.MonitorPingsOption(true))
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Roaster{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			if err := mdb.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Roaster.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
