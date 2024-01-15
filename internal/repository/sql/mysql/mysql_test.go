package mysql

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
	type args struct {
		db *sqlx.DB
	}
	tests := []struct {
		name string
		args args
		want *MysqlDB
	}{
		{
			name: "nil db",
			args: args{db: nil},
			want: &MysqlDB{db: nil},
		},
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			err = mdb.CreateSheet(tt.args.ctx, tt.args.sheet)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.CreateSheet() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBGetSheetById(t *testing.T) {
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetSheetById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.GetSheetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MysqlDB.GetSheetByName() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBGetSheetByName(t *testing.T) {
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetSheetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.GetSheetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MysqlDB.GetSheetByName() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBGetAllSheets(t *testing.T) {
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetAllSheets(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.GetAllSheets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MysqlDB.GetAllSheets() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBUpdateSheetById(t *testing.T) {
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.UpdateSheetById(tt.args.ctx, tt.args.id, tt.args.sheet)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.UpdateSheetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isEqual := func(a, b *sql.Sheet) bool {
				return a == b || (a != nil && b != nil && a.Id == b.Id && a.Name == b.Name)
			}

			if !isEqual(got, tt.want) {
				t.Errorf("MysqlDB.UpdateSheetById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBDeleteSheetById(t *testing.T) {
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.DeleteSheetById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.DeleteSheetById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMysqlDBPing(t *testing.T) {
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			if err := mdb.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestDBCreateBeans(t *testing.T) {
	now := time.Now()

	type args struct {
		ctx   context.Context
		beans *sql.Beans
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		wantErr     bool
	}{
		{
			name: "Unique beans - no error",
			args: args{ctx: context.TODO(), beans: &sql.Beans{RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO beans (roaster_name, beans_name, roast_date, roast_level) VALUES (?, ?, ?, ?)").
					WithArgs("roaster01", "beans01", now, sql.RoastLevelMediumToDark).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			wantErr: false,
		},
		{
			name: "Unique beans - error",
			args: args{ctx: context.TODO(), beans: &sql.Beans{RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO beans (roaster_name, beans_name, roast_date, roast_level) VALUES (?, ?, ?, ?)").
					WithArgs("roaster01", "beans01", now, sql.RoastLevelMediumToDark).
					WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			err = mdb.CreateBeans(tt.args.ctx, tt.args.beans)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.CreateBeans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBGetBeansById(t *testing.T) {
	now := time.Now()

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Beans
		wantErr     bool
	}{
		{
			name: "Beans exists",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM beans WHERE id = \\?$").WithArgs(1).WillReturnRows(
					sqlmock.NewRows(
						[]string{"id", "roaster_name", "beans_name", "roast_date", "roast_level"}).
						AddRow(1, "roaster01", "beans01", now, sql.RoastLevelMediumToDark),
				)
			},
			want:    &sql.Beans{Id: 1, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark},
			wantErr: false,
		},
		{
			name: "Beans does not exists",
			args: args{ctx: context.TODO(), id: 2},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM beans WHERE id = \\?$").WithArgs(2).WillReturnError(dbsql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error",
			args: args{ctx: context.TODO(), id: 3},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("^SELECT \\* FROM beans WHERE id = \\?$").WithArgs(3).WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetBeansById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.GetBeansById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MysqlDB.GetBeansById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBGetAllBeans(t *testing.T) {
	now := time.Now()

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        []sql.Beans
		wantErr     bool
	}{
		{
			name: "Empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM beans").WillReturnRows(
					sqlmock.NewRows([]string{"id", "roaster_name", "beans_name", "roast_date", "roast_level"}),
				)
			},
			want:    []sql.Beans{},
			wantErr: false,
		},
		{
			name: "Non empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM beans").WillReturnRows(
					sqlmock.NewRows([]string{"id", "roaster_name", "beans_name", "roast_date", "roast_level"}).
						AddRow(1, "roaster01", "beans01", now, sql.RoastLevelMediumToDark).
						AddRow(2, "roaster02", "beans02", now, sql.RoastLevelMediumToDark).
						AddRow(3, "roaster03", "beans03", now, sql.RoastLevelMediumToDark),
				)
			},
			want: []sql.Beans{
				{Id: 1, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark},
				{Id: 2, RoasterName: "roaster02", BeansName: "beans02", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark},
				{Id: 3, RoasterName: "roaster03", BeansName: "beans03", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark},
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT * FROM beans").WillReturnError(fmt.Errorf("mock error"))
			},
			want:    []sql.Beans{},
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetAllBeans(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.GetAllBeans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MysqlDB.GetAllBeans() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBUpdateBeansById(t *testing.T) {
	now := time.Now()
	type args struct {
		ctx   context.Context
		id    int
		beans *sql.Beans
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Beans
		wantErr     bool
	}{
		{
			name: "Beans.Id matching id - No error",
			args: args{ctx: context.TODO(), id: 1, beans: &sql.Beans{Id: 1, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET roaster_name = ?, beans_name = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("roaster01", "beans01", AnyTime{}, sql.RoastLevelMediumToDark, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Beans{Id: 1, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark},
			wantErr: false,
		},
		{
			name: "Beans.Id matching id - Error",
			args: args{ctx: context.TODO(), id: 1, beans: &sql.Beans{Id: 1, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET roaster_name = ?, beans_name = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("roaster01", "beans01", AnyTime{}, sql.RoastLevelMediumToDark, 1).
					WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Beans.Id not matching id - No error",
			args: args{ctx: context.TODO(), id: 1, beans: &sql.Beans{Id: 2, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET roaster_name = ?, beans_name = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("roaster01", "beans01", AnyTime{}, sql.RoastLevelMediumToDark, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Beans{Id: 1, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark},
			wantErr: false,
		},
		{
			name: "Beans.Id not matching id - Error",
			args: args{ctx: context.TODO(), id: 1, beans: &sql.Beans{Id: 2, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET roaster_name = ?, beans_name = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("roaster01", "beans01", AnyTime{}, sql.RoastLevelMediumToDark, 1).
					WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Beans does not exist",
			args: args{ctx: context.TODO(), id: 2, beans: &sql.Beans{Id: 2, RoasterName: "roaster01", BeansName: "beans01", RoastDate: now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET roaster_name = ?, beans_name = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("roaster01", "beans01", AnyTime{}, sql.RoastLevelMediumToDark, 2).
					WillReturnResult(sqlmock.NewResult(0, 0))
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.UpdateBeansById(tt.args.ctx, tt.args.id, tt.args.beans)
			if (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.UpdateBeansById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			isEqual := func(a, b *sql.Beans) bool {
				return a == b ||
					(a != nil &&
						b != nil &&
						a.Id == b.Id &&
						a.RoasterName == b.RoasterName &&
						a.BeansName == b.BeansName &&
						a.RoastDate == b.RoastDate &&
						a.RoastLevel == b.RoastLevel)
			}

			if !isEqual(got, tt.want) {
				t.Errorf("MysqlDB.UpdateBeansById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMysqlDBDeleteBeansById(t *testing.T) {
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
			name: "Beans found - no error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM beans WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "Beans found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from beans where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
			},
			wantErr: true,
		},
		{
			name: "Beans not found - No error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM beans WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "Beans not found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from beans where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &MysqlDB{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.DeleteBeansById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("MysqlDB.DeleteBeansById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
