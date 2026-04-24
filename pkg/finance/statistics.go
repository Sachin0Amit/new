package finance

import (
	"math"
	"sort"
)

// ============================================================================
// RISK & RETURN METRICS
// ============================================================================

const (
	TradingDaysPerYear = 252.0
)

// AnnualizeReturn converts a cumulative return to an annualized return.
func AnnualizeReturn(totalReturn float64, days int) float64 {
	if days <= 0 {
		return 0
	}
	years := float64(days) / TradingDaysPerYear
	return math.Pow(1.0+totalReturn, 1.0/years) - 1.0
}

// AnnualizeVolatility converts daily volatility to annualized volatility.
func AnnualizeVolatility(dailyVol float64) float64 {
	return dailyVol * math.Sqrt(TradingDaysPerYear)
}

// SharpeRatio calculates the risk-adjusted return relative to the risk-free rate.
func SharpeRatio(returns []float64, riskFreeRateAnnual float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	rfDaily := math.Pow(1.0+riskFreeRateAnnual, 1.0/TradingDaysPerYear) - 1.0
	
	excessReturns := make([]float64, len(returns))
	for i, r := range returns {
		excessReturns[i] = r - rfDaily
	}
	
	meanExcess := Mean(excessReturns)
	stdDevExcess := StdDev(excessReturns)
	
	if stdDevExcess == 0 {
		return 0
	}
	
	// Annualize the Sharpe Ratio
	return (meanExcess / stdDevExcess) * math.Sqrt(TradingDaysPerYear)
}

// SortinoRatio calculates the risk-adjusted return using downside deviation.
func SortinoRatio(returns []float64, riskFreeRateAnnual, targetReturnAnnual float64) float64 {
	if len(returns) < 2 {
		return 0
	}
	targetDaily := math.Pow(1.0+targetReturnAnnual, 1.0/TradingDaysPerYear) - 1.0
	rfDaily := math.Pow(1.0+riskFreeRateAnnual, 1.0/TradingDaysPerYear) - 1.0
	
	excessReturns := make([]float64, len(returns))
	downsideReturns := make([]float64, 0, len(returns))
	
	for i, r := range returns {
		excessReturns[i] = r - rfDaily
		if r < targetDaily {
			downsideReturns = append(downsideReturns, r-targetDaily)
		}
	}
	
	meanExcess := Mean(excessReturns)
	
	downsideSumSq := 0.0
	for _, dr := range downsideReturns {
		downsideSumSq += dr * dr
	}
	
	if len(downsideReturns) == 0 {
		return math.Inf(1)
	}
	
	downsideDeviation := math.Sqrt(downsideSumSq / float64(len(returns)))
	
	if downsideDeviation == 0 {
		return 0
	}
	
	return (meanExcess / downsideDeviation) * math.Sqrt(TradingDaysPerYear)
}

// MaximumDrawdown computes the largest peak-to-trough drop.
// Returns the drawdown percentage and the duration in days.
func MaximumDrawdown(equityCurve []float64) (float64, int) {
	if len(equityCurve) < 2 {
		return 0, 0
	}
	
	maxPeak := equityCurve[0]
	maxDrawdown := 0.0
	maxDuration := 0
	currentDuration := 0
	
	for _, v := range equityCurve {
		if v > maxPeak {
			maxPeak = v
			currentDuration = 0
		} else {
			currentDuration++
			dd := (maxPeak - v) / maxPeak
			if dd > maxDrawdown {
				maxDrawdown = dd
				maxDuration = currentDuration
			}
		}
	}
	return maxDrawdown, maxDuration
}

// ValueAtRisk computes the Historical VaR at the given confidence level (e.g., 95).
func ValueAtRisk(returns []float64, confidenceLevel float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	sorted := make([]float64, len(returns))
	copy(sorted, returns)
	sort.Float64s(sorted)
	
	percentile := 100.0 - confidenceLevel
	return -Percentile(sorted, percentile)
}

// ConditionalVaR computes the Expected Shortfall (CVaR).
func ConditionalVaR(returns []float64, confidenceLevel float64) float64 {
	if len(returns) == 0 {
		return 0
	}
	varValue := -ValueAtRisk(returns, confidenceLevel) // Get raw return threshold
	
	sum := 0.0
	count := 0
	for _, r := range returns {
		if r <= varValue {
			sum += r
			count++
		}
	}
	
	if count == 0 {
		return 0
	}
	return -(sum / float64(count))
}

// Beta calculates the volatility of an asset relative to a benchmark.
func Beta(assetReturns, benchmarkReturns []float64) float64 {
	cov := Covariance(assetReturns, benchmarkReturns)
	varBench := Variance(benchmarkReturns)
	if varBench == 0 {
		return 0
	}
	return cov / varBench
}

// Alpha calculates the excess return relative to the benchmark return and Beta.
func Alpha(assetReturns, benchmarkReturns []float64, riskFreeRateAnnual float64) float64 {
	beta := Beta(assetReturns, benchmarkReturns)

	
	meanAsset := Mean(assetReturns)
	meanBench := Mean(benchmarkReturns)
	
	// Annualize alpha
	assetAnn := math.Pow(1.0+meanAsset, TradingDaysPerYear) - 1.0
	benchAnn := math.Pow(1.0+meanBench, TradingDaysPerYear) - 1.0
	
	return assetAnn - (riskFreeRateAnnual + beta*(benchAnn-riskFreeRateAnnual))
}

// CalculateRiskMetrics compiles a comprehensive RiskMetrics struct.
func CalculateRiskMetrics(equityCurve, assetReturns, benchReturns []float64, rfAnnual float64) *RiskMetrics {
	if len(equityCurve) < 2 || len(assetReturns) < 2 {
		return &RiskMetrics{}
	}
	
	m := &RiskMetrics{}
	
	m.VolatilityDaily = StdDev(assetReturns)
	m.Volatility = AnnualizeVolatility(m.VolatilityDaily)
	m.SharpeRatio = SharpeRatio(assetReturns, rfAnnual)
	m.SortinoRatio = SortinoRatio(assetReturns, rfAnnual, 0.0) // 0% target
	m.MaxDrawdown, m.MaxDrawdownDur = MaximumDrawdown(equityCurve)
	m.ValueAtRisk95 = ValueAtRisk(assetReturns, 95.0)
	m.ValueAtRisk99 = ValueAtRisk(assetReturns, 99.0)
	m.CVaR95 = ConditionalVaR(assetReturns, 95.0)
	
	if len(benchReturns) == len(assetReturns) {
		m.Beta = Beta(assetReturns, benchReturns)
		m.Alpha = Alpha(assetReturns, benchReturns, rfAnnual)
		m.TreynorRatio = (math.Pow(1.0+Mean(assetReturns), TradingDaysPerYear) - 1.0 - rfAnnual) / m.Beta
	}
	
	// Calculate win rate
	wins := 0
	for _, r := range assetReturns {
		if r > 0 {
			wins++
		}
	}
	m.WinRate = float64(wins) / float64(len(assetReturns))
	
	return m
}
