#ifndef SOVEREIGN_NEURAL_CORE_PHI_HPP
#define SOVEREIGN_NEURAL_CORE_PHI_HPP

/**
 * ============================================================================
 *  SOVEREIGN NEURAL CORE — Sovereign-Phi (Phi-3 Architecture)
 *
 *  Core features:
 *    - Standard MHA/GQA with optimized head dimensions.
 *    - GeGLU or SwiGLU activation (implementation uses SwiGLU for consistency).
 *    - Compact footprint for single-machine deployment.
 * ============================================================================
 */

#include "sovereign_neural_config.hpp"
#include "sovereign_neural_layers.hpp"
#include "sovereign_neural_prime.hpp" // Can reuse GQA logic

namespace sovereign {
namespace neural {

// Phi-3 architecture is structurally similar to Llama-3 (GQA) but with 
// different training-time optimizations and smaller parameter counts.
// We implement it using a specialized GQA class to allow for Phi-specific 
// head dimension adjustments if needed in the future.

class PhiAttention : public GQA {
public:
    using GQA::GQA;
};

} // namespace neural
} // namespace sovereign

#endif // SOVEREIGN_NEURAL_CORE_PHI_HPP
