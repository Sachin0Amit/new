package finance

import (
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// STRATEGY INTERFACE
// ============================================================================

// Strategy defines the interface for a trading strategy used in backtesting.
type Strategy interface {
	// Name returns the strategy's identifier.
	Name() string
	// OnBar is called for each new bar. It returns a trade action or nil.
	OnBar(idx int, ts TimeSeries, indicators map[string][]float64) *TradeOrder
}

// TradeOrder represents a pending order from a strategy.
type TradeOrder struct {
	Action   TradeAction
	Quantity float64
	Reason   string
}

// ============================================================================
// BACKTEST ENGINE
// ============================================================================

// BacktestEngine runs strategy simulations over historical data.
type BacktestEngine struct {
	config   BacktestConfig
	strategy Strategy
}

// NewBacktestEngine creates a new backtesting engine.
func NewBacktestEngine(config BacktestConfig, strategy Strategy) *BacktestEngine {
	return &BacktestEngine{
		config:   config,
		strategy: strategy,
	}
}

// Run executes the backtest over the given time series.
func (be *BacktestEngine) Run(ts TimeSeries) *BacktestResult {
	startTime := time.Now()

	capital := be.config.InitialCapital
	position := 0.0
	avgEntryPrice := 0.0
	trades := make([]Trade, 0, 128)
	equityCurve := make([]float64, 0, len(ts))

	// Pre-compute indicators
	closes := ts.Closes()
	highs := ts.Highs()
	lows := ts.Lows()
	volumes := ts.Volumes()

	indicators := map[string][]float64{
		"sma_20":  SMA(closes, 20),
		"sma_50":  SMA(closes, 50),
		"sma_200": SMA(closes, 200),
		"ema_12":  EMA(closes, 12),
		"ema_26":  EMA(closes, 26),
		"rsi_14":  RSI(closes, 14),
		"atr_14":  ATR(highs, lows, closes, 14),
		"obv":     OBV(closes, volumes),
	}

	// Bollinger Bands
	bbUpper, bbMiddle, bbLower := BollingerBands(closes, 20, 2.0)
	indicators["bb_upper"] = bbUpper
	indicators["bb_middle"] = bbMiddle
	indicators["bb_lower"] = bbLower

	// MACD
	macdLine, signalLine, histogram := MACD(closes, 12, 26, 9)
	indicators["macd"] = macdLine
	indicators["macd_signal"] = signalLine
	indicators["macd_hist"] = histogram

	for i := 0; i < len(ts); i++ {
		bar := ts[i]
		currentPrice := bar.Close

		// Calculate equity at this point
		equity := capital + (position * currentPrice)
		equityCurve = append(equityCurve, equity)

		// Get strategy decision
		order := be.strategy.OnBar(i, ts, indicators)
		if order == nil {
			continue
		}

		switch order.Action {
		case ActionBuy:
			if position > 0 {
				continue // Already holding
			}

			// Position sizing: use config max position percentage
			maxInvest := capital * be.config.MaxPositionPct
			if maxInvest <= 0 {
				maxInvest = capital * 0.95 // Default to 95%
			}

			// Apply slippage
			effectivePrice := currentPrice * (1.0 + be.config.Slippage/100.0)
			commission := maxInvest * (be.config.Commission / 100.0)
			investable := maxInvest - commission

			qty := investable / effectivePrice
			if qty <= 0 {
				continue
			}

			position = qty
			avgEntryPrice = effectivePrice
			capital -= (qty * effectivePrice) + commission

			trades = append(trades, Trade{
				ID:        uuid.New(),
				Symbol:    be.config.Symbol,
				Action:    ActionBuy,
				Price:     effectivePrice,
				Quantity:  qty,
				Timestamp: bar.Timestamp,
				Reason:    order.Reason,
			})

		case ActionSell:
			if position <= 0 {
				continue // Nothing to sell
			}

			// Apply slippage (negative for selling)
			effectivePrice := currentPrice * (1.0 - be.config.Slippage/100.0)
			proceeds := position * effectivePrice
			commission := proceeds * (be.config.Commission / 100.0)

			pnl := (effectivePrice - avgEntryPrice) * position - commission

			capital += proceeds - commission

			trades = append(trades, Trade{
				ID:        uuid.New(),
				Symbol:    be.config.Symbol,
				Action:    ActionSell,
				Price:     effectivePrice,
				Quantity:  position,
				Timestamp: bar.Timestamp,
				PnL:       pnl,
				Reason:    order.Reason,
			})

			position = 0
			avgEntryPrice = 0
		}
	}

	// Close any open position at the last price
	if position > 0 && len(ts) > 0 {
		lastPrice := ts[len(ts)-1].Close
		proceeds := position * lastPrice
		commission := proceeds * (be.config.Commission / 100.0)
		pnl := (lastPrice - avgEntryPrice) * position - commission
		capital += proceeds - commission

		trades = append(trades, Trade{
			ID:        uuid.New(),
			Symbol:    be.config.Symbol,
			Action:    ActionSell,
			Price:     lastPrice,
			Quantity:  position,
			Timestamp: ts[len(ts)-1].Timestamp,
			PnL:       pnl,
			Reason:    "BACKTEST_EXIT",
		})
		position = 0
	}

	// Compile results
	result := &BacktestResult{
		Config:      be.config,
		Trades:      trades,
		FinalCapital: capital,
		EquityCurve: equityCurve,
		Duration:    time.Since(startTime),
	}

	result.TotalReturn = (capital - be.config.InitialCapital) / be.config.InitialCapital

	// Compute days for annualization
	if len(ts) >= 2 {
		days := ts[len(ts)-1].Timestamp.Sub(ts[0].Timestamp).Hours() / 24.0
		if days > 0 {
			result.AnnualReturn = AnnualizeReturn(result.TotalReturn, int(days))
		}
	}

	// Trade statistics
	result.TradeCount = len(trades)
	winTotal := 0.0
	lossTotal := 0.0
	for _, t := range trades {
		if t.Action == ActionSell {
			if t.PnL > 0 {
				result.WinCount++
				winTotal += t.PnL
				if t.PnL > result.LargestWin {
					result.LargestWin = t.PnL
				}
			} else if t.PnL < 0 {
				result.LossCount++
				lossTotal += math.Abs(t.PnL)
				if math.Abs(t.PnL) > math.Abs(result.LargestLoss) {
					result.LargestLoss = t.PnL
				}
			}
		}
	}

	if result.WinCount > 0 {
		result.AvgWin = winTotal / float64(result.WinCount)
	}
	if result.LossCount > 0 {
		result.AvgLoss = lossTotal / float64(result.LossCount)
	}

	// Risk metrics from equity curve
	if len(equityCurve) > 1 {
		eqReturns := make([]float64, len(equityCurve)-1)
		for i := 1; i < len(equityCurve); i++ {
			if equityCurve[i-1] > 0 {
				eqReturns[i-1] = math.Log(equityCurve[i] / equityCurve[i-1])
			}
		}
		benchReturns := make([]float64, len(eqReturns)) // Flat benchmark
		result.Risk = *CalculateRiskMetrics(equityCurve, eqReturns, benchReturns, 0.04)
	}

	return result
}

// ============================================================================
// BUILT-IN STRATEGIES
// ============================================================================

// SMACrossStrategy is a simple dual-SMA crossover strategy.
type SMACrossStrategy struct {
	FastPeriod int
	SlowPeriod int
}

func (s *SMACrossStrategy) Name() string {
	return fmt.Sprintf("SMA_CROSS_%d_%d", s.FastPeriod, s.SlowPeriod)
}

func (s *SMACrossStrategy) OnBar(idx int, ts TimeSeries, ind map[string][]float64) *TradeOrder {
	fastKey := fmt.Sprintf("sma_%d", s.FastPeriod)
	slowKey := fmt.Sprintf("sma_%d", s.SlowPeriod)

	fast, hasFast := ind[fastKey]
	slow, hasSlow := ind[slowKey]

	if !hasFast || !hasSlow || idx < 1 {
		return nil
	}

	if math.IsNaN(fast[idx]) || math.IsNaN(slow[idx]) || math.IsNaN(fast[idx-1]) || math.IsNaN(slow[idx-1]) {
		return nil
	}

	// Golden Cross
	if fast[idx-1] <= slow[idx-1] && fast[idx] > slow[idx] {
		return &TradeOrder{Action: ActionBuy, Reason: "GOLDEN_CROSS"}
	}
	// Death Cross
	if fast[idx-1] >= slow[idx-1] && fast[idx] < slow[idx] {
		return &TradeOrder{Action: ActionSell, Reason: "DEATH_CROSS"}
	}

	return nil
}

// RSIMeanReversionStrategy buys when RSI is oversold and sells when overbought.
type RSIMeanReversionStrategy struct {
	Period       int
	OversoldLvl  float64
	OverboughtLvl float64
}

func (s *RSIMeanReversionStrategy) Name() string {
	return fmt.Sprintf("RSI_MEANREV_%d", s.Period)
}

func (s *RSIMeanReversionStrategy) OnBar(idx int, ts TimeSeries, ind map[string][]float64) *TradeOrder {
	rsi, has := ind[fmt.Sprintf("rsi_%d", s.Period)]
	if !has || idx < 1 || math.IsNaN(rsi[idx]) {
		return nil
	}

	if rsi[idx] < s.OversoldLvl {
		return &TradeOrder{Action: ActionBuy, Reason: fmt.Sprintf("RSI_OVERSOLD_%.0f", rsi[idx])}
	}
	if rsi[idx] > s.OverboughtLvl {
		return &TradeOrder{Action: ActionSell, Reason: fmt.Sprintf("RSI_OVERBOUGHT_%.0f", rsi[idx])}
	}
	return nil
}

// BollingerBounceStrategy buys at lower band, sells at upper band.
type BollingerBounceStrategy struct{}

func (s *BollingerBounceStrategy) Name() string {
	return "BOLLINGER_BOUNCE"
}

func (s *BollingerBounceStrategy) OnBar(idx int, ts TimeSeries, ind map[string][]float64) *TradeOrder {
	lower, hasLower := ind["bb_lower"]
	upper, hasUpper := ind["bb_upper"]
	if !hasLower || !hasUpper || math.IsNaN(lower[idx]) || math.IsNaN(upper[idx]) {
		return nil
	}

	price := ts[idx].Close

	if price <= lower[idx] {
		return &TradeOrder{Action: ActionBuy, Reason: "BOLLINGER_LOWER_TOUCH"}
	}
	if price >= upper[idx] {
		return &TradeOrder{Action: ActionSell, Reason: "BOLLINGER_UPPER_TOUCH"}
	}
	return nil
}

// MACDStrategy trades on MACD histogram zero-line crossovers.
type MACDStrategy struct{}

func (s *MACDStrategy) Name() string {
	return "MACD_HISTOGRAM"
}

func (s *MACDStrategy) OnBar(idx int, ts TimeSeries, ind map[string][]float64) *TradeOrder {
	hist, has := ind["macd_hist"]
	if !has || idx < 1 || math.IsNaN(hist[idx]) || math.IsNaN(hist[idx-1]) {
		return nil
	}

	// Histogram crosses above zero
	if hist[idx-1] <= 0 && hist[idx] > 0 {
		return &TradeOrder{Action: ActionBuy, Reason: "MACD_HIST_BULLISH"}
	}
	// Histogram crosses below zero
	if hist[idx-1] >= 0 && hist[idx] < 0 {
		return &TradeOrder{Action: ActionSell, Reason: "MACD_HIST_BEARISH"}
	}
	return nil
}
