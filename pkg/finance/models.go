// Package finance implements a sovereign, local-first financial intelligence engine.
// Zero external AI dependencies. Pure mathematical analysis, statistical prediction,
// and autonomous market reasoning built from first principles.
package finance

import (
	"math"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// CORE MARKET DATA STRUCTURES
// ============================================================================

// OHLCV represents a single candlestick bar with Open, High, Low, Close, Volume.
type OHLCV struct {
	Timestamp time.Time `json:"timestamp"`
	Open      float64   `json:"open"`
	High      float64   `json:"high"`
	Low       float64   `json:"low"`
	Close     float64   `json:"close"`
	Volume    float64   `json:"volume"`
	AdjClose  float64   `json:"adj_close,omitempty"`
}

// TimeSeries is an ordered slice of OHLCV bars.
type TimeSeries []OHLCV

// Closes extracts a float64 slice of closing prices.
func (ts TimeSeries) Closes() []float64 {
	out := make([]float64, len(ts))
	for i, bar := range ts {
		out[i] = bar.Close
	}
	return out
}

// Highs extracts a float64 slice of high prices.
func (ts TimeSeries) Highs() []float64 {
	out := make([]float64, len(ts))
	for i, bar := range ts {
		out[i] = bar.High
	}
	return out
}

// Lows extracts a float64 slice of low prices.
func (ts TimeSeries) Lows() []float64 {
	out := make([]float64, len(ts))
	for i, bar := range ts {
		out[i] = bar.Low
	}
	return out
}

// Volumes extracts a float64 slice of volumes.
func (ts TimeSeries) Volumes() []float64 {
	out := make([]float64, len(ts))
	for i, bar := range ts {
		out[i] = bar.Volume
	}
	return out
}

// Returns computes the logarithmic returns of the closing prices.
func (ts TimeSeries) Returns() []float64 {
	if len(ts) < 2 {
		return nil
	}
	ret := make([]float64, len(ts)-1)
	for i := 1; i < len(ts); i++ {
		if ts[i-1].Close > 0 {
			ret[i-1] = math.Log(ts[i].Close / ts[i-1].Close)
		}
	}
	return ret
}

// Last returns the most recent bar, or a zero OHLCV if empty.
func (ts TimeSeries) Last() OHLCV {
	if len(ts) == 0 {
		return OHLCV{}
	}
	return ts[len(ts)-1]
}

// Slice returns a sub-range of the time series.
func (ts TimeSeries) Slice(start, end int) TimeSeries {
	if start < 0 {
		start = 0
	}
	if end > len(ts) {
		end = len(ts)
	}
	return ts[start:end]
}

// ============================================================================
// INSTRUMENT & MARKET DEFINITIONS
// ============================================================================

// AssetClass categorizes financial instruments.
type AssetClass string

const (
	AssetEquity     AssetClass = "EQUITY"
	AssetForex      AssetClass = "FOREX"
	AssetCrypto     AssetClass = "CRYPTO"
	AssetCommodity  AssetClass = "COMMODITY"
	AssetIndex      AssetClass = "INDEX"
	AssetBond       AssetClass = "BOND"
	AssetETF        AssetClass = "ETF"
	AssetOption     AssetClass = "OPTION"
	AssetFuture     AssetClass = "FUTURE"
)

// Timeframe represents the period of each candle.
type Timeframe string

const (
	TF1Min   Timeframe = "1m"
	TF5Min   Timeframe = "5m"
	TF15Min  Timeframe = "15m"
	TF30Min  Timeframe = "30m"
	TF1Hour  Timeframe = "1h"
	TF4Hour  Timeframe = "4h"
	TFDaily  Timeframe = "1d"
	TFWeekly Timeframe = "1w"
	TFMonth  Timeframe = "1M"
)

// Instrument represents a tradable financial entity.
type Instrument struct {
	Symbol    string     `json:"symbol"`
	Name      string     `json:"name"`
	Exchange  string     `json:"exchange"`
	Class     AssetClass `json:"asset_class"`
	Currency  string     `json:"currency"`
	Sector    string     `json:"sector,omitempty"`
	Industry  string     `json:"industry,omitempty"`
	Country   string     `json:"country,omitempty"`
	MarketCap float64    `json:"market_cap,omitempty"`
}

// MarketRegion defines geographic market groupings.
type MarketRegion string

const (
	RegionUS     MarketRegion = "US"
	RegionEU     MarketRegion = "EU"
	RegionAsia   MarketRegion = "ASIA"
	RegionIndia  MarketRegion = "INDIA"
	RegionGlobal MarketRegion = "GLOBAL"
)

// ============================================================================
// ANALYSIS RESULTS
// ============================================================================

// IndicatorResult holds the output of a single technical indicator.
type IndicatorResult struct {
	Name      string    `json:"name"`
	Values    []float64 `json:"values"`
	Signal    string    `json:"signal"` // "BUY", "SELL", "NEUTRAL"
	Strength  float64   `json:"strength"` // 0.0 to 1.0
	Timestamp time.Time `json:"timestamp"`
}

// PredictionResult holds the output of a statistical prediction model.
type PredictionResult struct {
	Model          string    `json:"model"` // "LINEAR_REGRESSION", "MONTE_CARLO", etc.
	PredictedPrice float64   `json:"predicted_price"`
	Confidence     float64   `json:"confidence"` // 0.0 to 1.0
	Direction      string    `json:"direction"`  // "BULLISH", "BEARISH", "NEUTRAL"
	Horizon        int       `json:"horizon_days"`
	UpperBound     float64   `json:"upper_bound"` // Confidence interval
	LowerBound     float64   `json:"lower_bound"`
	PricePath      []float64 `json:"price_path,omitempty"` // Simulated path
	R2Score        float64   `json:"r2_score,omitempty"`
	RMSE           float64   `json:"rmse,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// SignalType represents the direction of a trading signal.
type SignalType string

const (
	SignalBuy     SignalType = "BUY"
	SignalSell    SignalType = "SELL"
	SignalHold    SignalType = "HOLD"
	SignalNeutral SignalType = "NEUTRAL"
)

// TradingSignal is a composite signal derived from multiple indicators.
type TradingSignal struct {
	ID         uuid.UUID           `json:"id"`
	Symbol     string              `json:"symbol"`
	Type       SignalType          `json:"type"`
	Confidence float64             `json:"confidence"`
	Price      float64             `json:"price"`
	StopLoss   float64             `json:"stop_loss,omitempty"`
	TakeProfit float64             `json:"take_profit,omitempty"`
	Indicators []IndicatorResult   `json:"indicators"`
	Prediction *PredictionResult   `json:"prediction,omitempty"`
	Risk       *RiskMetrics        `json:"risk,omitempty"`
	Timestamp  time.Time           `json:"timestamp"`
	Reasoning  string              `json:"reasoning"` // Human-readable explanation
}

// ============================================================================
// RISK METRICS
// ============================================================================

// RiskMetrics contains comprehensive risk analysis results.
type RiskMetrics struct {
	ValueAtRisk95    float64 `json:"var_95"`     // 95% VaR
	ValueAtRisk99    float64 `json:"var_99"`     // 99% VaR
	CVaR95           float64 `json:"cvar_95"`    // Conditional VaR
	SharpeRatio      float64 `json:"sharpe"`
	SortinoRatio     float64 `json:"sortino"`
	MaxDrawdown      float64 `json:"max_drawdown"`
	MaxDrawdownDur   int     `json:"max_drawdown_duration_days"`
	Beta             float64 `json:"beta"`
	Alpha            float64 `json:"alpha"`
	Volatility       float64 `json:"volatility_annual"`
	VolatilityDaily  float64 `json:"volatility_daily"`
	CalmarRatio      float64 `json:"calmar"`
	InformationRatio float64 `json:"information_ratio"`
	TreynorRatio     float64 `json:"treynor"`
	OmegaRatio       float64 `json:"omega"`
	TailRatio        float64 `json:"tail_ratio"`
	Skewness         float64 `json:"skewness"`
	Kurtosis         float64 `json:"kurtosis"`
	WinRate          float64 `json:"win_rate"`
	ProfitFactor     float64 `json:"profit_factor"`
}

// ============================================================================
// PORTFOLIO STRUCTURES
// ============================================================================

// Position represents a single holding in a portfolio.
type Position struct {
	Symbol      string    `json:"symbol"`
	Quantity    float64   `json:"quantity"`
	AvgPrice    float64   `json:"avg_price"`
	CurrentPrice float64  `json:"current_price"`
	PnL         float64   `json:"pnl"`
	PnLPercent  float64   `json:"pnl_percent"`
	Weight      float64   `json:"weight"` // Portfolio weight
	OpenedAt    time.Time `json:"opened_at"`
}

// Portfolio represents a collection of positions.
type Portfolio struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Positions []Position `json:"positions"`
	Cash      float64    `json:"cash"`
	TotalValue float64   `json:"total_value"`
	Risk      *RiskMetrics `json:"risk,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TotalValue computes the total portfolio value including cash.
func (p *Portfolio) ComputeValue() float64 {
	total := p.Cash
	for _, pos := range p.Positions {
		total += pos.CurrentPrice * pos.Quantity
	}
	p.TotalValue = total
	return total
}

// ============================================================================
// BACKTESTING STRUCTURES
// ============================================================================

// TradeAction defines an action taken by a strategy.
type TradeAction string

const (
	ActionBuy  TradeAction = "BUY"
	ActionSell TradeAction = "SELL"
)

// Trade represents a single executed trade in a backtest.
type Trade struct {
	ID        uuid.UUID   `json:"id"`
	Symbol    string      `json:"symbol"`
	Action    TradeAction `json:"action"`
	Price     float64     `json:"price"`
	Quantity  float64     `json:"quantity"`
	Timestamp time.Time   `json:"timestamp"`
	PnL       float64     `json:"pnl"`
	Reason    string      `json:"reason"`
}

// BacktestConfig defines the parameters for a backtest run.
type BacktestConfig struct {
	Symbol         string    `json:"symbol"`
	StartDate      time.Time `json:"start_date"`
	EndDate        time.Time `json:"end_date"`
	InitialCapital float64  `json:"initial_capital"`
	Commission     float64  `json:"commission_percent"` // Per trade
	Slippage       float64  `json:"slippage_percent"`
	MaxPositionPct float64  `json:"max_position_pct"` // Max % of capital per position
}

// BacktestResult contains the output of a strategy backtest.
type BacktestResult struct {
	Config        BacktestConfig `json:"config"`
	Trades        []Trade        `json:"trades"`
	FinalCapital  float64        `json:"final_capital"`
	TotalReturn   float64        `json:"total_return"`
	AnnualReturn  float64        `json:"annual_return"`
	Risk          RiskMetrics    `json:"risk"`
	EquityCurve   []float64      `json:"equity_curve"`
	TradeCount    int            `json:"trade_count"`
	WinCount      int            `json:"win_count"`
	LossCount     int            `json:"loss_count"`
	AvgWin        float64        `json:"avg_win"`
	AvgLoss       float64        `json:"avg_loss"`
	LargestWin    float64        `json:"largest_win"`
	LargestLoss   float64        `json:"largest_loss"`
	Duration      time.Duration  `json:"duration"`
}

// ============================================================================
// PATTERN RECOGNITION
// ============================================================================

// PatternType defines recognized chart patterns.
type PatternType string

const (
	PatternHeadShoulders    PatternType = "HEAD_AND_SHOULDERS"
	PatternInvHeadShoulders PatternType = "INVERSE_HEAD_AND_SHOULDERS"
	PatternDoubleTop        PatternType = "DOUBLE_TOP"
	PatternDoubleBottom     PatternType = "DOUBLE_BOTTOM"
	PatternTripleTop        PatternType = "TRIPLE_TOP"
	PatternTripleBottom     PatternType = "TRIPLE_BOTTOM"
	PatternAscTriangle      PatternType = "ASCENDING_TRIANGLE"
	PatternDescTriangle     PatternType = "DESCENDING_TRIANGLE"
	PatternSymTriangle      PatternType = "SYMMETRICAL_TRIANGLE"
	PatternBullFlag         PatternType = "BULL_FLAG"
	PatternBearFlag         PatternType = "BEAR_FLAG"
	PatternBullPennant      PatternType = "BULL_PENNANT"
	PatternBearPennant      PatternType = "BEAR_PENNANT"
	PatternWedgeRising      PatternType = "RISING_WEDGE"
	PatternWedgeFalling     PatternType = "FALLING_WEDGE"
	PatternCupHandle        PatternType = "CUP_AND_HANDLE"
	PatternRounding         PatternType = "ROUNDING_BOTTOM"
	PatternChannel          PatternType = "CHANNEL"
	PatternDoji             PatternType = "DOJI"
	PatternHammer           PatternType = "HAMMER"
	PatternShootingStar     PatternType = "SHOOTING_STAR"
	PatternEngulfing        PatternType = "ENGULFING"
	PatternMorningStar      PatternType = "MORNING_STAR"
	PatternEveningStar      PatternType = "EVENING_STAR"
	PatternThreeWhiteSoldrs PatternType = "THREE_WHITE_SOLDIERS"
	PatternThreeBlackCrows  PatternType = "THREE_BLACK_CROWS"
	PatternHarami           PatternType = "HARAMI"
	PatternMarubozu         PatternType = "MARUBOZU"
	PatternSpinningTop      PatternType = "SPINNING_TOP"
	PatternTweezerTop       PatternType = "TWEEZER_TOP"
	PatternTweezerBottom    PatternType = "TWEEZER_BOTTOM"
)

// PatternDetection represents a detected chart or candlestick pattern.
type PatternDetection struct {
	Pattern    PatternType `json:"pattern"`
	Confidence float64    `json:"confidence"`
	Direction  string     `json:"direction"` // "BULLISH" or "BEARISH"
	StartIndex int        `json:"start_index"`
	EndIndex   int        `json:"end_index"`
	Timestamp  time.Time  `json:"timestamp"`
}

// ============================================================================
// MARKET SCREENING
// ============================================================================

// ScreenerFilter defines a single screening criterion.
type ScreenerFilter struct {
	Field    string  `json:"field"` // "rsi", "macd_signal", "volume_ratio", "price_change_pct"
	Operator string  `json:"operator"` // ">", "<", ">=", "<=", "==", "between"
	Value    float64 `json:"value"`
	Value2   float64 `json:"value2,omitempty"` // For "between" operator
}

// ScreenerResult represents one instrument that passed screening.
type ScreenerResult struct {
	Instrument Instrument  `json:"instrument"`
	Indicators map[string]float64 `json:"indicators"`
	Signal     SignalType  `json:"signal"`
	Score      float64     `json:"score"` // Composite score 0-100
}

// ============================================================================
// CORRELATION & MULTI-MARKET
// ============================================================================

// CorrelationMatrix stores pairwise correlation coefficients.
type CorrelationMatrix struct {
	Symbols []string    `json:"symbols"`
	Matrix  [][]float64 `json:"matrix"`
}

// MarketBreadth summarizes the overall market condition.
type MarketBreadth struct {
	AdvancingCount  int     `json:"advancing"`
	DecliningCount  int     `json:"declining"`
	UnchangedCount  int     `json:"unchanged"`
	AdvanceDecline  float64 `json:"advance_decline_ratio"`
	NewHighs        int     `json:"new_highs"`
	NewLows         int     `json:"new_lows"`
	HighLowRatio    float64 `json:"high_low_ratio"`
	BullishPercent  float64 `json:"bullish_percent"`
	MarketSentiment string  `json:"sentiment"` // "FEAR", "GREED", "NEUTRAL"
	FearGreedIndex  float64 `json:"fear_greed_index"` // 0=Extreme Fear, 100=Extreme Greed
	Timestamp       time.Time `json:"timestamp"`
}

// ============================================================================
// MARKET DATA CACHE (Thread-Safe)
// ============================================================================

// MarketDataCache provides concurrent-safe in-memory storage for market data.
type MarketDataCache struct {
	mu   sync.RWMutex
	data map[string]TimeSeries // key: "SYMBOL:TIMEFRAME"
}

// NewMarketDataCache initializes the cache.
func NewMarketDataCache() *MarketDataCache {
	return &MarketDataCache{
		data: make(map[string]TimeSeries),
	}
}

// Put stores a time series in the cache.
func (c *MarketDataCache) Put(symbol string, tf Timeframe, ts TimeSeries) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[symbol+":"+string(tf)] = ts
}

// Get retrieves a time series from the cache.
func (c *MarketDataCache) Get(symbol string, tf Timeframe) (TimeSeries, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ts, ok := c.data[symbol+":"+string(tf)]
	return ts, ok
}

// Symbols returns all cached symbols.
func (c *MarketDataCache) Symbols() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	seen := make(map[string]bool)
	for key := range c.data {
		for i, ch := range key {
			if ch == ':' {
				seen[key[:i]] = true
				break
			}
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

// ============================================================================
// STATISTICAL HELPERS (used across the package)
// ============================================================================

// Mean computes the arithmetic mean of a slice.
func Mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// StdDev computes the sample standard deviation.
func StdDev(data []float64) float64 {
	if len(data) < 2 {
		return 0
	}
	m := Mean(data)
	sum := 0.0
	for _, v := range data {
		d := v - m
		sum += d * d
	}
	return math.Sqrt(sum / float64(len(data)-1))
}

// Variance computes the sample variance.
func Variance(data []float64) float64 {
	s := StdDev(data)
	return s * s
}

// Covariance computes the sample covariance between two slices.
func Covariance(x, y []float64) float64 {
	n := len(x)
	if n != len(y) || n < 2 {
		return 0
	}
	mx, my := Mean(x), Mean(y)
	sum := 0.0
	for i := 0; i < n; i++ {
		sum += (x[i] - mx) * (y[i] - my)
	}
	return sum / float64(n-1)
}

// Correlation computes Pearson correlation coefficient.
func Correlation(x, y []float64) float64 {
	sx, sy := StdDev(x), StdDev(y)
	if sx == 0 || sy == 0 {
		return 0
	}
	return Covariance(x, y) / (sx * sy)
}

// Percentile computes the p-th percentile of a sorted slice (p in 0-100).
func Percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	k := (p / 100.0) * float64(len(sorted)-1)
	f := math.Floor(k)
	c := math.Ceil(k)
	if f == c {
		return sorted[int(k)]
	}
	d0 := sorted[int(f)] * (c - k)
	d1 := sorted[int(c)] * (k - f)
	return d0 + d1
}

// SortedCopy returns a sorted copy of the input slice.
func SortedCopy(data []float64) []float64 {
	cp := make([]float64, len(data))
	copy(cp, data)
	sort.Float64s(cp)
	return cp
}

// Min returns the minimum of a slice.
func Min(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	m := data[0]
	for _, v := range data[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

// Max returns the maximum of a slice.
func Max(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	m := data[0]
	for _, v := range data[1:] {
		if v > m {
			m = v
		}
	}
	return m
}

// Sum returns the sum of all elements.
func Sum(data []float64) float64 {
	s := 0.0
	for _, v := range data {
		s += v
	}
	return s
}

// CumulativeSum returns the running total.
func CumulativeSum(data []float64) []float64 {
	out := make([]float64, len(data))
	sum := 0.0
	for i, v := range data {
		sum += v
		out[i] = sum
	}
	return out
}

// LinearInterpolate performs simple linear interpolation.
func LinearInterpolate(x0, y0, x1, y1, x float64) float64 {
	if x1 == x0 {
		return y0
	}
	return y0 + (y1-y0)*(x-x0)/(x1-x0)
}
