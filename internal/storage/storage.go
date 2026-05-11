package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type FileStorage struct {
	rootPath string
}

func NewFileStorage(rootPath string) *FileStorage {
	return &FileStorage{rootPath: rootPath}
}

func (s *FileStorage) EnsureRoot() error {
	return os.MkdirAll(s.rootPath, 0755)
}

func (s *FileStorage) Save(accountID int64, header *multipart.FileHeader, src io.Reader) (string, int64, error) {
	attachmentDir := filepath.Join(s.rootPath, "attachments", fmt.Sprintf("%d", accountID))
	if err := os.MkdirAll(attachmentDir, 0755); err != nil {
		return "", 0, err
	}

	safeName := sanitizeFileName(header.Filename)
	storedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	fullPath := filepath.Join(attachmentDir, storedName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()

	n, err := io.Copy(dst, src)
	if err != nil {
		return "", 0, err
	}

	return storedName, n, nil
}

func (s *FileStorage) Open(accountID int64, storedName string) (*os.File, os.FileInfo, error) {
	fullPath := filepath.Join(s.rootPath, "attachments", fmt.Sprintf("%d", accountID), storedName)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	return file, info, nil
}

func (s *FileStorage) Delete(accountID int64, storedName string) error {
	fullPath := filepath.Join(s.rootPath, "attachments", fmt.Sprintf("%d", accountID), storedName)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *FileStorage) SavePersonalFile(header *multipart.FileHeader, src io.Reader) (string, int64, error) {
	filesDir := filepath.Join(s.rootPath, "files")
	if err := os.MkdirAll(filesDir, 0755); err != nil {
		return "", 0, err
	}

	safeName := sanitizeFileName(header.Filename)
	storedName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), safeName)
	fullPath := filepath.Join(filesDir, storedName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return "", 0, err
	}
	defer dst.Close()

	n, err := io.Copy(dst, src)
	if err != nil {
		return "", 0, err
	}

	return storedName, n, nil
}

func (s *FileStorage) OpenPersonalFile(storedName string) (*os.File, os.FileInfo, error) {
	fullPath := filepath.Join(s.rootPath, "files", storedName)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, nil, err
	}
	return file, info, nil
}

func (s *FileStorage) DeletePersonalFile(storedName string) error {
	fullPath := filepath.Join(s.rootPath, "files", storedName)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (s *FileStorage) UpdatePersonalFile(storedName string, content []byte) (int64, error) {
	fullPath := filepath.Join(s.rootPath, "files", storedName)

	dst, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	n, err := dst.Write(content)
	if err != nil {
		return 0, err
	}

	return int64(n), nil
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer("/", "_", "\\", "_", "..", "_", ":", "_")
	cleaned := strings.TrimSpace(replacer.Replace(name))
	if cleaned == "" {
		return "file.bin"
	}
	return cleaned
}
