package service

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"mime/multipart"
	"os"

	"passworder/internal/model"
	"passworder/internal/repository"
	"passworder/internal/storage"
)

type AttachmentService struct {
	repo    *repository.AttachmentRepository
	storage *storage.FileStorage
}

func NewAttachmentService(repo *repository.AttachmentRepository, storage *storage.FileStorage) *AttachmentService {
	return &AttachmentService{repo: repo, storage: storage}
}

func (s *AttachmentService) Upload(accountID int64, headers []*multipart.FileHeader) ([]model.Attachment, error) {
	var result []model.Attachment
	for _, header := range headers {
		src, err := header.Open()
		if err != nil {
			return nil, err
		}

		hash := sha256.New()
		tee := io.TeeReader(src, hash)

		storedName, size, err := s.storage.Save(accountID, header, tee)
		src.Close()
		if err != nil {
			return nil, err
		}

		sha256Hash := hex.EncodeToString(hash.Sum(nil))

		att := &model.Attachment{
			AccountID:    accountID,
			OriginalName: header.Filename,
			StoredName:   storedName,
			MimeType:     header.Header.Get("Content-Type"),
			SizeBytes:    size,
			Sha256:       sha256Hash,
		}

		if err := s.repo.Create(att); err != nil {
			_ = s.storage.Delete(accountID, storedName)
			return nil, err
		}
		result = append(result, *att)
	}
	return result, nil
}

func (s *AttachmentService) ListByAccount(accountID int64) ([]model.Attachment, error) {
	return s.repo.ListByAccount(accountID)
}

func (s *AttachmentService) GetFile(id int64) (*model.Attachment, *os.File, os.FileInfo, error) {
	att, err := s.repo.Get(id)
	if err != nil {
		return nil, nil, nil, err
	}
	file, info, err := s.storage.Open(att.AccountID, att.StoredName)
	if err != nil {
		return nil, nil, nil, err
	}
	return att, file, info, nil
}

func (s *AttachmentService) Delete(id int64) error {
	att, err := s.repo.Get(id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(id); err != nil {
		return err
	}
	_ = s.storage.Delete(att.AccountID, att.StoredName)
	return nil
}
