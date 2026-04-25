#ifndef SOVEREIGN_NEURAL_CORE_CONFIG_HPP
#define SOVEREIGN_NEURAL_CORE_CONFIG_HPP

/**
 * ============================================================================
 *  SOVEREIGN NEURAL CORE — Neural Architecture Configuration
 *  
 *  Model hyperparameters for the Sovereign Transformer architecture.
 *  Based on Multi-head Latent Attention (MLA) + Mixture-of-Experts (MoE).
 *
 *  Configuration tiers:
 *    - SOVEREIGN_MAX: Full-scale distributed (multi-node GPU)
 *    - SOVEREIGN_MID:  Workstation-scale (24GB+ VRAM)
 *    - SOVEREIGN_LOCAL: Single-machine CPU inference
 * ============================================================================
 */

#include <cstdint>
#include <cstddef>
#include <string>
#include <cmath>

namespace sovereign {
namespace neural {

enum class DType : uint8_t {
    BF16 = 0,
    FP8  = 1,
    FP32 = 2,
    INT8 = 3
};

enum class AttentionImpl : uint8_t {
    NAIVE  = 0,
    ABSORB = 1
};

enum class ScoreFunc : uint8_t {
    SOFTMAX = 0,
    SIGMOID = 1
};

enum class ModelType : uint8_t {
    SOVEREIGN_MLA   = 0, // Multi-Head Latent Attention + MoE
    SOVEREIGN_PRIME = 1, // GQA + RoPE (Llama-3 Base)
    SOVEREIGN_MIST  = 2, // Sliding Window Attention (Mistral Base)
    SOVEREIGN_PHI   = 3  // Compact Efficiency (Phi-3 Base)
};

inline const char* dtype_to_string(DType d) {
    switch (d) {
        case DType::BF16: return "bf16";
        case DType::FP8:  return "fp8";
        case DType::FP32: return "fp32";
        case DType::INT8: return "int8";
    }
    return "unknown";
}

// ============================================================================
// MODEL ARGUMENTS
// ============================================================================

struct ModelArgs {
    int max_batch_size      = 8;
    int max_seq_len         = 16384;
    DType dtype             = DType::BF16;
    AttentionImpl attn_impl = AttentionImpl::ABSORB;
    ModelType arch          = ModelType::SOVEREIGN_MLA;

    int vocab_size          = 129280;
    int dim                 = 7168;
    int inter_dim           = 18432;
    int moe_inter_dim       = 2048;

    int n_layers            = 61;
    int n_dense_layers      = 3;
    int n_heads             = 128;

    // MoE
    int n_routed_experts    = 256;
    int n_shared_experts    = 1;
    int n_activated_experts = 8;
    int n_expert_groups     = 8;
    int n_limited_groups    = 4;
    ScoreFunc score_func    = ScoreFunc::SIGMOID;
    double route_scale      = 2.5;

    // MLA
    int q_lora_rank         = 1536;
    int kv_lora_rank        = 512;
    int qk_nope_head_dim    = 128;
    int qk_rope_head_dim    = 64;
    int v_head_dim          = 128;
    int n_kv_heads          = 8;      // For GQA (Prime/Mist)
    int sliding_window      = 4096;   // For SWA (Mist)

    // YaRN RoPE
    int original_seq_len    = 4096;
    double rope_theta       = 10000.0;
    double rope_factor      = 40.0;
    int beta_fast           = 32;
    int beta_slow           = 1;
    double mscale           = 1.0;

    int quant_block_size    = 128;

    inline int qk_head_dim() const { return qk_nope_head_dim + qk_rope_head_dim; }
    inline int total_head_dim() const { return n_heads * qk_head_dim(); }
    inline int total_v_dim() const { return n_heads * v_head_dim; }

    inline int64_t total_params() const {
        int64_t embed = static_cast<int64_t>(vocab_size) * dim;
        int64_t attn;
        if (arch == ModelType::SOVEREIGN_MLA) {
            attn = static_cast<int64_t>(dim) * (q_lora_rank + kv_lora_rank + qk_rope_head_dim);
        } else {
            attn = static_cast<int64_t>(dim) * dim * 4; // Q, K, V, O
        }
        int64_t ffn;
        if (n_dense_layers >= n_layers) {
            ffn = static_cast<int64_t>(dim) * inter_dim * 3;
        } else {
            ffn = static_cast<int64_t>(dim) * moe_inter_dim * 3 * n_routed_experts;
        }
        return embed * 2 + n_layers * (attn + ffn);
    }

    inline int64_t activated_params() const {
        int64_t embed = static_cast<int64_t>(vocab_size) * dim;
        int64_t attn;
        if (arch == ModelType::SOVEREIGN_MLA) {
            attn = static_cast<int64_t>(dim) * (q_lora_rank + kv_lora_rank + qk_rope_head_dim);
        } else {
            attn = static_cast<int64_t>(dim) * dim * 4;
        }
        int64_t active_ffn;
        if (n_dense_layers >= n_layers) {
            active_ffn = static_cast<int64_t>(dim) * inter_dim * 3;
        } else {
            active_ffn = static_cast<int64_t>(dim) * moe_inter_dim * 3 * n_activated_experts;
        }
        return embed * 2 + n_layers * (attn + active_ffn);
    }

