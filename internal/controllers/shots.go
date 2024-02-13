package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
	"github.com/lescactus/espressoapi-go/internal/services/shot"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type CreateShotRequest struct {
	SheetId                       int                               `json:"sheet_id"`
	BeansId                       int                               `json:"beans_id"`
	GrindSetting                  int                               `json:"grind_setting"`
	QuantityIn                    float64                           `json:"quantity_in"`
	QuantityOut                   float64                           `json:"quantity_out"`
	ShotTime                      time.Duration                     `json:"shot_time"`
	WaterTemperature              float64                           `json:"water_temperature"`
	Rating                        float64                           `json:"rating"`
	IsTooBitter                   bool                              `json:"is_too_bitter"`
	IsTooSour                     bool                              `json:"is_too_sour"`
	ComparaisonWithPreviousResult sql.ComparaisonWithPreviousResult `json:"comparaison_with_previous_result"`
	AdditionalNotes               string                            `json:"additional_notes"`
}

func logShotFromRequest(r *http.Request, shot *shot.Shot, msg string) {
	hlog.FromRequest(r).Debug().Dict("shot", zerolog.Dict().
		Int("id", shot.Id).
		Dict("sheet", zerolog.Dict().
			Int("id", shot.Sheet.Id).
			Str("name", shot.Sheet.Name),
		).
		Dict("beans", zerolog.Dict().
			Int("id", shot.Beans.Id).
			Str("name", shot.Beans.Name).
			Dict("roaster", zerolog.Dict().
				Int("id", shot.Beans.Roaster.Id).
				Str("name", shot.Beans.Roaster.Name)),
		)).
		Int("grind_setting", shot.GrindSetting).
		Float64("quantity_in", shot.QuantityIn).
		Float64("quantity_out", shot.QuantityOut).
		Dur("shot_time", shot.ShotTime).
		Float64("water_temperature", shot.WaterTemperature).
		Float64("rating", shot.Rating).
		Bool("is_too_bitter", shot.IsTooBitter).
		Bool("is_too_sour", shot.IsTooSour).
		Uint8("comparaison_with_previous_result", uint8(shot.ComparaisonWithPreviousResult)).
		Str("additional_notes", shot.AdditionalNotes).
		Time("created_at", *shot.CreatedAt).
		Msg(msg)
}

func (h *Handler) CreateShot(w http.ResponseWriter, r *http.Request) {
	var shotReq CreateShotRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &shotReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	shot := &shot.Shot{
		Sheet:                         &sheet.Sheet{Id: shotReq.SheetId},
		Beans:                         &bean.Bean{Id: shotReq.BeansId},
		GrindSetting:                  shotReq.GrindSetting,
		QuantityIn:                    shotReq.QuantityIn,
		QuantityOut:                   shotReq.QuantityOut,
		ShotTime:                      shotReq.ShotTime,
		WaterTemperature:              shotReq.WaterTemperature,
		Rating:                        shotReq.Rating,
		IsTooBitter:                   shotReq.IsTooBitter,
		IsTooSour:                     shotReq.IsTooSour,
		ComparaisonWithPreviousResult: shotReq.ComparaisonWithPreviousResult,
		AdditionalNotes:               shotReq.AdditionalNotes,
	}

	shot, err := h.ShotService.CreateShot(r.Context(), shot)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	logShotFromRequest(r, shot, "shot successfully created")

	resp, err := json.Marshal(shot)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (h *Handler) GetShotById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	shot, err := h.ShotService.GetShotById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	logShotFromRequest(r, shot, "shot found by id")

	resp, err := json.Marshal(shot)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) GetAllShots(w http.ResponseWriter, r *http.Request) {
	shots, err := h.ShotService.GetAllShots(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	resp, err := json.Marshal(&shots)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type UpdateShotByIdRequest struct {
	SheetId                       int                               `json:"sheet_id"`
	BeansId                       int                               `json:"beans_id"`
	GrindSetting                  int                               `json:"grind_setting"`
	QuantityIn                    float64                           `json:"quantity_in"`
	QuantityOut                   float64                           `json:"quantity_out"`
	ShotTime                      time.Duration                     `json:"shot_time"`
	WaterTemperature              float64                           `json:"water_temperature"`
	Rating                        float64                           `json:"rating"`
	IsTooBitter                   bool                              `json:"is_too_bitter"`
	IsTooSour                     bool                              `json:"is_too_sour"`
	ComparaisonWithPreviousResult sql.ComparaisonWithPreviousResult `json:"comparaison_with_previous_result"`
	AdditionalNotes               string                            `json:"additional_notes"`
}

func (h *Handler) UpdateShotById(w http.ResponseWriter, r *http.Request) {
	var shotReq UpdateShotByIdRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &shotReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	shot := &shot.Shot{
		Id:                            id,
		Sheet:                         &sheet.Sheet{Id: shotReq.SheetId},
		Beans:                         &bean.Bean{Id: shotReq.BeansId},
		GrindSetting:                  shotReq.GrindSetting,
		QuantityIn:                    shotReq.QuantityIn,
		QuantityOut:                   shotReq.QuantityOut,
		ShotTime:                      shotReq.ShotTime,
		WaterTemperature:              shotReq.WaterTemperature,
		Rating:                        shotReq.Rating,
		IsTooBitter:                   shotReq.IsTooBitter,
		IsTooSour:                     shotReq.IsTooSour,
		ComparaisonWithPreviousResult: shotReq.ComparaisonWithPreviousResult,
		AdditionalNotes:               shotReq.AdditionalNotes,
	}

	shot, err = h.ShotService.UpdateShotById(r.Context(), id, shot)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	logShotFromRequest(r, shot, "shot successfully updated")

	resp, err := json.Marshal(shot)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) DeleteShotById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	err = h.ShotService.DeleteShotById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	hlog.FromRequest(r).Debug().Msg("shot successfully deleted")

	i := ItemDeletedResponse{
		Id:  id,
		Msg: fmt.Sprintf("shot %d deleted successfully", id),
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
