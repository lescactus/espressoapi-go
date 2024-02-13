package shot

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	svcbeans "github.com/lescactus/espressoapi-go/internal/services/bean"
	svcsheet "github.com/lescactus/espressoapi-go/internal/services/sheet"
)

var (
	now = time.Now()
)

type IsErrorCtxKey string
type IsEmptyCtxKey string

type MockShotRepository struct{}

func (m *MockShotRepository) CreateShot(ctx context.Context, shot *sql.Shot) (int, error) {
	switch shot.Id {
	case 1:
		return 1, nil
	case 3:
		return 3, nil
	default:
		return 0, fmt.Errorf("mock error")
	}
}

func (m *MockShotRepository) GetShotById(ctx context.Context, id int) (*sql.Shot, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == "isErrorFromUpdateShotById" {
		return nil, fmt.Errorf("mock error")
	}

	if id == 2 {
		return nil, errors.ErrShotDoesNotExist
	}

	if id == 3 {
		return &sql.Shot{
			Id:               id,
			Sheet:            &sql.Sheet{Id: 1, Name: "sheet01"},
			Beans:            &sql.Beans{Id: 1, Name: "beans01"},
			WaterTemperature: 93.0,
		}, nil
	}

	return &sql.Shot{
		Id:    id,
		Sheet: &sql.Sheet{Id: 1, Name: "sheet01"},
		Beans: &sql.Beans{Id: 1, Name: "beans01"},
	}, nil
}

func (m *MockShotRepository) GetAllShots(ctx context.Context) ([]sql.Shot, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if isEmpty := ctx.Value(IsEmptyCtxKey("isEmpty")); isEmpty == true {
		return []sql.Shot{}, nil
	}

	return []sql.Shot{
		{Id: 1, Sheet: &sql.Sheet{Id: 1, Name: "sheet01"}, Beans: &sql.Beans{Id: 1, Name: "beans01"}},
		{Id: 2, Sheet: &sql.Sheet{Id: 1, Name: "sheet02"}, Beans: &sql.Beans{Id: 1, Name: "beans02"}},
	}, nil
}

func (m *MockShotRepository) UpdateShotById(ctx context.Context, id int, beans *sql.Shot) (*sql.Shot, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if id == 1 {
		return &sql.Shot{
			Id:    id,
			Sheet: &sql.Sheet{Id: 1, Name: "sheet01"},
			Beans: &sql.Beans{Id: 1, Name: "beans01"},
		}, nil
	} else {
		return nil, errors.ErrShotDoesNotExist
	}
}

