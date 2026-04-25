#ifndef SOVEREIGN_FINANCE_PATTERNS_HPP
#define SOVEREIGN_FINANCE_PATTERNS_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Pattern Recognition
 *  Algorithmic detection of candlestick and chart patterns.
 *  21 distinct patterns supported.
 * ============================================================================
 */

#include "types.hpp"
#include <vector>
#include <cmath>
#include <algorithm>

namespace sovereign {
namespace finance {
namespace patterns {

// ============================================================================
// CANDLESTICK PATTERNS
// ============================================================================

/// Detect all candlestick patterns in a time series.
inline std::vector<PatternDetection> detect_candle_patterns(const TimeSeries& ts) {
    std::vector<PatternDetection> detections;
    if (ts.size() < 3) return detections;

    for (size_t i = 2; i < ts.size(); ++i) {
        const Bar& prev2 = ts[i - 2];
        const Bar& prev1 = ts[i - 1];
        const Bar& curr  = ts[i];

        double body_curr = curr.body();
        double range_curr = curr.range();
        double upper_curr = curr.upper_shadow();
        double lower_curr = curr.lower_shadow();

        // --- DOJI ---
        if (range_curr > 0.0 && body_curr <= range_curr * 0.1) {
            detections.push_back({PatternType::DOJI, 0.8, Signal::NEUTRAL, i, i});
        }

        // --- SPINNING TOP ---
        if (range_curr > 0.0 && body_curr <= range_curr * 0.3 &&
            upper_curr > body_curr * 0.5 && lower_curr > body_curr * 0.5) {
            detections.push_back({PatternType::SPINNING_TOP, 0.65, Signal::NEUTRAL, i, i});
        }

        // --- HAMMER --- (in downtrend)
        if (prev1.is_bearish() && lower_curr >= body_curr * 2.0 && upper_curr <= body_curr * 0.2) {
            detections.push_back({PatternType::HAMMER, 0.75, Signal::BUY, i, i});
        }

        // --- SHOOTING STAR --- (in uptrend)
        if (prev1.is_bullish() && upper_curr >= body_curr * 2.0 && lower_curr <= body_curr * 0.2) {
            detections.push_back({PatternType::SHOOTING_STAR, 0.75, Signal::SELL, i, i});
        }

        // --- BULLISH ENGULFING ---
        if (prev1.is_bearish() && curr.is_bullish() &&
            curr.open < prev1.close && curr.close > prev1.open) {
            detections.push_back({PatternType::BULLISH_ENGULFING, 0.85, Signal::BUY, i - 1, i});
        }

        // --- BEARISH ENGULFING ---
        if (prev1.is_bullish() && curr.is_bearish() &&
            curr.open > prev1.close && curr.close < prev1.open) {
            detections.push_back({PatternType::BEARISH_ENGULFING, 0.85, Signal::SELL, i - 1, i});
        }

        // --- MORNING STAR ---
        if (prev2.is_bearish() && prev2.body() > prev2.range() * 0.5 &&
            prev1.body() < prev1.range() * 0.3 &&
            curr.is_bullish() && curr.close > (prev2.open + prev2.close) / 2.0) {
            detections.push_back({PatternType::MORNING_STAR, 0.90, Signal::BUY, i - 2, i});
        }

        // --- EVENING STAR ---
        if (prev2.is_bullish() && prev2.body() > prev2.range() * 0.5 &&
            prev1.body() < prev1.range() * 0.3 &&
            curr.is_bearish() && curr.close < (prev2.open + prev2.close) / 2.0) {
            detections.push_back({PatternType::EVENING_STAR, 0.90, Signal::SELL, i - 2, i});
        }

        // --- BULLISH HARAMI ---
        if (prev1.is_bearish() && curr.is_bullish() &&
            curr.open > prev1.close && curr.close < prev1.open &&
            body_curr < prev1.body() * 0.5) {
            detections.push_back({PatternType::HARAMI_BULLISH, 0.70, Signal::BUY, i - 1, i});
        }

        // --- BEARISH HARAMI ---
        if (prev1.is_bullish() && curr.is_bearish() &&
            curr.open < prev1.close && curr.close > prev1.open &&
            body_curr < prev1.body() * 0.5) {
            detections.push_back({PatternType::HARAMI_BEARISH, 0.70, Signal::SELL, i - 1, i});
        }

        // --- MARUBOZU BULLISH (no shadows) ---
        if (curr.is_bullish() && upper_curr < range_curr * 0.02 && lower_curr < range_curr * 0.02) {
            detections.push_back({PatternType::MARUBOZU_BULLISH, 0.80, Signal::STRONG_BUY, i, i});
        }

        // --- MARUBOZU BEARISH ---
        if (curr.is_bearish() && upper_curr < range_curr * 0.02 && lower_curr < range_curr * 0.02) {
            detections.push_back({PatternType::MARUBOZU_BEARISH, 0.80, Signal::STRONG_SELL, i, i});
        }

        // --- TWEEZER TOP --- (equal highs in uptrend)
        if (prev1.is_bullish() && std::abs(prev1.high - curr.high) < range_curr * 0.01 &&
            curr.is_bearish()) {
            detections.push_back({PatternType::TWEEZER_TOP, 0.72, Signal::SELL, i - 1, i});
        }

        // --- TWEEZER BOTTOM --- (equal lows in downtrend)
        if (prev1.is_bearish() && std::abs(prev1.low - curr.low) < range_curr * 0.01 &&
            curr.is_bullish()) {
            detections.push_back({PatternType::TWEEZER_BOTTOM, 0.72, Signal::BUY, i - 1, i});
        }
    }

    // --- THREE WHITE SOLDIERS & THREE BLACK CROWS ---
    for (size_t i = 2; i < ts.size(); ++i) {
        const Bar& a = ts[i - 2];
        const Bar& b = ts[i - 1];
        const Bar& c = ts[i];

        // Three White Soldiers: 3 consecutive bullish bars with higher closes
        if (a.is_bullish() && b.is_bullish() && c.is_bullish() &&
            b.close > a.close && c.close > b.close &&
            b.open > a.open && c.open > b.open) {
            detections.push_back({PatternType::THREE_WHITE_SOLDIERS, 0.88, Signal::STRONG_BUY, i - 2, i});
        }

        // Three Black Crows: 3 consecutive bearish bars with lower closes
        if (a.is_bearish() && b.is_bearish() && c.is_bearish() &&
            b.close < a.close && c.close < b.close &&
            b.open < a.open && c.open < b.open) {
            detections.push_back({PatternType::THREE_BLACK_CROWS, 0.88, Signal::STRONG_SELL, i - 2, i});
        }
    }

    return detections;
}

// ============================================================================
// CHART PATTERNS (Support/Resistance based)
// ============================================================================

/// Find local peaks and troughs for chart pattern analysis.
struct PeakTrough {
    size_t index;
    double value;
    bool is_peak; // true = peak, false = trough
};

inline std::vector<PeakTrough> find_peaks_troughs(const std::vector<double>& data, int window = 5) {
    std::vector<PeakTrough> pts;
    if (data.size() < static_cast<size_t>(2 * window + 1)) return pts;

    for (size_t i = window; i < data.size() - window; ++i) {
        bool is_peak = true;
        bool is_trough = true;
        for (int j = -window; j <= window; ++j) {
            if (j == 0) continue;
            if (data[i + j] >= data[i]) is_peak = false;
            if (data[i + j] <= data[i]) is_trough = false;
        }
        if (is_peak) pts.push_back({i, data[i], true});
        if (is_trough) pts.push_back({i, data[i], false});
    }
    return pts;
}

/// Detect Double Top pattern.
inline std::vector<PatternDetection> detect_double_top(const std::vector<double>& closes, int window = 10) {
    std::vector<PatternDetection> results;
    auto pts = find_peaks_troughs(closes, window);

    // Find two peaks at roughly the same level
    for (size_t i = 0; i < pts.size(); ++i) {
        if (!pts[i].is_peak) continue;
        for (size_t j = i + 1; j < pts.size(); ++j) {
            if (!pts[j].is_peak) continue;
            double diff = std::abs(pts[i].value - pts[j].value);
            double avg = (pts[i].value + pts[j].value) / 2.0;
            if (diff / avg < 0.02) { // Within 2% of each other
                results.push_back({PatternType::DOUBLE_TOP, 0.78, Signal::SELL, pts[i].index, pts[j].index});
            }
        }
    }
    return results;
}

/// Detect Double Bottom pattern.
inline std::vector<PatternDetection> detect_double_bottom(const std::vector<double>& closes, int window = 10) {
    std::vector<PatternDetection> results;
    auto pts = find_peaks_troughs(closes, window);

    for (size_t i = 0; i < pts.size(); ++i) {
        if (pts[i].is_peak) continue;
        for (size_t j = i + 1; j < pts.size(); ++j) {
            if (pts[j].is_peak) continue;
            double diff = std::abs(pts[i].value - pts[j].value);
            double avg = (pts[i].value + pts[j].value) / 2.0;
            if (diff / avg < 0.02) {
                results.push_back({PatternType::DOUBLE_BOTTOM, 0.78, Signal::BUY, pts[i].index, pts[j].index});
            }
        }
    }
    return results;
}

/// Support & Resistance level detection using clustering.
struct SupportResistance {
    double level;
    int    touches;
    bool   is_support; // true=support, false=resistance
};

inline std::vector<SupportResistance> find_sr_levels(const std::vector<double>& closes,
                                                      double tolerance_pct = 1.0) {
    auto pts = find_peaks_troughs(closes, 5);
    std::vector<SupportResistance> levels;

    for (const auto& pt : pts) {
        bool merged = false;
        for (auto& lvl : levels) {
            if (std::abs(pt.value - lvl.level) / lvl.level * 100.0 < tolerance_pct) {
                lvl.level = (lvl.level * lvl.touches + pt.value) / (lvl.touches + 1);
                lvl.touches++;
                merged = true;
                break;
            }
        }
        if (!merged) {
            levels.push_back({pt.value, 1, !pt.is_peak});
        }
    }

    // Sort by number of touches (strongest first)
    std::sort(levels.begin(), levels.end(), [](const auto& a, const auto& b) {
        return a.touches > b.touches;
    });

    return levels;
}

} // namespace patterns
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_PATTERNS_HPP
