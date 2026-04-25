#ifndef SOVEREIGN_FINANCE_SENTIMENT_HPP
#define SOVEREIGN_FINANCE_SENTIMENT_HPP

/**
 * ============================================================================
 *  SOVEREIGN FINANCE ENGINE — Sentiment Analyzer
 *  High-speed lexicon-based NLP module for scoring financial news headlines.
 * ============================================================================
 */

#include <string>
#include <vector>
#include <unordered_map>
#include <algorithm>
#include <cctype>

namespace sovereign {
namespace finance {
namespace sentiment {

struct SentimentScore {
    double score;         // -1.0 to 1.0 (Negative to Positive)
    int positive_words;
    int negative_words;
    std::vector<std::string> triggers;
};

class LexiconAnalyzer {
public:
    LexiconAnalyzer() {
        // Elite Chartered Accountant / Quant Trader Lexicon

        // Bullish / Positive terms
        positive_ = {
            {"record", 0.8}, {"surge", 0.7}, {"jump", 0.6}, {"soar", 0.8},
            {"beat", 0.6}, {"outperform", 0.7}, {"upgrade", 0.7}, {"profit", 0.5},
            {"growth", 0.5}, {"dividend", 0.6}, {"buyback", 0.7}, {"acquire", 0.4},
            {"bullish", 0.8}, {"rally", 0.6}, {"breakout", 0.7}, {"strong", 0.4},
            {"raise", 0.5}, {"all-time high", 0.9}, {"exceed", 0.6}, {"optimism", 0.5}
        };

        // Bearish / Negative terms
        negative_ = {
            {"plunge", -0.8}, {"crash", -0.9}, {"tumble", -0.7}, {"fall", -0.5},
            {"miss", -0.6}, {"underperform", -0.7}, {"downgrade", -0.7}, {"loss", -0.6},
            {"debt", -0.4}, {"bankruptcy", -1.0}, {"lawsuit", -0.6}, {"scandal", -0.8},
            {"bearish", -0.8}, {"sell-off", -0.7}, {"collapse", -0.9}, {"weak", -0.4},
            {"cut", -0.5}, {"layoff", -0.5}, {"inflation", -0.4}, {"recession", -0.8}
        };
    }

    SentimentScore analyze(const std::string& text) const {
        SentimentScore result = {0.0, 0, 0, {}};
        if (text.empty()) return result;

        // Convert to lowercase for matching
        std::string lower_text = text;
        std::transform(lower_text.begin(), lower_text.end(), lower_text.begin(),
                       [](unsigned char c){ return std::tolower(c); });

        double total_score = 0.0;
        int word_count = 0;

        // Simple tokenization (by space)
        std::string current_word;
        std::vector<std::string> words;
        for (char c : lower_text) {
            if (std::isalnum(c) || c == '-') {
                current_word += c;
            } else if (!current_word.empty()) {
                words.push_back(current_word);
                current_word.clear();
            }
        }
        if (!current_word.empty()) words.push_back(current_word);

        // Score words
        for (const auto& w : words) {
            word_count++;
            auto p_it = positive_.find(w);
            if (p_it != positive_.end()) {
                total_score += p_it->second;
                result.positive_words++;
                result.triggers.push_back(w);
            } else {
                auto n_it = negative_.find(w);
                if (n_it != negative_.end()) {
                    total_score += n_it->second;
                    result.negative_words++;
                    result.triggers.push_back(w);
                }
            }
        }

        // Handle multi-word phrases manually for simplicity in this MVP
        if (lower_text.find("all-time high") != std::string::npos) {
            total_score += 0.9;
            result.positive_words++;
            result.triggers.push_back("all-time high");
        }

        if (word_count > 0) {
            // Normalize score
            result.score = total_score / std::max(1.0, (double)(result.positive_words + result.negative_words));
            
            // Clamp between -1.0 and 1.0
            if (result.score > 1.0) result.score = 1.0;
            if (result.score < -1.0) result.score = -1.0;
        }

        return result;
    }

private:
    std::unordered_map<std::string, double> positive_;
    std::unordered_map<std::string, double> negative_;
};

} // namespace sentiment
} // namespace finance
} // namespace sovereign

#endif // SOVEREIGN_FINANCE_SENTIMENT_HPP
