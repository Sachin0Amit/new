package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Sachin0Amit/new/internal/agent"
	"github.com/Sachin0Amit/new/internal/llm"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/Sachin0Amit/new/pkg/logger"
)

// WebSocketHandler manages WebSocket chat connections
type WebSocketHandler struct {
	llmClient       llm.Client
	agent           *agent.ReActAgent
	logger          logger.Logger
	upgrader        websocket.Upgrader
	sessions        map[string]*ChatSession
	sessionsMu      sync.RWMutex
}

// ChatSession represents an active WebSocket session
type ChatSession struct {
	ID        string
	UserID    string
	CreatedAt time.Time
	LastMsg   time.Time
	Conn      *websocket.Conn
	mu        sync.Mutex
}

// StreamMessage represents a message sent over WebSocket
type StreamMessage struct {
	Type      string      `json:"type"`                       // message, thought, action, observation, done, error
	Content   string      `json:"content"`
	Delta     string      `json:"delta,omitempty"`            // For streaming
	Timestamp int64       `json:"timestamp"`
	SessionID string      `json:"session_id,omitempty"`
	MessageID string      `json:"message_id,omitempty"`
	Step      *StepEvent  `json:"step,omitempty"`
	Error     string      `json:"error,omitempty"`
}

// StepEvent represents a single reasoning step
type StepEvent struct {
	Number      int    `json:"number"`
	Thought     string `json:"thought"`
	Action      string `json:"action,omitempty"`
	Observation string `json:"observation,omitempty"`
	Error       string `json:"error,omitempty"`
	Duration    int64  `json:"duration_ms"`
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler(llmClient llm.Client, agent *agent.ReActAgent, logger logger.Logger) *WebSocketHandler {
	return &WebSocketHandler{
		llmClient: llmClient,
		agent:     agent,
		logger:    logger,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// Allow all origins for now (should be restricted in production)
				return true
			},
		},
		sessions: make(map[string]*ChatSession),
	}
}

// HandleWebSocket handles WebSocket connections
func (wh *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := wh.upgrader.Upgrade(w, r, nil)
	if err != nil {
		wh.logger.Error("WebSocket upgrade failed", logger.ErrorF(err))
		return
	}
	defer conn.Close()

	sessionID := uuid.New().String()
	session := &ChatSession{
		ID:        sessionID,
		UserID:    r.Header.Get("X-User-ID"),
		CreatedAt: time.Now(),
		Conn:      conn,
	}

	wh.registerSession(session)
	defer wh.unregisterSession(sessionID)

	wh.logger.Info("WebSocket session opened", logger.String("session_id", sessionID))

	// Send session started message
	wh.sendMessage(conn, StreamMessage{
		Type:      "session_started",
		SessionID: sessionID,
		Timestamp: time.Now().UnixMilli(),
	})

	// Read messages from client
	for {
		var req ChatRequest
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				wh.logger.Error("WebSocket error", logger.ErrorF(err))
			}
			break
		}

		session.LastMsg = time.Now()

		// Handle the message
		wh.handleChatMessage(context.Background(), session, req)
	}
}

// handleChatMessage processes a chat message and streams response
func (wh *WebSocketHandler) handleChatMessage(ctx context.Context, session *ChatSession, req ChatRequest) {
	messageID := uuid.New().String()

	wh.logger.Info("Processing chat message",
		logger.String("session_id", session.ID),
		logger.String("message_id", messageID),
		logger.String("message_preview", truncate(req.Message, 80)),
	)

	// Process with agent if available, otherwise use simple LLM
	if wh.agent != nil {
		wh.handleWithAgent(ctx, session, req, messageID)
	} else {
		wh.handleWithLLM(ctx, session, req, messageID)
	}
}

