package finance

import (
	"math"
)

// ============================================================================
// VOLATILITY MODELS
// ============================================================================

// HistoricalVolatility computes the annualized volatility over a rolling window.
func HistoricalVolatility(closes []float64, window int) []float64 {
	if len(closes) < window+1 {
		return nil
	}

	// First compute returns
	returns := make([]float64, len(closes)-1)
	for i := 1; i < len(closes); i++ {
		if closes[i-1] > 0 {
			returns[i-1] = math.Log(closes[i] / closes[i-1])
		}
	}

	out := make([]float64, len(closes))
	for i := 0; i < window; i++ {
		out[i] = math.NaN()
	}

	for i := window; i < len(returns)+1; i++ {
		windowReturns := returns[i-window : i]
		dailyVol := StdDev(windowReturns)
		out[i] = dailyVol * math.Sqrt(TradingDaysPerYear) // Annualize
	}

	return out
}

// EWMA calculates Exponentially Weighted Moving Average volatility.
// This is a simplified version of the RiskMetrics model.
func EWMA(returns []float64, lambda float64) []float64 {
	if len(returns) < 2 {
		return nil
	}

	out := make([]float64, len(returns))
	// Initialize with the first return squared
	variance := returns[0] * returns[0]
	out[0] = math.Sqrt(variance) * math.Sqrt(TradingDaysPerYear)

	for i := 1; i < len(returns); i++ {
		variance = lambda*variance + (1.0-lambda)*returns[i]*returns[i]
		out[i] = math.Sqrt(variance) * math.Sqrt(TradingDaysPerYear) // Annualized
	}

	return out
}

// SimpleGARCH implements a GARCH(1,1) volatility model.
// σ²(t) = ω + α * r²(t-1) + β * σ²(t-1)
// Where:
//   - ω (omega) is the long-run variance constant
//   - α (alpha) is the coefficient for the lagged squared return
//   - β (beta) is the coefficient for the lagged variance
//   - α + β < 1 for stationarity
type GARCHParams struct {
	Omega float64 `json:"omega"`
	Alpha float64 `json:"alpha"`
	Beta  float64 `json:"beta"`
}

// DefaultGARCHParams returns commonly used GARCH(1,1) parameters.
func DefaultGARCHParams() GARCHParams {
	return GARCHParams{
		Omega: 0.000001,
		Alpha: 0.09,
		Beta:  0.90,
	}
}

// GARCHVolatility computes the GARCH(1,1) conditional volatility series.
func GARCHVolatility(returns []float64, params GARCHParams) []float64 {
	if len(returns) < 2 {
		return nil
	}

	out := make([]float64, len(returns))

	// Initialize with unconditional variance
	// For GARCH(1,1): σ² = ω / (1 - α - β)
	denom := 1.0 - params.Alpha - params.Beta
	if denom <= 0 {
		denom = 0.01 // Prevent instability
	}
	variance := params.Omega / denom
	out[0] = math.Sqrt(variance) * math.Sqrt(TradingDaysPerYear)

	for i := 1; i < len(returns); i++ {
		variance = params.Omega +
			params.Alpha*returns[i-1]*returns[i-1] +
			params.Beta*variance

		// Clamp to prevent numerical instability
		if variance < 0 {
			variance = 0.000001
		}

		out[i] = math.Sqrt(variance) * math.Sqrt(TradingDaysPerYear) // Annualized
	}

	return out
}

// ForecastGARCH projects volatility forward N periods using GARCH(1,1).
func ForecastGARCH(returns []float64, params GARCHParams, horizon int) []float64 {
	if len(returns) < 2 || horizon <= 0 {
		return nil
	}

	// Get the last conditional variance
	denom := 1.0 - params.Alpha - params.Beta
	if denom <= 0 {
		denom = 0.01
	}
	longRunVar := params.Omega / denom

	// Last period variance
	variance := longRunVar
	for i := 1; i < len(returns); i++ {
		variance = params.Omega + params.Alpha*returns[i-1]*returns[i-1] + params.Beta*variance
	}

	// Project forward
	forecast := make([]float64, horizon)
	for h := 0; h < horizon; h++ {
		// GARCH(1,1) multi-step forecast converges to unconditional variance
		// σ²(h) = σ² + (α+β)^h * (σ²(1) - σ²)
		abSum := params.Alpha + params.Beta
		forecast[h] = longRunVar + math.Pow(abSum, float64(h))*(variance-longRunVar)
		if forecast[h] < 0 {
			forecast[h] = 0.000001
		}
		forecast[h] = math.Sqrt(forecast[h]) * math.Sqrt(TradingDaysPerYear)
	}

	return forecast
}

// ParkinsonsVolatility uses the Parkinson estimator which leverages
// high-low price range and is more efficient than close-to-close volatility.
func ParkinsonsVolatility(highs, lows []float64, window int) []float64 {
	if len(highs) < window || len(lows) < window {
		return nil
	}

	out := make([]float64, len(highs))
	for i := 0; i < window-1; i++ {
		out[i] = math.NaN()
	}

	factor := 1.0 / (4.0 * math.Log(2.0))

	for i := window - 1; i < len(highs); i++ {
		sum := 0.0
		for j := i - window + 1; j <= i; j++ {
			if lows[j] > 0 {
				ratio := math.Log(highs[j] / lows[j])
				sum += ratio * ratio
			}
		}
		dailyVar := factor * sum / float64(window)
		out[i] = math.Sqrt(dailyVar) * math.Sqrt(TradingDaysPerYear)
	}

	return out
}

// VolatilityCone computes percentile ranges of realized volatility
// at different lookback windows, helping to assess if current vol is
// historically rich or cheap.
type VolatilityConePoint struct {
	Window     int     `json:"window_days"`
	Percentile10 float64 `json:"p10"`
	Percentile25 float64 `json:"p25"`
	Percentile50 float64 `json:"p50"`
	Percentile75 float64 `json:"p75"`
	Percentile90 float64 `json:"p90"`
	Current      float64 `json:"current"`
}

// BuildVolatilityCone constructs the volatility cone across windows.
func BuildVolatilityCone(closes []float64, windows []int) []VolatilityConePoint {
	points := make([]VolatilityConePoint, 0, len(windows))

	for _, w := range windows {
		hvSeries := HistoricalVolatility(closes, w)
		if hvSeries == nil {
			continue
		}

		// Collect valid values
		valid := make([]float64, 0)
		for _, v := range hvSeries {
			if !math.IsNaN(v) {
				valid = append(valid, v)
			}
		}

		if len(valid) < 5 {
			continue
		}

		sorted := SortedCopy(valid)
		current := valid[len(valid)-1]

		points = append(points, VolatilityConePoint{
			Window:       w,
			Percentile10: Percentile(sorted, 10),
			Percentile25: Percentile(sorted, 25),
			Percentile50: Percentile(sorted, 50),
			Percentile75: Percentile(sorted, 75),
			Percentile90: Percentile(sorted, 90),
			Current:      current,
		})
	}

	return points
}
