package sheet

import (
	"context"
	"fmt"
	"time"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
)

type Sheet struct {
	Id        int        `db:"id" json:"id"`
	Name      string     `db:"name" json:"name"`
	CreatedAt *time.Time `db:"created_at" json:"created_at"`
	UpdatedAt *time.Time `db:"updated_at" json:"updated_at"`
}

func (s *Sheet) SQLToSheet(sheet *sql.Sheet) *Sheet {
	s.Id = sheet.Id
	s.Name = sheet.Name
	s.CreatedAt = sheet.CreatedAt
	s.UpdatedAt = sheet.UpdatedAt

	return s
}

func (s *Sheet) SheetToSQL() *sql.Sheet {
	sqlSheet := new(sql.Sheet)

	sqlSheet.Id = s.Id
	sqlSheet.Name = s.Name
	sqlSheet.CreatedAt = s.CreatedAt
	sqlSheet.UpdatedAt = s.UpdatedAt

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
		return nil, fmt.Errorf("could not create sheet: %w", err)
	}

	// Will return the full Sheet as it exists in the DB instead of just the name
	return s.getSheetByName(ctx, name)
}

func (s *SheetService) GetSheetById(ctx context.Context, id int) (*Sheet, error) {
	sheet, err := s.repository.GetSheetById(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("could not get sheet by id: %w", err)
	}

	return (&Sheet{}).SQLToSheet(sheet), nil
}

func (s *SheetService) GetAllSheets(ctx context.Context) ([]Sheet, error) {
	sqlSheets, err := s.repository.GetAllSheets(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get all sheets: %w", err)
	}

	sheets := make([]Sheet, 0)

	for _, v := range sqlSheets {
		sheets = append(sheets, *(&Sheet{}).SQLToSheet(&v))
	}

	return sheets, nil
}

func (s *SheetService) UpdateSheetById(ctx context.Context, id int, sheet *Sheet) (*Sheet, error) {
	sheet.Id = id

	sqlSheet := sheet.SheetToSQL()

	sqlSheet, err := s.repository.UpdateSheetById(ctx, id, sqlSheet)
	if err != nil {
		return nil, fmt.Errorf("could not update sheet by id: %w", err)
	}

	return (&Sheet{}).SQLToSheet(sqlSheet), nil
}

func (s *SheetService) DeleteSheetById(ctx context.Context, id int) error {
	return s.repository.DeleteSheetById(ctx, id)
}

func (s *SheetService) Ping(ctx context.Context) error {
	return s.repository.Ping(ctx)
}

func (s *SheetService) getSheetByName(ctx context.Context, name string) (*Sheet, error) {
	sheet, err := s.repository.GetSheetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("could not get sheet by name: %w", err)
	}

	return (&Sheet{}).SQLToSheet(sheet), nil
}