    inline std::string arch_to_string() const {
        switch (arch) {
            case ModelType::SOVEREIGN_MLA: return "Sovereign-MLA";
            case ModelType::SOVEREIGN_PRIME: return "Sovereign-Prime";
            case ModelType::SOVEREIGN_MIST: return "Sovereign-Mist";
            case ModelType::SOVEREIGN_PHI: return "Sovereign-Phi";
            default: return "Sovereign-Core";
        }
    }

    inline std::string summary() const {
        double tb = static_cast<double>(total_params()) / 1e9;
        double ab = static_cast<double>(activated_params()) / 1e9;
        std::string s = arch_to_string() + "[" + std::to_string(n_layers) + "L, " +
               std::to_string(n_heads) + "H, d=" + std::to_string(dim);
        if (n_dense_layers < n_layers) {
            s += ", MoE=" + std::to_string(n_routed_experts) + "x" + std::to_string(n_activated_experts);
        }
        s += ", ~" + std::to_string(static_cast<int>(tb)) + "B/" +
               std::to_string(static_cast<int>(ab)) + "B active" +
               ", " + dtype_to_string(dtype) + "]";
        return s;
    }
};

// ============================================================================
// PRESET CONFIGURATIONS
// ============================================================================

inline ModelArgs sovereign_max_config() {
    ModelArgs a;
    a.vocab_size=129280; a.dim=7168; a.inter_dim=18432; a.moe_inter_dim=2048;
    a.n_layers=61; a.n_dense_layers=3; a.n_heads=128;
    a.n_routed_experts=256; a.n_shared_experts=1; a.n_activated_experts=8;
    a.n_expert_groups=8; a.n_limited_groups=4;
    a.score_func=ScoreFunc::SIGMOID; a.route_scale=2.5;
    a.q_lora_rank=1536; a.kv_lora_rank=512;
    a.qk_nope_head_dim=128; a.qk_rope_head_dim=64; a.v_head_dim=128;
    a.dtype=DType::FP8; a.max_seq_len=16384;
    return a;
}

inline ModelArgs sovereign_local_config() {
    ModelArgs a;
    a.vocab_size=32000; a.dim=512; a.inter_dim=1408; a.moe_inter_dim=256;
    a.n_layers=12; a.n_dense_layers=1; a.n_heads=8;
    a.n_routed_experts=16; a.n_shared_experts=1; a.n_activated_experts=2;
    a.n_expert_groups=2; a.n_limited_groups=1;
    a.score_func=ScoreFunc::SIGMOID; a.route_scale=2.5;
    a.q_lora_rank=128; a.kv_lora_rank=64;
    a.qk_nope_head_dim=64; a.qk_rope_head_dim=32; a.v_head_dim=64;
    a.dtype=DType::FP32; a.max_seq_len=4096; a.max_batch_size=4;
    return a;
}

inline ModelArgs sovereign_mid_config() {
    ModelArgs a;
    a.vocab_size=64000; a.dim=2048; a.inter_dim=5632; a.moe_inter_dim=768;
    a.n_layers=24; a.n_dense_layers=2; a.n_heads=16;
    a.n_routed_experts=64; a.n_shared_experts=1; a.n_activated_experts=4;
    a.n_expert_groups=4; a.n_limited_groups=2;
    a.score_func=ScoreFunc::SIGMOID; a.route_scale=2.5;
    a.q_lora_rank=512; a.kv_lora_rank=256;
    a.qk_nope_head_dim=128; a.qk_rope_head_dim=64; a.v_head_dim=128;
    a.dtype=DType::BF16; a.max_seq_len=8192; a.max_batch_size=8;
    return a;
}

inline ModelArgs sovereign_prime_config() {
    ModelArgs a;
    a.arch = ModelType::SOVEREIGN_PRIME;
    a.vocab_size=128256; a.dim=4096; a.n_layers=32; a.n_heads=32; a.n_kv_heads=8;
    a.n_dense_layers=32; // Fully dense
    a.rope_theta=500000.0; a.max_seq_len=8192;
    return a;
}

inline ModelArgs sovereign_mist_config() {
    ModelArgs a;
    a.arch = ModelType::SOVEREIGN_MIST;
    a.vocab_size=32000; a.dim=4096; a.n_layers=32; a.n_heads=32; a.n_kv_heads=8;
    a.n_dense_layers=32;
    a.sliding_window=4096; a.max_seq_len=32768;
    return a;
}

inline ModelArgs sovereign_phi_config() {
    ModelArgs a;
    a.arch = ModelType::SOVEREIGN_PHI;
    a.vocab_size=32064; a.dim=2560; a.n_layers=32; a.n_heads=32; a.n_kv_heads=32;
    a.n_dense_layers=32; a.max_seq_len=4096;
    return a;
}

} // namespace neural
} // namespace sovereign

#endif // SOVEREIGN_NEURAL_CORE_CONFIG_HPP
