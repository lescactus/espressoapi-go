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

// swagger:parameters createBeans
type CreateBeansParams struct {
	// The request body for creating beans
	// in: body
	// required: true
	Body CreateBeansRequest
}

// CreateBeansRequest represents the request body for creating beans
// swagger:model
type CreateBeansRequest struct {
	Name       string         `json:"name"`
	RoasterId  int            `json:"roaster_id"`
	RoastDate  RoastDate      `json:"roast_date"`
	RoastLevel sql.RoastLevel `json:"roast_level"`
}

// BeansResponse represents coffee beans for this application
//
// Beans have a name, a roaster, a roast date and a roast level.
//
// swagger:response BeansResponse
type BeansResponse struct {
	// swagger:allOf
	bean.Bean
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

// swagger:route POST /rest/v1/beans beans createBeans
//
// # Create beans
//
// This will create new beans.
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
//	  201: BeansResponse
//	  400: ErrorResponse
//	  409: ErrorResponse
//	  413: ErrorResponse
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
	beansResp := BeansResponse{*beans}
	logBeansFromRequest(r, beans, "beans successfully created")

	resp, err := json.Marshal(beansResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

// swagger:route GET /rest/v1/beans/{id} beans getBeans
//
// # Get beans
//
// This will get the beans with the given id.
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
//	    description: id of the beans to get
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: BeansResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
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
	BeansResp := BeansResponse{*beans}
	logBeansFromRequest(r, beans, "beans found by id")

	resp, err := json.Marshal(BeansResp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:route GET /rest/v1/beans beans getAllBeans
//
// # Get all beans
//
// This will show all beans by default.
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
//	  200: BeansResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
func (h *Handler) GetAllBeans(w http.ResponseWriter, r *http.Request) {
	beans, err := h.BeanService.GetAllBeans(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	beansResp := make([]BeansResponse, len(beans))
	for k, v := range beans {
		beansResp[k] = BeansResponse{v}
	}

	resp, err := json.Marshal(&beansResp)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:parameters updateBeansById
type UpdateBeansByIdRequestParams struct {
	// The request body for updating beans
	// in: body
	// required: true
	Body UpdateBeansByIdRequest
}

// UpdateBeansByIdRequest represents the request body for updating beans
// with the given id
// swagger:model
type UpdateBeansByIdRequest struct {
	Name       string         `json:"name"`
	RoasterId  int            `json:"roaster_id"`
	RoastDate  RoastDate      `json:"roast_date"`
	RoastLevel sql.RoastLevel `json:"roast_level"`
}

// swagger:route PUT /rest/v1/beans/{id} beans updateBeansById
//
// # Update beans
//
// This will update beans by its given id.
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
//	    description: id of the beans to update
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: BeansResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
//	  413: ErrorResponse
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
	beansResp := BeansResponse{*beans}
	logBeansFromRequest(r, beans, "beans successfully updated")

	resp, err := json.Marshal(beansResp)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

// swagger:route DELETE /rest/v1/beans/{id} beans deleteBeans
//
// # Delete beans
//
// This will delete beans by its given id.
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
//	    description: id of the beans to delete
//	    required: true
//	    type: integer
//	    format: int32
//
//	Responses:
//	  200: ItemDeletedResponse
//	  400: ErrorResponse
//	  404: ErrorResponse
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
