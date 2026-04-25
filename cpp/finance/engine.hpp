#ifndef SOVEREIGN_FINANCE_ENGINE_HPP
#define SOVEREIGN_FINANCE_ENGINE_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Master Orchestrator
 *  Unifies all modules: indicators, predictions, risk, patterns, backtest.
 *  Provides a single API surface for the Go backend via C FFI.
 * ============================================================================
 */

#include "types.hpp"
#include "math.hpp"
#include "indicators.hpp"
#include "predictions.hpp"
#include "risk.hpp"
#include "patterns.hpp"
#include "backtester.hpp"
#include "sentiment.hpp"
#include "fundamentals.hpp"

#include <unordered_map>
#include <mutex>
#include <string>
#include <vector>
#include <chrono>

namespace sovereign {
namespace finance {

// ============================================================================
// MARKET DATA CACHE
// ============================================================================

class MarketDataStore {
public:
    void put(const std::string& symbol, const TimeSeries& ts) {
        std::lock_guard<std::mutex> lock(mu_);
        data_[symbol] = ts;
    }

    bool get(const std::string& symbol, TimeSeries& out) const {
        std::lock_guard<std::mutex> lock(mu_);
        auto it = data_.find(symbol);
        if (it == data_.end()) return false;
        out = it->second;
        return true;
    }

    std::vector<std::string> symbols() const {
        std::lock_guard<std::mutex> lock(mu_);
        std::vector<std::string> syms;
        syms.reserve(data_.size());
        for (const auto& kv : data_) {
            syms.push_back(kv.first);
        }
        return syms;
    }

    size_t size() const {
        std::lock_guard<std::mutex> lock(mu_);
        return data_.size();
    }

private:
    mutable std::mutex mu_;
    std::unordered_map<std::string, TimeSeries> data_;
};

// ============================================================================
// FINANCE ENGINE — The Main Orchestrator
// ============================================================================

class FinanceEngine {
public:
    FinanceEngine() = default;

    /// Load market data into the cache.
    void load_data(const std::string& symbol, const TimeSeries& ts) {
        store_.put(symbol, ts);
    }

    /// Get full technical analysis for a symbol.
    struct TechnicalAnalysis {
        std::vector<double> sma_20, sma_50, sma_200;
        std::vector<double> ema_12, ema_26;
        std::vector<double> rsi_14;
        indicators::MACDResult macd;
        indicators::BollingerResult bollinger;
        indicators::ADXResult adx;
        std::vector<double> atr_14;
        std::vector<double> obv;
        std::vector<double> vwap;
        std::vector<double> cci;
        std::vector<double> williams_r;
        indicators::StochasticResult stochastic;
        std::vector<double> mfi;
        std::vector<double> hma;
        std::vector<double> dema;
    };

    TechnicalAnalysis analyze(const std::string& symbol) {
        TechnicalAnalysis ta;
        TimeSeries ts;
        if (!store_.get(symbol, ts) || ts.size() < 50) return ta;

        auto c = ts.closes();
        auto h = ts.highs();
        auto l = ts.lows();
        auto v = ts.volumes();

        ta.sma_20  = indicators::SMA(c, 20);
        ta.sma_50  = indicators::SMA(c, 50);
        ta.sma_200 = indicators::SMA(c, 200);
        ta.ema_12  = indicators::EMA(c, 12);
        ta.ema_26  = indicators::EMA(c, 26);
        ta.rsi_14  = indicators::RSI(c, 14);
        ta.macd    = indicators::MACD(c, 12, 26, 9);
        ta.bollinger = indicators::BollingerBands(c, 20, 2.0);
        ta.adx     = indicators::ADX(h, l, c, 14);
        ta.atr_14  = indicators::ATR(h, l, c, 14);
        ta.obv     = indicators::OBV(c, v);
        ta.vwap    = indicators::VWAP(h, l, c, v);
        ta.cci     = indicators::CCI(h, l, c, 20);
        ta.williams_r = indicators::WilliamsR(h, l, c, 14);
        ta.stochastic = indicators::Stochastic(h, l, c, 14, 3, 3);
        ta.mfi     = indicators::MFI(h, l, c, v, 14);
        ta.hma     = indicators::HMA(c, 20);
        ta.dema    = indicators::DEMA(c, 20);

        return ta;
    }

