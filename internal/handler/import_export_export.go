package handler

import (
	"archive/zip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func (h *ImportExportHandler) buildExportData() (ExportData, []ExportNote, error) {
	accounts, err := h.accountService.List()
	if err != nil {
		return ExportData{}, nil, err
	}

	categories, err := h.categoryService.List()
	if err != nil {
		return ExportData{}, nil, err
	}

	notes, err := h.personalFileService.List()
	if err != nil {
		return ExportData{}, nil, err
	}

	exportAccounts := make([]ExportAccount, 0, len(accounts))
	for _, acc := range accounts {
		fullAcc, err := h.accountService.Get(acc.ID)
		plainPassword := ""
		if err == nil && fullAcc != nil {
			plainPassword = fullAcc.Password
		}
		exportAccounts = append(exportAccounts, ExportAccount{
			ID:                  acc.ID,
			CategoryID:          acc.CategoryID,
			Title:               acc.Title,
			Website:             acc.Website,
			Username:            acc.Username,
			Password:            plainPassword,
			PasswordEncrypted:   base64.StdEncoding.EncodeToString(acc.PasswordEncrypted),
			Email:               acc.Email,
			ReminderEmail:       acc.ReminderEmail,
			RemindAt:            acc.RemindAt,
			ReminderPeriodType:  acc.ReminderPeriodType,
			ReminderPeriodValue: acc.ReminderPeriodValue,
			RegistrationTime:    acc.RegistrationTime,
			RegistrationNotes:   acc.RegistrationNotes,
			Phone:               acc.Phone,
			Notes:               acc.Notes,
			Tags:                acc.Tags,
			IsFavorite:          acc.IsFavorite,
			Status:              acc.Status,
			CreatedAt:           acc.CreatedAt,
			UpdatedAt:           acc.UpdatedAt,
		})
	}

	exportNotes := make([]ExportNote, 0, len(notes))
	for _, note := range notes {
		atts, _ := h.noteAttService.ListByFile(note.ID)
		exportAtts := make([]ExportAttachment, 0, len(atts))
		for _, att := range atts {
			exportAtts = append(exportAtts, ExportAttachment{
				ID:           att.ID,
				OriginalName: att.OriginalName,
				MimeType:     att.MimeType,
				SizeBytes:    att.SizeBytes,
				FileType:     att.FileType,
				Sha256:       att.Sha256,
			})
		}
		exportNotes = append(exportNotes, ExportNote{
			ID:           note.ID,
			Title:        note.Title,
			Remarks:      note.Remarks,
			Body:         note.Body,
			BodyFormat:   note.BodyFormat,
			OriginalName: note.OriginalName,
			MimeType:     note.MimeType,
			SizeBytes:    note.SizeBytes,
			FileType:     note.FileType,
			Sha256:       note.Sha256,
			Attachments:  exportAtts,
			CreatedAt:    note.CreatedAt,
			UpdatedAt:    note.UpdatedAt,
		})
	}

	settingsMap := make(map[string]string)
	if settings, err := h.settingService.List(); err == nil {
		for _, s := range settings {
			settingsMap[s.Key] = s.Value
		}
	}

	return ExportData{
		Version:    "2.0",
		ExportTime: time.Now().Unix(),
		Accounts:   exportAccounts,
		Categories: categories,
		Notes:      exportNotes,
		Settings:   settingsMap,
	}, exportNotes, nil
}

func (h *ImportExportHandler) writeExportZip(w http.ResponseWriter, data ExportData, notes []ExportNote) error {
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"passworder-export-%s.zip\"", time.Now().Format("2006-01-02")))

	zw := zip.NewWriter(w)
	defer zw.Close()

	manifestBytes, _ := json.MarshalIndent(data, "", "  ")
	manifestWriter, _ := zw.Create("manifest.json")
	_, _ = manifestWriter.Write(manifestBytes)

	for _, note := range notes {
		if note.OriginalName != "" {
			if _, rc, err := h.personalFileService.Open(note.ID); err == nil {
				entryName := fmt.Sprintf("notes/%d/primary/%s", note.ID, sanitizeZipName(note.OriginalName))
				entryWriter, _ := zw.Create(entryName)
				_, _ = io.Copy(entryWriter, rc)
				rc.Close()
			}
		}

		for _, att := range note.Attachments {
			if _, rc, err := h.noteAttService.Open(att.ID); err == nil {
				entryName := fmt.Sprintf("notes/%d/attachments/%d_%s", note.ID, att.ID, sanitizeZipName(att.OriginalName))
				entryWriter, _ := zw.Create(entryName)
				_, _ = io.Copy(entryWriter, rc)
				rc.Close()
			}
		}
	}

	return nil
}
