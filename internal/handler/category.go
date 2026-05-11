package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"passworder/internal/model"
	"passworder/internal/service"
)

type CategoryHandler struct {
	service *service.CategoryService
}

func NewCategoryHandler(service *service.CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.List()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to list categories")
		return
	}
	JSONResponse(w, http.StatusOK, categories)
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	category, err := h.service.Get(id)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to get category")
		return
	}
	if category == nil {
		JSONError(w, http.StatusNotFound, "Category not found")
		return
	}
	JSONResponse(w, http.StatusOK, category)
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var category model.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if category.Name == "" {
		JSONError(w, http.StatusBadRequest, "Name is required")
		return
	}

	if err := h.service.Create(&category); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to create category")
		return
	}
	JSONResponse(w, http.StatusCreated, category)
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	var category model.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	category.ID = id
	if err := h.service.Update(&category); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to update category")
		return
	}
	JSONResponse(w, http.StatusOK, category)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to delete category")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}
