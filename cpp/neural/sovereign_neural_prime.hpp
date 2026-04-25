#ifndef SOVEREIGN_NEURAL_CORE_PRIME_HPP
#define SOVEREIGN_NEURAL_CORE_PRIME_HPP

/**
 * ============================================================================
 *  SOVEREIGN NEURAL CORE — Sovereign-Prime (Llama-3 Architecture)
 *
 *  Core features:
 *    - Grouped-Query Attention (GQA): Reduces KV cache size.
 *    - RoPE Positional Embeddings: High-theta (500k+) for long context.
 *    - SwiGLU Activation: Enhanced MLP capacity.
 *    - RMSNorm: Pre-normalization for training stability.
 * ============================================================================
 */

#include "sovereign_neural_config.hpp"
#include "sovereign_neural_layers.hpp"
#include <vector>
#include <cmath>
#include <algorithm>

namespace sovereign {
namespace neural {

// ============================================================================
// GROUPED-QUERY ATTENTION (GQA)
// ============================================================================

class GQA {
public:
    GQA() = default;
    explicit GQA(const ModelArgs& args)
        : dim_(args.dim)
        , n_heads_(args.n_heads)
        , n_kv_heads_(args.n_kv_heads)
        , head_dim_(args.dim / args.n_heads)
    {
        wq_ = LinearProjection(dim_, n_heads_ * head_dim_);
        wk_ = LinearProjection(dim_, n_kv_heads_ * head_dim_);
        wv_ = LinearProjection(dim_, n_kv_heads_ * head_dim_);
        wo_ = LinearProjection(n_heads_ * head_dim_, dim_);

        // KV Cache allocation
        int max_cache = args.max_batch_size * args.max_seq_len;
        k_cache_.resize(max_cache * n_kv_heads_ * head_dim_);
        v_cache_.resize(max_cache * n_kv_heads_ * head_dim_);

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

        // Cache KV
        int kv_offset = pos * n_kv_heads_ * head_dim_;
        for (int i = 0; i < n_kv_heads_ * head_dim_; ++i) {
            k_cache_[kv_offset + i] = k[i];
            v_cache_[kv_offset + i] = v[i];
        }

        // Attention calculation
        Tensor output(n_heads_ * head_dim_);
        int kv_group_size = n_heads_ / n_kv_heads_;

        for (int h = 0; h < n_heads_; ++h) {
            int kv_h = h / kv_group_size;
            std::vector<float> scores(pos + 1);
            float max_score = -1e30f;

            for (int t = 0; t <= pos; ++t) {
                float score = 0.0f;
                int t_kv_offset = t * n_kv_heads_ * head_dim_ + kv_h * head_dim_;
                for (int i = 0; i < head_dim_; ++i) {
                    score += q[h * head_dim_ + i] * k_cache_[t_kv_offset + i];
                }
                score *= softmax_scale_;
                scores[t] = score;
                max_score = std::max(max_score, score);
            }

            float sum_exp = 0.0f;
            for (int t = 0; t <= pos; ++t) {
                scores[t] = std::exp(scores[t] - max_score);
                sum_exp += scores[t];
            }
            for (int t = 0; t <= pos; ++t) scores[t] /= (sum_exp + 1e-10f);

            for (int t = 0; t <= pos; ++t) {
                int t_v_offset = t * n_kv_heads_ * head_dim_ + kv_h * head_dim_;
                for (int i = 0; i < head_dim_; ++i) {
                    output[h * head_dim_ + i] += scores[t] * v_cache_[t_v_offset + i];
                }
            }
        }

        return wo_.forward(output);
    }

    size_t cache_memory_bytes() const {
        return (k_cache_.size() + v_cache_.size()) * sizeof(float);
    }

private:
    int dim_, n_heads_, n_kv_heads_, head_dim_;
    LinearProjection wq_, wk_, wv_, wo_;
    Tensor k_cache_, v_cache_;
    float softmax_scale_;
};

} // namespace neural
} // namespace sovereign

#endif // SOVEREIGN_NEURAL_CORE_PRIME_HPP
