package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"passworder/internal/model"
	"passworder/internal/service"
)

type ReminderHandler struct {
	service *service.ReminderService
}

func NewReminderHandler(service *service.ReminderService) *ReminderHandler {
	return &ReminderHandler{service: service}
}

type ReminderRequest struct {
	AccountID int64  `json:"accountId"`
	Title     string `json:"title"`
	RemindAt  int64  `json:"remindAt"`
	Email     string `json:"email"`
}

func (h *ReminderHandler) List(w http.ResponseWriter, r *http.Request) {
	reminders, err := h.service.List()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to list reminders")
		return
	}
	JSONResponse(w, http.StatusOK, reminders)
}

func (h *ReminderHandler) GetPending(w http.ResponseWriter, r *http.Request) {
	reminders, err := h.service.GetPending()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to get pending reminders")
		return
	}
	JSONResponse(w, http.StatusOK, reminders)
}

func (h *ReminderHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req ReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" || req.RemindAt == 0 {
		JSONError(w, http.StatusBadRequest, "Title and remind time are required")
		return
	}

	reminder := &model.Reminder{
		AccountID: req.AccountID,
		Title:     req.Title,
		RemindAt:  req.RemindAt,
		Email:     req.Email,
	}

	if err := h.service.Create(reminder); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create reminder")
		return
	}
	JSONResponse(w, http.StatusCreated, reminder)
}

func (h *ReminderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to delete reminder")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}

func (h *ReminderHandler) SendDue(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.SendDue()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	JSONResponse(w, http.StatusOK, result)
}
