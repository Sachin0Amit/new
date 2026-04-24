#ifndef SOVEREIGN_FINANCE_FUNDAMENTALS_HPP
#define SOVEREIGN_FINANCE_FUNDAMENTALS_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Fundamental Analyzer
 *  Analyzes company history, balance sheets, and income statements to 
 *  simulate "elite chartered accountant" knowledge.
 * ============================================================================
 */

#include <string>

namespace sovereign {
namespace finance {
namespace fundamentals {

struct FundamentalData {
    double pe_ratio;          // Price-to-Earnings
    double pb_ratio;          // Price-to-Book
    double debt_to_equity;    // D/E Ratio
    double roe;               // Return on Equity (%)
    double eps_growth_5y;     // Earnings Per Share Growth 5-Year (%)
    double current_ratio;     // Current Assets / Current Liabilities
    double operating_margin;  // Operating Income / Revenue (%)
};

struct FundamentalScore {
    double total_score;       // 0.0 to 100.0
    std::string rating;       // "AAA", "AA", "A", "BBB", "BB", "B", "C", "D"
    std::string reasoning;    // Explanation of the rating
};

class FundamentalAnalyzer {
public:
    FundamentalAnalyzer() = default;

    /// Evaluates raw fundamental data and assigns an elite-level score.
    FundamentalScore evaluate(const FundamentalData& data) const {
        double score = 0.0;
        std::string reasoning = "";

        // 1. Valuation (P/E) [Max 20 pts]
        if (data.pe_ratio > 0 && data.pe_ratio < 15) { score += 20; reasoning += "Deep value (Low P/E). "; }
        else if (data.pe_ratio >= 15 && data.pe_ratio < 25) { score += 15; reasoning += "Fair valuation. "; }
        else if (data.pe_ratio >= 25 && data.pe_ratio < 50) { score += 5; reasoning += "Growth valuation (High P/E). "; }
        else { reasoning += "Overvalued/Unprofitable. "; }

        // 2. Debt Management (Debt to Equity) [Max 20 pts]
        if (data.debt_to_equity < 0.5) { score += 20; reasoning += "Fortress balance sheet. "; }
        else if (data.debt_to_equity < 1.0) { score += 15; reasoning += "Healthy debt levels. "; }
        else if (data.debt_to_equity < 2.0) { score += 5; reasoning += "Moderate debt risk. "; }
        else { reasoning += "High insolvency risk. "; }

        // 3. Profitability (ROE) [Max 20 pts]
        if (data.roe > 20) { score += 20; reasoning += "Elite capital efficiency (ROE > 20%). "; }
        else if (data.roe > 10) { score += 15; reasoning += "Solid profitability. "; }
        else if (data.roe > 0) { score += 5; reasoning += "Weak profitability. "; }
        else { reasoning += "Burning cash. "; }

        // 4. Growth (EPS Growth 5Y) [Max 20 pts]
        if (data.eps_growth_5y > 15) { score += 20; reasoning += "Hyper-growth trajectory. "; }
        else if (data.eps_growth_5y > 5) { score += 15; reasoning += "Steady historical growth. "; }
        else if (data.eps_growth_5y > 0) { score += 5; reasoning += "Stagnant growth. "; }
        else { reasoning += "Earnings contraction. "; }

        // 5. Liquidity (Current Ratio) [Max 20 pts]
        if (data.current_ratio > 2.0) { score += 20; reasoning += "High short-term liquidity. "; }
        else if (data.current_ratio > 1.0) { score += 15; reasoning += "Adequate liquidity. "; }
        else { score += 0; reasoning += "Liquidity crisis risk. "; }

        FundamentalScore result;
        result.total_score = score;
        result.reasoning = reasoning;

        if (score >= 90) result.rating = "AAA";
        else if (score >= 80) result.rating = "AA";
        else if (score >= 70) result.rating = "A";
        else if (score >= 60) result.rating = "BBB";
        else if (score >= 50) result.rating = "BB";
        else if (score >= 40) result.rating = "B";
        else if (score >= 20) result.rating = "C";
        else result.rating = "D";

        return result;
    }

    /// Fetches mock fundamental data for specific symbols since we lack a paid API.
    FundamentalData get_mock_data(const std::string& symbol) const {
        if (symbol == "AAPL") {
            return {28.5, 45.2, 1.2, 160.5, 12.4, 0.99, 30.2}; // Apple
        } else if (symbol == "MSFT") {
            return {35.2, 12.1, 0.3, 38.4, 18.2, 1.77, 44.6}; // Microsoft
        } else if (symbol == "TSLA") {
            return {45.6, 9.4, 0.1, 22.1, 35.0, 1.63, 11.2}; // Tesla
        } else if (symbol == "BTC" || symbol == "ETH") {
            // Cryptocurrencies do not have traditional fundamentals, simulate pristine math
            return {10.0, 1.0, 0.0, 50.0, 100.0, 5.0, 100.0}; 
        }
        
        // Default generic data
        return {15.0, 2.0, 0.8, 12.0, 5.0, 1.5, 15.0};
    }
};

} // namespace fundamentals
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_FUNDAMENTALS_HPP
