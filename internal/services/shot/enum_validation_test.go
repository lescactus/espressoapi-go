package shot

import (
	"context"
	stderrors "errors"
	"testing"

	domainerrors "github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	svcbeans "github.com/lescactus/espressoapi-go/internal/services/bean"
	svcsheet "github.com/lescactus/espressoapi-go/internal/services/sheet"
)

type noWriteShotRepository struct {
	MockShotRepository
}

func (r *noWriteShotRepository) CreateShot(context.Context, *sql.Shot) (int, error) {
	panic("CreateShot must not be called for an invalid comparison result")
}

func (r *noWriteShotRepository) UpdateShotById(context.Context, int, *sql.Shot) (*sql.Shot, error) {
	panic("UpdateShotById must not be called for an invalid comparison result")
}

func TestShotServiceRejectsInvalidComparisonWithPreviousResult(t *testing.T) {
	invalidComparison := sql.ComparisonWithPreviousResult(4)
	newShot := func() *Shot {
		return &Shot{
			Sheet:                        &svcsheet.Sheet{Id: 1},
			Beans:                        &svcbeans.Bean{Id: 1},
			ComparisonWithPreviousResult: invalidComparison,
		}
	}

	tests := []struct {
		name string
		call func(*ShotService) error
	}{
		{
			name: "create",
			call: func(service *ShotService) error {
				_, err := service.CreateShot(context.Background(), newShot())
				return err
			},
		},
		{
			name: "update",
			call: func(service *ShotService) error {
				_, err := service.UpdateShotById(context.Background(), 1, newShot())
				return err
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.call(New(&noWriteShotRepository{}))
			if !stderrors.Is(err, domainerrors.ErrShotComparisonWithPreviousResultOutOfRange) {
				t.Errorf("error = %v, want %v", err, domainerrors.ErrShotComparisonWithPreviousResultOutOfRange)
			}
		})
	}
}
