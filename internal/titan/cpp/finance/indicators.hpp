#ifndef SOVEREIGN_FINANCE_INDICATORS_HPP
#define SOVEREIGN_FINANCE_INDICATORS_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Technical Indicators
 *  Ultra-fast C++ implementations of classical and modern indicators.
 *  Designed for cache-line efficiency and minimal heap allocation.
 * ============================================================================
 */

#include "types.hpp"
#include "math.hpp"
#include <cmath>
#include <vector>
#include <algorithm>

namespace sovereign {
namespace finance {
namespace indicators {

// ============================================================================
// MOVING AVERAGES
// ============================================================================

/// Simple Moving Average.
inline std::vector<double> SMA(const std::vector<double>& data, int period) {
    size_t n = data.size();
    std::vector<double> out(n, std::nan(""));
    if (n < static_cast<size_t>(period) || period <= 0) return out;

    double sum = 0.0;
    for (int i = 0; i < period; ++i) sum += data[i];
    out[period - 1] = sum / period;

    for (size_t i = period; i < n; ++i) {
        sum += data[i] - data[i - period];
        out[i] = sum / period;
    }
    return out;
}

/// Exponential Moving Average.
inline std::vector<double> EMA(const std::vector<double>& data, int period) {
    size_t n = data.size();
    std::vector<double> out(n, std::nan(""));
    if (n < static_cast<size_t>(period) || period <= 0) return out;

    double mult = 2.0 / (period + 1.0);
    double sum = 0.0;
    for (int i = 0; i < period; ++i) sum += data[i];
    out[period - 1] = sum / period;

    for (size_t i = period; i < n; ++i) {
        out[i] = (data[i] - out[i - 1]) * mult + out[i - 1];
    }
    return out;
}

/// Weighted Moving Average.
inline std::vector<double> WMA(const std::vector<double>& data, int period) {
    size_t n = data.size();
    std::vector<double> out(n, std::nan(""));
    if (n < static_cast<size_t>(period) || period <= 0) return out;

    double weight_sum = period * (period + 1) / 2.0;
    for (size_t i = period - 1; i < n; ++i) {
        double s = 0.0;
        for (int j = 0; j < period; ++j) {
            s += data[i - j] * (period - j);
        }
        out[i] = s / weight_sum;
    }
    return out;
}

/// Hull Moving Average (fast, low-lag).
inline std::vector<double> HMA(const std::vector<double>& data, int period) {
    auto wma_full = WMA(data, period);
    auto wma_half = WMA(data, period / 2);

    size_t n = data.size();
    std::vector<double> diff(n, std::nan(""));
    for (size_t i = 0; i < n; ++i) {
        if (!std::isnan(wma_full[i]) && !std::isnan(wma_half[i])) {
            diff[i] = 2.0 * wma_half[i] - wma_full[i];
        }
    }

    // Extract valid values for final WMA
    std::vector<double> valid;
    size_t start = 0;
    for (size_t i = 0; i < n; ++i) {
        if (!std::isnan(diff[i])) {
            if (valid.empty()) start = i;
            valid.push_back(diff[i]);
        }
    }

    int sqrt_period = static_cast<int>(std::sqrt(period));
    auto hma_raw = WMA(valid, sqrt_period);

    std::vector<double> out(n, std::nan(""));
    for (size_t i = 0; i < hma_raw.size(); ++i) {
        out[start + i] = hma_raw[i];
    }
    return out;
}

/// Double EMA (DEMA).
inline std::vector<double> DEMA(const std::vector<double>& data, int period) {
    auto ema1 = EMA(data, period);

    // EMA of EMA
    std::vector<double> valid;
    size_t start = 0;
    for (size_t i = 0; i < ema1.size(); ++i) {
        if (!std::isnan(ema1[i])) {
            if (valid.empty()) start = i;
            valid.push_back(ema1[i]);
        }
    }
    auto ema2 = EMA(valid, period);

    size_t n = data.size();
    std::vector<double> out(n, std::nan(""));
    for (size_t i = 0; i < ema2.size(); ++i) {
        if (!std::isnan(ema2[i])) {
            out[start + i] = 2.0 * ema1[start + i] - ema2[i];
        }
    }
    return out;
}

// ============================================================================
// MOMENTUM OSCILLATORS
// ============================================================================

/// Relative Strength Index.
inline std::vector<double> RSI(const std::vector<double>& data, int period) {
    size_t n = data.size();
    std::vector<double> out(n, std::nan(""));
    if (n <= static_cast<size_t>(period)) return out;

    double gain_sum = 0.0, loss_sum = 0.0;
    for (int i = 1; i <= period; ++i) {
        double change = data[i] - data[i - 1];
        if (change > 0) gain_sum += change;
        else loss_sum -= change;
    }

    double avg_gain = gain_sum / period;
    double avg_loss = loss_sum / period;
    out[period] = (avg_loss == 0.0) ? 100.0 : 100.0 - (100.0 / (1.0 + avg_gain / avg_loss));

    for (size_t i = period + 1; i < n; ++i) {
        double change = data[i] - data[i - 1];
        double g = (change > 0) ? change : 0.0;
        double l = (change < 0) ? -change : 0.0;
        avg_gain = (avg_gain * (period - 1) + g) / period;
        avg_loss = (avg_loss * (period - 1) + l) / period;
        out[i] = (avg_loss == 0.0) ? 100.0 : 100.0 - (100.0 / (1.0 + avg_gain / avg_loss));
    }
    return out;
}

/// Stochastic Oscillator (%K and %D).
struct StochasticResult {
    std::vector<double> k_line;
    std::vector<double> d_line;
};

inline StochasticResult Stochastic(const std::vector<double>& highs,
                                    const std::vector<double>& lows,
                                    const std::vector<double>& closes,
                                    int period_k, int smooth_k, int period_d) {
    size_t n = closes.size();
    std::vector<double> raw_k(n, std::nan(""));

    for (size_t i = period_k - 1; i < n; ++i) {
        double hh = *std::max_element(highs.begin() + i - period_k + 1, highs.begin() + i + 1);
        double ll = *std::min_element(lows.begin() + i - period_k + 1, lows.begin() + i + 1);
        raw_k[i] = (hh == ll) ? 50.0 : 100.0 * (closes[i] - ll) / (hh - ll);
    }

    // Smooth %K
    auto k = (smooth_k > 1) ? SMA(raw_k, smooth_k) : raw_k;
    // %D = SMA of %K
    auto d = SMA(k, period_d);

    return {k, d};
}

/// Williams %R.
inline std::vector<double> WilliamsR(const std::vector<double>& highs,
                                      const std::vector<double>& lows,
                                      const std::vector<double>& closes, int period) {
    size_t n = closes.size();
    std::vector<double> out(n, std::nan(""));
    for (size_t i = period - 1; i < n; ++i) {
        double hh = *std::max_element(highs.begin() + i - period + 1, highs.begin() + i + 1);
        double ll = *std::min_element(lows.begin() + i - period + 1, lows.begin() + i + 1);
        out[i] = (hh == ll) ? -50.0 : -100.0 * (hh - closes[i]) / (hh - ll);
    }
    return out;
}

/// Commodity Channel Index.
inline std::vector<double> CCI(const std::vector<double>& highs,
                                const std::vector<double>& lows,
                                const std::vector<double>& closes, int period) {
    size_t n = closes.size();
    std::vector<double> tp(n);
    for (size_t i = 0; i < n; ++i) {
        tp[i] = (highs[i] + lows[i] + closes[i]) / 3.0;
    }

    auto sma_tp = SMA(tp, period);
    std::vector<double> out(n, std::nan(""));

    for (size_t i = period - 1; i < n; ++i) {
        // Mean Absolute Deviation
        double mad = 0.0;
        for (size_t j = i - period + 1; j <= i; ++j) {
            mad += std::abs(tp[j] - sma_tp[i]);
        }
        mad /= period;
        out[i] = (mad == 0.0) ? 0.0 : (tp[i] - sma_tp[i]) / (0.015 * mad);
    }
    return out;
}

// ============================================================================
// TREND INDICATORS
// ============================================================================

/// MACD (Moving Average Convergence Divergence).
struct MACDResult {
    std::vector<double> macd_line;
    std::vector<double> signal_line;
    std::vector<double> histogram;
};

inline MACDResult MACD(const std::vector<double>& data, int fast, int slow, int signal) {
    auto fast_ema = EMA(data, fast);
    auto slow_ema = EMA(data, slow);

    size_t n = data.size();
    std::vector<double> macd_line(n, std::nan(""));
    for (size_t i = 0; i < n; ++i) {
        if (!std::isnan(fast_ema[i]) && !std::isnan(slow_ema[i])) {
            macd_line[i] = fast_ema[i] - slow_ema[i];
        }
    }

    // Extract valid MACD values for signal EMA
    std::vector<double> valid_macd;
    size_t valid_start = 0;
    for (size_t i = 0; i < n; ++i) {
        if (!std::isnan(macd_line[i])) {
            if (valid_macd.empty()) valid_start = i;
            valid_macd.push_back(macd_line[i]);
        }
    }

    auto sig_tmp = EMA(valid_macd, signal);

    std::vector<double> signal_line(n, std::nan(""));
    std::vector<double> histogram(n, std::nan(""));
    for (size_t i = 0; i < sig_tmp.size(); ++i) {
        signal_line[valid_start + i] = sig_tmp[i];
        if (!std::isnan(macd_line[valid_start + i]) && !std::isnan(sig_tmp[i])) {
            histogram[valid_start + i] = macd_line[valid_start + i] - sig_tmp[i];
        }
    }

    return {macd_line, signal_line, histogram};
}

/// Average Directional Index (ADX) — full implementation.
struct ADXResult {
    std::vector<double> adx;
    std::vector<double> plus_di;
    std::vector<double> minus_di;
};

inline ADXResult ADX(const std::vector<double>& highs,
                      const std::vector<double>& lows,
                      const std::vector<double>& closes, int period) {
    size_t n = closes.size();
    ADXResult res;
    res.adx.assign(n, std::nan(""));
    res.plus_di.assign(n, std::nan(""));
    res.minus_di.assign(n, std::nan(""));

    if (n < static_cast<size_t>(period + 1)) return res;

    // True Range, +DM, -DM
    std::vector<double> tr(n, 0.0);
    std::vector<double> plus_dm(n, 0.0);
    std::vector<double> minus_dm(n, 0.0);

    for (size_t i = 1; i < n; ++i) {
        double hl = highs[i] - lows[i];
        double hc = std::abs(highs[i] - closes[i - 1]);
        double lc = std::abs(lows[i] - closes[i - 1]);
        tr[i] = std::max({hl, hc, lc});

        double up = highs[i] - highs[i - 1];
        double down = lows[i - 1] - lows[i];

        plus_dm[i] = (up > down && up > 0) ? up : 0.0;
        minus_dm[i] = (down > up && down > 0) ? down : 0.0;
    }

    // Smoothed TR, +DM, -DM using Wilder's smoothing
    double atr = 0.0, sm_plus = 0.0, sm_minus = 0.0;
    for (int i = 1; i <= period; ++i) {
        atr += tr[i];
        sm_plus += plus_dm[i];
        sm_minus += minus_dm[i];
    }

    for (size_t i = period; i < n; ++i) {
        if (i == static_cast<size_t>(period)) {
            // First smoothed value
        } else {
            atr = atr - (atr / period) + tr[i];
            sm_plus = sm_plus - (sm_plus / period) + plus_dm[i];
            sm_minus = sm_minus - (sm_minus / period) + minus_dm[i];
        }

        double pdi = (atr > 0) ? 100.0 * sm_plus / atr : 0.0;
        double mdi = (atr > 0) ? 100.0 * sm_minus / atr : 0.0;
        res.plus_di[i] = pdi;
        res.minus_di[i] = mdi;

        double di_sum = pdi + mdi;
        double dx = (di_sum > 0) ? 100.0 * std::abs(pdi - mdi) / di_sum : 0.0;

        // Smooth DX into ADX
        if (i == static_cast<size_t>(period)) {
            res.adx[i] = dx; // First ADX = first DX
        } else if (!std::isnan(res.adx[i - 1])) {
            res.adx[i] = (res.adx[i - 1] * (period - 1) + dx) / period;
        }
    }

    return res;
}

// ============================================================================
// VOLATILITY INDICATORS
// ============================================================================

/// True Range vector.
inline std::vector<double> TrueRange(const std::vector<double>& highs,
                                      const std::vector<double>& lows,
                                      const std::vector<double>& closes) {
    size_t n = closes.size();
    std::vector<double> out(n, 0.0);
    if (n == 0) return out;
    out[0] = highs[0] - lows[0];
    for (size_t i = 1; i < n; ++i) {
        double hl = highs[i] - lows[i];
        double hc = std::abs(highs[i] - closes[i - 1]);
        double lc = std::abs(lows[i] - closes[i - 1]);
        out[i] = std::max({hl, hc, lc});
    }
    return out;
}

/// Average True Range (Wilder's smoothing).
inline std::vector<double> ATR(const std::vector<double>& highs,
                                const std::vector<double>& lows,
                                const std::vector<double>& closes, int period) {
    auto tr = TrueRange(highs, lows, closes);
    size_t n = closes.size();
    std::vector<double> out(n, std::nan(""));
    if (n < static_cast<size_t>(period)) return out;

    double sum = 0.0;
    for (int i = 0; i < period; ++i) sum += tr[i];
    out[period - 1] = sum / period;
    for (size_t i = period; i < n; ++i) {
        out[i] = (out[i - 1] * (period - 1) + tr[i]) / period;
    }
    return out;
}

/// Bollinger Bands.
struct BollingerResult {
    std::vector<double> upper;
    std::vector<double> middle;
    std::vector<double> lower;
    std::vector<double> bandwidth;
    std::vector<double> percent_b;
};

inline BollingerResult BollingerBands(const std::vector<double>& data,
                                       int period, double num_stddev) {
    auto middle = SMA(data, period);
    size_t n = data.size();
    BollingerResult res;
    res.upper.assign(n, std::nan(""));
    res.lower.assign(n, std::nan(""));
    res.middle = middle;
    res.bandwidth.assign(n, std::nan(""));
    res.percent_b.assign(n, std::nan(""));

    for (size_t i = period - 1; i < n; ++i) {
        double sd = math::stddev(&data[i - period + 1], period);
        res.upper[i] = middle[i] + num_stddev * sd;
        res.lower[i] = middle[i] - num_stddev * sd;
        if (middle[i] != 0.0) {
            res.bandwidth[i] = (res.upper[i] - res.lower[i]) / middle[i];
        }
        double band_width = res.upper[i] - res.lower[i];
        if (band_width != 0.0) {
            res.percent_b[i] = (data[i] - res.lower[i]) / band_width;
        }
    }
    return res;
}

/// Keltner Channels (EMA ± ATR multiplier).
struct KeltnerResult {
    std::vector<double> upper;
    std::vector<double> middle;
    std::vector<double> lower;
};

inline KeltnerResult KeltnerChannels(const std::vector<double>& highs,
                                      const std::vector<double>& lows,
                                      const std::vector<double>& closes,
                                      int ema_period, int atr_period, double mult) {
    auto mid = EMA(closes, ema_period);
    auto atr = ATR(highs, lows, closes, atr_period);
    size_t n = closes.size();

    KeltnerResult res;
    res.middle = mid;
    res.upper.assign(n, std::nan(""));
    res.lower.assign(n, std::nan(""));

    for (size_t i = 0; i < n; ++i) {
        if (!std::isnan(mid[i]) && !std::isnan(atr[i])) {
            res.upper[i] = mid[i] + mult * atr[i];
            res.lower[i] = mid[i] - mult * atr[i];
        }
    }
    return res;
}

// ============================================================================
// VOLUME INDICATORS
// ============================================================================

/// On-Balance Volume.
inline std::vector<double> OBV(const std::vector<double>& closes,
                                const std::vector<double>& volumes) {
    size_t n = closes.size();
    std::vector<double> out(n, 0.0);
    if (n == 0) return out;
    out[0] = volumes[0];
    for (size_t i = 1; i < n; ++i) {
        if (closes[i] > closes[i - 1]) out[i] = out[i - 1] + volumes[i];
        else if (closes[i] < closes[i - 1]) out[i] = out[i - 1] - volumes[i];
        else out[i] = out[i - 1];
    }
    return out;
}

/// Volume Weighted Average Price (cumulative).
inline std::vector<double> VWAP(const std::vector<double>& highs,
                                 const std::vector<double>& lows,
                                 const std::vector<double>& closes,
                                 const std::vector<double>& volumes) {
    size_t n = closes.size();
    std::vector<double> out(n, 0.0);
    double cum_vol = 0.0, cum_pv = 0.0;
    for (size_t i = 0; i < n; ++i) {
        double tp = (highs[i] + lows[i] + closes[i]) / 3.0;
        cum_vol += volumes[i];
        cum_pv += tp * volumes[i];
        out[i] = (cum_vol > 0) ? cum_pv / cum_vol : tp;
    }
    return out;
}

/// Money Flow Index (volume-weighted RSI).
inline std::vector<double> MFI(const std::vector<double>& highs,
                                const std::vector<double>& lows,
                                const std::vector<double>& closes,
                                const std::vector<double>& volumes, int period) {
    size_t n = closes.size();
    std::vector<double> out(n, std::nan(""));
    if (n <= static_cast<size_t>(period)) return out;

    std::vector<double> tp(n);
    for (size_t i = 0; i < n; ++i) {
        tp[i] = (highs[i] + lows[i] + closes[i]) / 3.0;
    }

    for (size_t i = period; i < n; ++i) {
        double pos_flow = 0.0, neg_flow = 0.0;
        for (size_t j = i - period + 1; j <= i; ++j) {
            double mf = tp[j] * volumes[j];
            if (tp[j] > tp[j - 1]) pos_flow += mf;
            else neg_flow += mf;
        }
        out[i] = (neg_flow == 0.0) ? 100.0 : 100.0 - (100.0 / (1.0 + pos_flow / neg_flow));
    }
    return out;
}

} // namespace indicators
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_INDICATORS_HPP
