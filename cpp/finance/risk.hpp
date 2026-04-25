#ifndef SOVEREIGN_FINANCE_RISK_HPP
#define SOVEREIGN_FINANCE_RISK_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Risk Analytics
 *  Quantitative risk measurement and volatility modeling.
 * ============================================================================
 */

#include "types.hpp"
#include "math.hpp"
#include <cmath>
#include <vector>
#include <algorithm>

namespace sovereign {
namespace finance {
namespace risk {

// ============================================================================
// VALUE AT RISK (VaR)
// ============================================================================

/// Historical VaR at a given confidence level (e.g., 95.0).
inline double historical_var(const std::vector<double>& returns, double confidence) {
    if (returns.empty()) return 0.0;
    std::vector<double> sorted = returns;
    std::sort(sorted.begin(), sorted.end());
    double pct = 100.0 - confidence;
    return -math::percentile(sorted.data(), sorted.size(), pct);
}

/// Parametric (Gaussian) VaR.
inline double parametric_var(const std::vector<double>& returns, double confidence) {
    double mu = math::mean(returns);
    double sigma = math::stddev(returns);
    // z-score for confidence level
    double z = 0.0;
    if (confidence >= 99.0) z = 2.326;
    else if (confidence >= 97.5) z = 1.960;
    else if (confidence >= 95.0) z = 1.645;
    else if (confidence >= 90.0) z = 1.282;
    else z = 1.0;
    return -(mu - z * sigma);
}

/// Conditional VaR (Expected Shortfall / CVaR).
inline double cvar(const std::vector<double>& returns, double confidence) {
    if (returns.empty()) return 0.0;
    std::vector<double> sorted = returns;
    std::sort(sorted.begin(), sorted.end());
    double pct = 100.0 - confidence;
    double threshold = math::percentile(sorted.data(), sorted.size(), pct);

    double sum = 0.0;
    int count = 0;
    for (double r : returns) {
        if (r <= threshold) {
            sum += r;
            count++;
        }
    }
    return (count > 0) ? -(sum / count) : 0.0;
}

// ============================================================================
// DRAWDOWN ANALYSIS
// ============================================================================

struct DrawdownResult {
    double max_drawdown = 0.0;
    int    max_dd_duration = 0;
    double avg_drawdown = 0.0;
    int    num_drawdowns = 0;
    std::vector<double> drawdown_series;
};

/// Compute the full drawdown profile from an equity curve.
inline DrawdownResult analyze_drawdowns(const std::vector<double>& equity) {
    DrawdownResult res;
    if (equity.size() < 2) return res;

    res.drawdown_series.resize(equity.size(), 0.0);
    double peak = equity[0];
    double sum_dd = 0.0;
    int dd_count = 0;
    int current_duration = 0;
    bool in_drawdown = false;

    for (size_t i = 0; i < equity.size(); ++i) {
        if (equity[i] > peak) {
            peak = equity[i];
            if (in_drawdown) {
                dd_count++;
                in_drawdown = false;
            }
            current_duration = 0;
        } else {
            double dd = (peak - equity[i]) / peak;
            res.drawdown_series[i] = dd;
            sum_dd += dd;
            current_duration++;
            in_drawdown = true;

            if (dd > res.max_drawdown) {
                res.max_drawdown = dd;
                res.max_dd_duration = current_duration;
            }
        }
    }

    res.num_drawdowns = dd_count;
    if (equity.size() > 0) {
        res.avg_drawdown = sum_dd / equity.size();
    }

    return res;
}

// ============================================================================
// PERFORMANCE RATIOS
// ============================================================================

/// Sharpe Ratio (annualized).
inline double sharpe_ratio(const std::vector<double>& returns, double rf_annual) {
    if (returns.size() < 2) return 0.0;
    double rf_daily = std::pow(1.0 + rf_annual, 1.0 / TRADING_DAYS_PER_YEAR) - 1.0;

    std::vector<double> excess(returns.size());
    for (size_t i = 0; i < returns.size(); ++i) {
        excess[i] = returns[i] - rf_daily;
    }

    double m = math::mean(excess);
    double s = math::stddev(excess);
    if (s == 0.0) return 0.0;
    return (m / s) * SQRT_252;
}

/// Sortino Ratio (using downside deviation).
inline double sortino_ratio(const std::vector<double>& returns, double rf_annual, double target = 0.0) {
    if (returns.size() < 2) return 0.0;
    double rf_daily = std::pow(1.0 + rf_annual, 1.0 / TRADING_DAYS_PER_YEAR) - 1.0;
    double target_daily = std::pow(1.0 + target, 1.0 / TRADING_DAYS_PER_YEAR) - 1.0;

    double sum_excess = 0.0;
    double sum_downside_sq = 0.0;
    int down_count = 0;
    for (double r : returns) {
        sum_excess += r - rf_daily;
        if (r < target_daily) {
            double d = r - target_daily;
            sum_downside_sq += d * d;
            down_count++;
        }
    }

    double mean_excess = sum_excess / returns.size();
    if (down_count == 0) return std::numeric_limits<double>::infinity();
    double downside_dev = std::sqrt(sum_downside_sq / returns.size());
    if (downside_dev == 0.0) return 0.0;
    return (mean_excess / downside_dev) * SQRT_252;
}

/// Calmar Ratio (annualized return / max drawdown).
inline double calmar_ratio(double annual_return, double max_drawdown) {
    if (max_drawdown == 0.0) return 0.0;
    return annual_return / max_drawdown;
}

/// Profit Factor (gross profit / gross loss).
inline double profit_factor(const std::vector<Trade>& trades) {
    double gross_profit = 0.0, gross_loss = 0.0;
    for (const auto& t : trades) {
        if (t.pnl > 0) gross_profit += t.pnl;
        else gross_loss += std::abs(t.pnl);
    }
    return (gross_loss > 0.0) ? gross_profit / gross_loss : 0.0;
}

/// Beta relative to a benchmark.
inline double beta(const std::vector<double>& asset_returns,
                   const std::vector<double>& bench_returns) {
    size_t n = std::min(asset_returns.size(), bench_returns.size());
    if (n < 2) return 0.0;
    double cov = math::covariance(asset_returns.data(), bench_returns.data(), n);
    double var_bench = math::variance(bench_returns.data(), n);
    return (var_bench > 0.0) ? cov / var_bench : 0.0;
}

/// Jensen's Alpha.
inline double alpha(const std::vector<double>& asset_returns,
                    const std::vector<double>& bench_returns, double rf_annual) {
    double b = beta(asset_returns, bench_returns);
    double asset_ann = std::pow(1.0 + math::mean(asset_returns), TRADING_DAYS_PER_YEAR) - 1.0;
    double bench_ann = std::pow(1.0 + math::mean(bench_returns), TRADING_DAYS_PER_YEAR) - 1.0;
    return asset_ann - (rf_annual + b * (bench_ann - rf_annual));
}

// ============================================================================
// GARCH(1,1) VOLATILITY
// ============================================================================

struct GARCHParams {
    double omega = 0.000001;
    double alpha = 0.09;
    double beta  = 0.90;
};

/// Compute GARCH(1,1) conditional volatility series (annualized).
inline std::vector<double> garch_volatility(const std::vector<double>& returns,
                                             const GARCHParams& params) {
    if (returns.size() < 2) return {};
    std::vector<double> out(returns.size());

    double denom = 1.0 - params.alpha - params.beta;
    if (denom <= 0) denom = 0.01;
    double variance = params.omega / denom;
    out[0] = std::sqrt(variance) * SQRT_252;

    for (size_t i = 1; i < returns.size(); ++i) {
        variance = params.omega +
                   params.alpha * returns[i - 1] * returns[i - 1] +
                   params.beta * variance;
        if (variance < 0) variance = 1e-8;
        out[i] = std::sqrt(variance) * SQRT_252;
    }
    return out;
}

/// Forecast GARCH volatility N periods ahead.
inline std::vector<double> garch_forecast(const std::vector<double>& returns,
                                           const GARCHParams& params, int horizon) {
    if (returns.size() < 2 || horizon <= 0) return {};

    double denom = 1.0 - params.alpha - params.beta;
    if (denom <= 0) denom = 0.01;
    double long_run_var = params.omega / denom;

    // Get last conditional variance
    double variance = long_run_var;
    for (size_t i = 1; i < returns.size(); ++i) {
        variance = params.omega + params.alpha * returns[i - 1] * returns[i - 1] + params.beta * variance;
    }

    std::vector<double> forecast(horizon);
    double ab_sum = params.alpha + params.beta;
    for (int h = 0; h < horizon; ++h) {
        double fv = long_run_var + std::pow(ab_sum, h) * (variance - long_run_var);
        if (fv < 0) fv = 1e-8;
        forecast[h] = std::sqrt(fv) * SQRT_252;
    }
    return forecast;
}

// ============================================================================
// COMPREHENSIVE RISK METRICS
// ============================================================================

/// Build a complete RiskMetrics struct from equity curve and return data.
inline RiskMetrics compute_full_risk(const std::vector<double>& equity,
                                      const std::vector<double>& returns,
                                      const std::vector<double>& bench_returns,
                                      double rf_annual) {
    RiskMetrics rm;
    if (returns.size() < 2) return rm;

    rm.volatility_daily = math::stddev(returns);
    rm.volatility_ann = rm.volatility_daily * SQRT_252;
    rm.sharpe_ratio = sharpe_ratio(returns, rf_annual);
    rm.sortino_ratio = sortino_ratio(returns, rf_annual);
    rm.var_95 = historical_var(returns, 95.0);
    rm.var_99 = historical_var(returns, 99.0);
    rm.cvar_95 = cvar(returns, 95.0);
    rm.skewness = math::skewness(returns);
    rm.kurtosis = math::kurtosis(returns);

    auto dd = analyze_drawdowns(equity);
    rm.max_drawdown = dd.max_drawdown;
    rm.max_dd_duration = dd.max_dd_duration;

    double ann_ret = std::pow(1.0 + math::mean(returns), TRADING_DAYS_PER_YEAR) - 1.0;
    rm.calmar_ratio = calmar_ratio(ann_ret, rm.max_drawdown);

    if (!bench_returns.empty()) {
        rm.beta = beta(returns, bench_returns);
        rm.alpha = alpha(returns, bench_returns, rf_annual);
    }

    int wins = 0;
    for (double r : returns) if (r > 0) wins++;
    rm.win_rate = static_cast<double>(wins) / returns.size();

    return rm;
}

} // namespace risk
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_RISK_HPP
