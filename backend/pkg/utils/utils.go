// Package utils provides small shared helpers: JSON responses and image
// upload handling used across the HTTP handlers.
package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

// JSON writes v as a JSON response with the given status code.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

// Error writes a JSON error envelope: {"error": "..."}.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}

// allowedImageTypes maps accepted MIME types to file extensions.
var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/gif":  ".gif",
}

// ErrUnsupportedImage is returned when an upload is not JPEG/PNG/GIF.
var ErrUnsupportedImage = errors.New("unsupported image type (allowed: jpeg, png, gif)")

// SaveImage validates and stores an uploaded image file, returning the public
// path (e.g. "/uploads/<uuid>.png") to store in the database. Returns ("", nil)
// when no file was provided.
func SaveImage(file multipart.File, header *multipart.FileHeader, uploadDir string) (string, error) {
	if file == nil {
		return "", nil
	}
	// Sniff the content type from the first 512 bytes.
	buf := make([]byte, 512)
	n, _ := file.Read(buf)
	contentType := http.DetectContentType(buf[:n])
	ext, ok := allowedImageTypes[contentType]
	if !ok {
		return "", ErrUnsupportedImage
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		return "", err
	}
	name := uuid.NewString() + ext
	dstPath := filepath.Join(uploadDir, name)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

// FormFileOptional returns the uploaded file for the given form field, or
// (nil, nil, nil) when the field is absent.
func FormFileOptional(r *http.Request, field string) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := r.FormFile(field)
	if errors.Is(err, http.ErrMissingFile) {
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, fmt.Errorf("read upload: %w", err)
	}
	return file, header, nil
}
