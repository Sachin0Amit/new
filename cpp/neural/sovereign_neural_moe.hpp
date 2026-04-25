#ifndef SOVEREIGN_NEURAL_CORE_MOE_HPP
#define SOVEREIGN_NEURAL_CORE_MOE_HPP

/**
 * ============================================================================
 *  SOVEREIGN NEURAL CORE — Mixture-of-Experts (MoE) Engine
 *
 *  Core innovation: Auxiliary-loss-free load balancing with sigmoid gating.
 *  Instead of softmax + auxiliary loss (which degrades performance), we use
 *  sigmoid scoring with learnable bias terms for natural load distribution.
 *
 *  Architecture:
 *    Gate:     x → score → group_select → top_k → weighted dispatch
 *    Expert:   x → w1 → SiLU → * w3 → w2 → y  (SwiGLU FFN)
 *    Shared:   Always-active expert(s) applied to every token
 *    Output:   sum(expert_i(x) * weight_i) + shared(x)
 * ============================================================================
 */

#include "sovereign_neural_layers.hpp"
#include <vector>
#include <algorithm>
#include <cmath>
#include <numeric>
#include <cassert>

namespace sovereign {
namespace neural {

// (Activation and Expert classes moved to sovereign_neural_layers.hpp)

// ============================================================================
// GATE — Expert Routing with Sigmoid Scoring
// ============================================================================

class Gate {
public:
    Gate() = default;
    Gate(const ModelArgs& args)
        : dim_(args.dim)
        , n_experts_(args.n_routed_experts)
        , topk_(args.n_activated_experts)
        , n_groups_(args.n_expert_groups)
        , topk_groups_(args.n_limited_groups)
        , score_func_(args.score_func)
        , route_scale_(args.route_scale)
        , gate_proj_(args.dim, args.n_routed_experts)
    {
        // Learnable bias for sigmoid gating (auxiliary-loss-free)
        if (args.dim == 7168) {
            bias_.resize(n_experts_);
            for (int i = 0; i < n_experts_; ++i) bias_[i] = 0.0f;
        }

        // Expert load tracking
        load_counts_.resize(n_experts_, 0);
    }

    struct GateResult {
        std::vector<int> selected_experts;
        std::vector<float> weights;
    };

    /// Route input to top-k experts with group selection.
    GateResult forward(const Tensor& x) {
        GateResult result;
        auto raw_scores = gate_proj_.forward(x);

        // Apply scoring function
        std::vector<float> scores(n_experts_);
        if (score_func_ == ScoreFunc::SIGMOID) {
            for (int i = 0; i < n_experts_; ++i) {
                scores[i] = sigmoid(raw_scores[i]);
            }
        } else {
            // Softmax
            float max_s = *std::max_element(raw_scores.data(),
                                             raw_scores.data() + n_experts_);
            float sum = 0.0f;
            for (int i = 0; i < n_experts_; ++i) {
                scores[i] = std::exp(raw_scores[i] - max_s);
                sum += scores[i];
            }
            for (int i = 0; i < n_experts_; ++i) scores[i] /= sum;
        }

        // Save original scores for weight computation
        std::vector<float> original_scores = scores;

        // Apply bias (auxiliary-loss-free load balancing)
        if (!bias_.empty()) {
            for (int i = 0; i < n_experts_; ++i) {
                scores[i] += bias_[i];
            }
        }

        // Group selection: partition experts into groups, select top groups
        if (n_groups_ > 1) {
            int experts_per_group = n_experts_ / n_groups_;

            // Compute group scores (max within each group)
            std::vector<float> group_scores(n_groups_);
            for (int g = 0; g < n_groups_; ++g) {
                float max_in_group = -1e30f;
                for (int e = 0; e < experts_per_group; ++e) {
                    int idx = g * experts_per_group + e;
                    max_in_group = std::max(max_in_group, scores[idx]);
                }
                group_scores[g] = max_in_group;
            }

            // Select top groups
            std::vector<int> group_indices(n_groups_);
            std::iota(group_indices.begin(), group_indices.end(), 0);
            std::partial_sort(group_indices.begin(),
                            group_indices.begin() + topk_groups_,
                            group_indices.end(),
                            [&](int a, int b) { return group_scores[a] > group_scores[b]; });

            // Mask out non-selected groups
            std::vector<bool> group_selected(n_groups_, false);
            for (int i = 0; i < topk_groups_; ++i) {
                group_selected[group_indices[i]] = true;
            }
            for (int g = 0; g < n_groups_; ++g) {
                if (!group_selected[g]) {
                    for (int e = 0; e < experts_per_group; ++e) {
                        scores[g * experts_per_group + e] = -1e30f;
                    }
                }
            }
        }

        // Select top-k experts globally
        std::vector<int> indices(n_experts_);
        std::iota(indices.begin(), indices.end(), 0);
        std::partial_sort(indices.begin(), indices.begin() + topk_, indices.end(),
                         [&](int a, int b) { return scores[a] > scores[b]; });

        result.selected_experts.resize(topk_);
        result.weights.resize(topk_);

        for (int k = 0; k < topk_; ++k) {
            result.selected_experts[k] = indices[k];
            result.weights[k] = original_scores[indices[k]];
        }

        // Normalize weights (for sigmoid)
        if (score_func_ == ScoreFunc::SIGMOID) {
            float w_sum = 0.0f;
            for (int k = 0; k < topk_; ++k) w_sum += result.weights[k];
            if (w_sum > 0.0f) {
                for (int k = 0; k < topk_; ++k) result.weights[k] /= w_sum;
            }
        }

        // Apply route scaling
        for (int k = 0; k < topk_; ++k) {
            result.weights[k] *= static_cast<float>(route_scale_);
        }

        // Update load tracking
        for (int k = 0; k < topk_; ++k) {
            load_counts_[result.selected_experts[k]]++;
        }

        return result;
    }

