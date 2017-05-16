package application

import (
	"crypto"

	"github.com/gosexy/checksum"
)

// Round rounds a float to the nearest whole number
func Round(val float64) int {
	if val < 0 {
		return int(val - 0.5)
	}
	return int(val + 0.5)
}

// ComputeFileMd5 computes the hash of the given file contents
func ComputeFileMd5(filename string) string {
	return checksum.File(filename, crypto.SHA256)
}

// ComputeMd5 computes the has of the given string contents
func ComputeMd5(string string) string {
	return checksum.String(string, crypto.MD5)
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
