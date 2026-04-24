package api

import (
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/papi-ai/sovereign-core/pkg/finance"
	"github.com/papi-ai/sovereign-core/pkg/finance/brokerage"
	"github.com/papi-ai/sovereign-core/pkg/finance/news"
	"github.com/papi-ai/sovereign-core/pkg/finance/trader"
	"github.com/papi-ai/sovereign-core/pkg/logger"
)

// FinanceHandler manages the RESTful interaction for the Finance Engine.
type FinanceHandler struct {
	logger  logger.Logger
	cache   *finance.MarketDataCache
	trader  *trader.SovereignTrader
	agg     *finance.SignalAggregator
	newsAgg *news.Aggregator
}

// NewFinanceHandler creates a new FinanceHandler.
func NewFinanceHandler(l logger.Logger) *FinanceHandler {
	cache := finance.NewMarketDataCache()
	agg := finance.NewSignalAggregator(cache)
	
	newsAgg := news.NewAggregator()
	go func() {
		newsAgg.FetchNews()
		ticker := time.NewTicker(15 * time.Minute)
		for range ticker.C {
			newsAgg.FetchNews()
		}
	}()
	
	// Default to mock broker
	broker, _ := brokerage.NewBroker(brokerage.BrokerMock)
	sTrader := trader.NewSovereignTrader(broker, cache, agg, newsAgg)

	handler := &FinanceHandler{
		logger:  l,
		cache:   cache,
		trader:  sTrader,
		agg:     agg,
		newsAgg: newsAgg,
	}
	handler.seedMockData() // Seed some data for the local dashboard
	return handler
}

// HandleMarketData returns OHLCV data for a given symbol.
func (h *FinanceHandler) HandleMarketData(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "AAPL"
	}

	ts, ok := h.cache.Get(symbol, finance.TFDaily)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Market data not found for symbol")
		return
	}

	writeJSON(w, http.StatusOK, ts)
}

// HandleIndicators returns calculated technical indicators.
func (h *FinanceHandler) HandleIndicators(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "AAPL"
	}

	ts, ok := h.cache.Get(symbol, finance.TFDaily)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Market data not found for symbol")
		return
	}

	closes := ts.Closes()

	sma20 := finance.SMA(closes, 20)
	ema50 := finance.EMA(closes, 50)
	rsi14 := finance.RSI(closes, 14)
	macd, signal, hist := finance.MACD(closes, 12, 26, 9)
	upper, middle, lower := finance.BollingerBands(closes, 20, 2.0)

	// Build response with latest values (last 50 data points to avoid huge payloads)
	limit := 50
	if len(closes) < limit {
		limit = len(closes)
	}

	start := len(closes) - limit

	response := map[string]interface{}{
		"symbol": symbol,
		"timestamps": extractTimestamps(ts[start:]),
		"prices": closes[start:],
		"indicators": map[string]interface{}{
			"SMA_20": sma20[start:],
			"EMA_50": ema50[start:],
			"RSI_14": rsi14[start:],
			"MACD": map[string]interface{}{
				"macd":   macd[start:],
				"signal": signal[start:],
				"hist":   hist[start:],
			},
			"Bollinger": map[string]interface{}{
				"upper":  upper[start:],
				"middle": middle[start:],
				"lower":  lower[start:],
			},
		},
	}

	writeJSON(w, http.StatusOK, response)
}

// HandlePrediction returns predictive modeling results.
func (h *FinanceHandler) HandlePrediction(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "AAPL"
	}
	
	horizonStr := r.URL.Query().Get("horizon")
	horizon := 30
	if h, err := strconv.Atoi(horizonStr); err == nil && h > 0 {
		horizon = h
	}

	ts, ok := h.cache.Get(symbol, finance.TFDaily)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Market data not found for symbol")
		return
	}

	closes := ts.Closes()
	
	// Use ensemble prediction
	prediction := finance.EnsemblePredict(closes, horizon)
	if prediction == nil {
		writeError(w, http.StatusInternalServerError, "PREDICTION_FAILED", "Failed to generate prediction")
		return
	}
	
	// Add current price for reference
	response := map[string]interface{}{
		"symbol": symbol,
		"current_price": closes[len(closes)-1],
		"prediction": prediction,
	}

	writeJSON(w, http.StatusOK, response)
}

// HandleRisk returns risk metrics.
func (h *FinanceHandler) HandleRisk(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "AAPL"
	}

	ts, ok := h.cache.Get(symbol, finance.TFDaily)
	if !ok {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Market data not found")
		return
	}

	returns := ts.Returns()
	
	// Dummy benchmark returns (e.g. SPY) and equity curve for simulation
	benchReturns := make([]float64, len(returns))
	equityCurve := finance.CumulativeSum(returns) // Simplified equity curve
	for i := range equityCurve {
		equityCurve[i] = 10000.0 * math.Exp(equityCurve[i]) // Base 10k
	}

	// Fake SPY returns for Beta/Alpha
	for i := range benchReturns {
		benchReturns[i] = returns[i] * 0.8 + (rand.Float64() - 0.5) * 0.01 
	}

	risk := finance.CalculateRiskMetrics(equityCurve, returns, benchReturns, 0.04) // 4% risk-free rate

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol": symbol,
		"risk":   risk,
	})
}

