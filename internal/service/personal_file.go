package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"path/filepath"
	"strings"

	"passworder/internal/model"
	"passworder/internal/repository"
	"passworder/internal/storage"
)

type PersonalFileService struct {
	repo           *repository.PersonalFileRepository
	noteAttachment *repository.NoteAttachmentRepository
	store          *storage.FileStorage
}

func NewPersonalFileService(repo *repository.PersonalFileRepository, noteAttachment *repository.NoteAttachmentRepository, store *storage.FileStorage) *PersonalFileService {
	return &PersonalFileService{repo: repo, noteAttachment: noteAttachment, store: store}
}

func (s *PersonalFileService) Create(title, remarks, body, bodyFormat string, header *multipart.FileHeader, src multipart.File) (*model.PersonalFile, error) {
	bodyFormat, err := normalizeBodyFormat(bodyFormat)
	if err != nil {
		return nil, err
	}

	originalName := ""
	storedName := ""
	mimeType := ""
	fileType := classifyFileType("", "")
	var size int64
	sha256Value := ""

	if header != nil && src != nil {
		originalName = header.Filename
		mimeType = header.Header.Get("Content-Type")
		fileType = classifyFileType(header.Filename, mimeType)

		hash := sha256.New()
		tee := io.TeeReader(src, hash)

		storedName, size, err = s.store.SavePersonalFile(header, tee)
		if err != nil {
			return nil, err
		}
		sha256Value = hex.EncodeToString(hash.Sum(nil))
	}

	now := model.Now()
	f := &model.PersonalFile{
		Title:        title,
		Remarks:      remarks,
		Body:         body,
		BodyFormat:   bodyFormat,
		OriginalName: originalName,
		StoredName:   storedName,
		MimeType:     mimeType,
		SizeBytes:    size,
		Sha256:       sha256Value,
		FileType:     fileType,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.repo.Create(f); err != nil {
		if storedName != "" {
			s.store.DeletePersonalFile(storedName)
		}
		return nil, err
	}

	return f, nil
}

func (s *PersonalFileService) CreateImported(title, remarks, body, bodyFormat string, header *multipart.FileHeader, src multipart.File, createdAt, updatedAt int64) (*model.PersonalFile, error) {
	bodyFormat, err := normalizeBodyFormat(bodyFormat)
	if err != nil {
		return nil, err
	}

	originalName := ""
	storedName := ""
	mimeType := ""
	fileType := classifyFileType("", "")
	var size int64
	sha256Value := ""

	if header != nil && src != nil {
		originalName = header.Filename
		mimeType = header.Header.Get("Content-Type")
		fileType = classifyFileType(header.Filename, mimeType)

		hash := sha256.New()
		tee := io.TeeReader(src, hash)

		storedName, size, err = s.store.SavePersonalFile(header, tee)
		if err != nil {
			return nil, err
		}
		sha256Value = hex.EncodeToString(hash.Sum(nil))
	}

	if createdAt == 0 {
		createdAt = model.Now()
	}
	if updatedAt == 0 {
		updatedAt = createdAt
	}

	f := &model.PersonalFile{
		Title:        title,
		Remarks:      remarks,
		Body:         body,
		BodyFormat:   bodyFormat,
		OriginalName: originalName,
		StoredName:   storedName,
		MimeType:     mimeType,
		SizeBytes:    size,
		Sha256:       sha256Value,
		FileType:     fileType,
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}

	if err := s.repo.Create(f); err != nil {
		if storedName != "" {
			s.store.DeletePersonalFile(storedName)
		}
		return nil, err
	}

	return f, nil
}

func (s *PersonalFileService) Update(id int64, title, remarks, body, bodyFormat string) error {
	bodyFormat, err := normalizeBodyFormat(bodyFormat)
	if err != nil {
		return err
	}
	return s.repo.Update(id, title, remarks, body, bodyFormat)
}

func (s *PersonalFileService) DeleteNote(id int64) error {
	f, err := s.repo.Get(id)
	if err != nil {
		return err
	}
	if f == nil {
		return fmt.Errorf("file not found")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}
	return nil
}

func (s *PersonalFileService) HardDeleteNote(id int64) error {
	f, err := s.repo.Get(id)
	if err != nil {
		return err
	}
	if f == nil {
		return fmt.Errorf("file not found")
	}

	attachments, err := s.noteAttachment.ListByFile(id)
	if err != nil {
		return err
	}
	for _, att := range attachments {
		if att.StoredName != "" {
			s.store.DeletePersonalFile(att.StoredName)
		}
	}
	if f.StoredName != "" {
		s.store.DeletePersonalFile(f.StoredName)
	}

	return s.repo.HardDelete(id)
}

func (s *PersonalFileService) Delete(id int64) error {
	return s.DeleteNote(id)
}

func (s *PersonalFileService) Get(id int64) (*model.PersonalFile, error) {
	return s.repo.Get(id)
}

func (s *PersonalFileService) List() ([]model.PersonalFile, error) {
	return s.repo.List()
}

func (s *PersonalFileService) ListDeletedNotes() ([]model.PersonalFile, error) {
	return s.repo.ListDeleted()
}

func (s *PersonalFileService) ListByType(fileType string) ([]model.PersonalFile, error) {
	return s.repo.ListByType(fileType)
}

func (s *PersonalFileService) RestoreNote(id int64) error {
	return s.repo.Restore(id)
}

func (s *PersonalFileService) EmptyTrash() error {
	deletedNotes, err := s.repo.ListDeleted()
	if err != nil {
		return err
	}

	for _, note := range deletedNotes {
		attachments, err := s.noteAttachment.ListByFile(note.ID)
		if err != nil {
			return err
		}

		for _, attachment := range attachments {
			if attachment.StoredName != "" {
				if err := s.store.DeletePersonalFile(attachment.StoredName); err != nil {
					return err
				}
			}
		}

		if note.StoredName != "" {
			if err := s.store.DeletePersonalFile(note.StoredName); err != nil {
				return err
			}
		}
	}

	return s.repo.EmptyTrash()
}

func (s *PersonalFileService) Open(id int64) (*model.PersonalFile, io.ReadCloser, error) {
	f, err := s.repo.Get(id)
	if err != nil {
		return nil, nil, err
	}
	if f == nil {
		return nil, nil, fmt.Errorf("file not found")
	}
	if f.StoredName == "" {
		return nil, nil, fmt.Errorf("file not found")
	}

	file, info, err := s.store.OpenPersonalFile(f.StoredName)
	if err != nil {
		return nil, nil, err
	}
	_ = info
	return f, file, nil
}

func (s *PersonalFileService) ReadContent(id int64) (string, error) {
	_, rc, err := s.Open(id)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *PersonalFileService) UpdateContent(id int64, content string) error {
	f, err := s.repo.Get(id)
	if err != nil {
		return err
	}
	if f == nil {
		return fmt.Errorf("file not found")
	}
	if f.FileType != "markdown" || f.StoredName == "" {
		return fmt.Errorf("非 Markdown 文件，无法编辑内容")
	}

	contentBytes := []byte(content)
	newSize, err := s.store.UpdatePersonalFile(f.StoredName, contentBytes)
	if err != nil {
		return err
	}

	hash := sha256.Sum256(contentBytes)
	newSha256 := hex.EncodeToString(hash[:])

	return s.repo.UpdateContent(id, newSize, newSha256)
}

func classifyFileType(filename, mimeType string) string {
	if filename == "" && mimeType == "" {
		return "none"
	}

	ext := strings.ToLower(filepath.Ext(filename))

	switch ext {
	case ".md", ".markdown":
		return "markdown"
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".svg":
		return "image"
	case ".zip", ".tar", ".gz", ".bz2", ".7z", ".rar":
		return "archive"
	case ".pdf", ".doc", ".docx", ".txt", ".csv", ".xls", ".xlsx", ".ppt", ".pptx":
		return "document"
	}

	if strings.HasPrefix(mimeType, "image/") {
		return "image"
	}
	if strings.HasPrefix(mimeType, "text/") {
		return "document"
	}

	return "other"
}

func mimeByFilename(name string) string {
	return mime.TypeByExtension(filepath.Ext(name))
}

func normalizeBodyFormat(bodyFormat string) (string, error) {
	format := strings.ToLower(strings.TrimSpace(bodyFormat))
	if format == "" {
		return "text", nil
	}
	if format != "markdown" && format != "text" {
		return "", fmt.Errorf("正文格式仅支持 markdown 或 text")
	}
	return format, nil
}
