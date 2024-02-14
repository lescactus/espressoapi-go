package shot

import (
	"context"
	"fmt"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	sqlshot "github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/rs/zerolog"
)

// Shot
//
// An espresso shot is made from coffee beans, ground at a specific setting,
// with a specific quantity of coffee in and out.
// It also has a specific shot time and water temperature.
//
// The result of a shot can be rated and compared to the previous shot.
// It can also be too bitter or too sour.
//
// swagger:model
type Shot struct {
	Id                            int                                   `json:"id"`
	Sheet                         *sheet.Sheet                          `json:"sheet"`
	Beans                         *bean.Bean                            `json:"beans"`
	GrindSetting                  int                                   `json:"grind_setting"`
	QuantityIn                    float64                               `json:"quantity_in"`
	QuantityOut                   float64                               `json:"quantity_out"`
	ShotTime                      time.Duration                         `json:"shot_time"`
	WaterTemperature              float64                               `json:"water_temperature"`
	Rating                        float64                               `json:"rating"`
	IsTooBitter                   bool                                  `json:"is_too_bitter"`
	IsTooSour                     bool                                  `json:"is_too_sour"`
	ComparaisonWithPreviousResult sqlshot.ComparaisonWithPreviousResult `json:"comparaison_with_previous_result"`
	AdditionalNotes               string                                `json:"additional_notes"`
	CreatedAt                     *time.Time                            `json:"created_at"`
	UpdatedAt                     *time.Time                            `json:"updated_at"`
}

func SQLToShot(shot *sqlshot.Shot) *Shot {
	if shot == nil {
		return nil
	}

	s := new(Shot)

	s.Id = shot.Id
	s.Sheet = sheet.SQLToSheet(shot.Sheet)
	s.Beans = bean.SQLToBean(shot.Beans)
	s.GrindSetting = shot.GrindSetting
	s.QuantityIn = shot.QuantityIn
	s.QuantityOut = shot.QuantityOut
	s.ShotTime = shot.ShotTime
	s.WaterTemperature = shot.WaterTemperature
	s.Rating = shot.Rating
	s.IsTooBitter = shot.IsTooBitter
	s.IsTooSour = shot.IsTooSour
	s.ComparaisonWithPreviousResult = shot.ComparaisonWithPreviousResult
	s.AdditionalNotes = shot.AdditionalNotes
	s.CreatedAt = shot.CreatedAt
	s.UpdatedAt = shot.UpdatedAt

	return s
}

func ShotToSQL(shot *Shot) *sqlshot.Shot {
	if shot == nil {
		return nil
	}

	sqlShot := new(sqlshot.Shot)

	sqlShot.Id = shot.Id
	sqlShot.Sheet = sheet.SheetToSQL(shot.Sheet)
	sqlShot.Beans = bean.BeanToSQL(shot.Beans)
	sqlShot.GrindSetting = shot.GrindSetting
	sqlShot.QuantityIn = shot.QuantityIn
	sqlShot.QuantityOut = shot.QuantityOut
	sqlShot.ShotTime = shot.ShotTime
	sqlShot.WaterTemperature = shot.WaterTemperature
	sqlShot.Rating = shot.Rating
	sqlShot.IsTooBitter = shot.IsTooBitter
	sqlShot.IsTooSour = shot.IsTooSour
	sqlShot.ComparaisonWithPreviousResult = shot.ComparaisonWithPreviousResult
	sqlShot.AdditionalNotes = shot.AdditionalNotes
	sqlShot.CreatedAt = shot.CreatedAt
	sqlShot.UpdatedAt = shot.UpdatedAt

	return sqlShot
}

type Service interface {
	CreateShot(ctx context.Context, shot *Shot) (*Shot, error)
	GetShotById(ctx context.Context, id int) (*Shot, error)
	GetAllShots(ctx context.Context) ([]Shot, error)
	UpdateShotById(ctx context.Context, id int, shot *Shot) (*Shot, error)
	DeleteShotById(ctx context.Context, id int) error
	Ping(ctx context.Context) error
}

type ShotService struct {
	repository repository.ShotRepository
}

var _ Service = (*ShotService)(nil)

func New(repo repository.ShotRepository) *ShotService {
	return &ShotService{repository: repo}
}

func (s *ShotService) CreateShot(ctx context.Context, shot *Shot) (*Shot, error) {
	if shot.WaterTemperature <= 0 {
		shot.WaterTemperature = 93.0
	}

	if !(shot.Rating >= 0.0 && shot.Rating <= 10.0) {
		return nil, errors.ErrShotRatingOutOfRange
	}

	id, err := s.repository.CreateShot(ctx, ShotToSQL(shot))
	if err != nil {
		msg := "could not create shot"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	createdShot, err := s.GetShotById(ctx, id)
	if err != nil {
		msg := "could not get newly created shot"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return createdShot, nil
}

func (s *ShotService) GetShotById(ctx context.Context, id int) (*Shot, error) {
	shot, err := s.repository.GetShotById(ctx, id)
	if err != nil {
		msg := "could not get shot by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToShot(shot), nil
}

func (s *ShotService) GetAllShots(ctx context.Context) ([]Shot, error) {
	sqlShots, err := s.repository.GetAllShots(ctx)
	if err != nil {
		msg := "could not get all shots"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	shots := make([]Shot, len(sqlShots))
	for i, v := range sqlShots {
		shots[i] = *SQLToShot(&v)
	}

	return shots, nil
}

func (s *ShotService) UpdateShotById(ctx context.Context, id int, shot *Shot) (*Shot, error) {
	shot.Id = id

	if shot.WaterTemperature <= 0 {
		shot.WaterTemperature = 93.0
	}

	if !(shot.Rating >= 0.0 && shot.Rating <= 10.0) {
		return nil, errors.ErrShotRatingOutOfRange
	}

	sqlShot := ShotToSQL(shot)

	if _, err := s.repository.UpdateShotById(ctx, id, sqlShot); err != nil {
		msg := "could not update shot by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	updatedShot, err := s.GetShotById(ctx, id)
	if err != nil {
		msg := "could not get updated shot"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return updatedShot, nil
}

func (s *ShotService) DeleteShotById(ctx context.Context, id int) error {
	if err := s.repository.DeleteShotById(ctx, id); err != nil {
		msg := "could not delete shot by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func (s *ShotService) Ping(ctx context.Context) error {
	if err := s.repository.Ping(ctx); err != nil {
		msg := "could not ping database"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}
