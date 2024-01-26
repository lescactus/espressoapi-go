package sheet

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
		want *Sheet
	}{
		{"Nil args", args{nil}, &Sheet{nil}},
		{"Non nil args", args{sqlxdb}, &Sheet{sqlxdb}},
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

func TestDBCreateSheet(t *testing.T) {
	type args struct {
		ctx   context.Context
		sheet *sql.Sheet
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "Unique sheet - no error",
			args: args{ctx: context.TODO(), sheet: &sql.Sheet{Name: "sheet01"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sheets (name) VALUES (?)").WithArgs("sheet01").WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Duplicate sheet - no error",
			args: args{ctx: context.TODO(), sheet: &sql.Sheet{Name: "sheetalreadyexists"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sheets (name) VALUES (?)").WithArgs("sheetalreadyexists").WillReturnError(&mysql.MySQLError{
					Number: 1062, // Error 1062 is "Duplicate entry"
				})
			},
			wantErr: true,
		},
		{
			name: "Unique sheet - error",
			args: args{ctx: context.TODO(), sheet: &sql.Sheet{Name: "sheet02"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO sheets (name) VALUES (?)").WithArgs("sheet02").WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &Sheet{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			err = mdb.CreateSheet(tt.args.ctx, tt.args.sheet)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.CreateSheet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSheetGetSheetById(t *testing.T) {
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Sheet
		wantErr     bool
	}{
		{
			name: "Sheet exists",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM sheets WHERE id = \\?$").WithArgs(1).WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "sheet01"),
				)
			},
			want:    &sql.Sheet{Id: 1, Name: "sheet01"},
			wantErr: false,
		},
		{
			name: "Sheet does not exists",
			args: args{ctx: context.TODO(), id: 2},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM sheets WHERE id = \\?$").WithArgs(2).WillReturnError(dbsql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error",
			args: args{ctx: context.TODO(), id: 3},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM sheets WHERE id = \\?$").WithArgs(3).WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &Sheet{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetSheetById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.GetSheetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.GetSheetById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSheetGetSheetByName(t *testing.T) {
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Sheet
		wantErr     bool
	}{
		{
			name: "Sheet exists",
			args: args{ctx: context.TODO(), name: "sheet01"},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM sheets WHERE name = \\?$").WithArgs("sheet01").WillReturnRows(
					sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "sheet01"),
				)
			},
			want:    &sql.Sheet{Id: 1, Name: "sheet01"},
			wantErr: false,
		},
		{
			name: "Sheet does not exists",
			args: args{ctx: context.TODO(), name: "sheet02"},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM sheets WHERE name = \\?$").WithArgs("sheet02").WillReturnError(dbsql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error",
			args: args{ctx: context.TODO(), name: "sheet03"},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM sheets WHERE name = \\?$").WithArgs("sheet03").WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &Sheet{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetSheetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.GetSheetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.GetSheetByName() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSheetGetAllSheets(t *testing.T) {
	now := time.Now()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        []sql.Sheet
		wantErr     bool
	}{
		{
			name: "Empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM sheets").WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}),
				)
			},
			want:    []sql.Sheet{},
			wantErr: false,
		},
		{
			name: "Non empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM sheets").WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "created_at", "updated_at"}).
						AddRow(1, "sheet01", now, nil).
						AddRow(2, "sheet02", now, now).
						AddRow(3, "sheet03", now, now),
				)
			},
			want: []sql.Sheet{
				{Id: 1, Name: "sheet01", CreatedAt: &now, UpdatedAt: nil},
				{Id: 2, Name: "sheet02", CreatedAt: &now, UpdatedAt: &now},
				{Id: 3, Name: "sheet03", CreatedAt: &now, UpdatedAt: &now},
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM sheets").WillReturnError(fmt.Errorf("mock error"))
			},
			want:    []sql.Sheet{},
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

			mdb := &Sheet{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetAllSheets(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.GetAllSheets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.GetAllSheets() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSheetUpdateSheetById(t *testing.T) {
	type args struct {
		ctx   context.Context
		id    int
		sheet *sql.Sheet
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Sheet
		wantErr     bool
	}{
		{
			name: "Sheet.Id matching id - No error",
			args: args{ctx: context.TODO(), id: 1, sheet: &sql.Sheet{Id: 1, Name: "sheetnewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sheets SET name = ?, updated_at = ? WHERE id = ?").WithArgs("sheetnewname", AnyTime{}, 1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Sheet{Id: 1, Name: "sheetnewname"},
			wantErr: false,
		},
		{
			name: "Sheet.Id matching id - Error",
			args: args{ctx: context.TODO(), id: 1, sheet: &sql.Sheet{Id: 1, Name: "sheetnewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sheets SET name = ?, updated_at = ? WHERE id = ?").WithArgs("sheetnewname", AnyTime{}, 1).WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Sheet.Id not matching id - No error",
			args: args{ctx: context.TODO(), id: 1, sheet: &sql.Sheet{Id: 2, Name: "sheetnewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sheets SET name = ?, updated_at = ? WHERE id = ?").WithArgs("sheetnewname", AnyTime{}, 1).WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Sheet{Id: 1, Name: "sheetnewname"},
			wantErr: false,
		},
		{
			name: "Sheet.Id not matching id - Error",
			args: args{ctx: context.TODO(), id: 1, sheet: &sql.Sheet{Id: 2, Name: "sheetnewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sheets SET name = ?, updated_at = ? WHERE id = ?").WithArgs("sheetnewname", AnyTime{}, 1).WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Sheet does not exist",
			args: args{ctx: context.TODO(), id: 2, sheet: &sql.Sheet{Id: 2, Name: "sheetnewname"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE sheets SET name = ?, updated_at = ? WHERE id = ?").WithArgs("sheetnewname", AnyTime{}, 2).WillReturnResult(sqlmock.NewResult(0, 0))
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

			mdb := &Sheet{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.UpdateSheetById(tt.args.ctx, tt.args.id, tt.args.sheet)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.UpdateSheetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isEqual := func(a, b *sql.Sheet) bool {
				return a == b || (a != nil && b != nil && a.Id == b.Id && a.Name == b.Name)
			}

			if !isEqual(got, tt.want) {
				t.Errorf("Sheet.UpdateSheetById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestSheetDeleteSheetById(t *testing.T) {
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
			name: "Sheet found - no error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM sheets WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "Sheet found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from sheets where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
			},
			wantErr: true,
		},
		{
			name: "Sheet not found - No error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM sheets WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "Sheet not found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from sheets where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &Sheet{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.DeleteSheetById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Sheet.DeleteSheetById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSheetPing(t *testing.T) {
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

			mdb := &Sheet{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			if err := mdb.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Sheet.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
