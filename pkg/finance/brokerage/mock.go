package brokerage

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/papi-ai/sovereign-core/pkg/finance"
)

// MockBroker implements BrokerAdapter for paper trading.
type MockBroker struct {
	mu            sync.Mutex
	authenticated bool
	portfolio     *finance.Portfolio
}

// NewMockBroker creates a new paper trading environment.
func NewMockBroker() *MockBroker {
	return &MockBroker{
		authenticated: false,
		portfolio:     finance.NewPortfolio("Paper Trading Account", 100000.0), // $100k starting capital
	}
}

func (m *MockBroker) Name() string {
	return "MOCK_BROKER"
}

func (m *MockBroker) Authenticate(apiKey, secret string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Mock authentication always succeeds
	m.authenticated = true
	return nil
}

func (m *MockBroker) IsAuthenticated() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.authenticated
}

func (m *MockBroker) GetMarketData(symbol string) (finance.TimeSeries, error) {
	if !m.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	// Generate realistic-looking mock data
	n := 100
	ts := make(finance.TimeSeries, n)
	
	price := 150.0
	if symbol == "BTC" {
		price = 60000.0
	} else if symbol == "MSFT" {
		price = 400.0
	}

	now := time.Now()
	for i := 0; i < n; i++ {
		change := 1.0 + (rand.Float64()*0.04 - 0.02)
		open := price
		close := price * change
		high := max(open, close) * (1.0 + rand.Float64()*0.01)
		low := min(open, close) * (1.0 - rand.Float64()*0.01)
		
		ts[i] = finance.OHLCV{
			Timestamp: now.AddDate(0, 0, i-n),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    1000000.0 + rand.Float64()*500000.0,
		}
		price = close
	}

	return ts, nil
}

func (m *MockBroker) ExecuteOrder(symbol string, action finance.TradeAction, qty float64) (*OrderReceipt, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.authenticated {
		return nil, fmt.Errorf("not authenticated")
	}

	// Simulate current market price
	price := 150.0 + rand.Float64()*5.0
	if symbol == "BTC" {
		price = 60000.0 + rand.Float64()*1000.0
	} else if symbol == "MSFT" {
		price = 400.0 + rand.Float64()*10.0
	}

	cost := qty * price

	if action == finance.ActionBuy {
		if m.portfolio.Cash < cost {
			return &OrderReceipt{
				OrderID:   uuid.New().String(),
				Symbol:    symbol,
				Action:    action,
				Quantity:  qty,
				Price:     price,
				Timestamp: time.Now().Unix(),
				Status:    "REJECTED",
				Message:   "Insufficient funds",
			}, nil
		}
		m.portfolio.AddPosition(symbol, qty, price)
	} else if action == finance.ActionSell {
		// Just execute the sell (mock broker allows short selling for simplicity, or we can check inventory)
		m.portfolio.AddPosition(symbol, -qty, price) // A negative qty denotes selling
	}

	return &OrderReceipt{
		OrderID:   uuid.New().String(),
		Symbol:    symbol,
		Action:    action,
		Quantity:  qty,
		Price:     price,
		Timestamp: time.Now().Unix(),
		Status:    "COMPLETED",
		Message:   "Order executed in paper trading",
	}, nil
}

func (m *MockBroker) GetPortfolio() (*finance.Portfolio, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if !m.authenticated {
		return nil, fmt.Errorf("not authenticated")
	}

	// Update mock prices
	prices := map[string]float64{
		"AAPL": 150.0 + rand.Float64()*5.0,
		"MSFT": 400.0 + rand.Float64()*10.0,
		"BTC":  60000.0 + rand.Float64()*1000.0,
	}
	m.portfolio.UpdatePrices(prices)

	return m.portfolio, nil
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
