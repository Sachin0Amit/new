package finance

import (
	"math"
)

// ============================================================================
// MOVING AVERAGES
// ============================================================================

// SMA calculates the Simple Moving Average.
func SMA(data []float64, period int) []float64 {
	if len(data) < period || period <= 0 {
		return nil
	}
	out := make([]float64, len(data))
	sum := 0.0

	// Initial sum
	for i := 0; i < period; i++ {
		sum += data[i]
		if i < period-1 {
			out[i] = math.NaN()
		}
	}
	out[period-1] = sum / float64(period)

	for i := period; i < len(data); i++ {
		sum = sum - data[i-period] + data[i]
		out[i] = sum / float64(period)
	}
	return out
}

// EMA calculates the Exponential Moving Average.
func EMA(data []float64, period int) []float64 {
	if len(data) < period || period <= 0 {
		return nil
	}
	out := make([]float64, len(data))
	multiplier := 2.0 / (float64(period) + 1.0)

	// SMA for the first valid value
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += data[i]
		if i < period-1 {
			out[i] = math.NaN()
		}
	}
	out[period-1] = sum / float64(period)

	for i := period; i < len(data); i++ {
		out[i] = (data[i]-out[i-1])*multiplier + out[i-1]
	}
	return out
}

// WMA calculates the Weighted Moving Average.
func WMA(data []float64, period int) []float64 {
	if len(data) < period || period <= 0 {
		return nil
	}
	out := make([]float64, len(data))
	weightSum := float64(period * (period + 1) / 2)

	for i := 0; i < period-1; i++ {
		out[i] = math.NaN()
	}

	for i := period - 1; i < len(data); i++ {
		sum := 0.0
		for j := 0; j < period; j++ {
			weight := float64(period - j)
			sum += data[i-j] * weight
		}
		out[i] = sum / weightSum
	}
	return out
}

// MACD calculates the Moving Average Convergence Divergence.
// Returns MACD line, Signal line, and Histogram.
func MACD(data []float64, fastPeriod, slowPeriod, signalPeriod int) ([]float64, []float64, []float64) {
	if len(data) < slowPeriod {
		return nil, nil, nil
	}
	fastEMA := EMA(data, fastPeriod)
	slowEMA := EMA(data, slowPeriod)

	macdLine := make([]float64, len(data))
	for i := range data {
		if math.IsNaN(fastEMA[i]) || math.IsNaN(slowEMA[i]) {
			macdLine[i] = math.NaN()
		} else {
			macdLine[i] = fastEMA[i] - slowEMA[i]
		}
	}

	// Calculate signal line which is EMA of MACD line.
	// Filter out NaNs for the EMA input.
	validMACD := make([]float64, 0)
	validStartIdx := 0
	for i, val := range macdLine {
		if !math.IsNaN(val) {
			if len(validMACD) == 0 {
				validStartIdx = i
			}
			validMACD = append(validMACD, val)
		}
	}

	signalLineTmp := EMA(validMACD, signalPeriod)
	signalLine := make([]float64, len(data))
	histogram := make([]float64, len(data))

	for i := 0; i < len(data); i++ {
		if i < validStartIdx {
			signalLine[i] = math.NaN()
			histogram[i] = math.NaN()
		} else {
			idx := i - validStartIdx
			if idx < len(signalLineTmp) {
				signalLine[i] = signalLineTmp[idx]
				if !math.IsNaN(macdLine[i]) && !math.IsNaN(signalLine[i]) {
					histogram[i] = macdLine[i] - signalLine[i]
				} else {
					histogram[i] = math.NaN()
				}
			} else {
				signalLine[i] = math.NaN()
				histogram[i] = math.NaN()
			}
		}
	}

	return macdLine, signalLine, histogram
}

// ============================================================================
// MOMENTUM & OSCILLATORS
// ============================================================================

// RSI calculates the Relative Strength Index.
func RSI(data []float64, period int) []float64 {
	if len(data) <= period || period <= 0 {
		return nil
	}
	out := make([]float64, len(data))
	out[0] = math.NaN()

	gains := 0.0
	losses := 0.0

	for i := 1; i <= period; i++ {
		out[i] = math.NaN()
		change := data[i] - data[i-1]
		if change > 0 {
			gains += change
		} else {
			losses -= change
		}
	}

	avgGain := gains / float64(period)
	avgLoss := losses / float64(period)

	if avgLoss == 0 {
		out[period] = 100.0
	} else {
		rs := avgGain / avgLoss
		out[period] = 100.0 - (100.0 / (1.0 + rs))
	}

	for i := period + 1; i < len(data); i++ {
		change := data[i] - data[i-1]
		gain := 0.0
		loss := 0.0
		if change > 0 {
			gain = change
		} else {
			loss = -change
		}

		avgGain = (avgGain*float64(period-1) + gain) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + loss) / float64(period)

		if avgLoss == 0 {
			out[i] = 100.0
		} else {
			rs := avgGain / avgLoss
			out[i] = 100.0 - (100.0 / (1.0 + rs))
		}
	}
	return out
}

