#ifndef TITAN_ENGINE_H
#define TITAN_ENGINE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

/**
 * Titan C++ Inference Core
 * Dual-path architecture:
 *   - Symbolic: Low-latency mathematical derivation
 *   - Neural: Sovereign MLA/MoE Transformer for language tasks
 */

typedef struct {
    char* data;
    float tokens_per_sec;
    int64_t memory_used;
} TitanResult;

typedef struct {
    char* text;
    int* token_ids;
    int token_count;
    float tokens_per_sec;
    double latency_ms;
    int64_t memory_used;
    char* model_info;
} TitanNeuralResult;

typedef void* TitanContext;
typedef void* TitanNeuralContext;

// ============ Symbolic Engine Lifecycle ============
TitanContext titan_init(const char* device);
void titan_free(TitanContext ctx);

// ============ Symbolic Inference ============
TitanResult titan_infer(TitanContext ctx, const char* payload);
void titan_free_result(TitanResult res);

// ============ Neural Engine Lifecycle ============
// config_tier: "local", "mid", "max"
TitanNeuralContext titan_neural_init(const char* config_tier);
void titan_neural_free(TitanNeuralContext ctx);

// ============ Neural Inference ============
TitanNeuralResult titan_neural_generate(TitanNeuralContext ctx,
                                         const char* prompt,
                                         int max_tokens,
                                         float temperature);
void titan_neural_free_result(TitanNeuralResult res);

// ============ Tokenization ============
int titan_tokenize(TitanNeuralContext ctx, const char* text,
                   int* output_ids, int max_ids);
int titan_detokenize(TitanNeuralContext ctx, const int* ids,
                     int count, char* output, int max_len);

// ============ Model Info ============
const char* titan_neural_model_info(TitanNeuralContext ctx);
int64_t titan_neural_cache_memory(TitanNeuralContext ctx);

#ifdef __cplusplus
}
#endif

#endif // TITAN_ENGINE_H
