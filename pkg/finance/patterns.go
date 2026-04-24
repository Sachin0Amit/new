package finance

import (
	"math"
)

// ============================================================================
// CANDLESTICK PATTERN RECOGNITION
// ============================================================================

// DetectCandlePatterns scans a time series for known candlestick patterns.
func DetectCandlePatterns(ts TimeSeries) []PatternDetection {
	var detections []PatternDetection

	if len(ts) < 3 {
		return detections
	}

	for i := 2; i < len(ts); i++ {
		prev2 := ts[i-2]
		prev1 := ts[i-1]
		curr := ts[i]

		// Doji
		if isDoji(curr) {
			detections = append(detections, PatternDetection{
				Pattern:    PatternDoji,
				Confidence: 0.8,
				Direction:  "NEUTRAL",
				StartIndex: i,
				EndIndex:   i,
				Timestamp:  curr.Timestamp,
			})
		}

		// Hammer
		if isHammer(curr, prev1) {
			detections = append(detections, PatternDetection{
				Pattern:    PatternHammer,
				Confidence: 0.75,
				Direction:  "BULLISH",
				StartIndex: i,
				EndIndex:   i,
				Timestamp:  curr.Timestamp,
			})
		}

		// Shooting Star
		if isShootingStar(curr, prev1) {
			detections = append(detections, PatternDetection{
				Pattern:    PatternShootingStar,
				Confidence: 0.75,
				Direction:  "BEARISH",
				StartIndex: i,
				EndIndex:   i,
				Timestamp:  curr.Timestamp,
			})
		}

		// Bullish Engulfing
		if isBullishEngulfing(prev1, curr) {
			detections = append(detections, PatternDetection{
				Pattern:    PatternEngulfing,
				Confidence: 0.85,
				Direction:  "BULLISH",
				StartIndex: i - 1,
				EndIndex:   i,
				Timestamp:  curr.Timestamp,
			})
		}

		// Bearish Engulfing
		if isBearishEngulfing(prev1, curr) {
			detections = append(detections, PatternDetection{
				Pattern:    PatternEngulfing,
				Confidence: 0.85,
				Direction:  "BEARISH",
				StartIndex: i - 1,
				EndIndex:   i,
				Timestamp:  curr.Timestamp,
			})
		}

		// Morning Star
		if isMorningStar(prev2, prev1, curr) {
			detections = append(detections, PatternDetection{
				Pattern:    PatternMorningStar,
				Confidence: 0.9,
				Direction:  "BULLISH",
				StartIndex: i - 2,
				EndIndex:   i,
				Timestamp:  curr.Timestamp,
			})
		}

		// Evening Star
		if isEveningStar(prev2, prev1, curr) {
			detections = append(detections, PatternDetection{
				Pattern:    PatternEveningStar,
				Confidence: 0.9,
				Direction:  "BEARISH",
				StartIndex: i - 2,
				EndIndex:   i,
				Timestamp:  curr.Timestamp,
			})
		}
	}

	return detections
}

// Helpers for candle analysis

func bodyLen(c OHLCV) float64 {
	return math.Abs(c.Close - c.Open)
}

func shadowUpper(c OHLCV) float64 {
	if c.Close > c.Open {
		return c.High - c.Close
	}
	return c.High - c.Open
}

func shadowLower(c OHLCV) float64 {
	if c.Close > c.Open {
		return c.Open - c.Low
	}
	return c.Close - c.Low
}

func isBullish(c OHLCV) bool {
	return c.Close > c.Open
}

func isBearish(c OHLCV) bool {
	return c.Close < c.Open
}

func isDoji(c OHLCV) bool {
	return bodyLen(c) <= (c.High-c.Low)*0.1
}

func isHammer(curr, prev OHLCV) bool {
	// Must be in a downtrend (simple check: prev is bearish)
	if !isBearish(prev) {
		return false
	}
	bLen := bodyLen(curr)
	sLower := shadowLower(curr)
	sUpper := shadowUpper(curr)

	// Long lower shadow (at least 2x body), short upper shadow
	return sLower >= bLen*2.0 && sUpper <= bLen*0.2
}

func isShootingStar(curr, prev OHLCV) bool {
	// Must be in an uptrend (simple check: prev is bullish)
	if !isBullish(prev) {
		return false
	}
	bLen := bodyLen(curr)
	sLower := shadowLower(curr)
	sUpper := shadowUpper(curr)

	// Long upper shadow, short lower shadow
	return sUpper >= bLen*2.0 && sLower <= bLen*0.2
}

func isBullishEngulfing(prev, curr OHLCV) bool {
	return isBearish(prev) && isBullish(curr) && curr.Open < prev.Close && curr.Close > prev.Open
}

func isBearishEngulfing(prev, curr OHLCV) bool {
	return isBullish(prev) && isBearish(curr) && curr.Open > prev.Close && curr.Close < prev.Open
}

func isMorningStar(prev2, prev1, curr OHLCV) bool {
	// Long red, small body (gap down), long green (closes into prev2 body)
	if !isBearish(prev2) || bodyLen(prev2) < (prev2.High-prev2.Low)*0.5 {
		return false
	}
	if bodyLen(prev1) > (prev1.High-prev1.Low)*0.3 {
		return false // prev1 should be small (like a doji/spinning top)
	}
	if !isBullish(curr) || curr.Close < (prev2.Open+prev2.Close)/2 {
		return false // curr must close above midpoint of prev2
	}
	return true
}

func isEveningStar(prev2, prev1, curr OHLCV) bool {
	// Long green, small body (gap up), long red (closes into prev2 body)
	if !isBullish(prev2) || bodyLen(prev2) < (prev2.High-prev2.Low)*0.5 {
		return false
	}
	if bodyLen(prev1) > (prev1.High-prev1.Low)*0.3 {
		return false // prev1 should be small
	}
	if !isBearish(curr) || curr.Close > (prev2.Open+prev2.Close)/2 {
		return false // curr must close below midpoint of prev2
	}
	return true
}
