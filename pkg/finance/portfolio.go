package finance

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// PORTFOLIO OPTIMIZER (Markowitz Mean-Variance)
// ============================================================================

// PortfolioOptimizer implements a simplified Markowitz mean-variance optimization.
type PortfolioOptimizer struct {
	assets       []string
	returnSeries map[string][]float64 // key=symbol, val=daily returns
}

// NewPortfolioOptimizer initializes the optimizer with asset return series.
func NewPortfolioOptimizer(assets map[string][]float64) *PortfolioOptimizer {
	names := make([]string, 0, len(assets))
	for k := range assets {
		names = append(names, k)
	}
	sort.Strings(names)

	return &PortfolioOptimizer{
		assets:       names,
		returnSeries: assets,
	}
}

// OptimizedAllocation contains the result of the optimization.
type OptimizedAllocation struct {
	Weights        map[string]float64 `json:"weights"`
	ExpectedReturn float64            `json:"expected_return_annual"`
	ExpectedRisk   float64            `json:"expected_risk_annual"`
	SharpeRatio    float64            `json:"sharpe_ratio"`
}

// MinimumVariancePortfolio finds the minimum-variance portfolio
// using a grid-search approach (suitable for small asset counts).
func (po *PortfolioOptimizer) MinimumVariancePortfolio(riskFreeRate float64) *OptimizedAllocation {
	n := len(po.assets)
	if n == 0 {
		return nil
	}
	if n == 1 {
		ret := po.returnSeries[po.assets[0]]
		meanReturn := Mean(ret)
		annReturn := math.Pow(1.0+meanReturn, TradingDaysPerYear) - 1.0
		annVol := StdDev(ret) * math.Sqrt(TradingDaysPerYear)
		return &OptimizedAllocation{
			Weights:        map[string]float64{po.assets[0]: 1.0},
			ExpectedReturn: annReturn,
			ExpectedRisk:   annVol,
			SharpeRatio:    (annReturn - riskFreeRate) / annVol,
		}
	}

	// Pre-compute means and covariance matrix
	means := make([]float64, n)
	for i, sym := range po.assets {
		means[i] = Mean(po.returnSeries[sym])
	}

	covMatrix := make([][]float64, n)
	for i := 0; i < n; i++ {
		covMatrix[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			covMatrix[i][j] = Covariance(po.returnSeries[po.assets[i]], po.returnSeries[po.assets[j]])
		}
	}

	// For 2-4 assets, do a fine grid search. For more, use equal weight.
	bestWeights := make([]float64, n)
	bestSharpe := math.Inf(-1)
	bestReturn := 0.0
	bestRisk := 0.0

	if n == 2 {
		// Grid search w1 from 0 to 1
		for w := 0.0; w <= 1.0; w += 0.01 {
			weights := []float64{w, 1.0 - w}
			ret, risk := po.portfolioStats(weights, means, covMatrix)
			sharpe := (ret - riskFreeRate) / risk
			if sharpe > bestSharpe {
				bestSharpe = sharpe
				bestReturn = ret
				bestRisk = risk
				copy(bestWeights, weights)
			}
		}
	} else if n == 3 {
		for w1 := 0.0; w1 <= 1.0; w1 += 0.05 {
			for w2 := 0.0; w2 <= 1.0-w1; w2 += 0.05 {
				w3 := 1.0 - w1 - w2
				weights := []float64{w1, w2, w3}
				ret, risk := po.portfolioStats(weights, means, covMatrix)
				sharpe := (ret - riskFreeRate) / risk
				if sharpe > bestSharpe {
					bestSharpe = sharpe
					bestReturn = ret
					bestRisk = risk
					copy(bestWeights, weights)
				}
			}
		}
	} else if n == 4 {
		for w1 := 0.0; w1 <= 1.0; w1 += 0.1 {
			for w2 := 0.0; w2 <= 1.0-w1; w2 += 0.1 {
				for w3 := 0.0; w3 <= 1.0-w1-w2; w3 += 0.1 {
					w4 := 1.0 - w1 - w2 - w3
					weights := []float64{w1, w2, w3, w4}
					ret, risk := po.portfolioStats(weights, means, covMatrix)
					sharpe := (ret - riskFreeRate) / risk
					if sharpe > bestSharpe {
						bestSharpe = sharpe
						bestReturn = ret
						bestRisk = risk
						copy(bestWeights, weights)
					}
				}
			}
		}
	} else {
		// Equal weight fallback for >4 assets
		w := 1.0 / float64(n)
		for i := range bestWeights {
			bestWeights[i] = w
		}
		bestReturn, bestRisk = po.portfolioStats(bestWeights, means, covMatrix)
		bestSharpe = (bestReturn - riskFreeRate) / bestRisk
	}

	result := &OptimizedAllocation{
		Weights:        make(map[string]float64, n),
		ExpectedReturn: bestReturn,
		ExpectedRisk:   bestRisk,
		SharpeRatio:    bestSharpe,
	}
	for i, sym := range po.assets {
		result.Weights[sym] = math.Round(bestWeights[i]*1000) / 1000
	}
	return result
}

