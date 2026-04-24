package finance

import (
	"math"
	"sort"
)

// ============================================================================
// MARKET SCANNER / SCREENER
// ============================================================================

// Scanner scans a universe of instruments and ranks them by composite score.
type Scanner struct {
	cache *MarketDataCache
}

// NewScanner creates a scanner backed by a market data cache.
func NewScanner(cache *MarketDataCache) *Scanner {
	return &Scanner{cache: cache}
}

// ScanResult contains an enriched result for one symbol.
type ScanResult struct {
	Symbol         string             `json:"symbol"`
	Price          float64            `json:"price"`
	Change1D       float64            `json:"change_1d_pct"`
	Change5D       float64            `json:"change_5d_pct"`
	Change20D      float64            `json:"change_20d_pct"`
	RSI            float64            `json:"rsi"`
	MACDSignal     string             `json:"macd_signal"` // "BULLISH", "BEARISH", "NEUTRAL"
	BBPosition     string             `json:"bb_position"` // "ABOVE_UPPER", "BELOW_LOWER", "WITHIN"
	VolumeRatio    float64            `json:"volume_ratio"` // Today vs 20d avg
	Trend          string             `json:"trend"`        // "UPTREND", "DOWNTREND", "SIDEWAYS"
	CompositeScore float64            `json:"score"`        // 0-100
	Signal         SignalType         `json:"signal"`
}

// ScanAll scans all symbols in the cache and returns ranked results.
func (sc *Scanner) ScanAll() []ScanResult {
	symbols := sc.cache.Symbols()
	results := make([]ScanResult, 0, len(symbols))

	for _, sym := range symbols {
		ts, ok := sc.cache.Get(sym, TFDaily)
		if !ok || len(ts) < 50 {
			continue
		}

		result := sc.analyzeSymbol(sym, ts)
		if result != nil {
			results = append(results, *result)
		}
	}

	// Sort by composite score descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].CompositeScore > results[j].CompositeScore
	})

	return results
}

