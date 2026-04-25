package testutil

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/Sachin0Amit/new/internal/auditor"
	"github.com/Sachin0Amit/new/internal/core"
	"github.com/Sachin0Amit/new/internal/guard"
	"github.com/Sachin0Amit/new/internal/mesh"
	"github.com/Sachin0Amit/new/internal/reflex"
	"github.com/Sachin0Amit/new/internal/titan"
)

// NewTestNode creates a fully wired Core with in-memory backends.
func NewTestNode(t *testing.T) *core.Core {
	// In-memory mesh
	tmpDir, _ := ioutil.TempDir("", "sov-test-*")
	m, err := mesh.NewKnowledgeMesh(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create test mesh: %v", err)
	}
	
	// Cleanup tmp dir on test completion
	t.Cleanup(func() {
		m.Close()
		os.RemoveAll(tmpDir)
	})

	// Initialize other modules with test defaults
	tn := titan.NewEngine("auto")
	g := guard.NewGuard("test_secret")
	a := auditor.NewAuditor()
	r := reflex.NewSelfHealer(100 * time.Millisecond)

	return core.New(tn, g, m, a, r)
}

// GoldenFile provides snapshot testing for large derivation outputs.
func GoldenFile(t *testing.T, name string, actual string) {
	t.Helper()
	goldenPath := filepath.Join("testdata", name+".golden")
	
	if os.Getenv("UPDATE_GOLDEN") == "true" {
		_ = os.MkdirAll(filepath.Dir(goldenPath), 0755)
		_ = ioutil.WriteFile(goldenPath, []byte(actual), 0644)
		return
	}

	expected, err := ioutil.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("Failed to read golden file: %v", err)
	}

	if string(expected) != actual {
		t.Errorf("Output does not match golden file %s.\nExpected: %s\nActual: %s", name, string(expected), actual)
	}
}

// WebSocketDialer provides a pre-configured dialer for integration tests.
func WebSocketDialer(t *testing.T, url string) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return conn
}
