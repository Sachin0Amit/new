package sensory

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVisionNormalization(t *testing.T) {
	// 1. Create a dummy 100x100 Red image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	buf := new(bytes.Buffer)
	png.Encode(buf, img)

	// 2. Process with VisionProcessor
	processor := NewVisionProcessor(224, 224)
	frame, err := processor.ProcessImage(buf.Bytes())

	assert.NoError(t, err)
	assert.Equal(t, 224, frame.Width)
	assert.Equal(t, 224, frame.Height)
	
	// First pixel should be roughly [1.0, 0.0, 0.0]
	assert.InDelta(t, 1.0, frame.Data[0], 0.01)
	assert.InDelta(t, 0.0, frame.Data[1], 0.01)
	assert.InDelta(t, 0.0, frame.Data[2], 0.01)
}

func TestAudioParsing(t *testing.T) {
	pulse := NewAudioPulse(16000, 1)

	// Create a fake 16-bit PCM WAV (minimal header + 4 bytes of silent data)
	fakeWav := make([]byte, 48) 
	copy(fakeWav[:4], "RIFF")
	// Index 44+ is the data
	fakeWav[44] = 0x00
	fakeWav[45] = 0x00

	audio, err := pulse.ParseWAV(bytes.NewReader(fakeWav))
	assert.NoError(t, err)
	assert.Greater(t, len(audio.Samples), 0)
}
