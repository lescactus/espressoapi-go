package mysqlerrors

import (
	"errors"
	"testing"

	"github.com/go-sql-driver/mysql"
	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
)

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
	fallback := errors.New("fallback error")
	unknownEntity := Entity("unknown")
	const (
		beansForeignKeyError      = "Cannot delete or update a parent row: a foreign key constraint fails (`espresso-api`.`beans`, CONSTRAINT `beans_ibfk_1` FOREIGN KEY (`roaster_id`) REFERENCES `roasters` (`id`))"
		sheetDoesNotExistError    = "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`shots`, CONSTRAINT `shots_ibfk_1` FOREIGN KEY (`sheet_id`) REFERENCES `sheets` (`id`))"
		roasterDoesNotExistError  = "Cannot add or update a child row: a foreign key constraint fails (`espresso-api`.`beans`, CONSTRAINT `beans_ibfk_1` FOREIGN KEY (`roaster_id`) REFERENCES `roasters` (`id`))"
		shotComparisonCheckError  = "Check constraint 'chk_shots_comparison_with_previous_result' is violated."
		beansRoastLevelCheckError = "Check constraint 'chk_beans_roast_level' is violated."
	)

	tests := []struct {
		name     string
		err      error
		entity   *Entity
		fallback error
		want     error
	}{
		{
			name: "nil error",
		},
		{
			name: "non-MySQL error returns fallback", err: errors.New("some error"), fallback: fallback, want: fallback,
		},
		{
			name: "duplicate with nil entity returns fallback", err: &mysql.MySQLError{Number: 1062}, fallback: fallback, want: fallback,
		},
		{
			name: "duplicate sheet", err: &mysql.MySQLError{Number: 1062}, entity: &EntitySheet, fallback: fallback, want: domainerrors.ErrSheetAlreadyExists,
		},
		{
			name: "duplicate roaster", err: &mysql.MySQLError{Number: 1062}, entity: &EntityRoaster, fallback: fallback, want: domainerrors.ErrRoasterAlreadyExists,
		},
		{
			name: "duplicate beans", err: &mysql.MySQLError{Number: 1062}, entity: &EntityBeans, fallback: fallback, want: domainerrors.ErrBeansAlreadyExists,
		},
		{
			name: "duplicate shot", err: &mysql.MySQLError{Number: 1062}, entity: &EntityShot, fallback: fallback, want: domainerrors.ErrShotAlreadyExists,
		},
		{
			name: "duplicate unknown entity returns fallback", err: &mysql.MySQLError{Number: 1062}, entity: &unknownEntity, fallback: fallback, want: fallback,
		},
		{
			name: "foreign key constraint inferred from beans table", err: &mysql.MySQLError{Number: 1451, Message: beansForeignKeyError}, fallback: fallback, want: domainerrors.ErrBeansForeignKeyConstraint,
		},
		{
			name: "foreign key constraint with explicit shot entity", err: &mysql.MySQLError{Number: 1451}, entity: &EntityShot, fallback: fallback, want: domainerrors.ErrShotForeignKeyConstraint,
		},
		{
			name: "foreign key constraint with malformed message returns fallback", err: &mysql.MySQLError{Number: 1451}, fallback: fallback, want: fallback,
		},
		{
			name: "foreign key constraint with unknown entity returns fallback", err: &mysql.MySQLError{Number: 1451}, entity: &unknownEntity, fallback: fallback, want: fallback,
		},
		{
			name: "missing sheet inferred from referenced table", err: &mysql.MySQLError{Number: 1452, Message: sheetDoesNotExistError}, fallback: fallback, want: domainerrors.ErrSheetDoesNotExist,
		},
		{
			name: "missing roaster inferred from referenced table", err: &mysql.MySQLError{Number: 1452, Message: roasterDoesNotExistError}, fallback: fallback, want: domainerrors.ErrRoasterDoesNotExist,
		},
		{
			name: "missing roaster with explicit entity", err: &mysql.MySQLError{Number: 1452}, entity: &EntityRoaster, fallback: fallback, want: domainerrors.ErrRoasterDoesNotExist,
		},
		{
			name: "missing roaster inferred from referenced table with explicit beans entity", err: &mysql.MySQLError{Number: 1452, Message: roasterDoesNotExistError}, entity: &EntityBeans, fallback: fallback, want: domainerrors.ErrRoasterDoesNotExist,
		},
		{
			name: "missing beans with explicit entity", err: &mysql.MySQLError{Number: 1452}, entity: &EntityBeans, fallback: fallback, want: domainerrors.ErrBeansDoesNotExist,
		},
		{
			name: "missing entity with malformed message returns fallback", err: &mysql.MySQLError{Number: 1452}, fallback: fallback, want: fallback,
		},
		{
			name: "missing unknown entity returns fallback", err: &mysql.MySQLError{Number: 1452}, entity: &unknownEntity, fallback: fallback, want: fallback,
		},
		{
			name: "shot comparison check constraint", err: &mysql.MySQLError{Number: 3819, Message: shotComparisonCheckError}, fallback: fallback, want: domainerrors.ErrShotComparisonWithPreviousResultOutOfRange,
		},
		{
			name: "beans roast level check constraint", err: &mysql.MySQLError{Number: 3819, Message: beansRoastLevelCheckError}, fallback: fallback, want: domainerrors.ErrBeansRoastLevelOutOfRange,
		},
		{
			name: "unknown check constraint returns fallback", err: &mysql.MySQLError{Number: 3819, Message: "Check constraint 'other_constraint' is violated."}, fallback: fallback, want: fallback,
		},
		{
			name: "unmapped MySQL error returns fallback", err: &mysql.MySQLError{Number: 9999}, fallback: fallback, want: fallback,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseMySQLError(tt.err, tt.entity, tt.fallback)
			if !errors.Is(got, tt.want) {
				t.Errorf("ParseMySQLError() = %v, want %v", got, tt.want)
			}
		})
	}
}
