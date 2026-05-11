package handler

import (
	"archive/zip"
	"bytes"
	"io"
	"net/http"
	"strings"

	"passworder/internal/auth"
	"passworder/internal/service"
)

type ImportExportHandler struct {
	accountService      *service.AccountService
	categoryService     *service.CategoryService
	reminderService     *service.ReminderService
	personalFileService *service.PersonalFileService
	noteAttService      *service.NoteAttachmentService
	authService         *auth.Service
	settingService      *service.SettingService
}

func NewImportExportHandler(
	accountService *service.AccountService,
	categoryService *service.CategoryService,
	reminderService *service.ReminderService,
	personalFileService *service.PersonalFileService,
	noteAttService *service.NoteAttachmentService,
	authService *auth.Service,
	settingService *service.SettingService,
) *ImportExportHandler {
	return &ImportExportHandler{
		accountService:      accountService,
		categoryService:     categoryService,
		reminderService:     reminderService,
		personalFileService: personalFileService,
		noteAttService:      noteAttService,
		authService:         authService,
		settingService:      settingService,
	}
}

func (h *ImportExportHandler) Export(w http.ResponseWriter, r *http.Request) {
	data, notes, err := h.buildExportData()
	if err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to export accounts")
		return
	}
	if err := h.writeExportZip(w, data, notes); err != nil {
		JSONError(w, http.StatusInternalServerError, "Failed to write export zip")
		return
	}
}

func (h *ImportExportHandler) Import(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Failed to read uploaded file")
		return
	}
	defer file.Close()

	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		JSONError(w, http.StatusBadRequest, "Only .zip files are supported")
		return
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Failed to read file content")
		return
	}

	zr, err := zip.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		JSONError(w, http.StatusBadRequest, "Invalid zip file")
		return
	}

	manifest, err := h.readImportManifest(zr)
	if err != nil {
		JSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	categoryMap := h.importCategories(manifest)
	accountsImported := h.importAccounts(manifest, categoryMap)
	notesImported := h.importNotes(manifest, zr)
	settingsImported := h.importSettings(manifest)

	JSONResponse(w, http.StatusOK, map[string]int{
		"categoriesImported": len(manifest.Categories),
		"accountsImported":   accountsImported,
		"notesImported":      notesImported,
		"settingsImported":   settingsImported,
	})
}