func (m *MockShotRepository) DeleteShotById(ctx context.Context, id int) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func (m *MockShotRepository) Ping(ctx context.Context) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func TestNew(t *testing.T) {
	type args struct {
		repo repository.ShotRepository
	}
	tests := []struct {
		name string
		args args
		want *ShotService
	}{
		{
			name: "nil args",
			args: args{nil},
			want: &ShotService{nil},
		},
		{
			name: "non nil args",
			args: args{&MockShotRepository{}},
			want: &ShotService{&MockShotRepository{}},
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

func TestSQLToShot(t *testing.T) {
	type args struct {
		shot *sql.Shot
	}
	tests := []struct {
		name string
		args args
		want *Shot
	}{
		{
			name: "Non nil",
			args: args{&sql.Shot{
				Id:                            1,
				Sheet:                         &sql.Sheet{Id: 1, Name: "sheet01"},
				Beans:                         &sql.Beans{Id: 1, Name: "beans01"},
				GrindSetting:                  10,
				QuantityIn:                    18.0,
				QuantityOut:                   36.0,
				ShotTime:                      25 * time.Second,
				WaterTemperature:              94.0,
				Rating:                        8.0,
				IsTooBitter:                   false,
				IsTooSour:                     false,
				ComparaisonWithPreviousResult: sql.Better,
				AdditionalNotes:               "This is a test",
				CreatedAt:                     &now,
				UpdatedAt:                     nil,
			}},
			want: &Shot{
				Id:                            1,
				Sheet:                         &svcsheet.Sheet{Id: 1, Name: "sheet01"},
				Beans:                         &svcbeans.Bean{Id: 1, Name: "beans01"},
				GrindSetting:                  10,
				QuantityIn:                    18.0,
				QuantityOut:                   36.0,
				ShotTime:                      25 * time.Second,
				WaterTemperature:              94.0,
				Rating:                        8.0,
				IsTooBitter:                   false,
				IsTooSour:                     false,
				ComparaisonWithPreviousResult: sql.Better,
				AdditionalNotes:               "This is a test",
				CreatedAt:                     &now,
				UpdatedAt:                     nil,
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
			if got := SQLToShot(tt.args.shot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLToShot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShotToSQL(t *testing.T) {
	type args struct {
		shot *Shot
	}
	tests := []struct {
		name string
		args args
		want *sql.Shot
	}{
		{
			name: "Non nil",
			args: args{&Shot{
				Id:                            1,
				Sheet:                         &svcsheet.Sheet{Id: 1, Name: "sheet01"},
				Beans:                         &svcbeans.Bean{Id: 1, Name: "beans01"},
				GrindSetting:                  10,
				QuantityIn:                    18.0,
				QuantityOut:                   36.0,
				ShotTime:                      25 * time.Second,
				WaterTemperature:              94.0,
				Rating:                        8.0,
				IsTooBitter:                   false,
				IsTooSour:                     false,
				ComparaisonWithPreviousResult: sql.Better,
				AdditionalNotes:               "This is a test",
				CreatedAt:                     &now,
				UpdatedAt:                     nil,
			}},
			want: &sql.Shot{
				Id:                            1,
				Sheet:                         &sql.Sheet{Id: 1, Name: "sheet01"},
				Beans:                         &sql.Beans{Id: 1, Name: "beans01"},
				GrindSetting:                  10,
				QuantityIn:                    18.0,
				QuantityOut:                   36.0,
				ShotTime:                      25 * time.Second,
				WaterTemperature:              94.0,
				Rating:                        8.0,
				IsTooBitter:                   false,
				IsTooSour:                     false,
				ComparaisonWithPreviousResult: sql.Better,
				AdditionalNotes:               "This is a test",
				CreatedAt:                     &now,
				UpdatedAt:                     nil,
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
			if got := ShotToSQL(tt.args.shot); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShotToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShotServiceCreateShot(t *testing.T) {
	type fields struct {
		repository repository.ShotRepository
	}
	type args struct {
		ctx  context.Context
		shot *Shot
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Shot
		wantErr bool
	}{
		{
			name:    "No error",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.TODO(), shot: &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}}},
			want:    &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			wantErr: false,
		},
		{
			name:    "No error - water temperature <= 0",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.TODO(), shot: &Shot{Id: 3, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}, WaterTemperature: -1}},
			want:    &Shot{Id: 3, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}, WaterTemperature: 93.0},
			wantErr: false,
		},
		{
			name:    "Error - Rating out of range",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.TODO(), shot: &Shot{Id: 3, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}, Rating: -1}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error creating shot",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.TODO(), shot: &Shot{Id: 2, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error getting new shot",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true), shot: &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShotService{
				repository: tt.fields.repository,
			}
			got, err := s.CreateShot(tt.args.ctx, tt.args.shot)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShotService.CreateShot() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShotService.CreateShot() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShotServiceGetShotById(t *testing.T) {
	type fields struct {
		repository repository.ShotRepository
	}
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Shot
		wantErr bool
	}{
		{
			name:    "Shot exists",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.TODO(), id: 1},
			want:    &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			wantErr: false,
		},

		{
			name:    "Shot does not exists",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.TODO(), id: 2},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShotService{
				repository: tt.fields.repository,
			}
			got, err := s.GetShotById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShotService.GetShotById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShotService.GetShotById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShotServiceGetAllShots(t *testing.T) {
	type fields struct {
		repository repository.ShotRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Shot
		wantErr bool
	}{
		{
			name:   "Empty result",
			fields: fields{&MockShotRepository{}},
			args: args{
				context.WithValue(context.WithValue(context.Background(), IsErrorCtxKey("isError"), false), IsEmptyCtxKey("isEmpty"), true)},
			want:    []Shot{},
			wantErr: false,
		},
		{
			name:   "Non empty result",
			fields: fields{&MockShotRepository{}},
			args:   args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			want: []Shot{
				{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
				{Id: 2, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet02"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans02"}},
			},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockShotRepository{}},
			args:    args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShotService{
				repository: tt.fields.repository,
			}
			got, err := s.GetAllShots(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShotService.GetAllShots() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShotService.GetAllShots() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShotServiceUpdateShotById(t *testing.T) {
	type fields struct {
		repository repository.ShotRepository
	}
	type args struct {
		ctx  context.Context
		id   int
		shot *Shot
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Shot
		wantErr bool
	}{
		{
			name:   "Shot.Id matching id - No error",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:   1,
				shot: &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			},
			want:    &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			wantErr: false,
		},
		{
			name:   "Shot.Id matching id - Error",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:   1,
				shot: &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Shot.Id matching id - Error Rating out of range",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:   1,
				shot: &Shot{Id: 1, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}, Rating: -1},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Shot.Id not matching id - Error",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:   1,
				shot: &Shot{Id: 2, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Shot.Id matching id - error from GetShotById",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), "isErrorFromUpdateShotById"),
				id:   1,
				shot: &Shot{Id: 2, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Shot does not exists",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:   2,
				shot: &Shot{Id: 2, Sheet: &svcsheet.Sheet{Id: 1, Name: "sheet01"}, Beans: &svcbeans.Bean{Id: 1, Name: "beans01"}},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShotService{
				repository: tt.fields.repository,
			}
			got, err := s.UpdateShotById(tt.args.ctx, tt.args.id, tt.args.shot)
			if (err != nil) != tt.wantErr {
				t.Errorf("ShotService.UpdateShotById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ShotService.UpdateShotById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShotServiceDeleteShotById(t *testing.T) {
	type fields struct {
		repository repository.ShotRepository
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
			name:   "Shot found - no error",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:  1,
			},
			wantErr: false,
		},
		{
			name:   "Shot found - Error",
			fields: fields{&MockShotRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:  1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShotService{
				repository: tt.fields.repository,
			}
			if err := s.DeleteShotById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("ShotService.DeleteShotById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestShotService_Ping(t *testing.T) {
	type fields struct {
		repository repository.ShotRepository
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
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockShotRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ShotService{
				repository: tt.fields.repository,
			}
			if err := s.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("ShotService.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