// portfolioStats computes annualized return and risk for given weights.
func (po *PortfolioOptimizer) portfolioStats(weights, means []float64, covMatrix [][]float64) (float64, float64) {
	n := len(weights)

	// Portfolio daily return: w . mu
	portReturn := 0.0
	for i := 0; i < n; i++ {
		portReturn += weights[i] * means[i]
	}

	// Portfolio daily variance: w' * Sigma * w
	portVariance := 0.0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			portVariance += weights[i] * weights[j] * covMatrix[i][j]
		}
	}

	annReturn := math.Pow(1.0+portReturn, TradingDaysPerYear) - 1.0
	annRisk := math.Sqrt(portVariance) * math.Sqrt(TradingDaysPerYear)

	if annRisk == 0 {
		annRisk = 0.0001 // Prevent division by zero
	}

	return annReturn, annRisk
}

// ============================================================================
// PORTFOLIO MANAGEMENT
// ============================================================================

// NewPortfolio creates an empty portfolio with initial cash.
func NewPortfolio(name string, cash float64) *Portfolio {
	return &Portfolio{
		ID:        uuid.New(),
		Name:      name,
		Positions: make([]Position, 0),
		Cash:      cash,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// AddPosition adds a new holding to the portfolio.
func (p *Portfolio) AddPosition(symbol string, qty, price float64) {
	// Check if position already exists
	for i := range p.Positions {
		if p.Positions[i].Symbol == symbol {
			// Average in
			totalQty := p.Positions[i].Quantity + qty
			totalCost := (p.Positions[i].Quantity * p.Positions[i].AvgPrice) + (qty * price)
			p.Positions[i].AvgPrice = totalCost / totalQty
			p.Positions[i].Quantity = totalQty
			p.Cash -= qty * price
			p.UpdatedAt = time.Now()
			return
		}
	}

	// New position
	p.Positions = append(p.Positions, Position{
		Symbol:   symbol,
		Quantity: qty,
		AvgPrice: price,
		OpenedAt: time.Now(),
	})
	p.Cash -= qty * price
	p.UpdatedAt = time.Now()
}

// ClosePosition sells an entire position.
func (p *Portfolio) ClosePosition(symbol string, currentPrice float64) float64 {
	for i := range p.Positions {
		if p.Positions[i].Symbol == symbol {
			proceeds := p.Positions[i].Quantity * currentPrice
			pnl := (currentPrice - p.Positions[i].AvgPrice) * p.Positions[i].Quantity
			p.Cash += proceeds
			// Remove position
			p.Positions = append(p.Positions[:i], p.Positions[i+1:]...)
			p.UpdatedAt = time.Now()
			return pnl
		}
	}
	return 0
}

// UpdatePrices refreshes current prices and PnL for all positions.
func (p *Portfolio) UpdatePrices(prices map[string]float64) {
	totalValue := p.Cash
	for i := range p.Positions {
		if price, ok := prices[p.Positions[i].Symbol]; ok {
			p.Positions[i].CurrentPrice = price
			p.Positions[i].PnL = (price - p.Positions[i].AvgPrice) * p.Positions[i].Quantity
			if p.Positions[i].AvgPrice > 0 {
				p.Positions[i].PnLPercent = (price - p.Positions[i].AvgPrice) / p.Positions[i].AvgPrice * 100
			}
		}
		totalValue += p.Positions[i].CurrentPrice * p.Positions[i].Quantity
	}
	p.TotalValue = totalValue

	// Update weights
	for i := range p.Positions {
		if totalValue > 0 {
			p.Positions[i].Weight = (p.Positions[i].CurrentPrice * p.Positions[i].Quantity) / totalValue
		}
	}
	p.UpdatedAt = time.Now()
}
