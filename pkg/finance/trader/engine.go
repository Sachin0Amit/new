package trader

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Sachin0Amit/new/pkg/finance"
	"github.com/Sachin0Amit/new/pkg/finance/brokerage"
	"github.com/Sachin0Amit/new/pkg/finance/news"
)

// SovereignTrader is the autonomous trading engine that bridges the
// C++ Quant Engine and the Brokerage Layer.
type SovereignTrader struct {
	mu             sync.Mutex
	isRunning      bool
	broker         brokerage.BrokerAdapter
	aggregator     *finance.SignalAggregator
	cache          *finance.MarketDataCache
	newsAgg        *news.Aggregator
	tradeLog       []brokerage.OrderReceipt
	symbolsToTrade []string
	pollInterval   time.Duration
	stopChan       chan struct{}

	// Risk limits
	minConfidence float64 // e.g. 0.85
	maxPositionPct float64
}

func NewSovereignTrader(broker brokerage.BrokerAdapter, cache *finance.MarketDataCache, aggregator *finance.SignalAggregator, newsAgg *news.Aggregator) *SovereignTrader {
	return &SovereignTrader{
		broker:         broker,
		aggregator:     aggregator,
		cache:          cache,
		newsAgg:        newsAgg,
		tradeLog:       make([]brokerage.OrderReceipt, 0),
		symbolsToTrade: []string{"AAPL", "MSFT", "BTC", "ETH"},
		pollInterval:   1 * time.Minute, // Poll every minute
		minConfidence:  0.80,            // 80% confidence required to trade
		maxPositionPct: 0.10,            // Max 10% of portfolio per trade
	}
}

// Start begins the autonomous trading loop.
func (t *SovereignTrader) Start() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.isRunning {
		return fmt.Errorf("trader is already running")
	}

	if !t.broker.IsAuthenticated() {
		return fmt.Errorf("broker is not authenticated. please login first")
	}

	t.isRunning = true
	t.stopChan = make(chan struct{})

	go t.tradingLoop()

	log.Printf("[TRADER] Sovereign Autonomous Engine started on %s", t.broker.Name())
	return nil
}

// Stop halts the autonomous trading loop.
func (t *SovereignTrader) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isRunning {
		return
	}

	close(t.stopChan)
	t.isRunning = false
	log.Printf("[TRADER] Sovereign Autonomous Engine stopped")
}

// IsRunning returns the current status.
func (t *SovereignTrader) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.isRunning
}

// GetTradeLogs returns recent trades.
func (t *SovereignTrader) GetTradeLogs() []brokerage.OrderReceipt {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.tradeLog
}

func (t *SovereignTrader) SetSymbols(symbols []string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.symbolsToTrade = symbols
}

// tradingLoop is the core autonomous function.
func (t *SovereignTrader) tradingLoop() {
	ticker := time.NewTicker(t.pollInterval)
	defer ticker.Stop()

	// Run immediately once
	t.executeCycle()

	for {
		select {
		case <-t.stopChan:
			return
		case <-ticker.C:
			t.executeCycle()
		}
	}
}

func (t *SovereignTrader) executeCycle() {
	log.Printf("[TRADER] Initiating evaluation cycle...")

	portfolio, err := t.broker.GetPortfolio()
	if err != nil {
		log.Printf("[TRADER] Error fetching portfolio: %v", err)
		return
	}

	t.mu.Lock()
	symbols := make([]string, len(t.symbolsToTrade))
	copy(symbols, t.symbolsToTrade)
	t.mu.Unlock()

	for _, symbol := range symbols {
		// 1. Pull market data from broker
		ts, err := t.broker.GetMarketData(symbol)
		if err != nil {
			log.Printf("[TRADER] Data error for %s: %v", symbol, err)
			continue
		}

		// 2. Feed to C++ engine / cache
		t.cache.Put(symbol, finance.TFDaily, ts)

		// 3. Evaluate Fundamentals
		fundamentals, err := finance.AnalyzeFundamentals(symbol)
		if err != nil {
			log.Printf("[TRADER] Fundamental analysis failed for %s: %v", symbol, err)
			continue
		}

		// 4. Evaluate News Sentiment
		articles := t.newsAgg.GetNewsForSymbol(symbol)
		var totalSentiment float64
		var validArticles int
		for _, article := range articles {
			sentiment, err := finance.AnalyzeSentiment(article.Title)
			if err == nil {
				totalSentiment += sentiment.Score
				validArticles++
			}
		}
		
		avgSentiment := 0.0
		if validArticles > 0 {
			avgSentiment = totalSentiment / float64(validArticles)
		}

		// 5. Get Technical Ensemble Signal
		signal := t.aggregator.GenerateSignal(symbol)
		if signal == nil {
			continue
		}

		log.Printf("[TRADER] %s | Tech: %v (%.2f) | Fund: %s (%.0f) | Sent: %.2f", 
			symbol, signal.Type, signal.Confidence, fundamentals.Rating, fundamentals.TotalScore, avgSentiment)

		// 6. Execution Logic (Combined Intelligence)
		// Only trade if Technicals, Fundamentals, and Sentiment align.
		if signal.Confidence >= t.minConfidence {
			action := finance.TradeAction("HOLD")
			
			// Bullish Alignment
			if signal.Type == finance.SignalBuy && fundamentals.TotalScore >= 50 && avgSentiment >= -0.2 {
				action = finance.ActionBuy
			} else if signal.Type == finance.SignalSell && fundamentals.TotalScore < 50 && avgSentiment <= 0.2 { // Bearish Alignment
				action = finance.ActionSell
			}

			if action != finance.TradeAction("HOLD") {
				// Calculate position size using Kelly or fixed % risk
				investAmount := portfolio.TotalValue * t.maxPositionPct
				qty := investAmount / signal.Price

				if qty > 0 {
					log.Printf("[TRADER] Executing %s %f %s at $%.2f", action, qty, symbol, signal.Price)
					receipt, err := t.broker.ExecuteOrder(symbol, action, qty)
					if err != nil {
						log.Printf("[TRADER] Execution Failed: %v", err)
						continue
					}

					t.mu.Lock()
					t.tradeLog = append(t.tradeLog, *receipt)
					// Keep log size manageable
					if len(t.tradeLog) > 100 {
						t.tradeLog = t.tradeLog[1:]
					}
					t.mu.Unlock()
					log.Printf("[TRADER] Execution Success: %s", receipt.OrderID)
				}
			}
		}
	}
}

// ManualOrder allows manual intervention.
func (t *SovereignTrader) ManualOrder(symbol string, action finance.TradeAction, qty float64) (*brokerage.OrderReceipt, error) {
	if !t.broker.IsAuthenticated() {
		return nil, fmt.Errorf("broker not authenticated")
	}
	
	receipt, err := t.broker.ExecuteOrder(symbol, action, qty)
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	t.tradeLog = append(t.tradeLog, *receipt)
	t.mu.Unlock()

	return receipt, nil
}

// ConnectBroker authenticates a new broker session.
func (t *SovereignTrader) ConnectBroker(apiKey, secret string) error {
	return t.broker.Authenticate(apiKey, secret)
}
