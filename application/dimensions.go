package application

import (
	"strconv"
	"strings"
)

// Dimensions holds the Width and Height and a display String
type Dimensions struct {
	String string
	Width  int
	Height int
}

// NewDimensions returns a new Dimensions struct with the Width and Height parsed
func NewDimensions(dimensions string) *Dimensions {
	parts := strings.Split(dimensions, "x")
	dw, _ := strconv.Atoi(parts[0])
	dh, _ := strconv.Atoi(parts[1])

	return &Dimensions{
		String: dimensions,
		Width:  dw,
		Height: dh,
	}
}
