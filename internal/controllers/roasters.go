package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type CreateRoasterRequest struct {
	Name string `json:"name"`
}

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
	hlog.FromRequest(r).Debug().Dict("roaster", zerolog.Dict().
		Int("id", roaster.Id).
		Str("name", roaster.Name).
		Time("created_at", *roaster.CreatedAt)).
		Msg("roaster successfully created")

	resp, em := json.Marshal(&roaster)
	if em != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

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
	hlog.FromRequest(r).Debug().Dict("roaster", zerolog.Dict().
		Int("id", roaster.Id).
		Str("name", roaster.Name).
		Time("created_at", *roaster.CreatedAt)).
		Msg("roaster found by id")

	resp, err := json.Marshal(roaster)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) GetAllRoasters(w http.ResponseWriter, r *http.Request) {
	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	resp, err := json.Marshal(&roasters)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type UpdateRoasterByIdRequest struct {
	Name string `json:"name"`
}

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
	hlog.FromRequest(r).Debug().Dict("roaster", zerolog.Dict().
		Int("id", roaster.Id).
		Str("name", roaster.Name)).
		Msg("roaster successfully updated")

	resp, err := json.Marshal(roaster)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

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