    /// Get load distribution across experts (for monitoring).
    std::vector<int64_t> load_distribution() const { return load_counts_; }

    /// Reset load counters.
    void reset_counters() { std::fill(load_counts_.begin(), load_counts_.end(), 0); }

private:
    int dim_ = 0, n_experts_ = 0, topk_ = 0;
    int n_groups_ = 0, topk_groups_ = 0;
    ScoreFunc score_func_ = ScoreFunc::SIGMOID;
    double route_scale_ = 1.0;
    LinearProjection gate_proj_;
    Tensor bias_;
    std::vector<int64_t> load_counts_;
};

// ============================================================================
// MOE — Full Mixture-of-Experts Layer
// ============================================================================

class MoE {
public:
    MoE() = default;
    MoE(const ModelArgs& args)
        : dim_(args.dim)
        , n_routed_experts_(args.n_routed_experts)
        , n_activated_(args.n_activated_experts)
        , gate_(args)
        , shared_expert_(args.dim, args.n_shared_experts * args.moe_inter_dim)
    {
        experts_.reserve(n_routed_experts_);
        for (int i = 0; i < n_routed_experts_; ++i) {
            experts_.emplace_back(args.dim, args.moe_inter_dim);
        }
    }

    /// Forward: route to top-k experts + shared expert.
    Tensor forward(const Tensor& x) {
        // 1. Gate: select experts and compute weights
        auto gate_result = gate_.forward(x);

        // 2. Compute weighted sum of routed expert outputs
        Tensor y(dim_);
        y.zero();

        for (int k = 0; k < n_activated_; ++k) {
            int expert_idx = gate_result.selected_experts[k];
            float weight = gate_result.weights[k];

            auto expert_out = experts_[expert_idx].forward(x);
            for (int d = 0; d < dim_; ++d) {
                y[d] += expert_out[d] * weight;
            }
        }

        // 3. Add shared expert output (always active)
        auto shared_out = shared_expert_.forward(x);
        y += shared_out;

        return y;
    }

    int n_experts() const { return n_routed_experts_; }
    int n_activated() const { return n_activated_; }
    Gate& gate() { return gate_; }
    const Gate& gate() const { return gate_; }

    /// Total expert parameters.
    int64_t total_expert_params() const {
        return static_cast<int64_t>(n_routed_experts_) * dim_ * 256 * 3; // approx
    }

private:
    int dim_ = 0;
    int n_routed_experts_ = 0;
    int n_activated_ = 0;
    Gate gate_;
    std::vector<Expert> experts_;
    Expert shared_expert_;
};

} // namespace neural
} // namespace sovereign

#endif // SOVEREIGN_NEURAL_CORE_MOE_HPP
