package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"

	"passworder/internal/model"
	"passworder/internal/repository"
	"passworder/internal/storage"
)

type NoteAttachmentService struct {
	repo  *repository.NoteAttachmentRepository
	store *storage.FileStorage
}

func NewNoteAttachmentService(repo *repository.NoteAttachmentRepository, store *storage.FileStorage) *NoteAttachmentService {
	return &NoteAttachmentService{repo: repo, store: store}
}

func (s *NoteAttachmentService) Create(personalFileID int64, header *multipart.FileHeader, src multipart.File) (*model.NoteAttachment, error) {
	if header == nil || src == nil {
		return nil, fmt.Errorf("no file provided")
	}

	hash := sha256.New()
	tee := io.TeeReader(src, hash)

	storedName, size, err := s.store.SavePersonalFile(header, tee)
	if err != nil {
		return nil, err
	}

	fileType := classifyFileType(header.Filename, header.Header.Get("Content-Type"))

	now := model.Now()
	a := &model.NoteAttachment{
		PersonalFileID: personalFileID,
		OriginalName:   header.Filename,
		StoredName:     storedName,
		MimeType:       header.Header.Get("Content-Type"),
		SizeBytes:      size,
		Sha256:         hex.EncodeToString(hash.Sum(nil)),
		FileType:       fileType,
		CreatedAt:      now,
	}

	if err := s.repo.Create(a); err != nil {
		s.store.DeletePersonalFile(storedName)
		return nil, err
	}

	return a, nil
}

func (s *NoteAttachmentService) Delete(id int64) error {
	a, err := s.repo.Get(id)
	if err != nil {
		return err
	}
	if a == nil {
		return fmt.Errorf("attachment not found")
	}

	if err := s.repo.Delete(id); err != nil {
		return err
	}

	if a.StoredName != "" {
		s.store.DeletePersonalFile(a.StoredName)
	}
	return nil
}

func (s *NoteAttachmentService) Get(id int64) (*model.NoteAttachment, error) {
	return s.repo.Get(id)
}

func (s *NoteAttachmentService) ListByFile(personalFileID int64) ([]model.NoteAttachment, error) {
	return s.repo.ListByFile(personalFileID)
}

func (s *NoteAttachmentService) Open(id int64) (*model.NoteAttachment, io.ReadCloser, error) {
	a, err := s.repo.Get(id)
	if err != nil {
		return nil, nil, err
	}
	if a == nil {
		return nil, nil, fmt.Errorf("attachment not found")
	}
	if a.StoredName == "" {
		return nil, nil, fmt.Errorf("attachment has no file")
	}

	file, info, err := s.store.OpenPersonalFile(a.StoredName)
	if err != nil {
		return nil, nil, err
	}
	_ = info
	return a, file, nil
}
