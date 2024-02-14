package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

// swagger:parameters createRoaster
type CreateRoasterParams struct {
	// The request body for creating a roaster
	// in: body
	// required: true
	Body CreateRoasterRequest
}

// CreateRoasterRequest represents the request body for creating a roaster
// swagger:model
type CreateRoasterRequest struct {
	Name string `json:"name"`
}

// RoasterResponse represents a roaster for this application
//
// A roaster is the professional who roasts coffee beans.
//
// swagger:response RoasterResponse
type RoasterResponse struct {
	// swagger:allOf
	roaster.Roaster
}

// swagger:route POST /rest/v1/roasters roasters createRoaster
//
// # Create roasters
//
// This will create a new roaster.
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
//	  201: RoasterResponse
//	  400: ErrorResponse
//	  409: ErrorResponse
//	  413: ErrorResponse
func (h *Handler) CreateRoaster(w http.ResponseWriter, r *http.Request) {
	var roasterReq CreateRoasterRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &roasterReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	roaster, err := h.RoasterService.CreateRoasterByName(r.Context(), roasterReq.Name)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	roasterResp := RoasterResponse{*roaster}

	hlog.FromRequest(r).Debug().Dict("roaster", zerolog.Dict().
		Int("id", roaster.Id).
		Str("name", roaster.Name).
		Time("created_at", *roaster.CreatedAt)).
		Msg("roaster successfully created")

	resp, em := json.Marshal(&roasterResp)
	if em != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// swagger:route GET /rest/v1/roasters/{id} roasters getRoaster
//
// # Get roasters
//
// This will get the roaster with the given id.
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
//	    description: id of the roaster to get
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: RoasterResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
func (h *Handler) GetRoasterById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	roaster, err := h.RoasterService.GetRoasterById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	roasterResp := RoasterResponse{*roaster}

	hlog.FromRequest(r).Debug().Dict("roaster", zerolog.Dict().
		Int("id", roaster.Id).
		Str("name", roaster.Name).
		Time("created_at", *roaster.CreatedAt)).
		Msg("roaster found by id")

	resp, err := json.Marshal(roasterResp)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:route GET /rest/v1/roasters roasters getAllRoasters
//
// # Get all roasters
//
// This will show all roasters by default.
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
//	  200: RoasterResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
func (h *Handler) GetAllRoasters(w http.ResponseWriter, r *http.Request) {
	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	roastersResp := make([]RoasterResponse, len(roasters))
	for k, v := range roasters {
		roastersResp[k] = RoasterResponse{v}
	}

	resp, err := json.Marshal(&roastersResp)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:parameters updateRoasterById
type UpdateRoasterByIdRequestParams struct {
	// The request body for updating a roaster
	// in: body
	// required: true
	Body UpdateRoasterByIdRequest
}

// UpdateRoasterByIdRequest represents the request body for updating a roaster
// with the given id
// swagger:model
type UpdateRoasterByIdRequest struct {
	Name string `json:"name"`
}

// swagger:route PUT /rest/v1/roasters/{id} roasters updateRoasterById
//
// # Update roasters
//
// This will update a roaster by its given id.
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
//	    description: id of the roaster to update
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: RoasterResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
//	  413: ErrorResponse
func (h *Handler) UpdateRoasterById(w http.ResponseWriter, r *http.Request) {
	var roasterReq UpdateRoasterByIdRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &roasterReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	roaster := &roaster.Roaster{
		Id:   id,
		Name: roasterReq.Name,
	}

	roaster, err = h.RoasterService.UpdateRoasterById(r.Context(), id, roaster)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	roasterResp := RoasterResponse{*roaster}

	hlog.FromRequest(r).Debug().Dict("roaster", zerolog.Dict().
		Int("id", roaster.Id).
		Str("name", roaster.Name)).
		Msg("roaster successfully updated")

	resp, err := json.Marshal(roasterResp)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:route DELETE /rest/v1/roasters/{id} roasters deleteRoaster
//
// # Delete roasters
//
// This will delete a roaster by its given id.
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
//	    description: id of the roaster to delete
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: ItemDeletedResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
func (h *Handler) DeleteRoasterById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	err = h.RoasterService.DeleteRoasterById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	hlog.FromRequest(r).Debug().Msg("roaster successfully deleted")

	i := ItemDeletedResponse{
		Id:  id,
		Msg: fmt.Sprintf("roaster %d deleted successfully", id),
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
