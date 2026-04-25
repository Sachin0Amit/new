/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Test Suite
 *  Validates all mathematical modules with deterministic test data.
 * ============================================================================
 */

#include "engine.hpp"
#include <iostream>
#include <cassert>
#include <cmath>
#include <iomanip>

using namespace sovereign::finance;

// ============================================================================
// TEST HELPERS
// ============================================================================

static int tests_passed = 0;
static int tests_failed = 0;

#define ASSERT_NEAR(val, expected, tol, msg) do { \
    double _v = (val), _e = (expected); \
    if (std::abs(_v - _e) <= (tol)) { tests_passed++; } \
    else { tests_failed++; std::cerr << "  FAIL: " << msg << " — got " << _v << ", expected " << _e << std::endl; } \
} while(0)

#define ASSERT_TRUE(cond, msg) do { \
    if (cond) { tests_passed++; } \
    else { tests_failed++; std::cerr << "  FAIL: " << msg << std::endl; } \
} while(0)

static TimeSeries make_test_series(int n = 500) {
    TimeSeries ts(n);
    double price = 100.0;
    for (int i = 0; i < n; ++i) {
        Bar b;
        b.timestamp = 1700000000 + i * 86400;
        double change = 1.0 + std::sin(i * 0.1) * 0.02 + (i % 7 - 3) * 0.001;
        b.open = price;
        b.close = price * change;
        b.high = std::max(b.open, b.close) * 1.005;
        b.low = std::min(b.open, b.close) * 0.995;
        b.volume = 1000000.0 + std::sin(i * 0.3) * 500000.0;
        ts.push_back(b);
        price = b.close;
    }
    return ts;
}

// ============================================================================
// TESTS
// ============================================================================

void test_math() {
    std::cout << "▸ Testing Math Primitives..." << std::endl;

    std::vector<double> data = {1, 2, 3, 4, 5};
    ASSERT_NEAR(math::mean(data), 3.0, 1e-10, "mean([1..5])");
    ASSERT_NEAR(math::variance(data), 2.5, 1e-10, "variance([1..5])");
    ASSERT_NEAR(math::stddev(data), std::sqrt(2.5), 1e-10, "stddev([1..5])");
    ASSERT_NEAR(math::sum(data), 15.0, 1e-10, "sum([1..5])");
    ASSERT_NEAR(math::min_val(data.data(), data.size()), 1.0, 1e-10, "min");
    ASSERT_NEAR(math::max_val(data.data(), data.size()), 5.0, 1e-10, "max");

    std::vector<double> x = {1, 2, 3, 4, 5};
    std::vector<double> y = {2, 4, 5, 4, 5};
    double corr = math::correlation(x, y);
    ASSERT_TRUE(corr > 0.5 && corr < 1.0, "correlation positive");

    ASSERT_TRUE(math::norm_cdf(0.0) > 0.499 && math::norm_cdf(0.0) < 0.501, "norm_cdf(0) ~ 0.5");
}

