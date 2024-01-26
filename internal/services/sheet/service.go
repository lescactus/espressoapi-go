package sheet

import (
	"context"
	"fmt"
	"time"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/rs/zerolog"
)

type Sheet struct {
	Id        int        `json:"id"`
	Name      string     `json:"name"`
	CreatedAt *time.Time `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}

func SQLToSheet(sheet *sql.Sheet) *Sheet {
	s := new(Sheet)
	s.Id = sheet.Id
	s.Name = sheet.Name
	s.CreatedAt = sheet.CreatedAt
	s.UpdatedAt = sheet.UpdatedAt

	return s
}

func SheetToSQL(sheet *Sheet) *sql.Sheet {
	sqlSheet := new(sql.Sheet)

	sqlSheet.Id = sheet.Id
	sqlSheet.Name = sheet.Name
	sqlSheet.CreatedAt = sheet.CreatedAt
	sqlSheet.UpdatedAt = sheet.UpdatedAt

	return sqlSheet
}

type Service interface {
	CreateSheetByName(ctx context.Context, name string) (*Sheet, error)
	GetSheetById(ctx context.Context, id int) (*Sheet, error)
	GetAllSheets(ctx context.Context) ([]Sheet, error)
	UpdateSheetById(ctx context.Context, id int, sheet *Sheet) (*Sheet, error)
	DeleteSheetById(ctx context.Context, id int) error
	Ping(ctx context.Context) error
}

type SheetService struct {
	repository repository.SheetRepository
}

var _ Service = (*SheetService)(nil)

func New(repo repository.SheetRepository) *SheetService {
	return &SheetService{repository: repo}
}

func (s *SheetService) CreateSheetByName(ctx context.Context, name string) (*Sheet, error) {
	sheet := sql.Sheet{Name: name}

	err := s.repository.CreateSheet(ctx, &sheet)
	if err != nil {
		msg := "could not create sheet"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	// Will return the full Sheet as it exists in the DB instead of just the name
	return s.getSheetByName(ctx, name)
}

func (s *SheetService) GetSheetById(ctx context.Context, id int) (*Sheet, error) {
	sheet, err := s.repository.GetSheetById(ctx, id)
	if err != nil {
		msg := "could not get sheet by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToSheet(sheet), nil
}

func (s *SheetService) GetAllSheets(ctx context.Context) ([]Sheet, error) {
	sqlSheets, err := s.repository.GetAllSheets(ctx)
	if err != nil {
		msg := "could not get all sheets"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	sheets := make([]Sheet, len(sqlSheets))
	for i, v := range sqlSheets {
		sheets[i] = *SQLToSheet(&v)
	}

	return sheets, nil
}

func (s *SheetService) UpdateSheetById(ctx context.Context, id int, sheet *Sheet) (*Sheet, error) {
	sheet.Id = id
	sqlSheet := SheetToSQL(sheet)

	sqlSheet, err := s.repository.UpdateSheetById(ctx, id, sqlSheet)
	if err != nil {
		msg := "could not update sheet by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToSheet(sqlSheet), nil
}

func (s *SheetService) DeleteSheetById(ctx context.Context, id int) error {
	if err := s.repository.DeleteSheetById(ctx, id); err != nil {
		msg := "could not delete sheet by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func (s *SheetService) Ping(ctx context.Context) error {
	if err := s.repository.Ping(ctx); err != nil {
		msg := "could not ping database"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func (s *SheetService) getSheetByName(ctx context.Context, name string) (*Sheet, error) {
	sheet, err := s.repository.GetSheetByName(ctx, name)
	if err != nil {
		msg := "could not get sheet by name"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToSheet(sheet), nil
}
