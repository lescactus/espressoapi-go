// Package web implements the server-rendered htmx + templ web UI. Handlers
// call the existing service interfaces directly (no REST round-trip).
package web

import (
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

func NewHandler(sheetService sheet.Service, roasterService roaster.Service, beanService bean.Service, shotService shot.Service) *Handler {
	return &Handler{
		SheetService:   sheetService,
		RoasterService: roasterService,
		BeanService:    beanService,
		ShotService:    shotService,
	}
}
