package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"passworder/internal/model"
	"passworder/internal/service"
)

type SettingHandler struct {
	service *service.SettingService
}

func NewSettingHandler(service *service.SettingService) *SettingHandler {
	return &SettingHandler{service: service}
}

func (h *SettingHandler) List(w http.ResponseWriter, r *http.Request) {
	settings, err := h.service.List()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to list settings")
		return
	}
	JSONResponse(w, http.StatusOK, settings)
}

func (h *SettingHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	value, err := h.service.Get(key)
	if err != nil {
		JSONError(w, http.StatusNotFound, "Setting not found")
		return
	}
	JSONResponse(w, http.StatusOK, model.Setting{Key: key, Value: value})
}

func (h *SettingHandler) Set(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["key"]

	var req struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.Set(key, req.Value); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to set setting")
		return
	}
	JSONResponse(w, http.StatusOK, model.Setting{Key: key, Value: req.Value})
}

func (h *SettingHandler) GetServerConfig(w http.ResponseWriter, r *http.Request) {
	config := h.service.GetServerConfig()
	JSONResponse(w, http.StatusOK, config)
}

func (h *SettingHandler) SetServerConfig(w http.ResponseWriter, r *http.Request) {
	var req model.ServerConfig
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.SetServerConfig(req); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to save server config")
		return
	}
	JSONResponse(w, http.StatusOK, req)
}
