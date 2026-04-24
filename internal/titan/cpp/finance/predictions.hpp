#ifndef SOVEREIGN_FINANCE_PREDICTIONS_HPP
#define SOVEREIGN_FINANCE_PREDICTIONS_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Predictive Models
 *  Statistical and stochastic models for price forecasting.
 *  No external ML frameworks — pure mathematical implementations.
 * ============================================================================
 */

#include "types.hpp"
#include "math.hpp"
#include <cmath>
#include <vector>
#include <algorithm>

namespace sovereign {
namespace finance {
namespace predictions {

// ============================================================================
// LINEAR REGRESSION
// ============================================================================

/// Fits y = mx + c to the price data and projects forward.
inline PredictionResult linear_regression(const std::vector<double>& prices, int horizon) {
    PredictionResult res;
    res.model = "LINEAR_REGRESSION";
    res.horizon_days = horizon;

    size_t n = prices.size();
    if (n < 2) return res;

    double nd = static_cast<double>(n);
    double sum_x = 0, sum_y = 0, sum_xy = 0, sum_x2 = 0;
    for (size_t i = 0; i < n; ++i) {
        double x = static_cast<double>(i);
        sum_x  += x;
        sum_y  += prices[i];
        sum_xy += x * prices[i];
        sum_x2 += x * x;
    }

    double denom = nd * sum_x2 - sum_x * sum_x;
    if (std::abs(denom) < 1e-15) return res;

    double m = (nd * sum_xy - sum_x * sum_y) / denom;
    double c = (sum_y - m * sum_x) / nd;

    // R-squared
    double mean_y = sum_y / nd;
    double ss_tot = 0, ss_res = 0;
    for (size_t i = 0; i < n; ++i) {
        double pred = m * i + c;
        ss_tot += (prices[i] - mean_y) * (prices[i] - mean_y);
        ss_res += (prices[i] - pred) * (prices[i] - pred);
    }
    res.r2_score = (ss_tot > 0) ? 1.0 - ss_res / ss_tot : 0.0;
    res.rmse = std::sqrt(ss_res / nd);

    // Forecast
    double target_x = static_cast<double>(n - 1 + horizon);
    res.predicted_price = m * target_x + c;
    res.confidence = std::max(0.0, std::min(1.0, res.r2_score));

    // Direction
    if (m > 0.001) res.direction = Signal::BUY;
    else if (m < -0.001) res.direction = Signal::SELL;
    else res.direction = Signal::NEUTRAL;

    // 95% confidence interval
    double se = res.rmse * std::sqrt(1.0 + 1.0 / nd +
        std::pow(target_x - sum_x / nd, 2) / (sum_x2 - sum_x * sum_x / nd));
    res.upper_bound = res.predicted_price + 1.96 * se;
    res.lower_bound = res.predicted_price - 1.96 * se;

    return res;
}

// ============================================================================
// POLYNOMIAL REGRESSION (degree 2)
// ============================================================================

/// Fits y = ax² + bx + c to the price data.
inline PredictionResult quadratic_regression(const std::vector<double>& prices, int horizon) {
    PredictionResult res;
    res.model = "QUADRATIC_REGRESSION";
    res.horizon_days = horizon;

    size_t n = prices.size();
    if (n < 3) return res;

    // Normal equations for degree-2 polynomial (3x3 system)
    // Using Cramer's rule for a simple 3-variable solve
    double s0 = static_cast<double>(n);
    double s1 = 0, s2 = 0, s3 = 0, s4 = 0;
    double t0 = 0, t1 = 0, t2 = 0;

    for (size_t i = 0; i < n; ++i) {
        double x = static_cast<double>(i);
        double x2 = x * x;
        s1 += x;
        s2 += x2;
        s3 += x2 * x;
        s4 += x2 * x2;
        t0 += prices[i];
        t1 += x * prices[i];
        t2 += x2 * prices[i];
    }

    // Solve using matrix determinant (Cramer's rule)
    // | s0 s1 s2 | | c |   | t0 |
    // | s1 s2 s3 | | b | = | t1 |
    // | s2 s3 s4 | | a |   | t2 |

    double D = s0 * (s2 * s4 - s3 * s3) - s1 * (s1 * s4 - s3 * s2) + s2 * (s1 * s3 - s2 * s2);
    if (std::abs(D) < 1e-15) return res;

    double Dc = t0 * (s2 * s4 - s3 * s3) - s1 * (t1 * s4 - s3 * t2) + s2 * (t1 * s3 - s2 * t2);
    double Db = s0 * (t1 * s4 - s3 * t2) - t0 * (s1 * s4 - s3 * s2) + s2 * (s1 * t2 - t1 * s2);
    double Da = s0 * (s2 * t2 - t1 * s3) - s1 * (s1 * t2 - t1 * s2) + t0 * (s1 * s3 - s2 * s2);

    double c_coeff = Dc / D;
    double b_coeff = Db / D;
    double a_coeff = Da / D;

    // R-squared
    double mean_y = t0 / s0;
    double ss_tot = 0, ss_res = 0;
    for (size_t i = 0; i < n; ++i) {
        double x = static_cast<double>(i);
        double pred = a_coeff * x * x + b_coeff * x + c_coeff;
        ss_tot += (prices[i] - mean_y) * (prices[i] - mean_y);
        ss_res += (prices[i] - pred) * (prices[i] - pred);
    }
    res.r2_score = (ss_tot > 0) ? 1.0 - ss_res / ss_tot : 0.0;
    res.rmse = std::sqrt(ss_res / s0);

    double target_x = static_cast<double>(n - 1 + horizon);
    res.predicted_price = a_coeff * target_x * target_x + b_coeff * target_x + c_coeff;
    res.confidence = std::max(0.0, std::min(1.0, res.r2_score));

    double last = prices.back();
    if (res.predicted_price > last * 1.01) res.direction = Signal::BUY;
    else if (res.predicted_price < last * 0.99) res.direction = Signal::SELL;
    else res.direction = Signal::NEUTRAL;

    double se = res.rmse * 2.0; // Rough bound
    res.upper_bound = res.predicted_price + 1.96 * se;
    res.lower_bound = res.predicted_price - 1.96 * se;

    return res;
}

// ============================================================================
// MONTE CARLO (Geometric Brownian Motion)
// ============================================================================

/// Simulates future price paths using GBM.
inline PredictionResult monte_carlo_gbm(const std::vector<double>& prices,
                                         int horizon, int num_simulations = 5000) {
    PredictionResult res;
    res.model = "MONTE_CARLO_GBM";
    res.horizon_days = horizon;

    size_t n = prices.size();
    if (n < 2) return res;

    // Compute log returns
    std::vector<double> returns(n - 1);
    for (size_t i = 1; i < n; ++i) {
        returns[i - 1] = std::log(prices[i] / prices[i - 1]);
    }

    double mu = math::mean(returns);
    double sigma = math::stddev(returns);
    double drift = mu - 0.5 * sigma * sigma;
    double last_price = prices.back();

    std::vector<double> final_prices(num_simulations);
    res.price_path.resize(horizon);

    for (int sim = 0; sim < num_simulations; ++sim) {
        double current = last_price;
        for (int step = 0; step < horizon; ++step) {
            double z = math::randn();
            current *= std::exp(drift + sigma * z);
            if (sim == 0) {
                res.price_path[step] = current;
            }
        }
        final_prices[sim] = current;
    }

    std::sort(final_prices.begin(), final_prices.end());

    res.predicted_price = math::mean(final_prices);
    res.lower_bound = math::percentile(final_prices.data(), final_prices.size(), 5.0);
    res.upper_bound = math::percentile(final_prices.data(), final_prices.size(), 95.0);

    double spread = (res.upper_bound - res.lower_bound) / res.predicted_price;
    res.confidence = std::max(0.0, 1.0 - spread);

    if (res.predicted_price > last_price * 1.02) res.direction = Signal::BUY;
    else if (res.predicted_price < last_price * 0.98) res.direction = Signal::SELL;
    else res.direction = Signal::NEUTRAL;

    return res;
}

// ============================================================================
// EXPONENTIAL SMOOTHING (Holt-Winters Double)
// ============================================================================

/// Double exponential smoothing for trend-based forecasting.
inline PredictionResult holt_linear(const std::vector<double>& prices,
                                     int horizon, double alpha = 0.3, double beta = 0.1) {
    PredictionResult res;
    res.model = "HOLT_LINEAR";
    res.horizon_days = horizon;

    size_t n = prices.size();
    if (n < 2) return res;

    // Initialize
    double level = prices[0];
    double trend = prices[1] - prices[0];

    std::vector<double> fitted(n);
    fitted[0] = level;

    for (size_t i = 1; i < n; ++i) {
        double prev_level = level;
        level = alpha * prices[i] + (1.0 - alpha) * (prev_level + trend);
        trend = beta * (level - prev_level) + (1.0 - beta) * trend;
        fitted[i] = level + trend;
    }

    // Forecast
    res.predicted_price = level + trend * horizon;
    res.price_path.resize(horizon);
    for (int h = 0; h < horizon; ++h) {
        res.price_path[h] = level + trend * (h + 1);
    }

    // Residual-based confidence
    double sse = 0.0;
    for (size_t i = 1; i < n; ++i) {
        double err = prices[i] - fitted[i];
        sse += err * err;
    }
    double rmse = std::sqrt(sse / (n - 1));
    res.rmse = rmse;
    res.upper_bound = res.predicted_price + 1.96 * rmse * std::sqrt(horizon);
    res.lower_bound = res.predicted_price - 1.96 * rmse * std::sqrt(horizon);
    res.confidence = std::max(0.0, 1.0 - rmse / prices.back());

    double last = prices.back();
    if (res.predicted_price > last * 1.01) res.direction = Signal::BUY;
    else if (res.predicted_price < last * 0.99) res.direction = Signal::SELL;
    else res.direction = Signal::NEUTRAL;

    return res;
}

// ============================================================================
// ENSEMBLE PREDICTOR
// ============================================================================

/// Combines multiple models for robust prediction.
inline PredictionResult ensemble_predict(const std::vector<double>& prices, int horizon) {
    auto lr = linear_regression(prices, horizon);
    auto qr = quadratic_regression(prices, horizon);
    auto mc = monte_carlo_gbm(prices, horizon, 3000);
    auto hl = holt_linear(prices, horizon);

    PredictionResult res;
    res.model = "ENSEMBLE_4_MODEL";
    res.horizon_days = horizon;

    // Weighted average (weight by confidence)
    double total_weight = lr.confidence + qr.confidence + mc.confidence + hl.confidence;
    if (total_weight < 0.01) total_weight = 1.0;

    res.predicted_price = (lr.predicted_price * lr.confidence +
                           qr.predicted_price * qr.confidence +
                           mc.predicted_price * mc.confidence +
                           hl.predicted_price * hl.confidence) / total_weight;

    res.upper_bound = (lr.upper_bound + qr.upper_bound + mc.upper_bound + hl.upper_bound) / 4.0;
    res.lower_bound = (lr.lower_bound + qr.lower_bound + mc.lower_bound + hl.lower_bound) / 4.0;
    res.confidence = (lr.confidence + qr.confidence + mc.confidence + hl.confidence) / 4.0;

    res.price_path = mc.price_path; // Use MC path for visualization

    double last = prices.back();
    if (res.predicted_price > last * 1.01) res.direction = Signal::BUY;
    else if (res.predicted_price < last * 0.99) res.direction = Signal::SELL;
    else res.direction = Signal::NEUTRAL;

    return res;
}

} // namespace predictions
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_PREDICTIONS_HPP
