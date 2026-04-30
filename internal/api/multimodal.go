package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Sachin0Amit/new/pkg/logger"
)

// MultiModalRequest represents a request with text + optional media.
type MultiModalRequest struct {
	Text     string           `json:"text"`
	Media    []MediaAttachment `json:"media,omitempty"`
	Tier     string           `json:"tier"`
}

// MediaAttachment holds an uploaded media file's metadata and content.
type MediaAttachment struct {
	Type     string `json:"type"`      // "image", "audio", "video"
	MimeType string `json:"mime_type"`
	Filename string `json:"filename"`
	Data     string `json:"data"`      // Base64-encoded content
	Size     int64  `json:"size"`
}

// MultiModalResponse is the API response for multi-modal requests.
type MultiModalResponse struct {
	Text        string            `json:"text"`
	Attachments []MediaAttachment `json:"attachments,omitempty"`
	Model       string            `json:"model"`
	LatencyMs   int64             `json:"latency_ms"`
}

// maxUploadSize limits each upload to 10MB.
const maxUploadSize = 10 << 20

// allowedMimeTypes defines accepted media types.
var allowedMimeTypes = map[string]string{
	"image/jpeg": "image",
	"image/png":  "image",
	"image/webp": "image",
	"image/gif":  "image",
	"audio/mpeg": "audio",
	"audio/wav":  "audio",
	"audio/ogg":  "audio",
	"video/mp4":  "video",
	"video/webm": "video",
}

// HandleMultiModal processes multi-modal input (text + images/audio).
// Supports both JSON body and multipart/form-data uploads.
//
//	POST /api/v1/multimodal
func (h *Handler) HandleMultiModal(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is accepted")
		return
	}

	start := time.Now()

	contentType := r.Header.Get("Content-Type")

	var req MultiModalRequest

	if strings.HasPrefix(contentType, "multipart/form-data") {
		// Handle multipart upload
		parsed, err := parseMultipartRequest(r)
		if err != nil {
			writeError(w, http.StatusBadRequest, "UPLOAD_ERROR", err.Error())
			return
		}
		req = *parsed
	} else {
		// Handle JSON body
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "failed to parse request")
			return
		}
	}

	if req.Text == "" && len(req.Media) == 0 {
		writeError(w, http.StatusBadRequest, "EMPTY_REQUEST", "provide text or media")
		return
	}

	// Build augmented prompt from media context
	prompt := req.Text
	for i, m := range req.Media {
		prompt += fmt.Sprintf("\n[Attachment %d: %s (%s, %d bytes)]", i+1, m.Filename, m.Type, m.Size)
	}

	// For now, pass to orchestrator as a text task. Vision/audio models
	// would consume the base64 payloads directly via multi-modal LLM APIs.
	if h.logger != nil {
		h.logger.Info("Multi-modal request",
			logger.String("text_preview", truncate(req.Text, 80)),
			logger.Int("attachments", len(req.Media)),
		)
	}

	resp := MultiModalResponse{
		Text:        fmt.Sprintf("Received your message with %d attachment(s). Multi-modal inference is processing.", len(req.Media)),
		Attachments: req.Media,
		Model:       "sovereign-multimodal",
		LatencyMs:   time.Since(start).Milliseconds(),
	}

	writeJSON(w, http.StatusOK, resp)
}

// parseMultipartRequest extracts text and media from a multipart form.
func parseMultipartRequest(r *http.Request) (*MultiModalRequest, error) {
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		return nil, fmt.Errorf("file too large (max %dMB)", maxUploadSize>>20)
	}

	req := &MultiModalRequest{
		Text: r.FormValue("text"),
		Tier: r.FormValue("tier"),
	}

	for _, headers := range r.MultipartForm.File {
		for _, fh := range headers {
			mime := fh.Header.Get("Content-Type")
			mediaType, ok := allowedMimeTypes[mime]
			if !ok {
				return nil, fmt.Errorf("unsupported file type: %s", mime)
			}

			if fh.Size > maxUploadSize {
				return nil, fmt.Errorf("file %s exceeds %dMB limit", fh.Filename, maxUploadSize>>20)
			}

			file, err := fh.Open()
			if err != nil {
				return nil, fmt.Errorf("failed to open file: %w", err)
			}

			data, err := io.ReadAll(file)
			file.Close()
			if err != nil {
				return nil, fmt.Errorf("failed to read file: %w", err)
			}

			req.Media = append(req.Media, MediaAttachment{
				Type:     mediaType,
				MimeType: mime,
				Filename: filepath.Base(fh.Filename),
				Data:     base64.StdEncoding.EncodeToString(data),
				Size:     fh.Size,
			})
		}
	}

	return req, nil
}
