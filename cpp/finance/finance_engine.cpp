/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — C FFI Bridge
 *  Exposes the C++ finance engine to Go via cgo-compatible C interface.
 * ============================================================================
 */

#include "engine.hpp"
#include <cstdlib>
#include <cstring>
#include <cstdio>

// Global engine instance
static sovereign::finance::FinanceEngine* g_engine = nullptr;

extern "C" {

// ============================================================================
// LIFECYCLE
// ============================================================================

void* finance_engine_init() {
    if (!g_engine) {
        g_engine = new sovereign::finance::FinanceEngine();
    }
    return static_cast<void*>(g_engine);
}

void finance_engine_free() {
    delete g_engine;
    g_engine = nullptr;
}

// ============================================================================
// DATA LOADING
// ============================================================================

/// Load OHLCV data for a symbol. Arrays must be of length `n`.
void finance_load_data(const char* symbol,
                       const int64_t* timestamps,
                       const double* opens,
                       const double* highs,
                       const double* lows,
                       const double* closes,
                       const double* volumes,
                       int n) {
    if (!g_engine || !symbol || n <= 0) return;

    sovereign::finance::TimeSeries ts(n);
    for (int i = 0; i < n; ++i) {
        sovereign::finance::Bar bar;
        bar.timestamp = timestamps[i];
        bar.open   = opens[i];
        bar.high   = highs[i];
        bar.low    = lows[i];
        bar.close  = closes[i];
        bar.volume = volumes[i];
        ts.push_back(bar);
    }

    g_engine->load_data(std::string(symbol), ts);
}

// ============================================================================
// ANALYSIS (Returns JSON strings — caller must free with finance_free_string)
// ============================================================================

/// Predict future price. Returns a JSON string.
char* finance_predict(const char* symbol, int horizon) {
    if (!g_engine || !symbol) return nullptr;

    auto pred = g_engine->predict(std::string(symbol), horizon);

    // Serialize to JSON manually (no external JSON lib dependency)
    char buf[2048];
    snprintf(buf, sizeof(buf),
        "{\"model\":\"%s\","
        "\"predicted_price\":%.4f,"
        "\"confidence\":%.4f,"
        "\"direction\":\"%s\","
        "\"horizon\":%d,"
        "\"upper_bound\":%.4f,"
        "\"lower_bound\":%.4f,"
        "\"r2_score\":%.4f,"
        "\"rmse\":%.4f}",
        pred.model.c_str(),
        pred.predicted_price,
        pred.confidence,
        sovereign::finance::signal_to_string(pred.direction),
        pred.horizon_days,
        pred.upper_bound,
        pred.lower_bound,
        pred.r2_score,
        pred.rmse);

    return strdup(buf);
}

/// Get risk metrics. Returns a JSON string.
char* finance_risk(const char* symbol) {
    if (!g_engine || !symbol) return nullptr;

    auto rm = g_engine->compute_risk(std::string(symbol));

    char buf[2048];
    snprintf(buf, sizeof(buf),
        "{\"sharpe\":%.4f,"
        "\"sortino\":%.4f,"
        "\"max_drawdown\":%.4f,"
        "\"max_dd_duration\":%d,"
        "\"var_95\":%.6f,"
        "\"var_99\":%.6f,"
        "\"cvar_95\":%.6f,"
        "\"volatility_annual\":%.4f,"
        "\"beta\":%.4f,"
        "\"alpha\":%.4f,"
        "\"calmar\":%.4f,"
        "\"win_rate\":%.4f,"
        "\"skewness\":%.4f,"
        "\"kurtosis\":%.4f}",
        rm.sharpe_ratio, rm.sortino_ratio, rm.max_drawdown, rm.max_dd_duration,
        rm.var_95, rm.var_99, rm.cvar_95, rm.volatility_ann,
        rm.beta, rm.alpha, rm.calmar_ratio, rm.win_rate,
        rm.skewness, rm.kurtosis);

    return strdup(buf);
}

/// Run a backtest. Returns a JSON string.
char* finance_backtest(const char* symbol, const char* strategy) {
    if (!g_engine || !symbol || !strategy) return nullptr;

    auto result = g_engine->backtest(std::string(symbol), std::string(strategy));

    char buf[2048];
    snprintf(buf, sizeof(buf),
        "{\"final_capital\":%.2f,"
        "\"total_return\":%.4f,"
        "\"trade_count\":%d,"
        "\"win_count\":%d,"
        "\"loss_count\":%d,"
        "\"avg_win\":%.2f,"
        "\"avg_loss\":%.2f,"
        "\"largest_win\":%.2f,"
        "\"largest_loss\":%.2f,"
        "\"sharpe\":%.4f,"
        "\"max_drawdown\":%.4f}",
        result.final_capital, result.total_return, result.trade_count,
        result.win_count, result.loss_count, result.avg_win, result.avg_loss,
        result.largest_win, result.largest_loss,
        result.risk.sharpe_ratio, result.risk.max_drawdown);

    return strdup(buf);
}

/// Analyze sentiment of a headline. Returns a JSON string.
char* finance_analyze_sentiment(const char* headline) {
    if (!g_engine || !headline) return nullptr;

    auto result = g_engine->analyze_sentiment(std::string(headline));

    char buf[1024];
    snprintf(buf, sizeof(buf),
        "{\"score\":%.4f,"
        "\"positive_words\":%d,"
        "\"negative_words\":%d}",
        result.score, result.positive_words, result.negative_words);

    return strdup(buf);
}

/// Analyze company fundamentals. Returns a JSON string.
char* finance_analyze_fundamentals(const char* symbol) {
    if (!g_engine || !symbol) return nullptr;

    auto result = g_engine->analyze_fundamentals(std::string(symbol));

    char buf[1024];
    snprintf(buf, sizeof(buf),
        "{\"total_score\":%.2f,"
        "\"rating\":\"%s\","
        "\"reasoning\":\"%s\"}",
        result.total_score, result.rating.c_str(), result.reasoning.c_str());

    return strdup(buf);
}

/// Free a string returned by the engine.
void finance_free_string(char* s) {
    free(s);
}

} // extern "C"
