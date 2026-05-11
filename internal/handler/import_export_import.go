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
	existingAccounts, _ := h.accountService.List()
	for _, acc := range manifest.Accounts {
		newCatID := categoryMap[acc.CategoryID]
		if newCatID == 0 {
			cats, _ := h.categoryService.List()
			if len(cats) > 0 {
				newCatID = cats[0].ID
			}
		}

		account := &model.Account{
			ID:                acc.ID,
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

		existingAccount, _ := h.accountService.Get(acc.ID)
		if existingAccount == nil {
			for i := range existingAccounts {
				candidate := existingAccounts[i]
				if candidate.Title == acc.Title && candidate.Website == acc.Website && candidate.Username == acc.Username && candidate.Email == acc.Email && candidate.CreatedAt == acc.CreatedAt {
					existingAccount = &candidate
					account.ID = candidate.ID
					break
				}
			}
		}
		var createErr error
		if acc.Password != "" {
			if existingAccount != nil {
				createErr = h.accountService.UpdateImported(account, acc.Password)
			} else {
				createErr = h.accountService.CreateImportedWithPassword(account, acc.Password)
			}
		} else if acc.PasswordEncrypted != "" {
			decoded, err := base64.StdEncoding.DecodeString(acc.PasswordEncrypted)
			if err == nil {
				account.PasswordEncrypted = decoded
				if existingAccount != nil {
					createErr = h.accountService.UpdateImported(account, "")
				} else {
					createErr = h.accountService.CreateImported(account)
				}
			} else {
				createErr = err
			}
		} else {
			if existingAccount != nil {
				createErr = h.accountService.UpdateImported(account, "")
			} else {
				createErr = h.accountService.CreateImportedWithPassword(account, "")
			}
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
	existingNotes, _ := h.personalFileService.List()
	for _, note := range manifest.Notes {
		var newNote *model.PersonalFile
		var err error

		existingNote, _ := h.personalFileService.Get(note.ID)
		if existingNote == nil {
			for i := range existingNotes {
				candidate := existingNotes[i]
				if candidate.Title == note.Title && candidate.Remarks == note.Remarks && candidate.Body == note.Body && candidate.BodyFormat == note.BodyFormat && candidate.OriginalName == note.OriginalName && candidate.CreatedAt == note.CreatedAt {
					existingNote = &candidate
					break
				}
			}
		}
		if existingNote != nil {
			if err := h.personalFileService.HardDeleteNote(existingNote.ID); err != nil {
				continue
			}
		}

		primaryHeader, primaryFile := h.buildPrimaryFile(note, zr)
		newNote, err = h.personalFileService.CreateImported(note.ID, note.Title, note.Remarks, note.Body, note.BodyFormat, primaryHeader, primaryFile, note.CreatedAt, note.UpdatedAt)
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
