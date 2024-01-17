package roaster

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

type MockRoasterRepository struct{}

func (m *MockRoasterRepository) CreateRoaster(ctx context.Context, roaster *sql.Roaster) error {
	switch roaster.Name {
	case "duplicateroaster":
		return errors.ErrRoasterAlreadyExists
	default:
		return nil
	}
}

func (m *MockRoasterRepository) GetRoasterById(ctx context.Context, id int) (*sql.Roaster, error) {
	if id == 1 {
		return &sql.Roaster{Id: id, Name: "roasterexists"}, nil
	} else {
		return nil, errors.ErrRoasterDoesNotExist
	}
}

func (m *MockRoasterRepository) GetRoasterByName(ctx context.Context, name string) (*sql.Roaster, error) {
	switch name {
	case "roasterdoesnotexists":
		return nil, errors.ErrRoasterDoesNotExist

	case "roastererror":
		return nil, fmt.Errorf("mock error")

	default:
		return &sql.Roaster{Id: 1, Name: "roaster01"}, nil
	}
}

func (m *MockRoasterRepository) GetAllRoasters(ctx context.Context) ([]sql.Roaster, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if isEmpty := ctx.Value(IsEmptyCtxKey("isEmpty")); isEmpty == true {
		return []sql.Roaster{}, nil
	}

	return []sql.Roaster{
		{Id: 1, Name: "roaster01"},
		{Id: 2, Name: "roaster02"},
	}, nil
}

func (m *MockRoasterRepository) UpdateRoasterById(ctx context.Context, id int, roaster *sql.Roaster) (*sql.Roaster, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if id == 1 {
		return &sql.Roaster{Id: id, Name: roaster.Name, UpdatedAt: &now}, nil
	} else {
		return nil, errors.ErrRoasterDoesNotExist
	}
}

