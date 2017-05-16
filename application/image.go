package application

import (
	"fmt"
	"mime"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/h2non/filetype.v0"
)

// Image holds all the information for an image
type Image struct {
	Key              string
	OriginalFilename string
	Filename         string
	Dimensions       *Dimensions
	Width            int
	Height           int
	Format           string
	Size             int64
	MimeType         string
	Extension        string
}

// Resize resizes an image to the given dimensions whilst preserving the aspect ratio
func (image *Image) Resize(dimensions *Dimensions, dest string) {
	fmt.Println("[Resize] ", image.Filename, dimensions)

	cmd := exec.Command("convert", image.Filename, "-auto-orient", "-resize", dimensions.String+">", "-quality", "80", "-strip", "-depth", "8", dest)
	fmt.Println("[Resize] execute:", cmd.Args)
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("[Resize] error: ", string(out[:]), err)
	}
}

// Fill crops and resizes an image to fill a space
func (image *Image) Fill(dimensions *Dimensions, dest string) {
	fmt.Println("[Fill] ", image.Filename, dimensions)

	cmd := exec.Command("convert", image.Filename, "-auto-orient", "-resize", dimensions.String+"^", "-gravity", "center", "-extent", dimensions.String, "-quality", "80", "-strip", "-depth", "8", dest)
	fmt.Println("[Fill] execute:", cmd.Args)
	out, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Println("[Fill] error: ", string(out[:]), err)
	}
}

// Identify determines the metadata for the image
func (image *Image) Identify() {
	if image.Dimensions != nil {
		return
	}

	image.Extension = filepath.Ext(image.OriginalFilename)
	image.MimeType = mime.TypeByExtension(image.Extension)

	// determine the size of the file
	file, err := os.Open(image.Filename)
	if err != nil {
		fmt.Println("[Identify] error:", err)
		return
	}
	defer file.Close()
	fi, err := file.Stat()
	if err != nil {
		fmt.Println("[Identify] error:", err)
		return
	}
	image.Size = fi.Size()

	// read first 261 bytes of file to check its type for image processing
	b1 := make([]byte, 261)
	_, err = file.Read(b1)
	if err != nil {
		fmt.Println("[Identify] error:", err)
		return
	}

	if filetype.IsImage(b1) {
		cmd := exec.Command("identify", "-format", "%m %wx%h %b", image.Filename)
		fmt.Println("[Identify] execute:", cmd.Args)
		out, err := cmd.CombinedOutput()

		data := string(out[:])

		if err != nil {
			fmt.Println("[Identify] error:", err)
			return
		}

		parts := strings.Split(data, " ")

		image.Dimensions = NewDimensions(parts[1])
		image.Width = image.Dimensions.Width
		image.Height = image.Dimensions.Height
		image.Format = parts[0]
		image.Extension = ExtensionForFormat(image.Format)
	}

	fmt.Println("[Identify] ", image)
}

// Optimize tries to optimize the image
func (image *Image) Optimize() {
	cmd := exec.Command("jpegoptim", image.Filename)
	out, err := cmd.CombinedOutput()

	fmt.Println("[Optimize]", string(out[:]))

	if err != nil {
		fmt.Println("[Optimize] jpegoptim error:", err)
	}
}
