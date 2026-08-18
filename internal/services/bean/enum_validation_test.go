package bean

import (
	"context"
	stderrors "errors"
	"testing"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
)

type noWriteBeanRepository struct {
	MockBeanRepository
}

func (r *noWriteBeanRepository) CreateBeans(context.Context, *sql.Beans) (int, error) {
	panic("CreateBeans must not be called for an invalid roast level")
}

func (r *noWriteBeanRepository) UpdateBeansById(context.Context, int, *sql.Beans) (*sql.Beans, error) {
	panic("UpdateBeansById must not be called for an invalid roast level")
}

func TestBeanServiceRejectsInvalidRoastLevel(t *testing.T) {
	invalidRoastLevel := sql.RoastLevel(5)
	newBean := func() *Bean {
		return &Bean{
			Name:       "beans",
			Roaster:    &roaster.Roaster{Id: 1},
			RoastLevel: invalidRoastLevel,
		}
	}

	tests := []struct {
		name string
		call func(*BeanService) error
	}{
		{
			name: "create",
			call: func(service *BeanService) error {
				_, err := service.CreateBean(context.Background(), newBean())
				return err
			},
		},
		{
			name: "update",
			call: func(service *BeanService) error {
				_, err := service.UpdateBeanById(context.Background(), 1, newBean())
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(New(&noWriteBeanRepository{}))
			if !stderrors.Is(err, domainerrors.ErrBeansRoastLevelOutOfRange) {
				t.Errorf("error = %v, want %v", err, domainerrors.ErrBeansRoastLevelOutOfRange)
			}
		})
	}
}