func (m *MockRoasterRepository) DeleteRoasterById(ctx context.Context, id int) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func (m *MockRoasterRepository) Ping(ctx context.Context) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func TestNew(t *testing.T) {
	type args struct {
		repo repository.RoasterRepository
	}
	tests := []struct {
		name string
		args args
		want *RoasterService
	}{
		{
			name: "nil args",
			args: args{nil},
			want: &RoasterService{nil},
		},
		{
			name: "non nil args",
			args: args{&MockRoasterRepository{}},
			want: &RoasterService{&MockRoasterRepository{}},
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

func TestRoasterGetRoasterByName(t *testing.T) {
	type fields struct {
		repository repository.RoasterRepository
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Roaster
		wantErr bool
	}{
		{
			name:    "Roaster exists",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.TODO(), name: "roaster01"},
			want:    &Roaster{Id: 1, Name: "roaster01"},
			wantErr: false,
		},
		{
			name:    "Roaster does not exists",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.TODO(), name: "roasterdoesnotexists"},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.TODO(), name: "roastererror"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RoasterService{
				repository: tt.fields.repository,
			}
			got, err := s.getRoasterByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.GetRoasterByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.GetRoasterByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoasterCreateRoasterByName(t *testing.T) {
	type fields struct {
		repository repository.RoasterRepository
	}
	type args struct {
		ctx  context.Context
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Roaster
		wantErr bool
	}{
		{
			name:    "Unique roaster",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.TODO(), name: "roaster01"},
			want:    &Roaster{Id: 1, Name: "roaster01"},
			wantErr: false,
		},
		{
			name:    "Duplicate roaster",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.TODO(), name: "duplicateroaster"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RoasterService{
				repository: tt.fields.repository,
			}
			got, err := s.CreateRoasterByName(tt.args.ctx, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.CreateRoasterByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.CreateRoasterByName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoasterGetRoasterById(t *testing.T) {
	type fields struct {
		repository repository.RoasterRepository
	}
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Roaster
		wantErr bool
	}{
		{
			name:    "Roaster exists",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.TODO(), id: 1},
			want:    &Roaster{Id: 1, Name: "roasterexists"},
			wantErr: false,
		},

		{
			name:    "Roaster does not exists",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.TODO(), id: 2},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RoasterService{
				repository: tt.fields.repository,
			}
			got, err := s.GetRoasterById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.GetRoasterById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.GetRoasterById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoasterGetAllRoasters(t *testing.T) {
	type fields struct {
		repository repository.RoasterRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Roaster
		wantErr bool
	}{
		{
			name:   "Empty result",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				context.WithValue(context.WithValue(context.Background(), IsErrorCtxKey("isError"), false), IsEmptyCtxKey("isEmpty"), true)},
			want:    []Roaster{},
			wantErr: false,
		},
		{
			name:   "Non empty result",
			fields: fields{&MockRoasterRepository{}},
			args:   args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			want: []Roaster{
				{Id: 1, Name: "roaster01"},
				{Id: 2, Name: "roaster02"},
			},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RoasterService{
				repository: tt.fields.repository,
			}
			got, err := s.GetAllRoasters(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.GetAllRoasters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.GetAllRoasters() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoasterUpdateRoasterById(t *testing.T) {
	type fields struct {
		repository repository.RoasterRepository
	}
	type args struct {
		ctx     context.Context
		id      int
		roaster *Roaster
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Roaster
		wantErr bool
	}{
		{
			name:   "Roaster.Id matching id - No error",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				ctx:     context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:      1,
				roaster: &Roaster{Id: 1, Name: "roaster01newname"},
			},
			want:    &Roaster{Id: 1, Name: "roaster01newname", UpdatedAt: &now},
			wantErr: false,
		},
		{
			name:   "Roaster.Id matching id - Error",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				ctx:     context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:      1,
				roaster: &Roaster{Id: 1, Name: "roaster01newname"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Roaster.Id not matching id - No error",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				ctx:     context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:      1,
				roaster: &Roaster{Id: 2, Name: "roaster01newname"},
			},
			want:    &Roaster{Id: 1, Name: "roaster01newname", UpdatedAt: &now},
			wantErr: false,
		},
		{
			name:   "Roaster.Id not matching id - Error",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				ctx:     context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:      1,
				roaster: &Roaster{Id: 2, Name: "roaster01newname"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Roaster does not exists",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				ctx:     context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:      2,
				roaster: &Roaster{Id: 2, Name: "roaster01newname"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RoasterService{
				repository: tt.fields.repository,
			}
			got, err := s.UpdateRoasterById(tt.args.ctx, tt.args.id, tt.args.roaster)
			if (err != nil) != tt.wantErr {
				t.Errorf("Roaster.UpdateRoasterById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Roaster.UpdateRoasterById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoasterServiceDeleteRoasterById(t *testing.T) {
	type fields struct {
		repository repository.RoasterRepository
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
			name:   "Roaster found - no error",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:  1,
			},
			wantErr: false,
		},
		{
			name:   "Roaster found - Error",
			fields: fields{&MockRoasterRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:  1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RoasterService{
				repository: tt.fields.repository,
			}
			if err := s.DeleteRoasterById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("RoasterService.DeleteRoasterById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoasterServicePing(t *testing.T) {
	type fields struct {
		repository repository.RoasterRepository
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
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockRoasterRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RoasterService{
				repository: tt.fields.repository,
			}
			if err := s.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("RoasterService.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSQLToRoaster(t *testing.T) {
	t.Run("SQLToRoaster", func(t *testing.T) {

		want := &Roaster{
			Id:        1,
			Name:      "roaster01",
			CreatedAt: &now,
			UpdatedAt: &now,
		}

		roaster := SQLToRoaster(&sql.Roaster{
			Id:        1,
			Name:      "roaster01",
			CreatedAt: &now,
			UpdatedAt: &now,
		})

		if !reflect.DeepEqual(roaster, want) {
			t.Errorf("SQLToRoaster() = %v, want %v", roaster, want)
		}
	})
}

func TestRoasterToSQL(t *testing.T) {
	t.Run("RoasterToSQL", func(t *testing.T) {
		roaster := RoasterToSQL(&Roaster{
			Id:        1,
			Name:      "roaster01",
			CreatedAt: &now,
			UpdatedAt: &now,
		})

		want := &sql.Roaster{
			Id:        1,
			Name:      "roaster01",
			CreatedAt: &now,
			UpdatedAt: &now,
		}

		if !reflect.DeepEqual(roaster, want) {
			t.Errorf("RoasterToSQL() = %v, want %v", roaster, want)
		}
	})
}