// Stochastic returns %K and %D lines.
func Stochastic(highs, lows, closes []float64, periodK, smoothK, periodD int) ([]float64, []float64) {
	if len(closes) < periodK {
		return nil, nil
	}
	
	rawK := make([]float64, len(closes))
	for i := 0; i < len(closes); i++ {
		if i < periodK-1 {
			rawK[i] = math.NaN()
			continue
		}
		
		highestHigh := Max(highs[i-periodK+1 : i+1])
		lowestLow := Min(lows[i-periodK+1 : i+1])
		
		if highestHigh == lowestLow {
			rawK[i] = 50.0
		} else {
			rawK[i] = 100.0 * ((closes[i] - lowestLow) / (highestHigh - lowestLow))
		}
	}
	
	// Smooth %K
	var kLine []float64
	if smoothK > 1 {
		kLine = SMA(rawK[periodK-1:], smoothK)
		// Pad front with NaNs
		paddedK := make([]float64, len(closes))
		for i := 0; i < periodK-1; i++ {
			paddedK[i] = math.NaN()
		}
		copy(paddedK[periodK-1:], kLine)
		kLine = paddedK
	} else {
		kLine = rawK
	}
	
	// Calculate %D (SMA of %K)
	validK := make([]float64, 0)
	validStartIdx := 0
	for i, val := range kLine {
		if !math.IsNaN(val) {
			if len(validK) == 0 {
				validStartIdx = i
			}
			validK = append(validK, val)
		}
	}
	
	dTmp := SMA(validK, periodD)
	dLine := make([]float64, len(closes))
	for i := 0; i < len(closes); i++ {
		if i < validStartIdx {
			dLine[i] = math.NaN()
		} else {
			idx := i - validStartIdx
			if idx < len(dTmp) {
				dLine[i] = dTmp[idx]
			} else {
				dLine[i] = math.NaN()
			}
		}
	}
	
	return kLine, dLine
}

// ============================================================================
// VOLATILITY
// ============================================================================

// TrueRange calculates the True Range.
func TrueRange(highs, lows, closes []float64) []float64 {
	out := make([]float64, len(closes))
	if len(closes) == 0 {
		return out
	}
	out[0] = highs[0] - lows[0]
	for i := 1; i < len(closes); i++ {
		hl := highs[i] - lows[i]
		hc := math.Abs(highs[i] - closes[i-1])
		lc := math.Abs(lows[i] - closes[i-1])
		out[i] = math.Max(hl, math.Max(hc, lc))
	}
	return out
}

// ATR calculates the Average True Range.
func ATR(highs, lows, closes []float64, period int) []float64 {
	if len(closes) < period {
		return nil
	}
	tr := TrueRange(highs, lows, closes)
	
	out := make([]float64, len(closes))
	sum := 0.0
	for i := 0; i < period; i++ {
		sum += tr[i]
		if i < period-1 {
			out[i] = math.NaN()
		}
	}
	out[period-1] = sum / float64(period)
	
	for i := period; i < len(closes); i++ {
		out[i] = (out[i-1]*float64(period-1) + tr[i]) / float64(period)
	}
	return out
}

// BollingerBands calculates upper, middle, and lower bands.
func BollingerBands(data []float64, period int, numStdDev float64) ([]float64, []float64, []float64) {
	if len(data) < period {
		return nil, nil, nil
	}
	
	middle := SMA(data, period)
	upper := make([]float64, len(data))
	lower := make([]float64, len(data))
	
	for i := 0; i < len(data); i++ {
		if math.IsNaN(middle[i]) {
			upper[i] = math.NaN()
			lower[i] = math.NaN()
			continue
		}
		
		stdDev := StdDev(data[i-period+1 : i+1])
		upper[i] = middle[i] + (numStdDev * stdDev)
		lower[i] = middle[i] - (numStdDev * stdDev)
	}
	
	return upper, middle, lower
}

// ============================================================================
// VOLUME INDICATORS
// ============================================================================

// OBV calculates On-Balance Volume.
func OBV(closes, volumes []float64) []float64 {
	if len(closes) == 0 {
		return nil
	}
	out := make([]float64, len(closes))
	out[0] = volumes[0]
	
	for i := 1; i < len(closes); i++ {
		if closes[i] > closes[i-1] {
			out[i] = out[i-1] + volumes[i]
		} else if closes[i] < closes[i-1] {
			out[i] = out[i-1] - volumes[i]
		} else {
			out[i] = out[i-1]
		}
	}
	return out
}

// VWAP calculates the Volume Weighted Average Price.
// This is typically reset daily, but here we provide a cumulative version.
func VWAP(highs, lows, closes, volumes []float64) []float64 {
	if len(closes) == 0 {
		return nil
	}
	out := make([]float64, len(closes))
	
	cumVol := 0.0
	cumPV := 0.0
	
	for i := 0; i < len(closes); i++ {
		typicalPrice := (highs[i] + lows[i] + closes[i]) / 3.0
		cumVol += volumes[i]
		cumPV += typicalPrice * volumes[i]
		
		if cumVol > 0 {
			out[i] = cumPV / cumVol
		} else {
			out[i] = typicalPrice
		}
	}
	return out
}
