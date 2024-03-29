package sheet

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
)

var (
	now = time.Now()
)

type IsErrorCtxKey string
type IsEmptyCtxKey string

type MockSheetRepository struct{}

func (m *MockSheetRepository) CreateSheet(ctx context.Context, sheet *sql.Sheet) error {
	switch sheet.Name {
	case "duplicatesheet":
		return errors.ErrSheetAlreadyExists
	default:
		return nil
	}
}

func (m *MockSheetRepository) GetSheetById(ctx context.Context, id int) (*sql.Sheet, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == "isErrorFromUpdateSheetById" {
		return nil, fmt.Errorf("mock error")
	}

	if id == 1 {
		return &sql.Sheet{Id: id, Name: "sheet01name", CreatedAt: &now, UpdatedAt: &now}, nil
	} else {
		return nil, errors.ErrSheetDoesNotExist
	}
}

func (m *MockSheetRepository) GetSheetByName(ctx context.Context, name string) (*sql.Sheet, error) {
	switch name {
	case "sheetdoesnotexists":
		return nil, errors.ErrSheetDoesNotExist

	case "sheeterror":
		return nil, fmt.Errorf("mock error")

	default:
		return &sql.Sheet{Id: 1, Name: "sheet01"}, nil
	}
}

func (m *MockSheetRepository) GetAllSheets(ctx context.Context) ([]sql.Sheet, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if isEmpty := ctx.Value(IsEmptyCtxKey("isEmpty")); isEmpty == true {
		return []sql.Sheet{}, nil
	}

	return []sql.Sheet{
		{Id: 1, Name: "sheet01"},
		{Id: 2, Name: "sheet02"},
	}, nil
}

func (m *MockSheetRepository) UpdateSheetById(ctx context.Context, id int, sheet *sql.Sheet) (*sql.Sheet, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if id == 1 {
		return &sql.Sheet{Id: id, Name: sheet.Name, UpdatedAt: &now}, nil
	} else {
		return nil, errors.ErrSheetDoesNotExist
	}
}

