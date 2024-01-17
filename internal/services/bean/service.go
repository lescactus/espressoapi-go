package beans

import (
	"context"
	"fmt"
	"time"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/rs/zerolog"
)

type Bean struct {
	Id          int            `json:"id"`
	RoasterName string         `json:"roaster_name"`
	BeansName   string         `json:"beans_name"`
	RoastDate   time.Time      `json:"roast_date"`
	RoastLevel  sql.RoastLevel `json:"roast_level"`
}

func SQLToBeans(beans *sql.Beans) *Bean {
	b := new(Bean)
	b.Id = beans.Id
	b.RoasterName = beans.RoasterName
	b.BeansName = beans.BeansName
	b.RoastDate = beans.RoastDate
	b.RoastLevel = beans.RoastLevel

	return b
}

func BeansToSQL(beans *Bean) *sql.Beans {
	sqlBeans := new(sql.Beans)

	sqlBeans.Id = beans.Id
	sqlBeans.RoasterName = beans.RoasterName
	sqlBeans.BeansName = beans.BeansName
	sqlBeans.RoastDate = beans.RoastDate
	sqlBeans.RoastLevel = beans.RoastLevel

	return sqlBeans
}

type Service interface {
	// CreateBeansByName(ctx context.Context, name string) (*Beans, error)
	// GetBeansById(ctx context.Context, id int) (*Beans, error)
	// GetAllBeanss(ctx context.Context) ([]Beans, error)
	// UpdateBeansById(ctx context.Context, id int, Beans *Beans) (*Beans, error)
	// DeleteBeansById(ctx context.Context, id int) error
	Ping(ctx context.Context) error
}

type BeansService struct {
	repository repository.BeansRepository
}

var _ Service = (*BeansService)(nil)

func New(repo repository.BeansRepository) *BeansService {
	return &BeansService{repository: repo}
}

// func (b *BeansService) CreateSheetByName(ctx context.Context, name string) (*Sheet, error) {
// 	sheet := sql.Sheet{Name: name}

// 	err := s.repository.CreateSheet(ctx, &sheet)
// 	if err != nil {
// 		msg := "could not create sheet"
// 		zerolog.Ctx(ctx).Err(err).Msg(msg)
// 		return nil, fmt.Errorf("%s: %w", msg, err)
// 	}

// 	// Will return the full Sheet as it exists in the DB instead of just the name
// 	return s.getSheetByName(ctx, name)
// }

// func (b *BeansService) GetSheetById(ctx context.Context, id int) (*Sheet, error) {
// 	sheet, err := s.repository.GetSheetById(ctx, id)
// 	if err != nil {
// 		msg := "could not get sheet by id"
// 		zerolog.Ctx(ctx).Err(err).Msg(msg)
// 		return nil, fmt.Errorf("%s: %w", msg, err)
// 	}

// 	return SQLToSheet(sheet), nil
// }

// func (b *BeansService) GetAllSheets(ctx context.Context) ([]Sheet, error) {
// 	sqlSheets, err := s.repository.GetAllSheets(ctx)
// 	if err != nil {
// 		msg := "could not get all sheets"
// 		zerolog.Ctx(ctx).Err(err).Msg(msg)
// 		return nil, fmt.Errorf("%s: %w", msg, err)
// 	}

// 	sheets := make([]Sheet, len(sqlSheets))
// 	for i, v := range sqlSheets {
// 		sheets[i] = *SQLToSheet(&v)
// 	}

// 	return sheets, nil
// }

// func (b *BeansService) UpdateSheetById(ctx context.Context, id int, sheet *Sheet) (*Sheet, error) {
// 	sheet.Id = id
// 	sqlSheet := SheetToSQL(sheet)

// 	sqlSheet, err := s.repository.UpdateSheetById(ctx, id, sqlSheet)
// 	if err != nil {
// 		msg := "could not update sheet by id"
// 		zerolog.Ctx(ctx).Err(err).Msg(msg)
// 		return nil, fmt.Errorf("%s: %w", msg, err)
// 	}

// 	return SQLToSheet(sqlSheet), nil
// }

// func (b *BeansService) DeleteSheetById(ctx context.Context, id int) error {
// 	if err := s.repository.DeleteSheetById(ctx, id); err != nil {
// 		msg := "could not delete sheet by id"
// 		zerolog.Ctx(ctx).Err(err).Msg(msg)
// 		return fmt.Errorf("%s: %w", msg, err)
// 	}
// 	return nil
// }

func (b *BeansService) Ping(ctx context.Context) error {
	if err := b.repository.Ping(ctx); err != nil {
		msg := "could not ping database"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

// func (b *BeansService) getSheetByName(ctx context.Context, name string) (*Sheet, error) {
// 	sheet, err := s.repository.GetSheetByName(ctx, name)
// 	if err != nil {
// 		msg := "could not get sheet by name"
// 		zerolog.Ctx(ctx).Err(err).Msg(msg)
// 		return nil, fmt.Errorf("%s: %w", msg, err)
// 	}

// 	return SQLToSheet(sheet), nil
// }
