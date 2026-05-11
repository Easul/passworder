package handler

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"

	"passworder/internal/model"
)

func (h *ImportExportHandler) readImportManifest(zr *zip.Reader) (ExportData, error) {
	var manifest ExportData
	for _, f := range zr.File {
		if f.Name != "manifest.json" {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return ExportData{}, err
		}
		defer rc.Close()
		if err := json.NewDecoder(rc).Decode(&manifest); err != nil {
			return ExportData{}, err
		}
		return manifest, nil
	}
	return ExportData{}, fmt.Errorf("manifest.json not found in zip")
}

func (h *ImportExportHandler) importCategories(manifest ExportData) map[int64]int64 {
	categoryMap := make(map[int64]int64)
	existingCategories, _ := h.categoryService.List()
	nameToID := make(map[string]int64)
	for _, c := range existingCategories {
		nameToID[c.Name] = c.ID
	}

	for _, cat := range manifest.Categories {
		if existingID, ok := nameToID[cat.Name]; ok {
			categoryMap[cat.ID] = existingID
			continue
		}
		newCat := &model.Category{Name: cat.Name}
		if err := h.categoryService.Create(newCat); err == nil {
			categoryMap[cat.ID] = newCat.ID
			nameToID[newCat.Name] = newCat.ID
		}
	}
	return categoryMap
}

func (h *ImportExportHandler) importAccounts(manifest ExportData, categoryMap map[int64]int64) int {
	if key := h.authService.GetCryptoKey(); key != nil {
		h.accountService.SetCryptoKey(key)
	}

	accountsImported := 0
	for _, acc := range manifest.Accounts {
		newCatID := categoryMap[acc.CategoryID]
		if newCatID == 0 {
			cats, _ := h.categoryService.List()
			if len(cats) > 0 {
				newCatID = cats[0].ID
			}
		}

		account := &model.Account{
			CategoryID:        newCatID,
			Title:             acc.Title,
			Website:           acc.Website,
			Username:          acc.Username,
			Email:             acc.Email,
			ReminderEmail:     acc.ReminderEmail,
			RemindAt:          acc.RemindAt,
			RegistrationTime:  acc.RegistrationTime,
			RegistrationNotes: acc.RegistrationNotes,
			Phone:             acc.Phone,
			Notes:             acc.Notes,
			Tags:              acc.Tags,
			IsFavorite:        acc.IsFavorite,
			Status:            acc.Status,
			CreatedAt:         acc.CreatedAt,
			UpdatedAt:         acc.UpdatedAt,
		}
		if account.Status == "" {
			account.Status = "active"
		}
		account.ReminderPeriodType = acc.ReminderPeriodType
		account.ReminderPeriodValue = acc.ReminderPeriodValue
		account.RemindAt, _ = h.reminderService.NormalizeSchedule(account.RemindAt, acc.ReminderPeriodType, acc.ReminderPeriodValue)

		var createErr error
		if acc.Password != "" {
			createErr = h.accountService.Create(account, acc.Password)
		} else if acc.PasswordEncrypted != "" {
			decoded, err := base64.StdEncoding.DecodeString(acc.PasswordEncrypted)
			if err == nil {
				account.PasswordEncrypted = decoded
				createErr = h.accountService.CreateImported(account)
			} else {
				createErr = err
			}
		} else {
			createErr = h.accountService.Create(account, "")
		}

		if createErr == nil {
			_ = h.reminderService.SyncAccountReminder(account, acc.ReminderPeriodType, acc.ReminderPeriodValue)
			accountsImported++
		}
	}
	return accountsImported
}

func (h *ImportExportHandler) buildPrimaryFile(note ExportNote, zr *zip.Reader) (*multipart.FileHeader, multipart.File) {
	if note.OriginalName == "" {
		return nil, nil
	}
	primaryPath := fmt.Sprintf("notes/%d/primary/%s", note.ID, sanitizeZipName(note.OriginalName))
	for _, f := range zr.File {
		if f.Name != primaryPath {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, nil
		}
		content, readErr := io.ReadAll(rc)
		rc.Close()
		if readErr != nil {
			return nil, nil
		}
		fakeFile := &fakeMultipartFile{name: note.OriginalName, content: content, size: int64(len(content)), mime: note.MimeType}
		return fakeFile.Header(), fakeFile
	}
	return nil, nil
}

func (h *ImportExportHandler) importNotes(manifest ExportData, zr *zip.Reader) int {
	notesImported := 0
	for _, note := range manifest.Notes {
		primaryHeader, primaryFile := h.buildPrimaryFile(note, zr)
		newNote, err := h.personalFileService.CreateImported(note.Title, note.Remarks, note.Body, note.BodyFormat, primaryHeader, primaryFile, note.CreatedAt, note.UpdatedAt)
		if err != nil {
			continue
		}
		notesImported++

		for _, att := range note.Attachments {
			attPath := fmt.Sprintf("notes/%d/attachments/%d_%s", note.ID, att.ID, sanitizeZipName(att.OriginalName))
			for _, f := range zr.File {
				if f.Name != attPath {
					continue
				}
				rc, err := f.Open()
				if err != nil {
					continue
				}
				var buf bytes.Buffer
				size, _ := io.Copy(&buf, rc)
				rc.Close()
				fakeFile := &fakeMultipartFile{name: att.OriginalName, content: buf.Bytes(), size: size, mime: att.MimeType}
				h.noteAttService.Create(newNote.ID, fakeFile.Header(), fakeFile)
				break
			}
		}
	}
	return notesImported
}

func (h *ImportExportHandler) importSettings(manifest ExportData) int {
	settingsImported := 0
	if manifest.Settings != nil {
		for key, value := range manifest.Settings {
			if err := h.settingService.Set(key, value); err == nil {
				settingsImported++
			}
		}
	}
	return settingsImported
}
