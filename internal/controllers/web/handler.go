package web

import (
	"context"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
)

type Handler struct {
	SheetService   sheet.Service
	RoasterService roaster.Service
	BeanService    bean.Service
	ShotService    shot.Service
}

func NewHandler(
	sheetService sheet.Service,
	roasterService roaster.Service,
	beanService bean.Service,
	ShotService shot.Service) *Handler {
	return &Handler{
		SheetService:   sheetService,
		RoasterService: roasterService,
		BeanService:    beanService,
		ShotService:    ShotService,
	}
}

// getIdFromParams extracts the ID parameter from the context and converts it to an integer.
// It returns the extracted ID and any error encountered during the process.
func (h *Handler) getIdFromParams(ctx context.Context) (int, error) {
	params := httprouter.ParamsFromContext(ctx)
	idParam := params.ByName("id")
	if idParam == "" {
		return 0, nil
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return 0, nil
	}

	return id, nil
}
