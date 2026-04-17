package sensory

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/papi-ai/sovereign-core/pkg/errors"
)

// AudioPulse handles the parsing and normalization of auditory waveforms.
type AudioPulse struct {
	SampleRate int
	Channels   int
}

// NormalizedAudio represents a PCM stream ready for cognitive analysis.
type NormalizedAudio struct {
	Samples []float32
}

// NewAudioPulse creates a listener with standard sampling configuration.
func NewAudioPulse(sampleRate, channels int) *AudioPulse {
	return &AudioPulse{
		SampleRate: sampleRate,
		Channels:   channels,
	}
}

// ParseWAV reads a raw WAV/PCM stream and normalizes the signal to float32.
func (a *AudioPulse) ParseWAV(r io.Reader) (*NormalizedAudio, error) {
	// Simple WAV Header Skip / Minimal Parser for prototype
	// In production, this would handle the full RIFF header specification.
	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	if len(buf) < 44 {
		return nil, errors.New(errors.CodeValidation, "invalid WAV buffer - too short", nil)
	}

	// For PCM 16-bit, samples start after the 44-byte header
	data := buf[44:]
	sampleCount := len(data) / 2
	samples := make([]float32, sampleCount)

	for i := 0; i < sampleCount; i++ {
		bits := binary.LittleEndian.Uint16(data[i*2 : i*2+2])
		val := int16(bits)
		samples[i] = float32(val) / 32768.0 // Normalize to [-1, 1]
	}

	return &NormalizedAudio{Samples: samples}, nil
}

// Spectrogram (simplified stub) would generate frequency-domain features.
func (a *AudioPulse) Spectrogram(audio *NormalizedAudio) ([][]float32, error) {
	fmt.Println("FFT Transformation active")
	return nil, nil // Advanced logic for subsequent phases
}
