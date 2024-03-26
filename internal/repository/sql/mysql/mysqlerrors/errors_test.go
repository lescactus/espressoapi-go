package mysqlerrors

import (
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestTtt(t *testing.T) {
	t.Log("Testttt")

	err := mysql.MySQLError{
		Number:  1451,
		Message: "Cannot delete or update a parent row: a foreign key constraint fails (`espresso-api`.`beans`, CONSTRAINT `beans_ibfk_1` FOREIGN KEY (`roaster_id`) REFERENCES `roasters` (`id`))",
	}

	table, _ := ExtractTableNameFromError1451(err)
	fmt.Println(table)
}

func TestExtractTableNameFromError1451(t *testing.T) {
	type args struct {
		err mysql.MySQLError
	}
	tests := []struct {
		name    string
		args    args
		want    Entity
		wantErr bool
	}{
		{
			name:    "Error is not 1451",
			args:    args{err: mysql.MySQLError{Number: 1234}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Error message does not match",
			args:    args{err: mysql.MySQLError{Number: 1451, Message: "Some other error"}},
			want:    "",
			wantErr: true,
		},
		{
			name:    "Error message match",
			args:    args{err: mysql.MySQLError{Number: 1451, Message: "Cannot delete or update a parent row: a foreign key constraint fails (`espresso-api`.`beans`, CONSTRAINT `beans_ibfk_1` FOREIGN KEY (`roaster_id`) REFERENCES `roasters` (`id`))"}},
			want:    "beans",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractTableNameFromError1451(tt.args.err)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractTableNameFromError1451() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractTableNameFromError1451() = %v, want %v", got, tt.want)
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
		want    Entity
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
			got, err := ExtractTableNameFromError1452(tt.args.err)
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

func TestParseMySQLError(t *testing.T) {
	type args struct {
		err      error
		entity   *Entity
		fallback error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "Error is nil",
			args:    args{err: nil, entity: nil, fallback: nil},
			wantErr: false,
		},
		{
			name:    "Error is not MySQLError",
			args:    args{err: fmt.Errorf("some error"), entity: nil, fallback: nil},
			wantErr: false,
		},
		{
			name:    "Error is MySQLError 1062 - nil entity",
			args:    args{err: &mysql.MySQLError{Number: 1062}, entity: nil, fallback: nil},
			wantErr: false,
		},
		{
			name:    "Error is MySQLError 1062 - non nil entity",
			args:    args{err: &mysql.MySQLError{Number: 1062}, entity: &EntitySheet, fallback: nil},
			wantErr: true,
		},
		{
			name:    "Error is MySQLError 1451 - nil entity - nil fallback",
			args:    args{err: &mysql.MySQLError{Number: 1451, Message: "Cannot delete or update a parent row: a foreign key constraint fails (`espresso-api`.`beans`, CONSTRAINT `beans_ibfk_1` FOREIGN KEY (`roaster_id`) REFERENCES `roasters` (`id`))"}, entity: nil, fallback: nil},
			wantErr: true,
		},
		{
			name:    "Error is MySQLError 1451 - nil entity - non nil fallback",
			args:    args{err: &mysql.MySQLError{Number: 1451}, entity: nil, fallback: fmt.Errorf("fallback error")},
			wantErr: true,
		},
		{
			name:    "Error is MySQLError 1451 - non nil entity - nil fallback",
			args:    args{err: &mysql.MySQLError{Number: 1451}, entity: &EntityBeans, fallback: nil},
			wantErr: true,
		},
		{
			name:    "Error is MySQLError 1452 - nil entity - nil fallback",
			args:    args{err: &mysql.MySQLError{Number: 1452, Message: "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `sheets` (`id`))"}, entity: nil, fallback: nil},
			wantErr: true,
		},
		{
			name:    "Error is MySQLError 1452 - nil entity - non nil fallback",
			args:    args{err: &mysql.MySQLError{Number: 1452}, entity: nil, fallback: fmt.Errorf("fallback error")},
			wantErr: true,
		},
		{
			name:    "Error is MySQLError 1452 - non nil entity - nil fallback",
			args:    args{err: &mysql.MySQLError{Number: 1452}, entity: &EntityBeans, fallback: nil},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ParseMySQLError(tt.args.err, tt.args.entity, tt.args.fallback); (err != nil) != tt.wantErr {
				t.Errorf("ParseMySQLError() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
