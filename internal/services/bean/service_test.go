package bean

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
)

var (
	now = time.Now()
)

type IsErrorCtxKey string
type IsEmptyCtxKey string

type MockBeanRepository struct{}

func (m *MockBeanRepository) CreateBeans(ctx context.Context, beans *sql.Beans) (int, error) {
	switch beans.Name {
	case "errorbeans":
		return 0, fmt.Errorf("mock error")
	default:
		return 1, nil
	}
}

func (m *MockBeanRepository) GetBeansById(ctx context.Context, id int) (*sql.Beans, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == "isErrorFromUpdateBeansById" {
		return nil, fmt.Errorf("mock error")
	}

	if id == 0 || id == 1 {
		return &sql.Beans{Id: id, Name: "bean01", Roaster: &sql.Roaster{Id: 1, Name: "roaster01"}, CreatedAt: nil, UpdatedAt: nil}, nil
	} else {
		return nil, errors.ErrBeansDoesNotExist
	}
}

func (m *MockBeanRepository) GetAllBeans(ctx context.Context) ([]sql.Beans, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if isEmpty := ctx.Value(IsEmptyCtxKey("isEmpty")); isEmpty == true {
		return []sql.Beans{}, nil
	}

	return []sql.Beans{
		{Id: 1, Name: "beans01"},
		{Id: 2, Name: "beans02"},
	}, nil
}

func (m *MockBeanRepository) UpdateBeansById(ctx context.Context, id int, beans *sql.Beans) (*sql.Beans, error) {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return nil, fmt.Errorf("mock error")
	}

	if id == 1 {
		return &sql.Beans{Id: id, Name: beans.Name, UpdatedAt: &now}, nil
	} else {
		return nil, errors.ErrBeansDoesNotExist
	}
}

func (m *MockBeanRepository) DeleteBeansById(ctx context.Context, id int) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func (m *MockBeanRepository) Ping(ctx context.Context) error {
	if isError := ctx.Value(IsErrorCtxKey("isError")); isError == true {
		return fmt.Errorf("mock error")
	}

	return nil
}

