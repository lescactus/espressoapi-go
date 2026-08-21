package bean

import (
	"context"
	stderrors "errors"
	"testing"
	"time"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
)

func TestBeanServiceRejectsRoastDateBefore1900(t *testing.T) {
	roastDate := time.Date(1899, time.December, 31, 0, 0, 0, 0, time.UTC)
	newBean := func() *Bean {
		return &Bean{
			Name:       "beans",
			Roaster:    &roaster.Roaster{Id: 1},
			RoastDate:  &roastDate,
			RoastLevel: sql.RoastLevelMedium,
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
			if !stderrors.Is(err, domainerrors.ErrBeansRoastDateOutOfRange) {
				t.Errorf("error = %v, want %v", err, domainerrors.ErrBeansRoastDateOutOfRange)
			}
		})
	}
}

func TestBeanServiceAcceptsRoastDateOn1900Floor(t *testing.T) {
	roastDate := time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
	bean := &Bean{
		Name:       "beans",
		Roaster:    &roaster.Roaster{Id: 1},
		RoastDate:  &roastDate,
		RoastLevel: sql.RoastLevelMedium,
	}

	if _, err := New(&MockBeanRepository{}).CreateBean(context.Background(), bean); err != nil {
		t.Fatalf("CreateBean() error = %v", err)
	}
}
