package application

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// Round rounds a float to the nearest whole number
func Round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

// ComputeFileMd5 computes the hash of the given file contents
func ComputeFileMd5(filename string) (string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ExtensionForFormat determines the best guess extension for the given
// filename
func ExtensionForFormat(format string) string {
	switch format {
	case "JPEG":
		return ".jpg"
	case "PNG":
		return ".png"
	case "GIF":
		return ".gif"
	}
	return ""
}
