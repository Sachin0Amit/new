#ifndef SOVEREIGN_FINANCE_MATH_HPP
#define SOVEREIGN_FINANCE_MATH_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Core Mathematical Primitives
 *  SIMD-ready statistical functions operating on raw double arrays.
 *  These are the fundamental building blocks for all indicators and models.
 * ============================================================================
 */

#include <cmath>
#include <vector>
#include <algorithm>
#include <numeric>
#include <random>
#include <cassert>

namespace sovereign {
namespace finance {
namespace math {

// ============================================================================
// DESCRIPTIVE STATISTICS
// ============================================================================

/// Arithmetic mean of [data, data+n).
inline double mean(const double* data, size_t n) {
    if (n == 0) return 0.0;
    double sum = 0.0;
    for (size_t i = 0; i < n; ++i) sum += data[i];
    return sum / static_cast<double>(n);
}

inline double mean(const std::vector<double>& v) {
    return mean(v.data(), v.size());
}

/// Sample variance (Bessel-corrected).
inline double variance(const double* data, size_t n) {
    if (n < 2) return 0.0;
    double m = mean(data, n);
    double sum_sq = 0.0;
    for (size_t i = 0; i < n; ++i) {
        double d = data[i] - m;
        sum_sq += d * d;
    }
    return sum_sq / static_cast<double>(n - 1);
}

inline double variance(const std::vector<double>& v) {
    return variance(v.data(), v.size());
}

/// Sample standard deviation.
inline double stddev(const double* data, size_t n) {
    return std::sqrt(variance(data, n));
}

inline double stddev(const std::vector<double>& v) {
    return stddev(v.data(), v.size());
}

/// Sample covariance between two arrays of equal length.
inline double covariance(const double* x, const double* y, size_t n) {
    if (n < 2) return 0.0;
    double mx = mean(x, n);
    double my = mean(y, n);
    double sum = 0.0;
    for (size_t i = 0; i < n; ++i) {
        sum += (x[i] - mx) * (y[i] - my);
    }
    return sum / static_cast<double>(n - 1);
}

/// Pearson correlation coefficient.
inline double correlation(const double* x, const double* y, size_t n) {
    double sx = stddev(x, n);
    double sy = stddev(y, n);
    if (sx == 0.0 || sy == 0.0) return 0.0;
    return covariance(x, y, n) / (sx * sy);
}

inline double correlation(const std::vector<double>& x, const std::vector<double>& y) {
    size_t n = std::min(x.size(), y.size());
    return correlation(x.data(), y.data(), n);
}

/// Skewness (Fisher's definition).
inline double skewness(const std::vector<double>& data) {
    size_t n = data.size();
    if (n < 3) return 0.0;
    double m = mean(data);
    double s = stddev(data);
    if (s == 0.0) return 0.0;

    double sum = 0.0;
    for (double v : data) {
        double d = (v - m) / s;
        sum += d * d * d;
    }
    return (static_cast<double>(n) / ((n - 1.0) * (n - 2.0))) * sum;
}

/// Excess kurtosis (Fisher's definition).
inline double kurtosis(const std::vector<double>& data) {
    size_t n = data.size();
    if (n < 4) return 0.0;
    double m = mean(data);
    double s = stddev(data);
    if (s == 0.0) return 0.0;

    double sum = 0.0;
    for (double v : data) {
        double d = (v - m) / s;
        sum += d * d * d * d;
    }
    double nd = static_cast<double>(n);
    double k = (nd * (nd + 1.0)) / ((nd - 1.0) * (nd - 2.0) * (nd - 3.0)) * sum;
    k -= (3.0 * (nd - 1.0) * (nd - 1.0)) / ((nd - 2.0) * (nd - 3.0));
    return k;
}

/// Percentile (linear interpolation, data must be sorted).
inline double percentile(const double* sorted, size_t n, double p) {
    if (n == 0) return 0.0;
    double k = (p / 100.0) * static_cast<double>(n - 1);
    size_t f = static_cast<size_t>(std::floor(k));
    size_t c = static_cast<size_t>(std::ceil(k));
    if (f == c) return sorted[f];
    return sorted[f] * (static_cast<double>(c) - k) +
           sorted[c] * (k - static_cast<double>(f));
}

/// Sum of elements.
inline double sum(const double* data, size_t n) {
    double s = 0.0;
    for (size_t i = 0; i < n; ++i) s += data[i];
    return s;
}

inline double sum(const std::vector<double>& v) {
    return sum(v.data(), v.size());
}

/// Min of elements.
inline double min_val(const double* data, size_t n) {
    if (n == 0) return 0.0;
    double m = data[0];
    for (size_t i = 1; i < n; ++i) if (data[i] < m) m = data[i];
    return m;
}

/// Max of elements.
inline double max_val(const double* data, size_t n) {
    if (n == 0) return 0.0;
    double m = data[0];
    for (size_t i = 1; i < n; ++i) if (data[i] > m) m = data[i];
    return m;
}

// ============================================================================
// CUMULATIVE & ROLLING
// ============================================================================

/// Cumulative sum.
inline std::vector<double> cumsum(const std::vector<double>& data) {
    std::vector<double> out(data.size());
    double s = 0.0;
    for (size_t i = 0; i < data.size(); ++i) {
        s += data[i];
        out[i] = s;
    }
    return out;
}

/// Rolling mean over window.
inline std::vector<double> rolling_mean(const std::vector<double>& data, size_t window) {
    std::vector<double> out(data.size(), std::nan(""));
    if (data.size() < window) return out;

    double s = 0.0;
    for (size_t i = 0; i < window; ++i) s += data[i];
    out[window - 1] = s / static_cast<double>(window);
    for (size_t i = window; i < data.size(); ++i) {
        s += data[i] - data[i - window];
        out[i] = s / static_cast<double>(window);
    }
    return out;
}

/// Rolling standard deviation over window.
inline std::vector<double> rolling_stddev(const std::vector<double>& data, size_t window) {
    std::vector<double> out(data.size(), std::nan(""));
    if (data.size() < window) return out;

    for (size_t i = window - 1; i < data.size(); ++i) {
        out[i] = stddev(&data[i - window + 1], window);
    }
    return out;
}

// ============================================================================
// RANDOM NUMBER GENERATION
// ============================================================================

/// Thread-local Mersenne Twister PRNG.
inline std::mt19937_64& get_rng() {
    thread_local std::mt19937_64 rng{std::random_device{}()};
    return rng;
}

/// Standard normal variate via Box-Muller.
inline double randn() {
    std::normal_distribution<double> dist(0.0, 1.0);
    return dist(get_rng());
}

/// Uniform [0, 1).
inline double randu() {
    std::uniform_real_distribution<double> dist(0.0, 1.0);
    return dist(get_rng());
}

/// Standard normal PDF.
inline double norm_pdf(double x) {
    return INV_SQRT_2PI * std::exp(-0.5 * x * x);
}

/// Standard normal CDF (Abramowitz & Stegun approximation).
inline double norm_cdf(double x) {
    static const double a1 =  0.254829592;
    static const double a2 = -0.284496736;
    static const double a3 =  1.421413741;
    static const double a4 = -1.453152027;
    static const double a5 =  1.061405429;
    static const double p  =  0.3275911;

    int sign = 1;
    if (x < 0) { sign = -1; x = -x; }
    double t = 1.0 / (1.0 + p * x);
    double y = 1.0 - (((((a5 * t + a4) * t) + a3) * t + a2) * t + a1) * t * std::exp(-x * x / 2.0);
    return 0.5 * (1.0 + sign * y);
}

} // namespace math
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_MATH_HPP
