package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// swagger:parameters createSheet
type CreateSheetParams struct {
	// The request body for creating a sheet
	// in: body
	// required: true
	Body CreateSheetRequest
}

// CreateSheetRequest represents the request body for creating a sheet
// swagger:model
type CreateSheetRequest struct {
	Name string `json:"name"`
}

// SheetResponse represents a sheet for this application
//
// A sheet is a collection of shots. It's used to group shots together
// in a logical way.
//
// swagger:response SheetResponse
type SheetResponse struct {
	// swagger:allOf
	sheet.Sheet
}

// swagger:route POST /rest/v1/sheets sheets createSheet
//
// # Create sheets
//
// This will create a new sheet.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Deprecated: false
//
//	Security:
//	  api_key:
//	  oauth:
//
//	Responses:
//	  201: SheetResponse
//	  400: ErrorResponse
//	  409: ErrorResponse
//	  413: ErrorResponse
func (h *Handler) CreateSheet(w http.ResponseWriter, r *http.Request) {
	var sheetReq CreateRoasterRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &sheetReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	sheet, err := h.SheetService.CreateSheetByName(r.Context(), sheetReq.Name)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	hlog.FromRequest(r).Debug().Dict("sheet", zerolog.Dict().
		Int("id", sheet.Id).
		Str("name", sheet.Name).
		Time("created_at", *sheet.CreatedAt)).
		Msg("sheet successfully created")

	resp, em := json.Marshal(&sheet)
	if em != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// swagger:route GET /rest/v1/sheets/{id} sheets getSheet
//
// # Get sheets
//
// This will get the sheet with the given id.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Deprecated: false
//
//	Security:
//	  api_key:
//	  oauth:
//
//	Parameters:
//	  + name: id
//	    in: path
//	    description: id of the sheet to get
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: SheetResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
func (h *Handler) GetSheetById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	sheet, err := h.SheetService.GetSheetById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	sheetResp := SheetResponse{*sheet}

	hlog.FromRequest(r).Debug().Dict("sheet", zerolog.Dict().
		Int("id", sheet.Id).
		Str("name", sheet.Name).
		Time("created_at", *sheet.CreatedAt)).
		Msg("sheet found by id")

	resp, err := json.Marshal(sheetResp)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:route GET /rest/v1/sheets sheets getAllSheets
//
// # Get all sheets
//
// This will show all sheets by default.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Deprecated: false
//
//	Security:
//	  api_key:
//	  oauth:
//
//	Responses:
//	  200: SheetResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
func (h *Handler) GetAllSheets(w http.ResponseWriter, r *http.Request) {
	sheets, err := h.SheetService.GetAllSheets(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	sheetsResp := make([]SheetResponse, len(sheets))
	for k, v := range sheets {
		sheetsResp[k] = SheetResponse{v}
	}

	resp, err := json.Marshal(&sheetsResp)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:parameters updateSheetById
type UpdateSheetByIdRequestParams struct {
	// The request body for updating a sheet
	// in: body
	// required: true
	Body UpdateSheetByIdRequest
}

// UpdateSheetByIdRequest represents the request body for updating a sheet
// with the given id
// swagger:model
type UpdateSheetByIdRequest struct {
	Name string `json:"name"`
}

// swagger:route PUT /rest/v1/sheets/{id} sheets updateSheetById
//
// # Update sheets
//
// This will update a sheet by its given id.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Deprecated: false
//
//	Security:
//	  api_key:
//	  oauth:
//
//	Parameters:
//	  + name: id
//	    in: path
//	    description: id of the sheet to update
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: SheetResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
//	  413: ErrorResponse
func (h *Handler) UpdateSheetById(w http.ResponseWriter, r *http.Request) {
	var sheetReq UpdateSheetByIdRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &sheetReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	sheet := &sheet.Sheet{
		Id:   id,
		Name: sheetReq.Name,
	}

	sheet, err = h.SheetService.UpdateSheetById(r.Context(), id, sheet)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	hlog.FromRequest(r).Debug().Dict("sheet", zerolog.Dict().
		Int("id", sheet.Id).
		Str("name", sheet.Name)).
		Msg("sheet successfully updated")

	resp, err := json.Marshal(sheet)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:route DELETE /rest/v1/sheets/{id} sheets deleteSheet
//
// # Delete sheets
//
// This will delete a sheet by its given id.
//
//	Consumes:
//	- application/json
//
//	Produces:
//	- application/json
//
//	Schemes: http, https
//
//	Deprecated: false
//
//	Security:
//	  api_key:
//	  oauth:
//
//	Parameters:
//	  + name: id
//	    in: path
//	    description: id of the sheet to delete
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: SheetResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
func (h *Handler) DeleteSheetById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	err = h.SheetService.DeleteSheetById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	hlog.FromRequest(r).Debug().Msg("sheet successfully deleted")

	i := ItemDeletedResponse{
		Id:  id,
		Msg: fmt.Sprintf("sheet %d deleted successfully", id),
	}

	resp, err := json.Marshal(i)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
