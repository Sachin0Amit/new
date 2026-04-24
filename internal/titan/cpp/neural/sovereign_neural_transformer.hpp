#ifndef SOVEREIGN_NEURAL_CORE_TRANSFORMER_HPP
#define SOVEREIGN_NEURAL_CORE_TRANSFORMER_HPP

/**
 * ============================================================================
 *  SOVEREIGN NEURAL CORE — Transformer Architecture
 *
 *  Full transformer stack integrating all sub-modules:
 *    - Token Embedding with vocabulary parallel support
 *    - N transformer blocks, each: RMSNorm → MLA → RMSNorm → (MLP|MoE)
 *    - First n_dense_layers use dense MLP, rest use MoE
 *    - Output: RMSNorm → Linear head → logits
 *    - Auto-regressive generation with KV cache
 *    - Temperature-controlled sampling
 * ============================================================================
 */

#include "sovereign_neural_moe.hpp"
#include "sovereign_neural_mla.hpp"
#include "sovereign_neural_prime.hpp"
#include "sovereign_neural_mist.hpp"
#include "sovereign_neural_phi.hpp"
#include <vector>
#include <memory>
#include <string>
#include <chrono>
#include <functional>
#include <random>
#include <iostream>

namespace sovereign {
namespace neural {

// ============================================================================
// EMBEDDING LAYER
// ============================================================================

class Embedding {
public:
    Embedding() = default;
    Embedding(int vocab_size, int dim) : vocab_size_(vocab_size), dim_(dim) {
        weight_.resize(static_cast<size_t>(vocab_size) * dim);
        float scale = 1.0f / std::sqrt(static_cast<float>(dim));
        std::mt19937 rng(42);
        std::normal_distribution<float> dist(0.0f, scale);
        for (size_t i = 0; i < weight_.size(); ++i) weight_[i] = dist(rng);
    }

    Tensor forward(int token_id) const {
        Tensor out(dim_);
        if (token_id < 0 || token_id >= vocab_size_) {
            out.zero();
            return out;
        }
        size_t offset = static_cast<size_t>(token_id) * dim_;
        for (int i = 0; i < dim_; ++i) out[i] = weight_[offset + i];
        return out;
    }

    int vocab_size() const { return vocab_size_; }
    int dim() const { return dim_; }

private:
    int vocab_size_ = 0, dim_ = 0;
    Tensor weight_;
};

// ============================================================================
// TRANSFORMER BLOCK
// ============================================================================

class Block {
public:
    Block() = default;
    Block(int layer_id, const ModelArgs& args)
        : layer_id_(layer_id)
        , arch_(args.arch)
        , is_dense_(layer_id < args.n_dense_layers)
        , attn_norm_(args.dim)
        , ffn_norm_(args.dim)
    {
        // Initialize attention based on architecture
        switch (arch_) {
            case ModelType::SOVEREIGN_MLA:
                mla_ = std::make_unique<MLA>(args);
                break;
            case ModelType::SOVEREIGN_PRIME:
                prime_ = std::make_unique<GQA>(args);
                break;
            case ModelType::SOVEREIGN_MIST:
                mist_ = std::make_unique<SWA>(args);
                break;
            case ModelType::SOVEREIGN_PHI:
                phi_ = std::make_unique<PhiAttention>(args);
                break;
        }

        if (is_dense_) {
            mlp_ = MLP(args.dim, args.inter_dim);
        } else {
            moe_ = MoE(args);
        }
    }

    Tensor forward(const Tensor& x, int pos, const FreqsCIS& freqs) {
        auto normed = attn_norm_.forward(x);
        Tensor attn_out;

        switch (arch_) {
            case ModelType::SOVEREIGN_MLA:   attn_out = mla_->forward(normed, pos, freqs); break;
            case ModelType::SOVEREIGN_PRIME: attn_out = prime_->forward(normed, pos, freqs); break;
            case ModelType::SOVEREIGN_MIST:  attn_out = mist_->forward(normed, pos, freqs); break;
            case ModelType::SOVEREIGN_PHI:   attn_out = phi_->forward(normed, pos, freqs); break;
        }

        Tensor h = x + attn_out;
        auto ffn_normed = ffn_norm_.forward(h);
        Tensor ffn_out;

        if (is_dense_) {
            ffn_out = mlp_.forward(ffn_normed);
        } else {
            ffn_out = moe_.forward(ffn_normed);
        }

        return h + ffn_out;
    }

    bool is_dense() const { return is_dense_; }
    int layer_id() const { return layer_id_; }
    
    size_t cache_memory_bytes() const {
        switch (arch_) {
            case ModelType::SOVEREIGN_MLA:   return mla_->cache_memory_bytes();
            case ModelType::SOVEREIGN_PRIME: return prime_->cache_memory_bytes();
            case ModelType::SOVEREIGN_MIST:  return mist_->cache_memory_bytes();
            case ModelType::SOVEREIGN_PHI:   return phi_->cache_memory_bytes();
        }
        return 0;
    }

    MoE& moe() { return moe_; }

private:
    int layer_id_ = 0;
    ModelType arch_;
    bool is_dense_ = false;
    
    std::unique_ptr<MLA> mla_;
    std::unique_ptr<GQA> prime_;
    std::unique_ptr<SWA> mist_;
    std::unique_ptr<PhiAttention> phi_;

