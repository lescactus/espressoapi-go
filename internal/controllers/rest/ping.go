package rest

import (
	"encoding/json"
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

	resp, err := json.Marshal(&p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}