func TestNew(t *testing.T) {
	type args struct {
		repo repository.BeansRepository
	}
	tests := []struct {
		name string
		args args
		want *BeanService
	}{
		{
			name: "nil args",
			args: args{nil},
			want: &BeanService{nil},
		},
		{
			name: "non nil args",
			args: args{&MockBeanRepository{}},
			want: &BeanService{&MockBeanRepository{}},
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

func TestBeanServiceCreateBean(t *testing.T) {
	type fields struct {
		repository repository.BeansRepository
	}
	type args struct {
		ctx  context.Context
		bean *Bean
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Bean
		wantErr bool
	}{
		{
			name:    "No error",
			fields:  fields{&MockBeanRepository{}},
			args:    args{ctx: context.TODO(), bean: &Bean{Id: 1, Name: "bean01", Roaster: &roaster.Roaster{Id: 1, Name: "roaster01"}}},
			want:    &Bean{Id: 1, Name: "bean01", Roaster: &roaster.Roaster{Id: 1, Name: "roaster01"}, CreatedAt: nil, UpdatedAt: nil},
			wantErr: false,
		},
		{
			name:    "Error creating bean",
			fields:  fields{&MockBeanRepository{}},
			args:    args{ctx: context.TODO(), bean: &Bean{Id: 1, Name: "errorbeans"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "Error getting new bean",
			fields:  fields{&MockBeanRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true), bean: &Bean{Id: 1, Name: "bean01"}},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeanService{
				repository: tt.fields.repository,
			}
			got, err := b.CreateBean(tt.args.ctx, tt.args.bean)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeanService.CreateBean() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeanService.CreateBean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeanServiceGetBeanById(t *testing.T) {
	type fields struct {
		repository repository.BeansRepository
	}
	type args struct {
		ctx context.Context
		id  int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Bean
		wantErr bool
	}{
		{
			name:    "Bean exists",
			fields:  fields{&MockBeanRepository{}},
			args:    args{ctx: context.TODO(), id: 1},
			want:    &Bean{Id: 1, Name: "bean01", Roaster: &roaster.Roaster{Id: 1, Name: "roaster01"}, CreatedAt: nil, UpdatedAt: nil},
			wantErr: false,
		},

		{
			name:    "Bean does not exists",
			fields:  fields{&MockBeanRepository{}},
			args:    args{ctx: context.TODO(), id: 2},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeanService{
				repository: tt.fields.repository,
			}
			got, err := b.GetBeanById(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeanService.GetBeanById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeanService.GetBeanById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeanServiceGetAllBeans(t *testing.T) {
	type fields struct {
		repository repository.BeansRepository
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Bean
		wantErr bool
	}{
		{
			name:   "Empty result",
			fields: fields{&MockBeanRepository{}},
			args: args{
				context.WithValue(context.WithValue(context.Background(), IsErrorCtxKey("isError"), false), IsEmptyCtxKey("isEmpty"), true)},
			want:    []Bean{},
			wantErr: false,
		},
		{
			name:   "Non empty result",
			fields: fields{&MockBeanRepository{}},
			args:   args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			want: []Bean{
				{Id: 1, Name: "beans01"},
				{Id: 2, Name: "beans02"},
			},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockBeanRepository{}},
			args:    args{context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeanService{
				repository: tt.fields.repository,
			}
			got, err := b.GetAllBeans(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeanService.GetAllBeans() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeanService.GetAllBeans() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeanServiceUpdateBeanById(t *testing.T) {
	type fields struct {
		repository repository.BeansRepository
	}
	type args struct {
		ctx  context.Context
		id   int
		bean *Bean
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Bean
		wantErr bool
	}{
		{
			name:   "Bean.Id matching id - No error",
			fields: fields{&MockBeanRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:   1,
				bean: &Bean{Id: 1, Name: "bean01"},
			},
			want:    &Bean{Id: 1, Name: "bean01", Roaster: &roaster.Roaster{Id: 1, Name: "roaster01"}, CreatedAt: nil, UpdatedAt: nil},
			wantErr: false,
		},
		{
			name:   "Bean.Id matching id - Error",
			fields: fields{&MockBeanRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:   1,
				bean: &Bean{Id: 1, Name: "bean01newname"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Bean.Id not matching id - Error",
			fields: fields{&MockBeanRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:   1,
				bean: &Bean{Id: 2, Name: "bean01newname"},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "Bean.Id matching id - error from GetBeansById",
			fields: fields{&MockBeanRepository{}},
			args: args{
				ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), "isErrorFromUpdateBeansById"),
				id:   1,
				bean: &Bean{Id: 1, Name: "bean01"},
			},
			want:    nil,
			wantErr: true,
		},
		// {
		// 	name:   "Bean does not exists",
		// 	fields: fields{&MockBeanRepository{}},
		// 	args: args{
		// 		ctx:  context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
		// 		id:   2,
		// 		bean: &Bean{Id: 2, Name: "bean01newname"},
		// 	},
		// 	want:    nil,
		// 	wantErr: true,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeanService{
				repository: tt.fields.repository,
			}
			got, err := b.UpdateBeanById(tt.args.ctx, tt.args.id, tt.args.bean)
			if (err != nil) != tt.wantErr {
				t.Errorf("BeanService.UpdateBeanById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeanService.UpdateBeanById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBeanServiceDeleteBeanById(t *testing.T) {
	type fields struct {
		repository repository.BeansRepository
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
			name:   "Bean found - no error",
			fields: fields{&MockBeanRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false),
				id:  1,
			},
			wantErr: false,
		},
		{
			name:   "Bean found - Error",
			fields: fields{&MockBeanRepository{}},
			args: args{
				ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true),
				id:  1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &BeanService{
				repository: tt.fields.repository,
			}
			if err := b.DeleteBeanById(tt.args.ctx, tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("BeanService.DeleteBeanById() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBeanServicePing(t *testing.T) {
	type fields struct {
		repository repository.BeansRepository
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
			fields:  fields{&MockBeanRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), false)},
			wantErr: false,
		},
		{
			name:    "Error",
			fields:  fields{&MockBeanRepository{}},
			args:    args{ctx: context.WithValue(context.Background(), IsErrorCtxKey("isError"), true)},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &BeanService{
				repository: tt.fields.repository,
			}
			if err := s.Ping(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("BeanService.Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBeanToSQL(t *testing.T) {
	type args struct {
		bean *Bean
	}
	tests := []struct {
		name string
		args args
		want *sql.Beans
	}{
		{
			name: "Non nil",
			args: args{&Bean{
				Id:         1,
				Roaster:    &roaster.Roaster{Id: 1, Name: "roaster01"},
				Name:       "bean01",
				RoastDate:  &now,
				RoastLevel: sql.RoastLevelLightToMedium,
			}},
			want: &sql.Beans{
				Id:         1,
				Roaster:    &sql.Roaster{Id: 1, Name: "roaster01"},
				Name:       "bean01",
				RoastDate:  &now,
				RoastLevel: sql.RoastLevelLightToMedium,
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
			if got := BeanToSQL(tt.args.bean); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BeanToSQL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSQLToBean(t *testing.T) {
	type args struct {
		beans *sql.Beans
	}
	tests := []struct {
		name string
		args args
		want *Bean
	}{
		{
			name: "Non nil",
			args: args{&sql.Beans{
				Id:         1,
				Roaster:    &sql.Roaster{Id: 1, Name: "roaster01", CreatedAt: nil, UpdatedAt: nil},
				Name:       "beans01",
				RoastDate:  &now,
				RoastLevel: sql.RoastLevelLightToMedium,
			}},
			want: &Bean{
				Id:         1,
				Roaster:    &roaster.Roaster{Id: 1, Name: "roaster01", CreatedAt: nil, UpdatedAt: nil},
				Name:       "beans01",
				RoastDate:  &now,
				RoastLevel: sql.RoastLevelLightToMedium,
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
			if got := SQLToBean(tt.args.beans); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SQLToBean() = %v, want %v", got, tt.want)
			}
		})
	}
}
