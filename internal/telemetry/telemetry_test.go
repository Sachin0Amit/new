package telemetry

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/time/rate"
)

func TestTelemetryConcurrecy(t *testing.T) {
	hub := NewHub("test-node")
	go hub.Run()

	server := httptest.NewServer(&Server{
		Hub:     hub,
		Limiter: rate.NewLimiter(rate.Limit(100), 100),
	})
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	const numClients = 50
	const numMessages = 1000

	var wg sync.WaitGroup
	wg.Add(numClients)

	clients := make([]*websocket.Conn, numClients)
	
	// Connect clients
	for i := 0; i < numClients; i++ {
		go func(idx int) {
			defer wg.Done()
			c, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
			if err != nil {
				t.Errorf("Failed to connect client %d: %v", idx, err)
				return
			}
			clients[idx] = c
		}(i)
	}
	wg.Wait()

	// Start reading for each client
	messageCounts := make([]int, numClients)
	var countMu sync.Mutex
	
	for i := 0; i < numClients; i++ {
		go func(idx int) {
			for {
				_, _, err := clients[idx].ReadMessage()
				if err != nil {
					return
				}
				countMu.Lock()
				messageCounts[idx]++
				countMu.Unlock()
			}
		}(i)
	}

	// Broadcast messages
	for i := 0; i < numMessages; i++ {
		time.Sleep(1 * time.Millisecond); hub.Broadcast([]byte(fmt.Sprintf("message-%d", i)))
	}

	// Give some time for delivery
	time.Sleep(2 * time.Second)

	// Verify counts
	countMu.Lock()
	defer countMu.Unlock()
	for i, count := range messageCounts {
		if count != numMessages {
			t.Errorf("Client %d received %d messages, want %d", i, count, numMessages)
		}
	}
}
