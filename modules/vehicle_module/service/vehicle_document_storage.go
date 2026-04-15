package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type VehicleDocumentStorage struct {
	rootDir string
	baseDir string
}

func NewVehicleDocumentStorage(rootDir string) *VehicleDocumentStorage {
	return &VehicleDocumentStorage{
		rootDir: rootDir,
		baseDir: filepath.Join(rootDir, "vehicle-documents"),
	}
}

func (s *VehicleDocumentStorage) EnsureDir(entityDir string) (string, error) {
	absDir := filepath.Join(s.baseDir, entityDir)
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", err
	}
	return absDir, nil
}

func (s *VehicleDocumentStorage) Save(entityDir string, entityID int64, fh *multipart.FileHeader, oldRelativePath string) (string, error) {
	if fh == nil {
		return "", errors.New("file is required")
	}

	if fh.Size > 20*1024*1024 {
		return "", errors.New("file size must be <= 20MB")
	}

	absDir, err := s.EnsureDir(entityDir)
	if err != nil {
		return "", fmt.Errorf("create document dir: %w", err)
	}

	src, err := fh.Open()
	if err != nil {
		return "", fmt.Errorf("open uploaded file: %w", err)
	}
	defer src.Close()

	head := make([]byte, 512)
	n, err := io.ReadFull(src, head)
	if err != nil && err != io.ErrUnexpectedEOF {
		return "", fmt.Errorf("read uploaded file header: %w", err)
	}

	contentType := http.DetectContentType(head[:n])
	ext, ok := documentExtensionByContentType(contentType, fh.Filename)
	if !ok {
		return "", errors.New("only pdf, images, doc, docx, xls, xlsx, txt files are allowed")
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("reset uploaded file reader: %w", err)
	}

	filename := fmt.Sprintf("%s_%d_%d%s", entityDir, entityID, time.Now().UnixNano(), ext)
	absPath := filepath.Join(absDir, filename)

	dst, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("save uploaded file: %w", err)
	}

	newRelativePath := filepath.ToSlash(filepath.Join("vehicle-documents", entityDir, filename))

	if strings.TrimSpace(oldRelativePath) != "" {
		_ = s.Delete(oldRelativePath)
	}

	return newRelativePath, nil
}

func (s *VehicleDocumentStorage) Delete(relativePath string) error {
	relativePath = strings.TrimSpace(relativePath)
	if relativePath == "" {
		return nil
	}

	cleanRel := filepath.Clean(relativePath)
	if cleanRel == "." || strings.HasPrefix(cleanRel, "..") {
		return errors.New("invalid file path")
	}

	absPath := filepath.Join(s.rootDir, cleanRel)
	if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

func parseFlexibleDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, errors.New("date is required")
	}

	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}

	for _, layout := range layouts {
		t, err := time.Parse(layout, value)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("invalid date format: %s", value)
}

func documentExtensionByContentType(contentType, filename string) (string, bool) {
	contentType = strings.ToLower(strings.TrimSpace(contentType))

	switch contentType {
	case "application/pdf":
		return ".pdf", true
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	case "text/plain; charset=utf-8", "text/plain":
		return ".txt", true
	case "application/msword":
		return ".doc", true
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		return ".docx", true
	case "application/vnd.ms-excel":
		return ".xls", true
	case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
		return ".xlsx", true
	}

	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".pdf", ".jpg", ".jpeg", ".png", ".webp", ".txt", ".doc", ".docx", ".xls", ".xlsx":
		if ext == ".jpeg" {
			return ".jpg", true
		}
		return ext, true
	default:
		return "", false
	}
}
