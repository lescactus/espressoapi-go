package controllers

import (
	"net/http"
)

// PingResponse represents the json response of a /ping endpoint
type PingResponse struct {
	Ping string `json:"ping"`
}

func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.SheetService.Ping(r.Context())
	if err != nil {
		h.SetErrorResponse(w, &ErrorResponse{
			status: http.StatusInternalServerError,
			Msg:    "unhealthy database",
		})
		return
	}

	p := PingResponse{Ping: "pong"}

	h.writeJSONResponse(w, http.StatusOK, &p)
}
