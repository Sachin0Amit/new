package finance

import (
	"math"
	"math/rand"
	"sort"
	"time"
)

// ============================================================================
// STATISTICAL PREDICTION MODELS
// ============================================================================

// LinearRegression fits a line y = mx + c to the data and predicts future values.
// x is time index (0, 1, 2...), y is price.
func LinearRegression(prices []float64, horizon int) *PredictionResult {
	if len(prices) < 2 {
		return nil
	}

	n := float64(len(prices))
	sumX := 0.0
	sumY := 0.0
	sumXY := 0.0
	sumX2 := 0.0

	for i, y := range prices {
		x := float64(i)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	denominator := (n * sumX2) - (sumX * sumX)
	if denominator == 0 {
		return nil
	}

	m := ((n * sumXY) - (sumX * sumY)) / denominator
	c := (sumY - (m * sumX)) / n

	// Calculate R-squared and RMSE for confidence
	sst := 0.0
	ssr := 0.0
	meanY := sumY / n
	mse := 0.0

	for i, y := range prices {
		predicted := m*float64(i) + c
		sst += (y - meanY) * (y - meanY)
		ssr += (predicted - meanY) * (predicted - meanY)
		mse += (y - predicted) * (y - predicted)
	}

	r2 := 0.0
	if sst != 0 {
		r2 = ssr / sst
	}
	rmse := math.Sqrt(mse / n)

	// Predict
	targetX := float64(len(prices) - 1 + horizon)
	predictedPrice := m*targetX + c

	// Direction
	direction := "NEUTRAL"
	if m > 0.001 { // Arbitrary small threshold
		direction = "BULLISH"
	} else if m < -0.001 {
		direction = "BEARISH"
	}

	// Calculate bounds based on standard error
	stdErr := rmse * math.Sqrt(1+1/n+math.Pow(targetX-sumX/n, 2)/(sumX2-sumX*sumX/n))
	zScore95 := 1.96

	return &PredictionResult{
		Model:          "LINEAR_REGRESSION",
		PredictedPrice: predictedPrice,
		Confidence:     math.Max(0, math.Min(1, r2)), // Use R2 as pseudo-confidence
		Direction:      direction,
		Horizon:        horizon,
		UpperBound:     predictedPrice + (zScore95 * stdErr),
		LowerBound:     predictedPrice - (zScore95 * stdErr),
		R2Score:        r2,
		RMSE:           rmse,
		Timestamp:      time.Now(),
	}
}

// MonteCarloSimulation runs N random walks based on historical volatility and drift.
// Returns the expected price, bounds, and one sample path.
func MonteCarloSimulation(prices []float64, horizon int, numSimulations int) *PredictionResult {
	if len(prices) < 2 {
		return nil
	}

	// Calculate daily returns
	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		returns[i-1] = math.Log(prices[i] / prices[i-1])
	}

	mu := Mean(returns)
	sigma := StdDev(returns)
	lastPrice := prices[len(prices)-1]

	// Drift = mu - (sigma^2 / 2)
	drift := mu - (0.5 * sigma * sigma)

	randSrc := rand.New(rand.NewSource(time.Now().UnixNano()))
	finalPrices := make([]float64, numSimulations)
	
	// Keep one path for visualization
	samplePath := make([]float64, horizon)

	for sim := 0; sim < numSimulations; sim++ {
		currentPrice := lastPrice
		for step := 0; step < horizon; step++ {
			// Geometric Brownian Motion: S_t = S_{t-1} * exp(drift + sigma * Z)
			// Z is standard normal
			z := randSrc.NormFloat64()
			shock := math.Exp(drift + sigma*z)
			currentPrice *= shock
			
			if sim == 0 {
				samplePath[step] = currentPrice
			}
		}
		finalPrices[sim] = currentPrice
	}

	// Analyze results
	sort.Float64s(finalPrices)
	expectedPrice := Mean(finalPrices)
	
	lowerBound := Percentile(finalPrices, 5.0)  // 90% confidence interval
	upperBound := Percentile(finalPrices, 95.0)

	direction := "NEUTRAL"
	if expectedPrice > lastPrice*1.02 {
		direction = "BULLISH"
	} else if expectedPrice < lastPrice*0.98 {
		direction = "BEARISH"
	}

	// Confidence based inversely on volatility spread
	spreadPct := (upperBound - lowerBound) / expectedPrice
	confidence := 1.0 - math.Min(spreadPct, 1.0)

	return &PredictionResult{
		Model:          "MONTE_CARLO_GBM",
		PredictedPrice: expectedPrice,
		Confidence:     confidence,
		Direction:      direction,
		Horizon:        horizon,
		UpperBound:     upperBound,
		LowerBound:     lowerBound,
		PricePath:      samplePath,
		Timestamp:      time.Now(),
	}
}

// MovingAverageCrossover builds a basic predictive signal based on MACD or SMA crossovers.
func MovingAverageCrossover(prices []float64, fastPeriod, slowPeriod int) SignalType {
	if len(prices) < slowPeriod {
		return SignalNeutral
	}

	fastSMA := SMA(prices, fastPeriod)
	slowSMA := SMA(prices, slowPeriod)

	currFast := fastSMA[len(fastSMA)-1]
	currSlow := slowSMA[len(slowSMA)-1]
	prevFast := fastSMA[len(fastSMA)-2]
	prevSlow := slowSMA[len(slowSMA)-2]

	if prevFast <= prevSlow && currFast > currSlow {
		return SignalBuy // Golden Cross
	}
	if prevFast >= prevSlow && currFast < currSlow {
		return SignalSell // Death Cross
	}
	return SignalHold
}

// EnsemblePredict combines multiple models for a robust prediction.
func EnsemblePredict(prices []float64, horizon int) *PredictionResult {
	lr := LinearRegression(prices, horizon)
	mc := MonteCarloSimulation(prices, horizon, 1000)

	if lr == nil && mc == nil {
		return nil
	}
	if lr == nil {
		return mc
	}
	if mc == nil {
		return lr
	}

	// Simple average ensemble
	expected := (lr.PredictedPrice + mc.PredictedPrice) / 2.0
	lower := (lr.LowerBound + mc.LowerBound) / 2.0
	upper := (lr.UpperBound + mc.UpperBound) / 2.0
	conf := (lr.Confidence + mc.Confidence) / 2.0

	lastPrice := prices[len(prices)-1]
	dir := "NEUTRAL"
	if expected > lastPrice {
		dir = "BULLISH"
	} else if expected < lastPrice {
		dir = "BEARISH"
	}

	return &PredictionResult{
		Model:          "ENSEMBLE_LR_MC",
		PredictedPrice: expected,
		Confidence:     conf,
		Direction:      dir,
		Horizon:        horizon,
		UpperBound:     upper,
		LowerBound:     lower,
		PricePath:      mc.PricePath,
		Timestamp:      time.Now(),
	}
}