// handleWithAgent processes the message using the ReAct agent
func (wh *WebSocketHandler) handleWithAgent(ctx context.Context, session *ChatSession, req ChatRequest, msgID string) {
	// Execute the ReAct loop
	result, err := wh.agent.Reason(ctx, req.Message, func(step *agent.Step) {
		// Stream each step
		wh.sendMessage(session.Conn, StreamMessage{
			Type:        "step",
			SessionID:   session.ID,
			MessageID:   msgID,
			Timestamp:   time.Now().UnixMilli(),
			Step: &StepEvent{
				Number:      step.Number,
				Thought:     step.Thought,
				Action:      fmt.Sprintf("%v", step.Action),
				Observation: step.Observation,
				Duration:    step.Duration.Milliseconds(),
			},
		})
	})

	if err != nil {
		wh.sendMessage(session.Conn, StreamMessage{
			Type:      "error",
			Error:     err.Error(),
			SessionID: session.ID,
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// Send final response
	wh.sendMessage(session.Conn, StreamMessage{
		Type:      "message",
		Content:   result.FinalResponse,
		SessionID: session.ID,
		MessageID: msgID,
		Timestamp: time.Now().UnixMilli(),
	})

	// Send completion
	wh.sendMessage(session.Conn, StreamMessage{
		Type:      "done",
		SessionID: session.ID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// handleWithLLM processes the message using the LLM with streaming
func (wh *WebSocketHandler) handleWithLLM(ctx context.Context, session *ChatSession, req ChatRequest, msgID string) {
	// Prepare the request
	messages := []llm.Message{
		{
			Role:    llm.RoleUser,
			Content: req.Message,
		},
	}

	llmReq := &llm.CompletionRequest{
		Model:       wh.llmClient.GetModel(),
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   2000,
		Stream:      true,
	}

	// Stream the response
	chunks, errors, err := wh.llmClient.Stream(ctx, llmReq)
	if err != nil {
		wh.sendMessage(session.Conn, StreamMessage{
			Type:      "error",
			Error:     fmt.Sprintf("Stream failed: %v", err),
			SessionID: session.ID,
			Timestamp: time.Now().UnixMilli(),
		})
		return
	}

	// Collect full response
	fullResponse := ""

	// Send chunks as they arrive
	for chunk := range chunks {
		fullResponse += chunk.Delta

		wh.sendMessage(session.Conn, StreamMessage{
			Type:      "chunk",
			Delta:     chunk.Delta,
			SessionID: session.ID,
			MessageID: msgID,
			Timestamp: time.Now().UnixMilli(),
		})
	}

	// Check for errors
	select {
	case err := <-errors:
		if err != nil {
			wh.sendMessage(session.Conn, StreamMessage{
				Type:      "error",
				Error:     err.Error(),
				SessionID: session.ID,
				Timestamp: time.Now().UnixMilli(),
			})
		}
	default:
	}

	// Send final message
	wh.sendMessage(session.Conn, StreamMessage{
		Type:      "message",
		Content:   fullResponse,
		SessionID: session.ID,
		MessageID: msgID,
		Timestamp: time.Now().UnixMilli(),
	})

	// Send completion
	wh.sendMessage(session.Conn, StreamMessage{
		Type:      "done",
		SessionID: session.ID,
		Timestamp: time.Now().UnixMilli(),
	})
}

// sendMessage sends a message on the WebSocket
func (wh *WebSocketHandler) sendMessage(conn *websocket.Conn, msg StreamMessage) error {
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return conn.WriteJSON(msg)
}

// registerSession registers a session
func (wh *WebSocketHandler) registerSession(session *ChatSession) {
	wh.sessionsMu.Lock()
	defer wh.sessionsMu.Unlock()
	wh.sessions[session.ID] = session
}

// unregisterSession unregisters a session
func (wh *WebSocketHandler) unregisterSession(sessionID string) {
	wh.sessionsMu.Lock()
	defer wh.sessionsMu.Unlock()
	delete(wh.sessions, sessionID)
}

// BroadcastMessage broadcasts a message to all sessions
func (wh *WebSocketHandler) BroadcastMessage(msg StreamMessage) {
	wh.sessionsMu.RLock()
	defer wh.sessionsMu.RUnlock()

	for _, session := range wh.sessions {
		go wh.sendMessage(session.Conn, msg)
	}
}

// GetSessionCount returns the number of active sessions
func (wh *WebSocketHandler) GetSessionCount() int {
	wh.sessionsMu.RLock()
	defer wh.sessionsMu.RUnlock()
	return len(wh.sessions)
}
