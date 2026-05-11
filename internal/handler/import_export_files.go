package handler

import (
	"bytes"
	"mime/multipart"
	"strings"
)

func sanitizeZipName(name string) string {
	if name == "" {
		return "file.bin"
	}
	cleaned := strings.ReplaceAll(name, "..", "_")
	cleaned = strings.ReplaceAll(cleaned, "/", "_")
	cleaned = strings.ReplaceAll(cleaned, "\\", "_")
	return cleaned
}

type fakeMultipartFile struct {
	name    string
	content []byte
	size    int64
	mime    string
	reader  *bytes.Reader
}

func (f *fakeMultipartFile) Read(p []byte) (int, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.content)
	}
	return f.reader.Read(p)
}

func (f *fakeMultipartFile) ReadAt(p []byte, off int64) (int, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.content)
	}
	return f.reader.ReadAt(p, off)
}

func (f *fakeMultipartFile) Seek(offset int64, whence int) (int64, error) {
	if f.reader == nil {
		f.reader = bytes.NewReader(f.content)
	}
	return f.reader.Seek(offset, whence)
}

func (f *fakeMultipartFile) Close() error {
	return nil
}

func (f *fakeMultipartFile) Header() *multipart.FileHeader {
	h := &multipart.FileHeader{
		Filename: f.name,
		Size:     f.size,
	}
	h.Header = make(map[string][]string)
	if f.mime != "" {
		h.Header.Set("Content-Type", f.mime)
	}
	return h
}
