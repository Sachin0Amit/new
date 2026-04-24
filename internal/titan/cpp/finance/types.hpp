#ifndef SOVEREIGN_FINANCE_TYPES_H
#define SOVEREIGN_FINANCE_TYPES_H

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Core Data Types
 *  High-performance C++ financial data structures optimized for cache locality,
 *  SIMD-friendly alignment, and zero-copy operations.
 * ============================================================================
 */

#include <cstdint>
#include <cstddef>
#include <cmath>
#include <vector>
#include <string>
#include <array>
#include <algorithm>
#include <numeric>
#include <limits>
#include <chrono>
#include <unordered_map>

namespace sovereign {
namespace finance {

// ============================================================================
// CONSTANTS
// ============================================================================

constexpr double TRADING_DAYS_PER_YEAR = 252.0;
constexpr double PI = 3.14159265358979323846;
constexpr double SQRT_252 = 15.874507866387544; // sqrt(252)
constexpr double INV_SQRT_2PI = 0.3989422804014327; // 1 / sqrt(2*pi)

// ============================================================================
// OHLCV BAR
// ============================================================================

/// A single price bar. 48 bytes, aligned for vectorized access.
struct alignas(64) Bar {
    int64_t timestamp;  // Unix timestamp (seconds)
    double  open;
    double  high;
    double  low;
    double  close;
    double  volume;

    /// Returns the body length (absolute difference between open and close).
    inline double body() const { return std::abs(close - open); }

    /// Returns the upper shadow.
    inline double upper_shadow() const {
        return high - std::max(open, close);
    }

    /// Returns the lower shadow.
    inline double lower_shadow() const {
        return std::min(open, close) - low;
    }

    /// Returns the full range (high - low).
    inline double range() const { return high - low; }

    /// Returns the typical price (H+L+C)/3.
    inline double typical_price() const { return (high + low + close) / 3.0; }

    /// Returns true if this is a bullish (green) bar.
    inline bool is_bullish() const { return close > open; }

    /// Returns true if this is a bearish (red) bar.
    inline bool is_bearish() const { return close < open; }

    /// Returns true if this is a doji (body < 10% of range).
    inline bool is_doji() const {
        double r = range();
        return r > 0.0 && body() < r * 0.1;
    }
};

// ============================================================================
// TIME SERIES
// ============================================================================

/// A contiguous, cache-friendly time series of OHLCV bars.
class TimeSeries {
public:
    TimeSeries() = default;
    explicit TimeSeries(size_t capacity) { bars_.reserve(capacity); }

    // Core accessors
    inline size_t size() const { return bars_.size(); }
    inline bool empty() const { return bars_.empty(); }
    inline const Bar& operator[](size_t i) const { return bars_[i]; }
    inline Bar& operator[](size_t i) { return bars_[i]; }
    inline const Bar& back() const { return bars_.back(); }
    inline const Bar& front() const { return bars_.front(); }

    // Data access
    inline const std::vector<Bar>& bars() const { return bars_; }
    inline void push_back(const Bar& bar) { bars_.push_back(bar); }
    inline void reserve(size_t n) { bars_.reserve(n); }
    inline void clear() { bars_.clear(); }

    /// Extract closing prices into a flat vector (SIMD-friendly).
    std::vector<double> closes() const {
        std::vector<double> out(bars_.size());
        for (size_t i = 0; i < bars_.size(); ++i) {
            out[i] = bars_[i].close;
        }
        return out;
    }

    /// Extract high prices.
    std::vector<double> highs() const {
        std::vector<double> out(bars_.size());
        for (size_t i = 0; i < bars_.size(); ++i) {
            out[i] = bars_[i].high;
        }
        return out;
    }

    /// Extract low prices.
    std::vector<double> lows() const {
        std::vector<double> out(bars_.size());
        for (size_t i = 0; i < bars_.size(); ++i) {
            out[i] = bars_[i].low;
        }
        return out;
    }

    /// Extract volumes.
    std::vector<double> volumes() const {
        std::vector<double> out(bars_.size());
        for (size_t i = 0; i < bars_.size(); ++i) {
            out[i] = bars_[i].volume;
        }
        return out;
    }

    /// Compute log returns.
    std::vector<double> log_returns() const {
        if (bars_.size() < 2) return {};
        std::vector<double> ret(bars_.size() - 1);
        for (size_t i = 1; i < bars_.size(); ++i) {
            if (bars_[i - 1].close > 0.0) {
                ret[i - 1] = std::log(bars_[i].close / bars_[i - 1].close);
            }
        }
        return ret;
    }

    /// Compute simple returns.
    std::vector<double> simple_returns() const {
        if (bars_.size() < 2) return {};
        std::vector<double> ret(bars_.size() - 1);
        for (size_t i = 1; i < bars_.size(); ++i) {
            if (bars_[i - 1].close > 0.0) {
                ret[i - 1] = (bars_[i].close - bars_[i - 1].close) / bars_[i - 1].close;
            }
        }
        return ret;
    }

