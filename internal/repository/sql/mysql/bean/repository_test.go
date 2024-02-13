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
		want        int
		wantErr     bool
	}{
		{
			name: "Beans - no error",
			args: args{ctx: context.TODO(), beans: &sql.Beans{Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)").
					WithArgs("beans01", 1, now, sql.RoastLevelMediumToDark).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Beans - LastInsertId error",
			args: args{ctx: context.TODO(), beans: &sql.Beans{Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)").
					WithArgs("beans01", 1, now, sql.RoastLevelMediumToDark).
					WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("mock error")))
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Beans - foreign key constraint error - roaster does not exist",
			args: args{ctx: context.TODO(), beans: &sql.Beans{Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)").
					WithArgs("beans01", 1, now, sql.RoastLevelMediumToDark).
					WillReturnError(&mysql.MySQLError{
						Number: 1452, // Error 1452 is "Cannot add or update a child row: a foreign key constraint fails"
					})
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Beans - error",
			args: args{ctx: context.TODO(), beans: &sql.Beans{Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO beans (name, roaster_id, roast_date, roast_level) VALUES (?, ?, ?, ?)").
					WithArgs("beans01", 1, now, sql.RoastLevelMediumToDark).
					WillReturnError(fmt.Errorf("mock error"))
			},
			want:    0,
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

			id, err := mdb.CreateBeans(tt.args.ctx, tt.args.beans)
			if (err != nil) != tt.wantErr {
				t.Errorf("Bean.CreateBeans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(id, tt.want) {
				t.Errorf("Bean.CreateBeans() = %v, want %v", id, tt.want)
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

	expectQuery := `
	SELECT
		beans.id,
		beans.name,
		beans.roast_date,
		beans.roast_level,
		beans.created_at,
		beans.updated_at,
		roaster.id AS "roaster.id",
		roaster.name AS "roaster.name",
		roaster.created_at AS "roaster.created_at",
		roaster.updated_at AS "roaster.updated_at"
	FROM beans
		INNER JOIN roasters roaster
			ON beans.roaster_id = roaster.id
	WHERE
		beans.id = ?`

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
				mock.ExpectQuery(expectQuery).WithArgs(1).WillReturnRows(
					sqlmock.NewRows(
						[]string{"id", "name", "roaster.id", "roaster.name", "roast_date", "roast_level"}).
						AddRow(1, "beans01", 1, "roaster01", now, sql.RoastLevelMediumToDark),
				)
			},
			want:    &sql.Beans{Id: 1, Roaster: &sql.Roaster{Id: 1, Name: "roaster01"}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark},
			wantErr: false,
		},
		{
			name: "Beans does not exists",
			args: args{ctx: context.TODO(), id: 2},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WithArgs(2).WillReturnError(dbsql.ErrNoRows)
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Error",
			args: args{ctx: context.TODO(), id: 3},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WithArgs(3).WillReturnError(fmt.Errorf("mock error"))
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

	expectQuery := `
	SELECT
		beans.id,
		beans.name,
		beans.roast_date,
		beans.roast_level,
		beans.created_at,
		beans.updated_at,
		roaster.id AS "roaster.id",
		roaster.name AS "roaster.name",
		roaster.created_at AS "roaster.created_at",
		roaster.updated_at AS "roaster.updated_at"
	FROM beans
		INNER JOIN roasters roaster
			ON beans.roaster_id = roaster.id`

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
				mock.ExpectQuery(expectQuery).WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "roast_date", "roast_level", "roaster.id", "roaster.name"}),
				)
			},
			want:    []sql.Beans{},
			wantErr: false,
		},
		{
			name: "Non empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WillReturnRows(
					sqlmock.NewRows([]string{"id", "name", "roast_date", "roast_level", "roaster.id", "roaster.name"}).
						AddRow(1, "beans01", now, sql.RoastLevelMediumToDark, 1, "roaster01").
						AddRow(2, "beans02", now, sql.RoastLevelMediumToDark, 2, "roaster02").
						AddRow(3, "beans03", now, sql.RoastLevelMediumToDark, 3, "roaster03"),
				)
			},
			want: []sql.Beans{
				{Id: 1, Roaster: &sql.Roaster{Id: 1, Name: "roaster01"}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark},
				{Id: 2, Roaster: &sql.Roaster{Id: 2, Name: "roaster02"}, Name: "beans02", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark},
				{Id: 3, Roaster: &sql.Roaster{Id: 3, Name: "roaster03"}, Name: "beans03", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark},
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WillReturnError(fmt.Errorf("mock error"))
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
			args: args{ctx: context.TODO(), id: 1, beans: &sql.Beans{Id: 1, Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET name = ?, roaster_id = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("beans01", 1, AnyTime{}, sql.RoastLevelMediumToDark, 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Beans{Id: 1, Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark},
			wantErr: false,
		},
		{
			name: "Beans.Id matching id - Error",
			args: args{ctx: context.TODO(), id: 1, beans: &sql.Beans{Id: 1, Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET name = ?, roaster_id = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("beans01", 1, AnyTime{}, sql.RoastLevelMediumToDark, 1).
					WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Beans.Id not matching id - Error",
			args: args{ctx: context.TODO(), id: 1, beans: &sql.Beans{Id: 2, Roaster: &sql.Roaster{Id: 1}, Name: "beans01", RoastDate: &now, RoastLevel: sql.RoastLevelMediumToDark}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("UPDATE beans SET name = ?, roaster_id = ?, roast_date = ?, roast_level = ? WHERE id = ?").
					WithArgs("beans01", 1, AnyTime{}, sql.RoastLevelMediumToDark, 1).
					WillReturnError(fmt.Errorf("mock error"))
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
						reflect.DeepEqual(a.Roaster, b.Roaster) &&
						a.Name == b.Name &&
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

func TestBeanPing(t *testing.T) {
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
			db, mock, err := sqlmock.New(
				sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual),
				sqlmock.MonitorPingsOption(true),
			)
			if err != nil {
				t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
			}
			defer db.Close()

			mdb := &Bean{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Bean.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
