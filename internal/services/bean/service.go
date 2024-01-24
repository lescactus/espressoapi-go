package bean

import (
	"context"
	"fmt"
	"time"

	"github.com/lescactus/espressoapi-go/internal/services/roaster"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/repository"
	"github.com/rs/zerolog"
)

type Bean struct {
	Id         int              `json:"id"`
	Roaster    *roaster.Roaster `json:"roaster"`
	Name       string           `json:"name"`
	RoastDate  *time.Time       `json:"roast_date"`
	RoastLevel sql.RoastLevel   `json:"roast_level"`
	CreatedAt  *time.Time       `json:"created_at"`
	UpdatedAt  *time.Time       `json:"updated_at"`
}

func SQLToBean(bean *sql.Beans) *Bean {
	if bean == nil {
		return nil
	}

	b := new(Bean)

	b.Id = bean.Id
	b.Roaster = roaster.SQLToRoaster(bean.Roaster)
	b.Name = bean.Name
	b.RoastDate = bean.RoastDate
	b.RoastLevel = bean.RoastLevel
	b.CreatedAt = bean.CreatedAt
	b.UpdatedAt = bean.UpdatedAt

	return b
}

func BeanToSQL(bean *Bean) *sql.Beans {
	if bean == nil {
		return nil
	}

	sqlBeans := new(sql.Beans)

	sqlBeans.Id = bean.Id
	sqlBeans.Roaster = roaster.RoasterToSQL(bean.Roaster)
	sqlBeans.Name = bean.Name
	sqlBeans.RoastDate = bean.RoastDate
	sqlBeans.RoastLevel = bean.RoastLevel
	sqlBeans.CreatedAt = bean.CreatedAt
	sqlBeans.UpdatedAt = bean.UpdatedAt

	return sqlBeans
}

type Service interface {
	CreateBean(ctx context.Context, bean *Bean) (*Bean, error)
	GetBeanById(ctx context.Context, id int) (*Bean, error)
	GetAllBeans(ctx context.Context) ([]Bean, error)
	UpdateBeanById(ctx context.Context, id int, bean *Bean) (*Bean, error)
	DeleteBeanById(ctx context.Context, id int) error
	Ping(ctx context.Context) error
}

type BeanService struct {
	repository repository.BeansRepository
}

var _ Service = (*BeanService)(nil)

func New(repo repository.BeansRepository) *BeanService {
	return &BeanService{repository: repo}
}

func (b *BeanService) CreateBean(ctx context.Context, bean *Bean) (*Bean, error) {
	if err := b.repository.CreateBeans(ctx, BeanToSQL(bean)); err != nil {
		msg := "could not create bean"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return bean, nil
}

func (b *BeanService) GetBeanById(ctx context.Context, id int) (*Bean, error) {
	bean, err := b.repository.GetBeansById(ctx, id)
	if err != nil {
		msg := "could not get bean by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToBean(bean), nil
}

func (b *BeanService) GetAllBeans(ctx context.Context) ([]Bean, error) {
	sqlBeans, err := b.repository.GetAllBeans(ctx)
	if err != nil {
		msg := "could not get all beans"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	beans := make([]Bean, len(sqlBeans))
	for i, v := range sqlBeans {
		beans[i] = *SQLToBean(&v)
	}

	return beans, nil
}

func (b *BeanService) UpdateBeanById(ctx context.Context, id int, bean *Bean) (*Bean, error) {
	bean.Id = id
	sqlBean := BeanToSQL(bean)

	sqlBean, err := b.repository.UpdateBeansById(ctx, id, sqlBean)
	if err != nil {
		msg := "could not update bean by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return nil, fmt.Errorf("%s: %w", msg, err)
	}

	return SQLToBean(sqlBean), nil
}

func (b *BeanService) DeleteBeanById(ctx context.Context, id int) error {
	if err := b.repository.DeleteBeansById(ctx, id); err != nil {
		msg := "could not delete bean by id"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}

func (b *BeanService) Ping(ctx context.Context) error {
	if err := b.repository.Ping(ctx); err != nil {
		msg := "could not ping database"
		zerolog.Ctx(ctx).Err(err).Msg(msg)
		return fmt.Errorf("%s: %w", msg, err)
	}
	return nil
}
