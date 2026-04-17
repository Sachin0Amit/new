package sensory

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"github.com/papi-ai/sovereign-core/internal/models"
	"github.com/papi-ai/sovereign-core/pkg/errors"
)

// VisionProcessor handles the local transformation of visual data into inference tensors.
type VisionProcessor struct {
	TargetWidth  int
	TargetHeight int
}

// NewVisionProcessor creates a processor with standard input dimensions for the Titan core.
func NewVisionProcessor(width, height int) *VisionProcessor {
	return &VisionProcessor{
		TargetWidth:  width,
		TargetHeight: height,
	}
}

// ProcessImage decodes and normalizes a raw image buffer into a VisualFrame.
func (v *VisionProcessor) ProcessImage(raw []byte) (*models.VisualFrame, error) {
	img, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, errors.New(errors.CodeValidation, "failed to decode image buffer", err)
	}

	bounds := img.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	
	// Normalize to TargetWidth x TargetHeight x 3 (RGB)
	// For this prototype, we'll perform a basic 1:1 mapping if already sized, 
	// or center-crop/resize logic could go here.
	data := make([]float32, v.TargetWidth*v.TargetHeight*3)
	
	for y := 0; y < v.TargetHeight; y++ {
		for x := 0; x < v.TargetWidth; x++ {
			// Basic sampling (nearest neighbor simplification for V1)
			srcX := x * width / v.TargetWidth
			srcY := y * height / v.TargetHeight
			
			r, g, b, _ := img.At(srcX, srcY).RGBA()
			
			pixelIdx := (y*v.TargetWidth + x) * 3
			data[pixelIdx] = float32(r>>8) / 255.0
			data[pixelIdx+1] = float32(g>>8) / 255.0
			data[pixelIdx+2] = float32(b>>8) / 255.0
		}
	}

	return &models.VisualFrame{
		Width:    v.TargetWidth,
		Height:   v.TargetHeight,
		Channels: 3,
		Data:     data,
	}, nil
}

// StreamToBuffer reads a sensory stream into a memory buffer.
func StreamToBuffer(r io.Reader) ([]byte, error) {
	return io.ReadAll(r)
}
