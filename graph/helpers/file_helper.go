package helpers

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProcessBase64Image processes a base64 encoded image string, saves it to disk, and returns the public URL.
// It handles prefix removal, decoding, extension detection, and file saving.
// Returns the URL path for the saved image.
func ProcessBase64Image(base64Data string, entityID int, uploadDir string) (string, error) {
	// Ensure upload directory exists
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Remove data:image/...;base64, prefix if present
	if idx := strings.Index(base64Data, "base64,"); idx != -1 {
		base64Data = base64Data[idx+7:] // +7 to skip "base64,"
	}

	// Decode base64 string to bytes
	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("invalid base64 data: %w", err)
	}

	// Determine file extension from content type or default to png
	extension := "png"
	if strings.Contains(base64Data, "image/jpeg") || strings.Contains(base64Data, "image/jpg") {
		extension = "jpg"
	} else if strings.Contains(base64Data, "image/gif") {
		extension = "gif"
	}

	// Create unique filename and path
	filename := fmt.Sprintf("%d_%d.%s", entityID, time.Now().Unix(), extension)
	filePath := filepath.Join(uploadDir, filename)

	// Save file to disk
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to save file: %w", err)
	}

	// Build URL (adjust static serving accordingly)
	fileURL := fmt.Sprintf("/uploads/%s/%s", filepath.Base(uploadDir), filename)

	return fileURL, nil
}

// isLikelyBase64Image returns true if the string looks like a base64-encoded image (data URI)
func IsLikelyBase64Image(s string) bool {
	ss := strings.TrimSpace(s)
	if ss == "" {
		return false
	}
	// Common data URI prefix and marker
	if strings.HasPrefix(ss, "data:image/") && strings.Contains(ss, "base64,") {
		return true
	}
	// Fallback heuristic
	if strings.Contains(ss, "base64,") {
		return true
	}
	return false
}
