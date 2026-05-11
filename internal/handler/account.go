package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"passworder/internal/auth"
	"passworder/internal/model"
	"passworder/internal/service"
)

type AccountHandler struct {
	service         *service.AccountService
	reminderService *service.ReminderService
	authService     *auth.Service
	categoryService *service.CategoryService
}

func NewAccountHandler(service *service.AccountService, reminderService *service.ReminderService, authService *auth.Service, categoryService *service.CategoryService) *AccountHandler {
	return &AccountHandler{
		service:         service,
		reminderService: reminderService,
		authService:     authService,
		categoryService: categoryService,
	}
}

type AccountRequest struct {
	CategoryID          int64  `json:"categoryId"`
	Title               string `json:"title"`
	Website             string `json:"website"`
	Username            string `json:"username"`
	Password            string `json:"password"`
	Email               string `json:"email"`
	ReminderEmail       string `json:"reminderEmail"`
	RemindAt            int64  `json:"remindAt"`
	ReminderPeriodType  string `json:"reminderPeriodType"`
	ReminderPeriodValue int    `json:"reminderPeriodValue"`
	RegistrationTime    int64  `json:"registrationTime"`
	RegistrationNotes   string `json:"registrationNotes"`
	Phone               string `json:"phone"`
	Notes               string `json:"notes"`
	Tags                string `json:"tags"`
	IsFavorite          int    `json:"isFavorite"`
	Status              string `json:"status"`
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.service.List()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to list accounts")
		return
	}
	JSONResponse(w, http.StatusOK, accounts)
}

func (h *AccountHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.List(w, r)
		return
	}
	accounts, err := h.service.Search(query)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to search accounts")
		return
	}
	JSONResponse(w, http.StatusOK, accounts)
}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	if key := h.authService.GetCryptoKey(); key != nil {
		h.service.SetCryptoKey(key)
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	account, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to get account")
		return
	}
	if account == nil {
		JSONError(w, http.StatusNotFound, "Account not found")
		return
	}
	JSONResponse(w, http.StatusOK, account)
}

func (h *AccountHandler) GetPassword(w http.ResponseWriter, r *http.Request) {
	if key := h.authService.GetCryptoKey(); key != nil {
		h.service.SetCryptoKey(key)
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	account, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to get account")
		return
	}
	if account == nil {
		JSONError(w, http.StatusNotFound, "Account not found")
		return
	}
	JSONResponse(w, http.StatusOK, map[string]string{"password": account.Password})
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	key := h.authService.GetCryptoKey()
	if key == nil {
		JSONError(w, http.StatusUnauthorized, "Session key not available")
		return
	}
	h.service.SetCryptoKey(key)

	var req AccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Title == "" || req.Password == "" {
		JSONError(w, http.StatusBadRequest, "Title and password are required")
		return
	}

	categoryID, err := h.resolveCategoryID(req.CategoryID)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	account := &model.Account{
		CategoryID:        categoryID,
		Title:             req.Title,
		Website:           req.Website,
		Username:          req.Username,
		Email:             req.Email,
		ReminderEmail:     req.ReminderEmail,
		RemindAt:          req.RemindAt,
		RegistrationTime:  req.RegistrationTime,
		RegistrationNotes: req.RegistrationNotes,
		Phone:             req.Phone,
		Notes:             req.Notes,
		Tags:              req.Tags,
		IsFavorite:        req.IsFavorite,
		Status:            "active",
	}
	account.ReminderPeriodType = req.ReminderPeriodType
	account.ReminderPeriodValue = req.ReminderPeriodValue
	account.RemindAt, _ = h.reminderService.NormalizeSchedule(account.RemindAt, req.ReminderPeriodType, req.ReminderPeriodValue)

	if err := h.service.Create(account, req.Password); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create account")
		return
	}
	if err := h.reminderService.SyncAccountReminder(account, req.ReminderPeriodType, req.ReminderPeriodValue); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to sync reminder")
		return
	}
	JSONResponse(w, http.StatusCreated, account)
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	if key := h.authService.GetCryptoKey(); key != nil {
		h.service.SetCryptoKey(key)
	}

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var req AccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	existing, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to get account")
		return
	}
	if existing == nil {
		JSONError(w, http.StatusNotFound, "Account not found")
		return
	}

	categoryID, err := h.resolveCategoryID(req.CategoryID)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	account := &model.Account{
		ID:                id,
		CategoryID:        categoryID,
		Title:             req.Title,
		Website:           req.Website,
		Username:          req.Username,
		Email:             req.Email,
		ReminderEmail:     req.ReminderEmail,
		RemindAt:          req.RemindAt,
		RegistrationTime:  req.RegistrationTime,
		RegistrationNotes: req.RegistrationNotes,
		Phone:             req.Phone,
		Notes:             req.Notes,
		Tags:              req.Tags,
		IsFavorite:        req.IsFavorite,
		Status:            req.Status,
	}
	account.ReminderPeriodType = req.ReminderPeriodType
	account.ReminderPeriodValue = req.ReminderPeriodValue
	account.RemindAt, _ = h.reminderService.NormalizeSchedule(account.RemindAt, req.ReminderPeriodType, req.ReminderPeriodValue)

	if err := h.service.Update(account, req.Password); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to update account")
		return
	}
	if err := h.reminderService.SyncAccountReminder(account, req.ReminderPeriodType, req.ReminderPeriodValue); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to sync reminder")
		return
	}
	JSONResponse(w, http.StatusOK, account)
}

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to delete account")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}

func (h *AccountHandler) resolveCategoryID(categoryID int64) (int64, error) {
	if categoryID > 0 {
		category, err := h.categoryService.Get(categoryID)
		if err != nil {
			return 0, err
		}
		if category != nil {
			return categoryID, nil
		}
	}

	categories, err := h.categoryService.List()
	if err != nil {
		return 0, err
	}
	if len(categories) == 0 {
		return 0, errors.New("请先创建分类")
	}
	return categories[0].ID, nil
}
