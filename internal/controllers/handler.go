package controllers

import (
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

const (
	// ContentTypeApplicationJSON represent the applcation/json Content-Type value
	ContentTypeApplicationJSON = "application/json"
)

type Handler struct {
	SheetService sheet.Service
}

func NewHandler(sheetService sheet.Service) *Handler {
	return &Handler{SheetService: sheetService}
}
