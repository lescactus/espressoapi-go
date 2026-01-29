package web

import (
	"fmt"
	"log/slog"
	"net/http"
	"text/template"

	"github.com/lescactus/espressoapi-go/cmd/app"
	"github.com/lescactus/espressoapi-go/internal/services/roaster"
)

func (h *Handler) IsHtmxRequest(r *http.Request) bool {
	return r.Header.Get("HX-Request") == "true"
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	templateFS := app.App.TemplatesFS
	tmpl := template.Must(template.Must(template.ParseFS(templateFS, "views/templates/index.html.tmpl")).ParseFS(templateFS, "views/templates/roasters/*.html.tmpl"))

	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		// h.SetErrorResponse(w, err)
		return
	}

	var data struct {
		PageTitle    string
		RoasterTitle string
		Roasters     []roaster.Roaster
		IsRoasterAdd bool
		IsError      bool
		Error        string
	} = struct {
		PageTitle    string
		RoasterTitle string
		Roasters     []roaster.Roaster
		IsRoasterAdd bool
		IsError      bool
		Error        string
	}{PageTitle: "Hello world", RoasterTitle: "Roasters", Roasters: roasters}

	w.Header().Add("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func (h *Handler) GetRoasters(w http.ResponseWriter, r *http.Request) {
	if !h.IsHtmxRequest(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
		return
	}

	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		// h.SetErrorResponse(w, err)
		return
	}

	var data struct {
		Roasters     []roaster.Roaster
		IsRoasterAdd bool
		IsError      bool
		Error        string
	} = struct {
		Roasters     []roaster.Roaster
		IsRoasterAdd bool
		IsError      bool
		Error        string
	}{Roasters: roasters}

	tmpl := template.Must(template.ParseFS(
		app.App.TemplatesFS,
		"views/templates/roasters/roasters.html.tmpl",
		"views/templates/roasters/row.html.tmpl",
	))
	w.Header().Add("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func (h *Handler) GetRoasterById(w http.ResponseWriter, r *http.Request) {
	if !h.IsHtmxRequest(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
		return
	}

	id, _ := h.getIdFromParams(r.Context())

	roaster, err := h.RoasterService.GetRoasterById(r.Context(), id)
	if err != nil {
		return
	}

	tmpl := template.Must(template.ParseFS(
		app.App.TemplatesFS,
		"views/templates/roasters/row.html.tmpl",
	))
	w.Header().Add("Content-Type", "text/html")
	tmpl.Execute(w, roaster)
}

func (h *Handler) UpdateRoasterById(w http.ResponseWriter, r *http.Request) {
	if !h.IsHtmxRequest(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
		return
	}

	id, _ := h.getIdFromParams(r.Context())

	switch r.Method {
	case http.MethodGet:
		roaster, err := h.RoasterService.GetRoasterById(r.Context(), id)
		if err != nil {
			return
		}

		tmpl := template.Must(template.ParseFS(
			app.App.TemplatesFS,
			"views/templates/roasters/row-edit.html.tmpl",
		))
		tmpl.Execute(w, roaster)

	case http.MethodPut:
		r.ParseForm()

		roaster := &roaster.Roaster{
			Id:   id,
			Name: r.FormValue("name"),
		}

		roaster, err := h.RoasterService.UpdateRoasterById(r.Context(), id, roaster)
		if err != nil {
			return
		}

		tmpl := template.Must(template.ParseFS(
			app.App.TemplatesFS,
			"views/templates/roasters/row.html.tmpl",
		))
		tmpl.Execute(w, roaster)
	}
}

func (h *Handler) CreateRoaster(w http.ResponseWriter, r *http.Request) {
	if !h.IsHtmxRequest(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
		return
	}

	var data struct {
		Roasters     []roaster.Roaster
		IsRoasterAdd bool
		IsError      bool
		Error        string
	} = struct {
		Roasters     []roaster.Roaster
		IsRoasterAdd bool
		IsError      bool
		Error        string
	}{}

	if r.Method == http.MethodPost {
		r.ParseForm()
		name := r.FormValue("name")

		_, err := h.RoasterService.CreateRoasterByName(r.Context(), name)
		if err != nil {
			data.IsError = true
			data.Error = err.Error()
		}

		roasters, err := h.RoasterService.GetAllRoasters(r.Context())
		if err != nil {
			// h.SetErrorResponse(w, err)
			return
		}
		data.Roasters = roasters

		var templateFiles []string
		templateFiles = append(templateFiles, "views/templates/roasters/roasters.html.tmpl")
		templateFiles = append(templateFiles, "views/templates/roasters/row.html.tmpl")

		if data.IsError {
			templateFiles = append(templateFiles, "views/templates/roasters/row-add.html.tmpl")
			data.IsRoasterAdd = true
		} else {
			data.IsRoasterAdd = false
		}

		tmpl := template.Must(template.ParseFS(app.App.TemplatesFS, templateFiles...))

		w.Header().Add("Content-Type", "text/html")
		if err := tmpl.Execute(w, data); err != nil {
			slog.Error(err.Error())
		}
		return
	}

	roasters, err := h.RoasterService.GetAllRoasters(r.Context())
	if err != nil {
		// h.SetErrorResponse(w, err)
		return
	}

	data.Roasters = roasters
	data.IsRoasterAdd = true

	tmpl := template.Must(template.ParseFS(
		app.App.TemplatesFS,
		"views/templates/roasters/roasters-add.html.tmpl",
		"views/templates/roasters/row.html.tmpl",
		"views/templates/roasters/row-add.html.tmpl",
	))
	w.Header().Add("Content-Type", "text/html")
	if err := tmpl.Execute(w, data); err != nil {
		slog.Error(err.Error())
	}
}

func (h *Handler) DeleteRoasterById(w http.ResponseWriter, r *http.Request) {
	if !h.IsHtmxRequest(r) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("404 page not found"))
		return
	}

	id, _ := h.getIdFromParams(r.Context())

	if err := h.RoasterService.DeleteRoasterById(r.Context(), id); err != nil {
		roasters, errg := h.RoasterService.GetAllRoasters(r.Context())
		if errg != nil {
			// h.SetErrorResponse(w, err)
			return
		}
		var data struct {
			Roasters     []roaster.Roaster
			IsRoasterAdd bool
			IsError      bool
			Error        string
		} = struct {
			Roasters     []roaster.Roaster
			IsRoasterAdd bool
			IsError      bool
			Error        string
		}{Roasters: roasters, IsError: true, Error: fmt.Sprintf("Could not delete roaster. Maybe it is still associated with some beans? (technical error: id=%d, err=%s)", id, err.Error())}

		tmpl := template.Must(template.ParseFS(
			app.App.TemplatesFS,
			"views/templates/roasters/roasters.html.tmpl",
			"views/templates/roasters/row.html.tmpl",
		))
		w.Header().Add("Content-Type", "text/html")
		tmpl.Execute(w, data)

		return
	}

	h.GetRoasters(w, r)
}