    MLP mlp_;
    MoE moe_;
    RMSNorm attn_norm_;
    RMSNorm ffn_norm_;
};

// ============================================================================
// GENERATION RESULT
// ============================================================================

struct GenerationResult {
    std::vector<int> tokens;
    std::string text;
    int prompt_tokens = 0;
    int generated_tokens = 0;
    double tokens_per_second = 0.0;
    double latency_ms = 0.0;
    int64_t memory_used_bytes = 0;
    std::string model_name;
};

// ============================================================================
// TRANSFORMER — Full Model
// ============================================================================

class Transformer {
public:
    Transformer() = default;
    explicit Transformer(const ModelArgs& args)
        : args_(args)
        , embed_(args.vocab_size, args.dim)
        , norm_(args.dim)
        , head_(args.dim, args.vocab_size)
        , freqs_(precompute_freqs_cis(args))
    {
        layers_.reserve(args.n_layers);
        for (int i = 0; i < args.n_layers; ++i) {
            layers_.emplace_back(i, args);
        }
    }

    /// Forward pass: tokens → logits (for a single token at position pos).
    Tensor forward(int token, int pos) {
        // 1. Embed token
        auto h = embed_.forward(token);

        // 2. Pass through all transformer blocks
        for (auto& layer : layers_) {
            h = layer.forward(h, pos, freqs_);
        }

        // 3. Final normalization
        h = norm_.forward(h);

        // 4. Project to vocabulary (logits)
        return head_.forward(h);
    }

    /// Sample a token from logits with temperature.
    int sample(const Tensor& logits, float temperature = 1.0f) {
        int vocab = args_.vocab_size;

        if (temperature <= 0.0f) {
            // Greedy: argmax
            int best = 0;
            float best_val = logits[0];
            for (int i = 1; i < vocab; ++i) {
                if (logits[i] > best_val) { best_val = logits[i]; best = i; }
            }
            return best;
        }

        // Temperature-scaled softmax sampling
        std::vector<float> probs(vocab);
        float max_val = *std::max_element(logits.data(), logits.data() + vocab);
        float sum = 0.0f;
        for (int i = 0; i < vocab; ++i) {
            probs[i] = std::exp((logits[i] - max_val) / temperature);
            sum += probs[i];
        }
        for (int i = 0; i < vocab; ++i) probs[i] /= sum;

        // Gumbel-max trick (fast sampling)
        std::uniform_real_distribution<float> udist(0.0f, 1.0f);
        float best_g = -1e30f;
        int best_idx = 0;
        for (int i = 0; i < vocab; ++i) {
            if (probs[i] < 1e-10f) continue;
            float u = udist(rng_);
            float g = -std::log(-std::log(u + 1e-10f) + 1e-10f) + std::log(probs[i] + 1e-10f);
            if (g > best_g) { best_g = g; best_idx = i; }
        }

        return best_idx;
    }

    /// Auto-regressive generation from prompt tokens.
    GenerationResult generate(const std::vector<int>& prompt_tokens,
                              int max_new_tokens,
                              float temperature = 0.7f,
                              int eos_token = -1) {
        GenerationResult result;
        result.model_name = args_.summary();
        result.prompt_tokens = static_cast<int>(prompt_tokens.size());

        auto start = std::chrono::high_resolution_clock::now();

        // Process prompt (prefill)
        std::vector<int> all_tokens = prompt_tokens;
        int pos = 0;

        for (int i = 0; i < static_cast<int>(prompt_tokens.size()); ++i) {
            forward(prompt_tokens[i], pos++);
        }

        // Generate new tokens
        int generated = 0;
        for (int step = 0; step < max_new_tokens; ++step) {
            int last_token = all_tokens.back();
            auto logits = forward(last_token, pos++);
            int next_token = sample(logits, temperature);

            if (next_token == eos_token) break;

            all_tokens.push_back(next_token);
            generated++;
        }

        auto end = std::chrono::high_resolution_clock::now();
        double elapsed_ms = std::chrono::duration<double, std::milli>(end - start).count();

        result.tokens = all_tokens;
        result.generated_tokens = generated;
        result.latency_ms = elapsed_ms;
        result.tokens_per_second = (elapsed_ms > 0) ?
            static_cast<double>(generated) / (elapsed_ms / 1000.0) : 0.0;

        // Estimate memory usage
        size_t cache_mem = 0;
        for (auto& layer : layers_) {
            cache_mem += layer.cache_memory_bytes();
        }
        result.memory_used_bytes = static_cast<int64_t>(cache_mem);

        return result;
    }

    /// Get model info.
    const ModelArgs& args() const { return args_; }
    int n_layers() const { return static_cast<int>(layers_.size()); }

    /// Get MoE load distribution across all layers.
    struct LayerStats {
        int layer_id;
        bool is_dense;
        std::vector<int64_t> expert_loads;
    };

    std::vector<LayerStats> get_layer_stats() const {
        std::vector<LayerStats> stats;
        for (size_t i = 0; i < layers_.size(); ++i) {
            LayerStats ls;
            ls.layer_id = static_cast<int>(i);
            ls.is_dense = layers_[i].is_dense();
            if (!ls.is_dense) {
                ls.expert_loads = const_cast<Block&>(layers_[i]).moe().gate().load_distribution();
            }
            stats.push_back(ls);
        }
        return stats;
    }

private:
    ModelArgs args_;
    Embedding embed_;
    RMSNorm norm_;
    LinearProjection head_;
    FreqsCIS freqs_;
    std::vector<Block> layers_;
    std::mt19937 rng_{42};
};

} // namespace neural
} // namespace sovereign

#endif // SOVEREIGN_NEURAL_CORE_TRANSFORMER_HPP