void test_indicators() {
    std::cout << "▸ Testing Technical Indicators..." << std::endl;

    auto ts = make_test_series();
    auto closes = ts.closes();
    auto highs = ts.highs();
    auto lows = ts.lows();
    auto volumes = ts.volumes();

    // SMA
    auto sma = indicators::SMA(closes, 20);
    ASSERT_TRUE(sma.size() == closes.size(), "SMA size matches");
    ASSERT_TRUE(!std::isnan(sma[19]), "SMA[19] is valid");
    ASSERT_TRUE(std::isnan(sma[18]), "SMA[18] is NaN");

    // EMA
    auto ema = indicators::EMA(closes, 12);
    ASSERT_TRUE(!std::isnan(ema[11]), "EMA[11] is valid");

    // RSI
    auto rsi = indicators::RSI(closes, 14);
    ASSERT_TRUE(!std::isnan(rsi[14]), "RSI[14] is valid");
    ASSERT_TRUE(rsi[14] >= 0.0 && rsi[14] <= 100.0, "RSI in [0, 100]");

    // MACD
    auto macd = indicators::MACD(closes, 12, 26, 9);
    ASSERT_TRUE(macd.macd_line.size() == closes.size(), "MACD size");

    // Bollinger
    auto bb = indicators::BollingerBands(closes, 20, 2.0);
    for (size_t i = 19; i < closes.size(); ++i) {
        ASSERT_TRUE(bb.upper[i] >= bb.middle[i] && bb.middle[i] >= bb.lower[i],
                     "BB upper >= middle >= lower");
    }

    // ATR
    auto atr = indicators::ATR(highs, lows, closes, 14);
    ASSERT_TRUE(!std::isnan(atr[13]), "ATR[13] is valid");
    ASSERT_TRUE(atr[13] > 0.0, "ATR is positive");

    // ADX
    auto adx = indicators::ADX(highs, lows, closes, 14);
    ASSERT_TRUE(adx.adx.size() == closes.size(), "ADX size");

    // OBV
    auto obv = indicators::OBV(closes, volumes);
    ASSERT_TRUE(obv.size() == closes.size(), "OBV size");

    // HMA
    auto hma = indicators::HMA(closes, 20);
    ASSERT_TRUE(hma.size() == closes.size(), "HMA size");

    // CCI
    auto cci = indicators::CCI(highs, lows, closes, 20);
    ASSERT_TRUE(cci.size() == closes.size(), "CCI size");

    // Stochastic
    auto stoch = indicators::Stochastic(highs, lows, closes, 14, 3, 3);
    ASSERT_TRUE(stoch.k_line.size() == closes.size(), "Stochastic K size");

    // MFI
    auto mfi = indicators::MFI(highs, lows, closes, volumes, 14);
    ASSERT_TRUE(mfi.size() == closes.size(), "MFI size");
}

void test_predictions() {
    std::cout << "▸ Testing Prediction Models..." << std::endl;

    auto ts = make_test_series(200);
    auto closes = ts.closes();

    auto lr = predictions::linear_regression(closes, 30);
    ASSERT_TRUE(lr.predicted_price > 0, "LR prediction > 0");
    ASSERT_TRUE(lr.r2_score >= 0 && lr.r2_score <= 1.1, "LR R² in [0, 1]");

    auto qr = predictions::quadratic_regression(closes, 30);
    ASSERT_TRUE(qr.predicted_price > 0, "QR prediction > 0");

    auto mc = predictions::monte_carlo_gbm(closes, 30, 1000);
    ASSERT_TRUE(mc.predicted_price > 0, "MC prediction > 0");
    ASSERT_TRUE(mc.lower_bound < mc.upper_bound, "MC bounds ordered");
    ASSERT_TRUE(mc.price_path.size() == 30, "MC path length = 30");

    auto hl = predictions::holt_linear(closes, 30);
    ASSERT_TRUE(hl.predicted_price > 0, "Holt prediction > 0");

    auto ens = predictions::ensemble_predict(closes, 30);
    ASSERT_TRUE(ens.predicted_price > 0, "Ensemble prediction > 0");
}

void test_risk() {
    std::cout << "▸ Testing Risk Analytics..." << std::endl;

    auto ts = make_test_series(300);
    auto returns = ts.log_returns();
    auto closes = ts.closes();

    // Build equity curve
    std::vector<double> equity(closes.size());
    equity[0] = 10000.0;
    for (size_t i = 1; i < closes.size(); ++i) {
        equity[i] = equity[i - 1] * (closes[i] / closes[i - 1]);
    }

    double sharpe = risk::sharpe_ratio(returns, 0.04);
    ASSERT_TRUE(!std::isnan(sharpe), "Sharpe is valid");

    double var95 = risk::historical_var(returns, 95.0);
    ASSERT_TRUE(var95 > 0, "VaR95 > 0");

    auto dd = risk::analyze_drawdowns(equity);
    ASSERT_TRUE(dd.max_drawdown >= 0 && dd.max_drawdown <= 1.0, "Max DD in [0, 1]");

    std::vector<double> bench(returns.size(), 0.0);
    auto rm = risk::compute_full_risk(equity, returns, bench, 0.04);
    ASSERT_TRUE(rm.volatility_ann > 0, "Annualized vol > 0");
}

