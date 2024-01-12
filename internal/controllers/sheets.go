package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/lescactus/espressoapi-go/internal/services/sheet"
)

type CreateSheetRequest struct {
	Name string `json:"name"`
}

func (h *Handler) CreateSheet(w http.ResponseWriter, r *http.Request) {
	var sheetReq CreateSheetRequest

	err := jsonDecodeBody(w, r, &sheetReq)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	sheet, err := h.SheetService.CreateSheetByName(r.Context(), sheetReq.Name)
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	resp, em := json.Marshal(&sheet)
	if em != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusCreated)
	w.Write(resp)
}

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

	resp, err := json.Marshal(sheet)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) GetAllSheets(w http.ResponseWriter, r *http.Request) {
	sheets, err := h.SheetService.GetAllSheets(r.Context())
	if err != nil {
		h.SetErrorResponse(w, err)
		return
	}

	resp, err := json.Marshal(&sheets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type UpdateSheetByIdRequest struct {
	Name string `json:"name"`
}

func (h *Handler) UpdateSheetById(w http.ResponseWriter, r *http.Request) {
	var sheetReq UpdateSheetByIdRequest

	err := jsonDecodeBody(w, r, &sheetReq)
	if err != nil {
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

	resp, err := json.Marshal(sheet)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

type SheetDeletedResponse struct {
	Id  int    `json:"id"`
	Msg string `json:"msg"`
}

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

	s := SheetDeletedResponse{
		Id:  id,
		Msg: fmt.Sprintf("sheet %d deleted successfully", id),
	}

	resp, err := json.Marshal(s)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", ContentTypeApplicationJSON)
	w.WriteHeader(http.StatusOK)
	w.Write(resp)
}

func (h *Handler) getIdFromParams(ctx context.Context) (int, error) {
	params := httprouter.ParamsFromContext(ctx)
	idParam := params.ByName("id")
	if idParam == "" {
		return 0, ErrIDNotFound
	}

	id, err := strconv.Atoi(idParam)
	if err != nil {
		return 0, ErrIDNotInteger
	}

	return id, nil
}