    /// Run prediction models.
    PredictionResult predict(const std::string& symbol, int horizon = 30) {
        TimeSeries ts;
        if (!store_.get(symbol, ts) || ts.size() < 50) return {};
        return predictions::ensemble_predict(ts.closes(), horizon);
    }

    /// Compute risk metrics.
    RiskMetrics compute_risk(const std::string& symbol) {
        TimeSeries ts;
        if (!store_.get(symbol, ts) || ts.size() < 50) return {};

        auto returns = ts.log_returns();
        auto closes = ts.closes();

        // Build simple equity curve
        std::vector<double> equity(closes.size());
        equity[0] = 10000.0;
        for (size_t i = 1; i < closes.size(); ++i) {
            equity[i] = equity[i - 1] * (1.0 + (closes[i] - closes[i - 1]) / closes[i - 1]);
        }

        std::vector<double> bench(returns.size(), 0.0); // Flat benchmark
        return risk::compute_full_risk(equity, returns, bench, 0.04);
    }

    /// Detect patterns.
    std::vector<PatternDetection> detect_patterns(const std::string& symbol) {
        TimeSeries ts;
        if (!store_.get(symbol, ts)) return {};
        return patterns::detect_candle_patterns(ts);
    }

    /// Analyze sentiment of a news headline.
    sentiment::SentimentScore analyze_sentiment(const std::string& headline) const {
        sentiment::LexiconAnalyzer analyzer;
        return analyzer.analyze(headline);
    }

    /// Analyze fundamentals of a company.
    fundamentals::FundamentalScore analyze_fundamentals(const std::string& symbol) const {
        fundamentals::FundamentalAnalyzer analyzer;
        auto data = analyzer.get_mock_data(symbol);
        return analyzer.evaluate(data);
    }

    /// Run a backtest.
    BacktestResult backtest(const std::string& symbol, const std::string& strategy_name) {
        TimeSeries ts;
        if (!store_.get(symbol, ts)) return {};

        BacktestConfig cfg;
        cfg.symbol = symbol;
        cfg.initial_capital = 100000.0;
        cfg.commission_pct = 0.1;
        cfg.slippage_pct = 0.05;
        cfg.max_position_pct = 0.95;

        backtest::StrategyFn strat;
        if (strategy_name == "SMA_CROSS") {
            strat = backtest::make_sma_cross(20, 50);
        } else if (strategy_name == "RSI") {
            strat = backtest::make_rsi_meanrev();
        } else if (strategy_name == "BOLLINGER") {
            strat = backtest::make_bb_bounce();
        } else if (strategy_name == "MACD") {
            strat = backtest::make_macd_histogram();
        } else if (strategy_name == "ADX") {
            strat = backtest::make_adx_trend();
        } else {
            strat = backtest::make_sma_cross(20, 50); // Default
        }

        backtest::Backtester bt(cfg, strat);
        return bt.run(ts);
    }

    /// Compute GARCH volatility forecast.
    std::vector<double> volatility_forecast(const std::string& symbol, int horizon = 30) {
        TimeSeries ts;
        if (!store_.get(symbol, ts)) return {};
        auto returns = ts.log_returns();
        risk::GARCHParams params;
        return risk::garch_forecast(returns, params, horizon);
    }

    /// Get the data store for direct access.
    MarketDataStore& store() { return store_; }
    const MarketDataStore& store() const { return store_; }

private:
    MarketDataStore store_;
};

} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_ENGINE_HPP