// HandleIntelligence returns fundamentals and news sentiment.
func (h *FinanceHandler) HandleIntelligence(w http.ResponseWriter, r *http.Request) {
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		symbol = "AAPL"
	}

	fundamentals, err := finance.AnalyzeFundamentals(symbol)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "FUND_ERROR", err.Error())
		return
	}

	articles := h.newsAgg.GetNewsForSymbol(symbol)
	var scoredNews = make([]map[string]interface{}, 0)

	if len(articles) == 0 {
		// Mock news for demonstration when RSS fails
		articles = []news.Article{
			{Title: symbol + " announces groundbreaking AI integration", Description: "The company is pushing the boundaries of AI.", PubDate: time.Now()},
			{Title: "Analysts upgrade " + symbol + " to strong buy", Description: "Earnings beat expectations by 20%.", PubDate: time.Now().Add(-2 * time.Hour)},
			{Title: "Regulatory scrutiny increases for " + symbol, Description: "Government evaluating anti-trust concerns.", PubDate: time.Now().Add(-24 * time.Hour)},
		}
	}

	for _, article := range articles {
		score, err := finance.AnalyzeSentiment(article.Title)
		if err == nil {
			scoredNews = append(scoredNews, map[string]interface{}{
				"title":       article.Title,
				"description": article.Description,
				"pub_date":    article.PubDate.Format(time.RFC3339),
				"score":       score.Score,
				"sentiment":   score,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"symbol":       symbol,
		"fundamentals": fundamentals,
		"news":         scoredNews,
	})
}

// extractTimestamps gets ISO strings.
func extractTimestamps(ts finance.TimeSeries) []string {
	out := make([]string, len(ts))
	for i, b := range ts {
		out[i] = b.Timestamp.Format(time.RFC3339)
	}
	return out
}

// HandleTraderStatus returns the status of the trader engine and logs.
func (h *FinanceHandler) HandleTraderStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"is_running": h.trader.IsRunning(),
		"logs":       h.trader.GetTradeLogs(),
	}
	writeJSON(w, http.StatusOK, status)
}

// HandleTraderToggle turns the autonomous trader on/off.
func (h *FinanceHandler) HandleTraderToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Use POST")
		return
	}

	action := r.URL.Query().Get("action")
	if action == "start" {
		if err := h.trader.Start(); err != nil {
			writeError(w, http.StatusBadRequest, "TRADER_ERROR", err.Error())
			return
		}
	} else if action == "stop" {
		h.trader.Stop()
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "success", "is_running": strconv.FormatBool(h.trader.IsRunning())})
}

// HandleManualOrder places a manual trade.
func (h *FinanceHandler) HandleManualOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Use POST")
		return
	}

	symbol := r.FormValue("symbol")
	actionStr := r.FormValue("action")
	qtyStr := r.FormValue("qty")

	qty, err := strconv.ParseFloat(qtyStr, 64)
	if err != nil || qty <= 0 {
		writeError(w, http.StatusBadRequest, "INVALID_QTY", "Invalid quantity")
		return
	}

	action := finance.TradeAction(actionStr)
	if action != finance.ActionBuy && action != finance.ActionSell {
		writeError(w, http.StatusBadRequest, "INVALID_ACTION", "Action must be BUY or SELL")
		return
	}

	receipt, err := h.trader.ManualOrder(symbol, action, qty)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "EXECUTION_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":  "success",
		"receipt": receipt,
	})
}

// HandleBrokerLogin authenticates a broker session.
func (h *FinanceHandler) HandleBrokerLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "Use POST")
		return
	}

	apiKey := r.FormValue("api_key")
	apiSecret := r.FormValue("api_secret")
	// brokerType := r.FormValue("broker")

	// For MVP, we just connect the mock broker (which accepts anything)
	// In production, we'd initialize the correct broker based on brokerType
	if err := h.trader.ConnectBroker(apiKey, apiSecret); err != nil {
		writeError(w, http.StatusUnauthorized, "LOGIN_FAILED", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "authenticated"})
}

// seedMockData generates fake historical data for testing the local dashboard.
func (h *FinanceHandler) seedMockData() {
	symbols := []string{"AAPL", "MSFT", "BTC", "ETH"}
	
	for _, sym := range symbols {
		ts := make(finance.TimeSeries, 500)
		basePrice := 150.0
		if sym == "BTC" { basePrice = 40000.0 }
		if sym == "ETH" { basePrice = 2000.0 }
		
		volatility := 0.02
		if sym == "BTC" || sym == "ETH" { volatility = 0.05 }

		now := time.Now()
		currentPrice := basePrice

		// Generate backwards
		for i := 499; i >= 0; i-- {
			date := now.AddDate(0, 0, -(499-i))
			
			// Random walk
			change := 1.0 + (rand.NormFloat64() * volatility)
			open := currentPrice
			close := currentPrice * change
			high := math.Max(open, close) * (1.0 + rand.Float64()*volatility*0.5)
			low := math.Min(open, close) * (1.0 - rand.Float64()*volatility*0.5)
			
			ts[i] = finance.OHLCV{
				Timestamp: date,
				Open:      open,
				High:      high,
				Low:       low,
				Close:     close,
				Volume:    1000000 * rand.Float64(),
			}
			currentPrice = close
		}
		
		h.cache.Put(sym, finance.TFDaily, ts)
	}
}
