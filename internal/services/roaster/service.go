package roaster

import (
	"context"
	"fmt"
	"time"

	"github.com/lescactus/espressoapi-go/internal/errors"
	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/rs/zerolog"
)

// Roaster
//
// # Represents a roaster for this application
//
// A roaster is the professional who roasts coffee beans.
//
// swagger:model
type Roaster struct {
	// The id for the roaster
	Id int `json:"id"`

	// The name for the roaster
	Name string `json:"name"`

	// The creation date of the roaster
	CreatedAt *time.Time `json:"created_at"`

	// The last update date of the roaster
	UpdatedAt *time.Time `json:"updated_at"`
}

func SQLToRoaster(roaster *sql.Roaster) *Roaster {
	if roaster == nil {
		return nil
	}

	s := new(Roaster)
	s.Id = roaster.Id
	s.Name = roaster.Name
	s.CreatedAt = roaster.CreatedAt
	s.UpdatedAt = roaster.UpdatedAt

	return s
}

func RoasterToSQL(roaster *Roaster) *sql.Roaster {
	if roaster == nil {
		return nil
	}
	sqlRoaster := new(sql.Roaster)

	sqlRoaster.Id = roaster.Id
	sqlRoaster.Name = roaster.Name
	sqlRoaster.CreatedAt = roaster.CreatedAt
	sqlRoaster.UpdatedAt = roaster.UpdatedAt

	return sqlRoaster
}

type Service interface {
	CreateRoasterByName(ctx context.Context, name string) (*Roaster, error)
	GetRoasterById(ctx context.Context, id int) (*Roaster, error)
	GetAllRoasters(ctx context.Context) ([]Roaster, error)
	UpdateRoasterById(ctx context.Context, id int, roaster *Roaster) (*Roaster, error)
	DeleteRoasterById(ctx context.Context, id int) error
	Ping(ctx context.Context) error
}

type RoasterService struct {
	repository repository.RoasterRepository
}

var _ Service = (*RoasterService)(nil)

func New(repo repository.RoasterRepository) *RoasterService {
	return &RoasterService{repository: repo}
}

func (s *RoasterService) CreateRoasterByName(ctx context.Context, name string) (*Roaster, error) {
	if name == "" {
		err := errors.ErrRoasterNameIsEmpty
		msg := "could not create roaster"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	roaster := sql.Roaster{Name: name}

	err := s.repository.CreateRoaster(ctx, &roaster)
	if err != nil {
		msg := "could not create roaster"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	// Will return the full Roaster as it exists in the DB instead of just the name
	return s.getRoasterByName(ctx, name)
}

func (s *RoasterService) GetRoasterById(ctx context.Context, id int) (*Roaster, error) {
	roaster, err := s.repository.GetRoasterById(ctx, id)
	if err != nil {
		msg := "could not get roaster by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToRoaster(roaster), nil
}

func (s *RoasterService) GetAllRoasters(ctx context.Context) ([]Roaster, error) {
	sqlRoasters, err := s.repository.GetAllRoasters(ctx)
	if err != nil {
		msg := "could not get all roasters"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	roasters := make([]Roaster, len(sqlRoasters))
	for i, v := range sqlRoasters {
		roasters[i] = *SQLToRoaster(&v)
	}

	return roasters, nil
}

func (s *RoasterService) UpdateRoasterById(ctx context.Context, id int, roaster *Roaster) (*Roaster, error) {
	if roaster.Name == "" {
		err := errors.ErrRoasterNameIsEmpty
		msg := "could not update roaster by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	roaster.Id = id
	sqlRoaster := RoasterToSQL(roaster)

	if _, err := s.repository.UpdateRoasterById(ctx, id, sqlRoaster); err != nil {
		msg := "could not update roaster by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	updatedRoaster, err := s.GetRoasterById(ctx, roaster.Id)
	if err != nil {
		msg := "could not get updated roaster"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return updatedRoaster, nil
}

func (s *RoasterService) DeleteRoasterById(ctx context.Context, id int) error {
	if err := s.repository.DeleteRoasterById(ctx, id); err != nil {
		msg := "could not delete roaster by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func (s *RoasterService) Ping(ctx context.Context) error {
	if err := s.repository.Ping(ctx); err != nil {
		msg := "could not ping database"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func (s *RoasterService) getRoasterByName(ctx context.Context, name string) (*Roaster, error) {
	roaster, err := s.repository.GetRoasterByName(ctx, name)
	if err != nil {
		msg := "could not get roaster by name"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToRoaster(roaster), nil
}
