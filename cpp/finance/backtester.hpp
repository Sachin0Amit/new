#ifndef SOVEREIGN_FINANCE_BACKTESTER_HPP
#define SOVEREIGN_FINANCE_BACKTESTER_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Backtesting Engine
 *  Event-driven strategy backtester with position sizing, commission,
 *  slippage, and comprehensive performance reporting.
 * ============================================================================
 */

#include "types.hpp"
#include "indicators.hpp"
#include "risk.hpp"
#include "math.hpp"
#include <cmath>
#include <vector>
#include <functional>
#include <string>

namespace sovereign {
namespace finance {
namespace backtest {

// ============================================================================
// STRATEGY INTERFACE
// ============================================================================

/// A strategy callback: given bar index, time series, and indicator cache,
/// return a trade action (+1 buy, -1 sell, 0 hold).
using StrategyFn = std::function<int(size_t idx, const TimeSeries& ts,
                                      const std::unordered_map<std::string, std::vector<double>>& indicators)>;

// ============================================================================
// BACKTESTER
// ============================================================================

class Backtester {
public:
    Backtester(const BacktestConfig& config, StrategyFn strategy)
        : config_(config), strategy_(std::move(strategy)) {}

    BacktestResult run(const TimeSeries& ts) {
        BacktestResult result;
        if (ts.size() < 50) return result;

        double capital = config_.initial_capital;
        double position = 0.0;
        double avg_entry = 0.0;

        // Pre-compute indicators
        auto closes  = ts.closes();
        auto highs   = ts.highs();
        auto lows    = ts.lows();
        auto volumes = ts.volumes();

        std::unordered_map<std::string, std::vector<double>> ind;
        ind["sma_20"]  = indicators::SMA(closes, 20);
        ind["sma_50"]  = indicators::SMA(closes, 50);
        ind["sma_200"] = indicators::SMA(closes, 200);
        ind["ema_12"]  = indicators::EMA(closes, 12);
        ind["ema_26"]  = indicators::EMA(closes, 26);
        ind["rsi_14"]  = indicators::RSI(closes, 14);
        ind["atr_14"]  = indicators::ATR(highs, lows, closes, 14);
        ind["obv"]     = indicators::OBV(closes, volumes);

        auto bb = indicators::BollingerBands(closes, 20, 2.0);
        ind["bb_upper"]  = bb.upper;
        ind["bb_middle"] = bb.middle;
        ind["bb_lower"]  = bb.lower;

        auto macd = indicators::MACD(closes, 12, 26, 9);
        ind["macd"]        = macd.macd_line;
        ind["macd_signal"] = macd.signal_line;
        ind["macd_hist"]   = macd.histogram;

        auto adx = indicators::ADX(highs, lows, closes, 14);
        ind["adx"]      = adx.adx;
        ind["plus_di"]  = adx.plus_di;
        ind["minus_di"] = adx.minus_di;

        result.equity_curve.reserve(ts.size());

        for (size_t i = 0; i < ts.size(); ++i) {
            double price = ts[i].close;
            double equity = capital + position * price;
            result.equity_curve.push_back(equity);

            int action = strategy_(i, ts, ind);

            if (action == 1 && position == 0.0) {
                // BUY
                double max_invest = capital * config_.max_position_pct;
                double eff_price = price * (1.0 + config_.slippage_pct / 100.0);
                double commission = max_invest * (config_.commission_pct / 100.0);
                double investable = max_invest - commission;
                double qty = investable / eff_price;

                if (qty > 0) {
                    position = qty;
                    avg_entry = eff_price;
                    capital -= qty * eff_price + commission;

                    Trade t;
                    t.symbol = config_.symbol;
                    t.action = TradeAction::BUY;
                    t.price = eff_price;
                    t.quantity = qty;
                    t.timestamp = ts[i].timestamp;
                    t.reason = "STRATEGY_BUY";
                    result.trades.push_back(t);
                }
            } else if (action == -1 && position > 0.0) {
                // SELL
                double eff_price = price * (1.0 - config_.slippage_pct / 100.0);
                double proceeds = position * eff_price;
                double commission = proceeds * (config_.commission_pct / 100.0);
                double pnl = (eff_price - avg_entry) * position - commission;

                capital += proceeds - commission;

                Trade t;
                t.symbol = config_.symbol;
                t.action = TradeAction::SELL;
                t.price = eff_price;
                t.quantity = position;
                t.pnl = pnl;
                t.timestamp = ts[i].timestamp;
                t.reason = "STRATEGY_SELL";
                result.trades.push_back(t);

                position = 0.0;
                avg_entry = 0.0;
            }
        }

        // Close open position at last price
        if (position > 0.0 && ts.size() > 0) {
            double last = ts.back().close;
            double pnl = (last - avg_entry) * position;
            capital += position * last;

            Trade t;
            t.symbol = config_.symbol;
            t.action = TradeAction::SELL;
            t.price = last;
            t.quantity = position;
            t.pnl = pnl;
            t.timestamp = ts.back().timestamp;
            t.reason = "BACKTEST_EXIT";
            result.trades.push_back(t);
            position = 0.0;
        }

        result.final_capital = capital;
        result.total_return = (capital - config_.initial_capital) / config_.initial_capital;
        result.trade_count = static_cast<int>(result.trades.size());

        // Count wins/losses
        double total_win = 0.0, total_loss = 0.0;
        for (const auto& t : result.trades) {
            if (t.action == TradeAction::SELL) {
                if (t.pnl > 0) {
                    result.win_count++;
                    total_win += t.pnl;
                    if (t.pnl > result.largest_win) result.largest_win = t.pnl;
                } else if (t.pnl < 0) {
                    result.loss_count++;
                    total_loss += std::abs(t.pnl);
                    if (std::abs(t.pnl) > std::abs(result.largest_loss))
                        result.largest_loss = t.pnl;
                }
            }
        }
        if (result.win_count > 0) result.avg_win = total_win / result.win_count;
        if (result.loss_count > 0) result.avg_loss = total_loss / result.loss_count;

        // Risk metrics from equity curve
        if (result.equity_curve.size() > 1) {
            std::vector<double> eq_returns(result.equity_curve.size() - 1);
            for (size_t i = 1; i < result.equity_curve.size(); ++i) {
                if (result.equity_curve[i - 1] > 0) {
                    eq_returns[i - 1] = std::log(result.equity_curve[i] / result.equity_curve[i - 1]);
                }
            }
            std::vector<double> bench(eq_returns.size(), 0.0);
            result.risk = risk::compute_full_risk(result.equity_curve, eq_returns, bench, 0.04);
        }

        return result;
    }

private:
    BacktestConfig config_;
    StrategyFn strategy_;
};

// ============================================================================
// BUILT-IN STRATEGIES (as lambdas)
// ============================================================================

/// SMA Crossover Strategy factory.
inline StrategyFn make_sma_cross(int fast, int slow) {
    std::string fast_key = "sma_" + std::to_string(fast);
    std::string slow_key = "sma_" + std::to_string(slow);
    return [fast_key, slow_key](size_t idx, const TimeSeries&,
                                 const std::unordered_map<std::string, std::vector<double>>& ind) -> int {
        if (idx < 1) return 0;
        auto it_f = ind.find(fast_key);
        auto it_s = ind.find(slow_key);
        if (it_f == ind.end() || it_s == ind.end()) return 0;
        const auto& f = it_f->second;
        const auto& s = it_s->second;
        if (std::isnan(f[idx]) || std::isnan(s[idx]) || std::isnan(f[idx-1]) || std::isnan(s[idx-1])) return 0;
        if (f[idx-1] <= s[idx-1] && f[idx] > s[idx]) return 1;  // Golden Cross
        if (f[idx-1] >= s[idx-1] && f[idx] < s[idx]) return -1; // Death Cross
        return 0;
    };
}

/// RSI Mean Reversion factory.
inline StrategyFn make_rsi_meanrev(double oversold = 30.0, double overbought = 70.0) {
    return [oversold, overbought](size_t idx, const TimeSeries&,
                                   const std::unordered_map<std::string, std::vector<double>>& ind) -> int {
        auto it = ind.find("rsi_14");
        if (it == ind.end() || std::isnan(it->second[idx])) return 0;
        double rsi = it->second[idx];
        if (rsi < oversold) return 1;
        if (rsi > overbought) return -1;
        return 0;
    };
}

/// Bollinger Band Bounce factory.
inline StrategyFn make_bb_bounce() {
    return [](size_t idx, const TimeSeries& ts,
              const std::unordered_map<std::string, std::vector<double>>& ind) -> int {
        auto it_u = ind.find("bb_upper");
        auto it_l = ind.find("bb_lower");
        if (it_u == ind.end() || it_l == ind.end()) return 0;
        if (std::isnan(it_u->second[idx]) || std::isnan(it_l->second[idx])) return 0;
        double price = ts[idx].close;
        if (price <= it_l->second[idx]) return 1;
        if (price >= it_u->second[idx]) return -1;
        return 0;
    };
}

/// MACD Histogram zero-cross factory.
inline StrategyFn make_macd_histogram() {
    return [](size_t idx, const TimeSeries&,
              const std::unordered_map<std::string, std::vector<double>>& ind) -> int {
        auto it = ind.find("macd_hist");
        if (it == ind.end() || idx < 1) return 0;
        const auto& h = it->second;
        if (std::isnan(h[idx]) || std::isnan(h[idx-1])) return 0;
        if (h[idx-1] <= 0 && h[idx] > 0) return 1;
        if (h[idx-1] >= 0 && h[idx] < 0) return -1;
        return 0;
    };
}

/// ADX Trend Following factory: Buy when ADX > threshold and +DI > -DI.
inline StrategyFn make_adx_trend(double threshold = 25.0) {
    return [threshold](size_t idx, const TimeSeries&,
                        const std::unordered_map<std::string, std::vector<double>>& ind) -> int {
        auto it_adx = ind.find("adx");
        auto it_pdi = ind.find("plus_di");
        auto it_mdi = ind.find("minus_di");
        if (it_adx == ind.end() || it_pdi == ind.end() || it_mdi == ind.end()) return 0;
        double adx = it_adx->second[idx];
        double pdi = it_pdi->second[idx];
        double mdi = it_mdi->second[idx];
        if (std::isnan(adx) || std::isnan(pdi) || std::isnan(mdi)) return 0;
        if (adx > threshold && pdi > mdi) return 1;
        if (adx > threshold && mdi > pdi) return -1;
        return 0;
    };
}

} // namespace backtest
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_BACKTESTER_HPP