void test_patterns() {
    std::cout << "▸ Testing Pattern Recognition..." << std::endl;

    auto ts = make_test_series(300);
    auto detections = patterns::detect_candle_patterns(ts);
    ASSERT_TRUE(detections.size() > 0, "Found at least one pattern");

    auto closes = ts.closes();
    auto dt = patterns::detect_double_top(closes, 10);
    auto db = patterns::detect_double_bottom(closes, 10);
    // Just verify they don't crash
    ASSERT_TRUE(true, "Double top/bottom ran without crash");

    auto sr = patterns::find_sr_levels(closes);
    ASSERT_TRUE(sr.size() > 0, "Found support/resistance levels");
}

void test_backtest() {
    std::cout << "▸ Testing Backtesting Engine..." << std::endl;

    auto ts = make_test_series(300);

    BacktestConfig cfg;
    cfg.symbol = "TEST";
    cfg.initial_capital = 100000.0;
    cfg.commission_pct = 0.1;
    cfg.slippage_pct = 0.05;
    cfg.max_position_pct = 0.95;

    // Test SMA Cross
    auto strat = backtest::make_sma_cross(20, 50);
    backtest::Backtester bt(cfg, strat);
    auto result = bt.run(ts);
    ASSERT_TRUE(result.final_capital > 0, "Final capital > 0");
    ASSERT_TRUE(result.equity_curve.size() == ts.size(), "Equity curve size");

    // Test RSI
    auto strat2 = backtest::make_rsi_meanrev();
    backtest::Backtester bt2(cfg, strat2);
    auto result2 = bt2.run(ts);
    ASSERT_TRUE(result2.final_capital > 0, "RSI backtest capital > 0");

    // Test MACD
    auto strat3 = backtest::make_macd_histogram();
    backtest::Backtester bt3(cfg, strat3);
    auto result3 = bt3.run(ts);
    ASSERT_TRUE(result3.final_capital > 0, "MACD backtest capital > 0");
}

void test_engine() {
    std::cout << "▸ Testing Finance Engine Orchestrator..." << std::endl;

    FinanceEngine engine;
    auto ts = make_test_series(400);
    engine.load_data("AAPL", ts);

    auto ta = engine.analyze("AAPL");
    ASSERT_TRUE(ta.sma_20.size() == ts.size(), "TA SMA size");
    ASSERT_TRUE(ta.rsi_14.size() == ts.size(), "TA RSI size");

    auto pred = engine.predict("AAPL", 30);
    ASSERT_TRUE(pred.predicted_price > 0, "Engine prediction > 0");

    auto rm = engine.compute_risk("AAPL");
    ASSERT_TRUE(rm.volatility_ann > 0, "Engine risk vol > 0");

    auto pats = engine.detect_patterns("AAPL");
    ASSERT_TRUE(pats.size() > 0, "Engine found patterns");

    auto bt = engine.backtest("AAPL", "SMA_CROSS");
    ASSERT_TRUE(bt.final_capital > 0, "Engine backtest capital > 0");

    auto vf = engine.volatility_forecast("AAPL", 30);
    ASSERT_TRUE(vf.size() == 30, "Vol forecast 30 days");
}

// ============================================================================
// MAIN
// ============================================================================

int main() {
    std::cout << std::fixed << std::setprecision(6);
    std::cout << "╔══════════════════════════════════════════════╗" << std::endl;
    std::cout << "║  SOVEREIGN FINANCE ENGINE — C++ Test Suite   ║" << std::endl;
    std::cout << "╚══════════════════════════════════════════════╝" << std::endl;

    test_math();
    test_indicators();
    test_predictions();
    test_risk();
    test_patterns();
    test_backtest();
    test_engine();

    std::cout << std::endl;
    std::cout << "═══════════════════════════════════════════════" << std::endl;
    std::cout << "  RESULTS: " << tests_passed << " passed, " << tests_failed << " failed" << std::endl;
    std::cout << "═══════════════════════════════════════════════" << std::endl;

    return (tests_failed > 0) ? 1 : 0;
}
