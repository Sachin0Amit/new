package finance

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// COMPOSITE SIGNAL AGGREGATOR
// ============================================================================

// SignalAggregator combines multiple technical indicators, pattern detections,
// and statistical predictions into a single, weighted trading signal with
// human-readable reasoning.
type SignalAggregator struct {
	cache *MarketDataCache
}

// NewSignalAggregator creates a new aggregator.
func NewSignalAggregator(cache *MarketDataCache) *SignalAggregator {
	return &SignalAggregator{cache: cache}
}

// GenerateSignal produces a comprehensive trading signal for a symbol.
func (sa *SignalAggregator) GenerateSignal(symbol string) *TradingSignal {
	ts, ok := sa.cache.Get(symbol, TFDaily)
	if !ok || len(ts) < 50 {
		return nil
	}

	closes := ts.Closes()
	highs := ts.Highs()
	lows := ts.Lows()
	volumes := ts.Volumes()
	n := len(closes)

	signal := &TradingSignal{
		ID:        uuid.New(),
		Symbol:    symbol,
		Price:     closes[n-1],
		Timestamp: time.Now(),
	}

	// Collect indicator signals with weights
	type weightedSignal struct {
		name      string
		direction float64 // -1 to +1
		weight    float64
		detail    string
	}

	signals := make([]weightedSignal, 0, 10)

	// 1. RSI (14) — weight: 0.15
	rsi := RSI(closes, 14)
	if rsi != nil && !math.IsNaN(rsi[n-1]) {
		rsiVal := rsi[n-1]
		dir := 0.0
		detail := fmt.Sprintf("RSI(14) = %.1f", rsiVal)
		if rsiVal < 30 {
			dir = 0.8
			detail += " — Oversold"
		} else if rsiVal < 40 {
			dir = 0.3
			detail += " — Approaching Oversold"
		} else if rsiVal > 70 {
			dir = -0.8
			detail += " — Overbought"
		} else if rsiVal > 60 {
			dir = -0.3
			detail += " — Approaching Overbought"
		}
		signals = append(signals, weightedSignal{"RSI", dir, 0.15, detail})
		signal.Indicators = append(signal.Indicators, IndicatorResult{
			Name:   "RSI_14",
			Values: rsi[n-5:],
			Signal: dirToSignalString(dir),
			Strength: math.Abs(dir),
		})
	}

	// 2. MACD — weight: 0.15
	macdLine, macdSig, macdHist := MACD(closes, 12, 26, 9)
	if macdHist != nil && n >= 2 && !math.IsNaN(macdHist[n-1]) {
		dir := 0.0
		detail := fmt.Sprintf("MACD Hist = %.4f", macdHist[n-1])
		if macdHist[n-1] > 0 && macdHist[n-2] <= 0 {
			dir = 1.0
			detail += " — Bullish Crossover"
		} else if macdHist[n-1] < 0 && macdHist[n-2] >= 0 {
			dir = -1.0
			detail += " — Bearish Crossover"
		} else if macdHist[n-1] > 0 {
			dir = 0.3
			detail += " — Bullish Momentum"
		} else {
			dir = -0.3
			detail += " — Bearish Momentum"
		}
		signals = append(signals, weightedSignal{"MACD", dir, 0.15, detail})
		_ = macdLine
		_ = macdSig
	}

	// 3. Bollinger Bands — weight: 0.10
	bbUpper, bbMiddle, bbLower := BollingerBands(closes, 20, 2.0)
	if bbUpper != nil && !math.IsNaN(bbUpper[n-1]) {
		dir := 0.0
		pctB := 0.0
		if bbUpper[n-1] != bbLower[n-1] {
			pctB = (closes[n-1] - bbLower[n-1]) / (bbUpper[n-1] - bbLower[n-1])
		}
		detail := fmt.Sprintf("%%B = %.2f", pctB)
		if pctB <= 0.0 {
			dir = 0.8
			detail += " — Below Lower Band (Mean Reversion)"
		} else if pctB >= 1.0 {
			dir = -0.5
			detail += " — Above Upper Band (Extended)"
		}
		signals = append(signals, weightedSignal{"Bollinger", dir, 0.10, detail})
		_ = bbMiddle
	}

	// 4. SMA Trend Alignment — weight: 0.20
	sma20 := SMA(closes, 20)
	sma50 := SMA(closes, 50)
	if sma20 != nil && sma50 != nil && !math.IsNaN(sma20[n-1]) && !math.IsNaN(sma50[n-1]) {
		dir := 0.0
		detail := ""
		if closes[n-1] > sma20[n-1] && sma20[n-1] > sma50[n-1] {
			dir = 0.8
			detail = "Price > SMA20 > SMA50 — Strong Uptrend"
		} else if closes[n-1] < sma20[n-1] && sma20[n-1] < sma50[n-1] {
			dir = -0.8
			detail = "Price < SMA20 < SMA50 — Strong Downtrend"
		} else {
			dir = 0.0
			detail = "Mixed SMA alignment — Consolidation"
		}
		signals = append(signals, weightedSignal{"SMA_TREND", dir, 0.20, detail})
	}

	// 5. Volume Confirmation — weight: 0.10
	if n >= 21 {
		avgVol := Mean(volumes[n-20 : n])
		if avgVol > 0 {
			volRatio := volumes[n-1] / avgVol
			dir := 0.0
			detail := fmt.Sprintf("Volume Ratio = %.2f", volRatio)
			if volRatio > 1.5 && closes[n-1] > closes[n-2] {
				dir = 0.5
				detail += " — High Volume Bullish"
			} else if volRatio > 1.5 && closes[n-1] < closes[n-2] {
				dir = -0.5
				detail += " — High Volume Bearish"
			}
			signals = append(signals, weightedSignal{"VOLUME", dir, 0.10, detail})
		}
	}

	// 6. ATR Stop-Loss Calculation — weight: 0.0 (info only)
	atr := ATR(highs, lows, closes, 14)
	if atr != nil && !math.IsNaN(atr[n-1]) {
		signal.StopLoss = closes[n-1] - (2.0 * atr[n-1])
		signal.TakeProfit = closes[n-1] + (3.0 * atr[n-1])
	}

	// 7. Pattern Detection — weight: 0.15
	patterns := DetectCandlePatterns(ts)
	recentPatterns := make([]PatternDetection, 0)
	for _, p := range patterns {
		if p.EndIndex >= n-3 { // Last 3 bars
			recentPatterns = append(recentPatterns, p)
		}
	}
	if len(recentPatterns) > 0 {
		avgDir := 0.0
		patNames := make([]string, 0)
		for _, p := range recentPatterns {
			if p.Direction == "BULLISH" {
				avgDir += p.Confidence
			} else {
				avgDir -= p.Confidence
			}
			patNames = append(patNames, string(p.Pattern))
		}
		avgDir /= float64(len(recentPatterns))
		detail := "Patterns: " + strings.Join(patNames, ", ")
		signals = append(signals, weightedSignal{"PATTERNS", avgDir, 0.15, detail})
	}

	// 8. Prediction — weight: 0.15
	prediction := EnsemblePredict(closes, 7) // 7-day prediction
	if prediction != nil {
		dir := 0.0
		detail := fmt.Sprintf("7d Prediction: $%.2f (%.0f%% conf)", prediction.PredictedPrice, prediction.Confidence*100)
		if prediction.Direction == "BULLISH" {
			dir = prediction.Confidence * 0.8
		} else if prediction.Direction == "BEARISH" {
			dir = -prediction.Confidence * 0.8
		}
		signals = append(signals, weightedSignal{"PREDICTION", dir, 0.15, detail})
		signal.Prediction = prediction
	}

	// Aggregate weighted score
	totalWeight := 0.0
	weightedScore := 0.0
	reasons := make([]string, 0, len(signals))

	for _, ws := range signals {
		weightedScore += ws.direction * ws.weight
		totalWeight += ws.weight
		if ws.detail != "" {
			reasons = append(reasons, ws.detail)
		}
	}

	if totalWeight > 0 {
		weightedScore /= totalWeight
	}

	// Map score to signal type and confidence
	signal.Confidence = math.Abs(weightedScore)
	if weightedScore > 0.2 {
		signal.Type = SignalBuy
	} else if weightedScore < -0.2 {
		signal.Type = SignalSell
	} else {
		signal.Type = SignalHold
	}

	signal.Reasoning = strings.Join(reasons, " | ")

	return signal
}

func dirToSignalString(dir float64) string {
	if dir > 0.2 {
		return "BUY"
	}
	if dir < -0.2 {
		return "SELL"
	}
	return "NEUTRAL"
}
