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

type VehiclePhotoStorage struct {
	rootDir     string
	vehiclesDir string
}

func NewVehiclePhotoStorage(rootDir string) *VehiclePhotoStorage {
	return &VehiclePhotoStorage{
		rootDir:     rootDir,
		vehiclesDir: filepath.Join(rootDir, "vehicles"),
	}
}

func (s *VehiclePhotoStorage) EnsureDirs() error {
	return os.MkdirAll(s.vehiclesDir, 0o755)
}

func (s *VehiclePhotoStorage) Save(vehicleID int64, fh *multipart.FileHeader, oldRelativePath string) (string, error) {
	if fh == nil {
		return "", errors.New("photo file is required")
	}

	if err := s.EnsureDirs(); err != nil {
		return "", fmt.Errorf("create vehicle photo dir: %w", err)
	}

	if fh.Size > 10*1024*1024 {
		return "", errors.New("photo size must be <= 10MB")
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
	ext, ok := vehiclePhotoExtensionByContentType(contentType)
	if !ok {
		return "", errors.New("only jpg, jpeg, png and webp images are allowed")
	}

	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("reset uploaded file reader: %w", err)
	}

	filename := fmt.Sprintf("vehicle_%d_%d%s", vehicleID, time.Now().UnixNano(), ext)
	absPath := filepath.Join(s.vehiclesDir, filename)

	dst, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("create destination file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("save uploaded photo: %w", err)
	}

	newRelativePath := filepath.ToSlash(filepath.Join("vehicles", filename))

	if strings.TrimSpace(oldRelativePath) != "" {
		_ = s.Delete(oldRelativePath)
	}

	return newRelativePath, nil
}

func (s *VehiclePhotoStorage) Delete(relativePath string) error {
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

func vehiclePhotoExtensionByContentType(contentType string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return ".jpg", true
	case "image/png":
		return ".png", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
