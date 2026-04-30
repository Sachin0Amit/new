package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ConversationMessage represents a single message in an exported conversation.
type ConversationMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// ConversationExport is the full export payload.
type ConversationExport struct {
	SessionID   string                `json:"session_id"`
	Title       string                `json:"title"`
	Messages    []ConversationMessage `json:"messages"`
	CreatedAt   string                `json:"created_at"`
	ExportedAt  string                `json:"exported_at"`
	MessageCount int                  `json:"message_count"`
	Format      string                `json:"format"` // "json" or "markdown"
}

// HandleExport handles conversation export requests.
// POST /api/v1/export - exports a conversation as JSON or Markdown.
func (h *Handler) HandleExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "only POST is accepted")
		return
	}

	var req struct {
		SessionID string                `json:"session_id"`
		Messages  []ConversationMessage `json:"messages"`
		Title     string                `json:"title"`
		Format    string                `json:"format"` // "json" or "markdown"
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_PAYLOAD", "failed to parse request")
		return
	}

	if len(req.Messages) == 0 {
		writeError(w, http.StatusBadRequest, "EMPTY_CONVERSATION", "no messages to export")
		return
	}

	if req.Format == "" {
		req.Format = "json"
	}

	export := ConversationExport{
		SessionID:    req.SessionID,
		Title:        req.Title,
		Messages:     req.Messages,
		CreatedAt:    time.Now().Format(time.RFC3339),
		ExportedAt:   time.Now().Format(time.RFC3339),
		MessageCount: len(req.Messages),
		Format:       req.Format,
	}

	if req.Format == "markdown" {
		md := fmt.Sprintf("# %s\n\n_Exported: %s_\n\n---\n\n", export.Title, export.ExportedAt)
		for _, msg := range export.Messages {
			role := "**User**"
			if msg.Role == "assistant" {
				role = "**Sovereign**"
			}
			md += fmt.Sprintf("### %s\n%s\n\n", role, msg.Content)
		}
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sovereign-chat-%s.md\"", req.SessionID))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(md))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"sovereign-chat-%s.json\"", req.SessionID))
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(export)
}
