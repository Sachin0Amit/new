#ifndef TITAN_ENGINE_H
#define TITAN_ENGINE_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>

/**
 * Titan C++ Inference Core
 * Optimized for low-latency symbolic derivation.
 */

typedef struct {
    char* data;
    float tokens_per_sec;
    int64_t memory_used;
} TitanResult;

typedef void* TitanContext;

// Engine Lifecycle
TitanContext titan_init(const char* device);
void titan_free(TitanContext ctx);

// Inference Operations
TitanResult titan_infer(TitanContext ctx, const char* payload);
void titan_free_result(TitanResult res);

#ifdef __cplusplus
}
#endif

#endif // TITAN_ENGINE_H