    /// Sub-slice (view-like, copies for now).
    TimeSeries slice(size_t start, size_t end) const {
        TimeSeries ts;
        if (start >= bars_.size()) return ts;
        if (end > bars_.size()) end = bars_.size();
        ts.bars_.assign(bars_.begin() + start, bars_.begin() + end);
        return ts;
    }

private:
    std::vector<Bar> bars_;
};

// ============================================================================
// SIGNAL TYPES
// ============================================================================

enum class Signal : int8_t {
    STRONG_SELL = -2,
    SELL        = -1,
    NEUTRAL     =  0,
    BUY         =  1,
    STRONG_BUY  =  2
};

inline const char* signal_to_string(Signal s) {
    switch (s) {
        case Signal::STRONG_SELL: return "STRONG_SELL";
        case Signal::SELL:        return "SELL";
        case Signal::NEUTRAL:     return "NEUTRAL";
        case Signal::BUY:         return "BUY";
        case Signal::STRONG_BUY:  return "STRONG_BUY";
    }
    return "UNKNOWN";
}

// ============================================================================
// INDICATOR RESULT
// ============================================================================

struct IndicatorResult {
    std::string name;
    std::vector<double> values;
    Signal signal = Signal::NEUTRAL;
    double strength = 0.0; // 0.0 to 1.0
};

// ============================================================================
// PREDICTION RESULT
// ============================================================================

struct PredictionResult {
    std::string model;
    double predicted_price = 0.0;
    double confidence      = 0.0;
    Signal direction       = Signal::NEUTRAL;
    int    horizon_days    = 0;
    double upper_bound     = 0.0;
    double lower_bound     = 0.0;
    double r2_score        = 0.0;
    double rmse            = 0.0;
    std::vector<double> price_path;
};

// ============================================================================
// RISK METRICS
// ============================================================================

struct RiskMetrics {
    double var_95           = 0.0;
    double var_99           = 0.0;
    double cvar_95          = 0.0;
    double sharpe_ratio     = 0.0;
    double sortino_ratio    = 0.0;
    double max_drawdown     = 0.0;
    int    max_dd_duration  = 0;
    double beta             = 0.0;
    double alpha            = 0.0;
    double volatility_ann   = 0.0;
    double volatility_daily = 0.0;
    double calmar_ratio     = 0.0;
    double win_rate         = 0.0;
    double profit_factor    = 0.0;
    double skewness         = 0.0;
    double kurtosis         = 0.0;
};

// ============================================================================
// TRADE & BACKTEST TYPES
// ============================================================================

enum class TradeAction : int8_t {
    BUY  = 1,
    SELL = -1
};

struct Trade {
    std::string symbol;
    TradeAction action;
    double price     = 0.0;
    double quantity   = 0.0;
    double pnl       = 0.0;
    int64_t timestamp = 0;
    std::string reason;
};

struct BacktestConfig {
    std::string symbol;
    double initial_capital   = 100000.0;
    double commission_pct    = 0.1;
    double slippage_pct      = 0.05;
    double max_position_pct  = 0.95;
};

struct BacktestResult {
    double final_capital   = 0.0;
    double total_return    = 0.0;
    double annual_return   = 0.0;
    int    trade_count     = 0;
    int    win_count       = 0;
    int    loss_count      = 0;
    double avg_win         = 0.0;
    double avg_loss        = 0.0;
    double largest_win     = 0.0;
    double largest_loss    = 0.0;
    RiskMetrics risk;
    std::vector<double> equity_curve;
    std::vector<Trade> trades;
};

// ============================================================================
// PATTERN TYPES
// ============================================================================

enum class PatternType : uint8_t {
    DOJI,
    HAMMER,
    SHOOTING_STAR,
    BULLISH_ENGULFING,
    BEARISH_ENGULFING,
    MORNING_STAR,
    EVENING_STAR,
    THREE_WHITE_SOLDIERS,
    THREE_BLACK_CROWS,
    TWEEZER_TOP,
    TWEEZER_BOTTOM,
    HARAMI_BULLISH,
    HARAMI_BEARISH,
    MARUBOZU_BULLISH,
    MARUBOZU_BEARISH,
    SPINNING_TOP,
    DOUBLE_TOP,
    DOUBLE_BOTTOM,
    HEAD_AND_SHOULDERS,
    ASCENDING_TRIANGLE,
    DESCENDING_TRIANGLE
};

struct PatternDetection {
    PatternType type;
    double confidence = 0.0;
    Signal direction  = Signal::NEUTRAL;
    size_t start_idx  = 0;
    size_t end_idx    = 0;
};

} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_TYPES_H
