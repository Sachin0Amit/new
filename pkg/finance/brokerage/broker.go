package brokerage

import (
	"fmt"

	"github.com/papi-ai/sovereign-core/pkg/finance"
)

// BrokerType identifies the brokerage platform.
type BrokerType string

const (
	BrokerMock    BrokerType = "MOCK"
	BrokerZerodha BrokerType = "ZERODHA"
	BrokerGroww   BrokerType = "GROWW" // Note: Official API generally not available, kept for UI consistency
	BrokerBinance BrokerType = "BINANCE"
)

// OrderReceipt represents the confirmation of an executed trade.
type OrderReceipt struct {
	OrderID   string
	Symbol    string
	Action    finance.TradeAction
	Quantity  float64
	Price     float64
	Timestamp int64
	Status    string // "COMPLETED", "REJECTED", "PENDING"
	Message   string
}

// BrokerAdapter defines the standard interface for interacting with any brokerage.
type BrokerAdapter interface {
	// Name returns the broker's identifier.
	Name() string

	// Authenticate validates credentials and establishes a session.
	Authenticate(apiKey, secret string) error

	// IsAuthenticated returns true if a valid session exists.
	IsAuthenticated() bool

	// GetMarketData fetches recent OHLCV data for a symbol.
	GetMarketData(symbol string) (finance.TimeSeries, error)

	// ExecuteOrder places a market order.
	ExecuteOrder(symbol string, action finance.TradeAction, qty float64) (*OrderReceipt, error)

	// GetPortfolio fetches current cash balance and positions.
	GetPortfolio() (*finance.Portfolio, error)
}

// Factory function to create the appropriate broker adapter.
func NewBroker(brokerType BrokerType) (BrokerAdapter, error) {
	switch brokerType {
	case BrokerMock:
		return NewMockBroker(), nil
	case BrokerZerodha:
		return NewZerodhaBroker(), nil
	default:
		return nil, fmt.Errorf("unsupported broker type: %s", brokerType)
	}
}
