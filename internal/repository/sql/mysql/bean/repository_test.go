package bean

import (
	"context"
	dbsql "database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
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
		want *Bean
	}{
		{
			name: "nil db",
			args: args{db: nil},
			want: &Bean{db: nil},
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

			mdb := &Bean{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			err = mdb.CreateBeans(tt.args.ctx, tt.args.beans)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bean.CreateBeans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestBeanGetBeansById(t *testing.T) {
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

			mdb := &Bean{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetBeansById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bean.GetBeansById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bean.GetBeansById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestBeanGetAllBeans(t *testing.T) {
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

			mdb := &Bean{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetAllBeans(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bean.GetAllBeans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Bean.GetAllBeans() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestBeanUpdateBeansById(t *testing.T) {
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

			mdb := &Bean{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.UpdateBeansById(tt.args.ctx, tt.args.id, tt.args.beans)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bean.UpdateBeansById() error = %v, wantErr %v", err, tt.wantErr)
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
				t.Errorf("Bean.UpdateBeansById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestBeanDeleteBeansById(t *testing.T) {
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

			mdb := &Bean{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.DeleteBeansById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Bean.DeleteBeansById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}