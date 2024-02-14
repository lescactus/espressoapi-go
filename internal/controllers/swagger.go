package controllers

import (
	"bytes"
	"io"
	"net/http"

	"github.com/lescactus/espressoapi-go/cmd/app"
)

func (h *Handler) Swagger(w http.ResponseWriter, r *http.Request) {
	f, err := app.App.SwaggerFS.Open("docs/swagger.json")
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}
	defer f.Close()

	var b bytes.Buffer
	io.Copy(&b, f)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(b.Bytes())
}
