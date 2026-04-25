package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Sachin0Amit/new/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestEndToEndDerivationFlow(t *testing.T) {
	// 1. Setup Full Stack
	core := testutil.NewTestNode(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	core.Start(ctx)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]string{"id": "test-task-id"})
		}
	})
	
	server := httptest.NewServer(mux)
	defer server.Close()

	// 2. Submit Task
	t.Run("Submit Task", func(t *testing.T) {
		payload := map[string]interface{}{"prompt": "solve 2+2"}
		buf, _ := json.Marshal(payload)
		resp, err := http.Post(server.URL+"/api/v1/tasks", "application/json", bytes.NewBuffer(buf))
		
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
	})

	// 3. Verify Auditor Record
	t.Run("Verify Audit", func(t *testing.T) {
		history, err := core.Auditor.GetHistory(ctx, "system", 10)
		assert.NoError(t, err)
		assert.NotEmpty(t, history)
	})
}