func (m *MockSheetRepository) DeleteSheetById(ctx context.Context, id int) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func (m *MockSheetRepository) Ping(ctx context.Context) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func TestNew(t *testing.T) {
	type args struct {
		repo repository.SheetRepository
	}
	tests := []struct {
		name string
		args args
		want *SheetService
	}{
		{
			name: "nil args",
			args: args{nil},
			want: &SheetService{nil},
		},
		{
			name: "non nil args",
			args: args{&MockSheetRepository{}},
			want: &SheetService{&MockSheetRepository{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.repo); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSheetGetSheetByName(t *testing.T) {
	type fields struct {
		repository repository.SheetRepository
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Sheet
		wantErr bool
	}{
		{
			name:    "Sheet exists",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), name: "sheet01"},
			want:    &Sheet{Id: 1, Name: "sheet01"},
			wantErr: false,
		},
		{
			name:    "Sheet does not exists",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), name: "sheetdoesnotexists"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), name: "sheeterror"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SheetService{
				repository: tt.fields.repository,
			}
			got, err := s.getSheetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.GetSheetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.GetSheetByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSheetCreateSheetByName(t *testing.T) {
	type fields struct {
		repository repository.SheetRepository
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Sheet
		wantErr bool
	}{
		{
			name:    "Name is empty",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), name: ""},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Unique sheet",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), name: "sheet01"},
			want:    &Sheet{Id: 1, Name: "sheet01"},
			wantErr: false,
		},
		{
			name:    "Duplicate sheet",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), name: "duplicatesheet"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SheetService{
				repository: tt.fields.repository,
			}
			got, err := s.CreateSheetByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.CreateSheetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.CreateSheetByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSheetGetSheetById(t *testing.T) {
	type fields struct {
		repository repository.SheetRepository
	}
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Sheet
		wantErr bool
	}{
		{
			name:    "Sheet exists",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), id: 1},
			want:    &Sheet{Id: 1, Name: "sheet01name", CreatedAt: &now, UpdatedAt: &now},
			wantErr: false,
		},

		{
			name:    "Sheet does not exists",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.TODO(), id: 2},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SheetService{
				repository: tt.fields.repository,
			}
			got, err := s.GetSheetById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.GetSheetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.GetSheetById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSheetGetAllSheets(t *testing.T) {
	type fields struct {
		repository repository.SheetRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Sheet
		wantErr bool
	}{
		{
			name:   "Empty result",
			fields: fields{&MockSheetRepository{}},
			args: args{
				context.WithValue(context.WithValue(context.Background(), IsErrorCtxKey("isError"), false), IsEmptyCtxKey("isEmpty"), true)},
			want:    []Sheet{},
			wantErr: false,
		},
		{
			name:   "Non empty result",
			fields: fields{&MockSheetRepository{}},
			args:   args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			want: []Sheet{
				{Id: 1, Name: "sheet01"},
				{Id: 2, Name: "sheet02"},
			},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockSheetRepository{}},
			args:    args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SheetService{
				repository: tt.fields.repository,
			}
			got, err := s.GetAllSheets(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.GetAllSheets() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.GetAllSheets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSheetUpdateSheetById(t *testing.T) {
	type fields struct {
		repository repository.SheetRepository
	}
	type args struct {
		ctx   context.Context
		id    int
		sheet *Sheet
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Sheet
		wantErr bool
	}{
		{
			name:   "Sheet.Name is empty",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx:   context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:    1,
				sheet: &Sheet{Id: 1, Name: ""},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Sheet.Id matching id - No error",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx:   context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:    1,
				sheet: &Sheet{Id: 1, Name: "sheet01name"},
			},
			want:    &Sheet{Id: 1, Name: "sheet01name", CreatedAt: &now, UpdatedAt: &now},
			wantErr: false,
		},
		{
			name:   "Sheet.Id matching id - Error",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx:   context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:    1,
				sheet: &Sheet{Id: 1, Name: "sheet01name"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Sheet.Id not matching id - No error",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx:   context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:    1,
				sheet: &Sheet{Id: 2, Name: "sheet01name"},
			},
			want:    &Sheet{Id: 1, Name: "sheet01name", CreatedAt: &now, UpdatedAt: &now},
			wantErr: false,
		},
		{
			name:   "Sheet.Id not matching id - Error",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx:   context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:    1,
				sheet: &Sheet{Id: 2, Name: "sheet01newname"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Sheet.Id matching id - error from GetSheetById",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx:   context.WithValue(context.Background(), IsErrorCtxKey("isError"), "isErrorFromUpdateSheetById"),
				id:    1,
				sheet: &Sheet{Id: 1, Name: "sheet01newname"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Sheet does not exists",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx:   context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:    2,
				sheet: &Sheet{Id: 2, Name: "sheet01newname"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SheetService{
				repository: tt.fields.repository,
			}
			got, err := s.UpdateSheetById(tt.args.ctx, tt.args.id, tt.args.sheet)
			if (err != nil) != tt.wantErr {
				t.Errorf("Sheet.UpdateSheetById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sheet.UpdateSheetById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSheetServiceDeleteSheetById(t *testing.T) {
	type fields struct {
		repository repository.SheetRepository
	}
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "Sheet found - no error",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:  1,
			},
			wantErr: false,
		},
		{
			name:   "Sheet found - Error",
			fields: fields{&MockSheetRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:  1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SheetService{
				repository: tt.fields.repository,
			}
			if err := s.DeleteSheetById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("SheetService.DeleteSheetById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSheetServicePing(t *testing.T) {
	type fields struct {
		repository repository.SheetRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "No error",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockSheetRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SheetService{
				repository: tt.fields.repository,
			}
			if err := s.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("SheetService.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLToSheet(t *testing.T) {
	type args struct {
		sheet *sql.Sheet
	}
	tests := []struct {
		name string
		args args
		want *Sheet
	}{
		{
			name: "Non nil",
			args: args{&sql.Sheet{
				Id:        1,
				Name:      "sheet01",
				CreatedAt: &now,
				UpdatedAt: &now,
			}},
			want: &Sheet{
				Id:        1,
				Name:      "sheet01",
				CreatedAt: &now,
				UpdatedAt: &now,
			},
		},
		{
			name: "Nil",
			args: args{nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SQLToSheet(tt.args.sheet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLToSheet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSheetToSQL(t *testing.T) {
	type args struct {
		sheet *Sheet
	}
	tests := []struct {
		name string
		args args
		want *sql.Sheet
	}{
		{
			name: "Non nil",
			args: args{&Sheet{
				Id:        1,
				Name:      "sheet01",
				CreatedAt: &now,
				UpdatedAt: &now,
			}},
			want: &sql.Sheet{
				Id:        1,
				Name:      "sheet01",
				CreatedAt: &now,
				UpdatedAt: &now,
			},
		},
		{
			name: "Nil",
			args: args{nil},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SheetToSQL(tt.args.sheet); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SheetToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}
