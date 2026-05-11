package handler

import (
	"fmt"
	"mime"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"passworder/internal/service"
)

type AttachmentHandler struct {
	service        *service.AttachmentService
	accountService *service.AccountService
}

func NewAttachmentHandler(service *service.AttachmentService, accountService *service.AccountService) *AttachmentHandler {
	return &AttachmentHandler{
		service:        service,
		accountService: accountService,
	}
}

const maxFileSize = 10 << 20

func (h *AttachmentHandler) List(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	attachments, err := h.service.ListByAccount(accountID)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to list attachments")
		return
	}
	JSONResponse(w, http.StatusOK, attachments)
}

func (h *AttachmentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	accountID, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid account ID")
		return
	}

	if err := r.ParseMultipartForm(maxFileSize); err != nil {
		JSONError(w, http.StatusBadRequest, "File too large")
		return
	}

	files := r.MultipartForm.File["files"]
	if len(files) == 0 {
		files = r.MultipartForm.File["file"]
	}
	if len(files) == 0 {
		JSONError(w, http.StatusBadRequest, "No files provided")
		return
	}

	attachments, err := h.service.Upload(accountID, files)
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to upload files")
		return
	}
	JSONResponse(w, http.StatusCreated, attachments)
}

func (h *AttachmentHandler) Download(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	att, file, info, err := h.service.GetFile(id)
	if err != nil {
		JSONError(w, http.StatusNotFound, "Attachment not found")
		return
	}
	defer file.Close()

	contentType := att.MimeType
	if contentType == "" {
		contentType = mime.TypeByExtension(att.OriginalName)
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, att.OriginalName))
	http.ServeContent(w, r, att.OriginalName, info.ModTime(), file)
}

func (h *AttachmentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid ID")
		return
	}

	if err := h.service.Delete(id); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to delete attachment")
		return
	}
	JSONResponse(w, http.StatusOK, nil)
}
