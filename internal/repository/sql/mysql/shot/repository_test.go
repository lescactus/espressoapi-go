package shot

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
		want *Shot
	}{
		{
			name: "nil db",
			args: args{db: nil},
			want: &Shot{db: nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.db); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTableNameFromError1452(t *testing.T) {
	type args struct {
		err mysql.MySQLError
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "Error is not 1452",
			args:    args{err: mysql.MySQLError{Number: 1234}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Error message does not match",
			args:    args{err: mysql.MySQLError{Number: 1452, Message: "Some other error"}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Error message match",
			args:    args{err: mysql.MySQLError{Number: 1452, Message: "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `sheets` (`id`))"}},
			want:    "sheets",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := extractTableNameFromError1452(tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractTableNameFromError1452() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("extractTableNameFromError1452() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShotCreateShot(t *testing.T) {
	type args struct {
		ctx  context.Context
		shot *sql.Shot
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        int
		wantErr     bool
	}{
		{
			name: "Shots - no error",
			args: args{ctx: context.TODO(), shot: &sql.Shot{Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1},
				ComparaisonWithPreviousResult: sql.Unknown, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO 
				shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Unknown, "This is a test").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    1,
			wantErr: false,
		},
		{
			name: "Shots - LastInsertId error",
			args: args{ctx: context.TODO(), shot: &sql.Shot{Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1},
				ComparaisonWithPreviousResult: sql.Unknown, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO 
				shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Unknown, "This is a test").
					WillReturnResult(sqlmock.NewErrorResult(fmt.Errorf("mock error")))
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Shots - foreign key constraint error - sheets does not exist",
			args: args{ctx: context.TODO(), shot: &sql.Shot{Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1},
				ComparaisonWithPreviousResult: sql.Unknown, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO 
				shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Unknown, "This is a test").
					WillReturnError(&mysql.MySQLError{
						Message: "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `sheets` (`id`))",
						Number:  1452, // Error 1452 is "Cannot add or update a child row: a foreign key constraint fails"
					})
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Shots - foreign key constraint error - beans does not exist",
			args: args{ctx: context.TODO(), shot: &sql.Shot{Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1},
				ComparaisonWithPreviousResult: sql.Unknown, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO 
				shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Unknown, "This is a test").
					WillReturnError(&mysql.MySQLError{
						Message: "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `beans` (`id`))",
						Number:  1452, // Error 1452 is "Cannot add or update a child row: a foreign key constraint fails"
					})
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Shots - error 01",
			args: args{ctx: context.TODO(), shot: &sql.Shot{Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1},
				ComparaisonWithPreviousResult: sql.Unknown, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO 
				shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Unknown, "This is a test").
					WillReturnError(fmt.Errorf("mock error"))
			},
			want:    0,
			wantErr: true,
		},
		{
			name: "Shots - error 02",
			args: args{ctx: context.TODO(), shot: &sql.Shot{Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1},
				ComparaisonWithPreviousResult: sql.Unknown, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`INSERT INTO 
				shots (sheet_id, beans_id, grind_setting, quantity_in, quantity_out, shot_time, water_temperature, rating, is_too_bitter, is_too_sour, comparaison_with_previous_result, additional_notes)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Unknown, "This is a test").
					WillReturnError(&mysql.MySQLError{
						Message: "unparsable error message",
						Number:  1452, // Error 1452 is "Cannot add or update a child row: a foreign key constraint fails"
					})
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

			mdb := &Shot{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			id, err := mdb.CreateShot(tt.args.ctx, tt.args.shot)
			if (err != nil) != tt.wantErr {
				t.Errorf("Shot.CreateShot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(id, tt.want) {
				t.Errorf("Shot.CreateShot() = %v, want %v", id, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestShotGetShotById(t *testing.T) {
	now := time.Now()

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
	shots.comparaison_with_previous_result,
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
WHERE shots.id = ?`

	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Shot
		wantErr     bool
	}{
		{
			name: "Shot exists",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WithArgs(1).WillReturnRows(
					sqlmock.NewRows(
						[]string{"id", "grind_setting", "quantity_in", "quantity_out", "shot_time", "water_temperature", "rating", "is_too_bitter", "is_too_sour", "comparaison_with_previous_result", "additional_notes", "sheet.id", "sheet.name", "beans.id", "beans.name", "beans.roast_date", "beans.roast_level"}).
						AddRow(1, 11, 18.0, 36.0, 25*time.Second, 90.0, 4.5, false, true, sql.Better, "This is a test", 1, "sheet01", 1, "beans01", now, sql.RoastLevelLight))
			},
			want: &sql.Shot{
				Id:                            1,
				GrindSetting:                  11,
				QuantityIn:                    18.0,
				QuantityOut:                   36.0,
				ShotTime:                      25 * time.Second,
				WaterTemperature:              90.0,
				Rating:                        4.5,
				IsTooBitter:                   false,
				IsTooSour:                     true,
				ComparaisonWithPreviousResult: sql.Better,
				AdditionalNotes:               "This is a test",
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
			wantErr: false,
		},
		{
			name: "Shot does not exists",
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

			mdb := &Shot{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetShotById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Shot.GetShotById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Shot.GetShotById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestShotGetAllShots(t *testing.T) {
	now := time.Now()

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
	shots.comparaison_with_previous_result,
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
	roasters roaster ON beans.roaster_id = roaster.id`

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        []sql.Shot
		wantErr     bool
	}{
		{
			name: "Empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WillReturnRows(
					sqlmock.NewRows([]string{"id", "grind_setting", "quantity_in", "quantity_out", "shot_time", "water_temperature", "rating", "is_too_bitter", "is_too_sour", "comparaison_with_previous_result", "additional_notes", "sheet.id", "sheet.name", "beans.id", "beans.name", "beans.roast_date", "beans.roast_level", "beans.roaster.id", "beans.roaster.name", "beans.roaster.created_at", "beans.roaster.updated_at"}),
				)
			},
			want:    []sql.Shot{},
			wantErr: false,
		},
		{
			name: "Non empty result",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WillReturnRows(
					sqlmock.NewRows([]string{"id", "grind_setting", "quantity_in", "quantity_out", "shot_time", "water_temperature", "rating", "is_too_bitter", "is_too_sour", "comparaison_with_previous_result", "additional_notes", "sheet.id", "sheet.name", "beans.id", "beans.name", "beans.roast_date", "beans.roast_level"}).
						AddRow(1, 11, 18.0, 36.0, 25*time.Second, 90.0, 4.5, false, true, sql.Better, "This is a test", 1, "sheet01", 1, "beans01", now, sql.RoastLevelLight),
				)
			},
			want: []sql.Shot{
				{
					Id:                            1,
					GrindSetting:                  11,
					QuantityIn:                    18.0,
					QuantityOut:                   36.0,
					ShotTime:                      25 * time.Second,
					WaterTemperature:              90.0,
					Rating:                        4.5,
					IsTooBitter:                   false,
					IsTooSour:                     true,
					ComparaisonWithPreviousResult: sql.Better,
					AdditionalNotes:               "This is a test",
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
			},
			wantErr: false,
		},
		{
			name: "Error",
			args: args{context.TODO()},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(expectQuery).WillReturnError(fmt.Errorf("mock error"))
			},
			want:    []sql.Shot{},
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

			mdb := &Shot{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.GetAllShots(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Shot.GetAllShots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Shot.GetAllShots() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestShotUpdateShotById(t *testing.T) {
	expectQuery := `UPDATE shots SET
	sheet_id = ?,
	beans_id = ?,
	grind_setting = ?,
	quantity_in = ?,
	quantity_out = ?,
	shot_time = ?,
	water_temperature = ?,
	rating = ?,
	is_too_bitter = ?,
	is_too_sour = ?,
	comparaison_with_previous_result = ?,
	additional_notes = ?
	WHERE id = ?`

	type args struct {
		ctx  context.Context
		id   int
		shot *sql.Shot
	}
	tests := []struct {
		name        string
		args        args
		mockClosure func(mock sqlmock.Sqlmock)
		want        *sql.Shot
		wantErr     bool
	}{
		{
			name: "Shot.Id matching id - No error",
			args: args{ctx: context.TODO(), id: 1, shot: &sql.Shot{Id: 1, Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1}, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(expectQuery).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "This is a test", 1).
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
			want:    &sql.Shot{Id: 1, Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1}, AdditionalNotes: "This is a test"},
			wantErr: false,
		},
		{
			name: "Shot.Id matching id - Error",
			args: args{ctx: context.TODO(), id: 1, shot: &sql.Shot{Id: 1, Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1}, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(expectQuery).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "This is a test", 1).
					WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Shot.Id not matching id - Error",
			args: args{ctx: context.TODO(), id: 1, shot: &sql.Shot{Id: 2, Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1}, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(expectQuery).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "This is a test", 1).
					WillReturnError(fmt.Errorf("mock error"))
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Shots - foreign key constraint error - sheets does not exist",
			args: args{ctx: context.TODO(), id: 1, shot: &sql.Shot{Id: 1, Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1}, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(expectQuery).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "This is a test", 1).
					WillReturnError(&mysql.MySQLError{
						Message: "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `sheets` (`id`))",
						Number:  1452, // Error 1452 is "Cannot add or update a child row: a foreign key constraint fails"
					})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Shots - foreign key constraint error - beans does not exist",
			args: args{ctx: context.TODO(), id: 1, shot: &sql.Shot{Id: 1, Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1}, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(expectQuery).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "This is a test", 1).
					WillReturnError(&mysql.MySQLError{
						Message: "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`beans_id`) REFERENCES `beans` (`id`))",
						Number:  1452, // Error 1452 is "Cannot add or update a child row: a foreign key constraint fails"
					})
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "Shots - generic error",
			args: args{ctx: context.TODO(), id: 1, shot: &sql.Shot{Id: 1, Sheet: &sql.Sheet{Id: 1}, Beans: &sql.Beans{Id: 1}, AdditionalNotes: "This is a test"}},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(expectQuery).
					WithArgs(1, 1, 0, 0.0, 0.0, 0, 0.0, 0.0, false, false, sql.Worst, "This is a test", 1).
					WillReturnError(&mysql.MySQLError{
						Message: "mock generic error",
						Number:  1452, // Error 1452 is "Cannot add or update a child row: a foreign key constraint fails"
					})
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

			mdb := &Shot{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)

			got, err := mdb.UpdateShotById(tt.args.ctx, tt.args.id, tt.args.shot)
			if (err != nil) != tt.wantErr {
				t.Errorf("Shot.UpdateShotById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Shot.UpdateBeansById() = %v, want %v", got, tt.want)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestShotDeleteShotById(t *testing.T) {
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
			name: "Shots found - no error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM shots WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 1))
			},
			wantErr: false,
		},
		{
			name: "Shots found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from shots where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
			},
			wantErr: true,
		},
		{
			name: "Shots not found - No error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE FROM shots WHERE id = ?").WithArgs(1).WillReturnResult(sqlmock.NewResult(0, 0))
			},
			wantErr: true,
		},
		{
			name: "Shots not found - Error",
			args: args{ctx: context.TODO(), id: 1},
			mockClosure: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("DELETE from shots where id = ?").WithArgs(1).WillReturnError(fmt.Errorf("mock error"))
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

			mdb := &Shot{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.DeleteShotById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("Shot.DeleteShotById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShotPing(t *testing.T) {
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

			mdb := &Shot{
				db: sqlx.NewDb(db, "sqlmock"),
			}

			// Set mock expectations
			tt.mockClosure(mock)
			if err := mdb.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Shot.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Make sure all expectations were met
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