// ScanWithFilters applies filters and returns matching results.
func (sc *Scanner) ScanWithFilters(filters []ScreenerFilter) []ScanResult {
	all := sc.ScanAll()
	filtered := make([]ScanResult, 0)

	for _, r := range all {
		if matchesFilters(r, filters) {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

func (sc *Scanner) analyzeSymbol(symbol string, ts TimeSeries) *ScanResult {
	closes := ts.Closes()
	highs := ts.Highs()
	lows := ts.Lows()
	volumes := ts.Volumes()
	n := len(closes)

	if n < 50 {
		return nil
	}

	result := &ScanResult{
		Symbol: symbol,
		Price:  closes[n-1],
	}

	// Price changes
	if n >= 2 && closes[n-2] > 0 {
		result.Change1D = (closes[n-1] - closes[n-2]) / closes[n-2] * 100
	}
	if n >= 6 && closes[n-6] > 0 {
		result.Change5D = (closes[n-1] - closes[n-6]) / closes[n-6] * 100
	}
	if n >= 21 && closes[n-21] > 0 {
		result.Change20D = (closes[n-1] - closes[n-21]) / closes[n-21] * 100
	}

	// RSI
	rsi := RSI(closes, 14)
	if rsi != nil && !math.IsNaN(rsi[n-1]) {
		result.RSI = rsi[n-1]
	}

	// MACD signal
	_, _, hist := MACD(closes, 12, 26, 9)
	if hist != nil && n >= 2 && !math.IsNaN(hist[n-1]) && !math.IsNaN(hist[n-2]) {
		if hist[n-1] > 0 && hist[n-2] <= 0 {
			result.MACDSignal = "BULLISH"
		} else if hist[n-1] < 0 && hist[n-2] >= 0 {
			result.MACDSignal = "BEARISH"
		} else {
			result.MACDSignal = "NEUTRAL"
		}
	}

	// Bollinger Bands position
	upper, _, lower := BollingerBands(closes, 20, 2.0)
	if upper != nil && !math.IsNaN(upper[n-1]) {
		if closes[n-1] >= upper[n-1] {
			result.BBPosition = "ABOVE_UPPER"
		} else if closes[n-1] <= lower[n-1] {
			result.BBPosition = "BELOW_LOWER"
		} else {
			result.BBPosition = "WITHIN"
		}
	}

	// Volume ratio (today vs 20d average)
	if n >= 21 {
		vol20d := Mean(volumes[n-20 : n])
		if vol20d > 0 {
			result.VolumeRatio = volumes[n-1] / vol20d
		}
	}

	// Trend detection using SMA alignment
	sma20 := SMA(closes, 20)
	sma50 := SMA(closes, 50)
	if sma20 != nil && sma50 != nil && !math.IsNaN(sma20[n-1]) && !math.IsNaN(sma50[n-1]) {
		if closes[n-1] > sma20[n-1] && sma20[n-1] > sma50[n-1] {
			result.Trend = "UPTREND"
		} else if closes[n-1] < sma20[n-1] && sma20[n-1] < sma50[n-1] {
			result.Trend = "DOWNTREND"
		} else {
			result.Trend = "SIDEWAYS"
		}
	}

	// ATR for volatility awareness
	atr := ATR(highs, lows, closes, 14)
	_ = atr // Available for advanced scoring

	// Composite score (multi-factor model)
	result.CompositeScore = sc.computeScore(result)

	// Final signal
	if result.CompositeScore >= 70 {
		result.Signal = SignalBuy
	} else if result.CompositeScore <= 30 {
		result.Signal = SignalSell
	} else {
		result.Signal = SignalHold
	}

	return result
}

func (sc *Scanner) computeScore(r *ScanResult) float64 {
	score := 50.0 // Start neutral

	// Momentum component (RSI)
	if r.RSI > 0 {
		if r.RSI > 50 && r.RSI < 70 {
			score += 10 // Bullish momentum
		} else if r.RSI < 30 {
			score += 15 // Oversold bounce potential
		} else if r.RSI > 80 {
			score -= 15 // Overbought risk
		}
	}

	// Trend component
	switch r.Trend {
	case "UPTREND":
		score += 15
	case "DOWNTREND":
		score -= 15
	}

	// MACD component
	switch r.MACDSignal {
	case "BULLISH":
		score += 10
	case "BEARISH":
		score -= 10
	}

	// Bollinger component
	switch r.BBPosition {
	case "BELOW_LOWER":
		score += 10 // Mean reversion opportunity
	case "ABOVE_UPPER":
		score -= 5  // Extended
	}

	// Volume confirmation
	if r.VolumeRatio > 1.5 && r.Change1D > 0 {
		score += 5 // High volume bullish
	}

	// Recent price momentum
	if r.Change5D > 0 {
		score += 5
	}
	if r.Change20D > 0 {
		score += 5
	}

	// Clamp
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

func matchesFilters(r ScanResult, filters []ScreenerFilter) bool {
	for _, f := range filters {
		val := getFieldValue(r, f.Field)
		switch f.Operator {
		case ">":
			if val <= f.Value {
				return false
			}
		case "<":
			if val >= f.Value {
				return false
			}
		case ">=":
			if val < f.Value {
				return false
			}
		case "<=":
			if val > f.Value {
				return false
			}
		case "==":
			if math.Abs(val-f.Value) > 0.001 {
				return false
			}
		case "between":
			if val < f.Value || val > f.Value2 {
				return false
			}
		}
	}
	return true
}

func getFieldValue(r ScanResult, field string) float64 {
	switch field {
	case "rsi":
		return r.RSI
	case "price":
		return r.Price
	case "change_1d":
		return r.Change1D
	case "change_5d":
		return r.Change5D
	case "change_20d":
		return r.Change20D
	case "volume_ratio":
		return r.VolumeRatio
	case "score":
		return r.CompositeScore
	}
	return 0
}
