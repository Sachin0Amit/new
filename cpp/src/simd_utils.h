#ifndef SIMD_UTILS_H
#define SIMD_UTILS_H

#include <vector>
#include <cmath>

#if defined(__AVX2__)
#include <immintrin.h>
#endif

namespace titan {

/**
 * High-performance cosine similarity using SIMD (AVX2) with scalar fallback.
 */
inline float cosine_similarity(const float* a, const float* b, size_t size) {
#if defined(__AVX2__)
    __m256 dot_v = _mm256_setzero_ps();
    __m256 a_sq_v = _mm256_setzero_ps();
    __m256 b_sq_v = _mm256_setzero_ps();

    size_t i = 0;
    for (; i + 7 < size; i += 8) {
        __m256 va = _mm256_loadu_ps(a + i);
        __m256 vb = _mm256_loadu_ps(b + i);

        dot_v = _mm256_add_ps(dot_v, _mm256_mul_ps(va, vb));
        a_sq_v = _mm256_add_ps(a_sq_v, _mm256_mul_ps(va, va));
        b_sq_v = _mm256_add_ps(b_sq_v, _mm256_mul_ps(vb, vb));
    }

    float dot = 0, a_sq = 0, b_sq = 0;
    float temp_dot[8], temp_a[8], temp_b[8];
    _mm256_storeu_ps(temp_dot, dot_v);
    _mm256_storeu_ps(temp_a, a_sq_v);
    _mm256_storeu_ps(temp_b, b_sq_v);

    for (int j = 0; j < 8; ++j) {
        dot += temp_dot[j];
        a_sq += temp_a[j];
        b_sq += temp_b[j];
    }

    // Scalar fallback for remaining elements
    for (; i < size; ++i) {
        dot += a[i] * b[i];
        a_sq += a[i] * a[i];
        b_sq += b[i] * b[i];
    }
#else
    float dot = 0, a_sq = 0, b_sq = 0;
    for (size_t i = 0; i < size; ++i) {
        dot += a[i] * b[i];
        a_sq += a[i] * a[i];
        b_sq += b[i] * b[i];
    }
#endif

    if (a_sq == 0 || b_sq == 0) return 0;
    return dot / (std::sqrt(a_sq) * std::sqrt(b_sq));
}

} // namespace titan

#endif // SIMD_UTILS_H
