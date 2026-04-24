package finance

import (
	"math"
	"time"
)

// ============================================================================
// CROSS-MARKET CORRELATION ANALYSIS
// ============================================================================

// BuildCorrelationMatrix computes a pairwise Pearson correlation matrix
// across all provided time series.
func BuildCorrelationMatrix(assets map[string]TimeSeries) *CorrelationMatrix {
	symbols := make([]string, 0, len(assets))
	for s := range assets {
		symbols = append(symbols, s)
	}

	n := len(symbols)
	matrix := make([][]float64, n)

	for i := 0; i < n; i++ {
		matrix[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			if i == j {
				matrix[i][j] = 1.0
				continue
			}

			retI := assets[symbols[i]].Returns()
			retJ := assets[symbols[j]].Returns()

			// Align lengths (take the minimum)
			minLen := len(retI)
			if len(retJ) < minLen {
				minLen = len(retJ)
			}
			if minLen < 2 {
				matrix[i][j] = 0
				continue
			}

			// Use the most recent data
			aI := retI[len(retI)-minLen:]
			aJ := retJ[len(retJ)-minLen:]

			matrix[i][j] = Correlation(aI, aJ)
		}
	}

	return &CorrelationMatrix{
		Symbols: symbols,
		Matrix:  matrix,
	}
}

// RollingCorrelation computes the correlation between two return series
// over a sliding window, producing a time series of correlation values.
func RollingCorrelation(returnsA, returnsB []float64, window int) []float64 {
	if len(returnsA) != len(returnsB) {
		minLen := len(returnsA)
		if len(returnsB) < minLen {
			minLen = len(returnsB)
		}
		returnsA = returnsA[:minLen]
		returnsB = returnsB[:minLen]
	}

	n := len(returnsA)
	if n < window {
		return nil
	}

	out := make([]float64, n)
	for i := 0; i < window-1; i++ {
		out[i] = math.NaN()
	}

	for i := window - 1; i < n; i++ {
		windowA := returnsA[i-window+1 : i+1]
		windowB := returnsB[i-window+1 : i+1]
		out[i] = Correlation(windowA, windowB)
	}

	return out
}

// ============================================================================
// MARKET BREADTH ANALYSIS
// ============================================================================

// CalculateMarketBreadth computes the breadth indicators for a collection
// of assets' daily price changes.
func CalculateMarketBreadth(assets map[string]TimeSeries) *MarketBreadth {
	mb := &MarketBreadth{
		Timestamp: time.Now(),
	}

	for _, ts := range assets {
		if len(ts) < 2 {
			mb.UnchangedCount++
			continue
		}

		lastClose := ts[len(ts)-1].Close
		prevClose := ts[len(ts)-2].Close
		changePct := (lastClose - prevClose) / prevClose

		if changePct > 0.001 {
			mb.AdvancingCount++
		} else if changePct < -0.001 {
			mb.DecliningCount++
		} else {
			mb.UnchangedCount++
		}

		// 52-week high/low approximation (use last 252 bars if available)
		lookback := 252
		if len(ts) < lookback {
			lookback = len(ts)
		}

		recentHighs := ts[len(ts)-lookback:].Highs()
		recentLows := ts[len(ts)-lookback:].Lows()

		highest := Max(recentHighs)
		lowest := Min(recentLows)

		// Is today a new high or new low?
		if lastClose >= highest*0.99 {
			mb.NewHighs++
		}
		if lastClose <= lowest*1.01 {
			mb.NewLows++
		}
	}

	// Ratios
	if mb.DecliningCount > 0 {
		mb.AdvanceDecline = float64(mb.AdvancingCount) / float64(mb.DecliningCount)
	}
	if mb.NewLows > 0 {
		mb.HighLowRatio = float64(mb.NewHighs) / float64(mb.NewLows)
	}

	total := float64(mb.AdvancingCount + mb.DecliningCount + mb.UnchangedCount)
	if total > 0 {
		mb.BullishPercent = float64(mb.AdvancingCount) / total * 100
	}

	// Fear/Greed Index (simplified model)
	// Based on: advance/decline ratio, high/low ratio, bullish percent
	score := 50.0 // Start neutral
	score += (mb.AdvanceDecline - 1.0) * 20.0
	score += (mb.HighLowRatio - 1.0) * 10.0
	score += (mb.BullishPercent - 50.0) * 0.5

	// Clamp to 0-100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	mb.FearGreedIndex = score

	if score < 30 {
		mb.MarketSentiment = "FEAR"
	} else if score > 70 {
		mb.MarketSentiment = "GREED"
	} else {
		mb.MarketSentiment = "NEUTRAL"
	}

	return mb
}
