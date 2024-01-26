package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lescactus/espressoapi-go/internal/models/sql"
	"github.com/lescactus/espressoapi-go/internal/services/bean"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type CreateBeansRequest struct {
	Name       string         `json:"name"`
	RoasterId  int            `json:"roaster_id"`
	RoastDate  RoastDate      `json:"roast_date"`
	RoastLevel sql.RoastLevel `json:"roast_level"`
}

func logBeansFromRequest(r *http.Request, beans *bean.Bean, msg string) {
	hlog.FromRequest(r).Debug().Dict("beans", zerolog.Dict().
		Int("id", beans.Id).
		Str("name", beans.Name).
		Dict("roaster", zerolog.Dict().
			Int("id", beans.Roaster.Id).
			Str("name", beans.Roaster.Name),
		).
		Time("roast_date", *beans.RoastDate).
		Uint8("roast_level", uint8(beans.RoastLevel)).
		Time("created_at", *beans.CreatedAt)).
		Msg(msg)
}

func (h *Handler) CreateBeans(w http.ResponseWriter, r *http.Request) {
	var beansReq CreateBeansRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &beansReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	beans := &bean.Bean{
		Name: beansReq.Name,
		Roaster: &roaster.Roaster{
			Id: beansReq.RoasterId,
		},
		RoastDate:  (*time.Time)(&beansReq.RoastDate),
		RoastLevel: beansReq.RoastLevel,
	}

	beans, err := h.BeanService.CreateBean(r.Context(), beans)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	logBeansFromRequest(r, beans, "beans successfully created")

	resp, err := json.Marshal(beans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

func (h *Handler) GetBeansById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	beans, err := h.BeanService.GetBeanById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	logBeansFromRequest(r, beans, "beans found by id")

	resp, err := json.Marshal(beans)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) GetAllBeans(w http.ResponseWriter, r *http.Request) {
	beans, err := h.BeanService.GetAllBeans(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	resp, err := json.Marshal(&beans)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type UpdateBeansByIdRequest struct {
	Name       string         `json:"name"`
	RoasterId  int            `json:"roaster_id"`
	RoastDate  RoastDate      `json:"roast_date"`
	RoastLevel sql.RoastLevel `json:"roast_level"`
}

func (h *Handler) UpdateBeanById(w http.ResponseWriter, r *http.Request) {
	var beansReq UpdateBeansByIdRequest

	if err := h.parseContentType(r); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	if err := jsonDecodeBody(r, &beansReq); err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	beans := &bean.Bean{
		Id:   id,
		Name: beansReq.Name,
		Roaster: &roaster.Roaster{
			Id: beansReq.RoasterId,
		},
		RoastDate:  (*time.Time)(&beansReq.RoastDate),
		RoastLevel: beansReq.RoastLevel,
	}

	beans, err = h.BeanService.UpdateBeanById(r.Context(), id, beans)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	logBeansFromRequest(r, beans, "beans successfully updated")

	resp, err := json.Marshal(beans)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) DeleteBeansById(w http.ResponseWriter, r *http.Request) {
	id, err := h.getIdFromParams(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	err = h.BeanService.DeleteBeanById(r.Context(), id)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	hlog.FromRequest(r).Debug().Msg("beans successfully deleted")

	i := ItemDeletedResponse{
		Id:  id,
		Msg: fmt.Sprintf("beans %d deleted successfully", id),
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
