#ifndef SOVEREIGN_NEURAL_CORE_MIST_HPP
#define SOVEREIGN_NEURAL_CORE_MIST_HPP

/**
 * ============================================================================
 *  SOVEREIGN NEURAL CORE — Sovereign-Mist (Mistral Architecture)
 *
 *  Core features:
 *    - Sliding Window Attention (SWA): Constant memory for KV cache.
 *    - Grouped-Query Attention (GQA).
 *    - Rolling Buffer Cache logic.
 * ============================================================================
 */

#include "sovereign_neural_config.hpp"
#include "sovereign_neural_layers.hpp"
#include "sovereign_neural_prime.hpp" // Reuse GQA base logic if needed, but we'll specialize
#include <vector>
#include <cmath>
#include <algorithm>

namespace sovereign {
namespace neural {

// ============================================================================
// SLIDING WINDOW ATTENTION (SWA)
// ============================================================================

class SWA {
public:
    SWA() = default;
    explicit SWA(const ModelArgs& args)
        : dim_(args.dim)
        , n_heads_(args.n_heads)
        , n_kv_heads_(args.n_kv_heads)
        , head_dim_(args.dim / args.n_heads)
        , window_size_(args.sliding_window)
    {
        wq_ = LinearProjection(dim_, n_heads_ * head_dim_);
        wk_ = LinearProjection(dim_, n_kv_heads_ * head_dim_);
        wv_ = LinearProjection(dim_, n_kv_heads_ * head_dim_);
        wo_ = LinearProjection(n_heads_ * head_dim_, dim_);

        // Rolling KV Cache allocation (only window_size_ positions)
        k_cache_.resize(window_size_ * n_kv_heads_ * head_dim_);
        v_cache_.resize(window_size_ * n_kv_heads_ * head_dim_);

        softmax_scale_ = 1.0f / std::sqrt(static_cast<float>(head_dim_));
    }

    Tensor forward(const Tensor& x, int pos, const FreqsCIS& freqs) {
        auto q = wq_.forward(x);
        auto k = wk_.forward(x);
        auto v = wv_.forward(x);

        // Apply RoPE
        for (int h = 0; h < n_heads_; ++h) {
            Tensor q_h(head_dim_);
            for (int i = 0; i < head_dim_; ++i) q_h[i] = q[h * head_dim_ + i];
            apply_rotary_emb(q_h, freqs, pos, head_dim_);
            for (int i = 0; i < head_dim_; ++i) q[h * head_dim_ + i] = q_h[i];
        }

        for (int h = 0; h < n_kv_heads_; ++h) {
            Tensor k_h(head_dim_);
            for (int i = 0; i < head_dim_; ++i) k_h[i] = k[h * head_dim_ + i];
            apply_rotary_emb(k_h, freqs, pos, head_dim_);
            for (int i = 0; i < head_dim_; ++i) k[h * head_dim_ + i] = k_h[i];
        }

        // Cache KV using Rolling Buffer logic
        int cache_pos = pos % window_size_;
        int kv_offset = cache_pos * n_kv_heads_ * head_dim_;
        for (int i = 0; i < n_kv_heads_ * head_dim_; ++i) {
            k_cache_[kv_offset + i] = k[i];
            v_cache_[kv_offset + i] = v[i];
        }

        // Attention calculation with sliding window
        Tensor output(n_heads_ * head_dim_);
        int kv_group_size = n_heads_ / n_kv_heads_;
        int start_pos = std::max(0, pos - window_size_ + 1);

        for (int h = 0; h < n_heads_; ++h) {
            int kv_h = h / kv_group_size;
            std::vector<float> scores(pos - start_pos + 1);
            float max_score = -1e30f;

            for (int t = start_pos; t <= pos; ++t) {
                float score = 0.0f;
                int t_cache_pos = t % window_size_;
                int t_kv_offset = t_cache_pos * n_kv_heads_ * head_dim_ + kv_h * head_dim_;
                for (int i = 0; i < head_dim_; ++i) {
                    score += q[h * head_dim_ + i] * k_cache_[t_kv_offset + i];
                }
                score *= softmax_scale_;
                scores[t - start_pos] = score;
                max_score = std::max(max_score, score);
            }

            float sum_exp = 0.0f;
            for (size_t i = 0; i < scores.size(); ++i) {
                scores[i] = std::exp(scores[i] - max_score);
                sum_exp += scores[i];
            }
            for (size_t i = 0; i < scores.size(); ++i) scores[i] /= (sum_exp + 1e-10f);

            for (int t = start_pos; t <= pos; ++t) {
                int t_cache_pos = t % window_size_;
                int t_v_offset = t_cache_pos * n_kv_heads_ * head_dim_ + kv_h * head_dim_;
                for (int i = 0; i < head_dim_; ++i) {
                    output[h * head_dim_ + i] += scores[t - start_pos] * v_cache_[t_v_offset + i];
                }
            }
        }

        return wo_.forward(output);
    }

    size_t cache_memory_bytes() const {
        return (k_cache_.size() + v_cache_.size()) * sizeof(float);
    }

private:
    int dim_, n_heads_, n_kv_heads_, head_dim_, window_size_;
    LinearProjection wq_, wk_, wv_, wo_;
    Tensor k_cache_, v_cache_;
    float softmax_scale_;
};

} // namespace neural
} // namespace sovereign

#endif // SOVEREIGN_NEURAL_CORE_MIST_HPP
