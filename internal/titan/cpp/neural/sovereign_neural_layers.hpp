#ifndef SOVEREIGN_NEURAL_CORE_LAYERS_HPP
#define SOVEREIGN_NEURAL_CORE_LAYERS_HPP

#include <vector>
#include <cmath>
#include <algorithm>
#include <random>
#include <complex>

namespace sovereign {
namespace neural {

// ============================================================================
// ACTIVATION FUNCTIONS
// ============================================================================

inline float silu(float x) { return x / (1.0f + std::exp(-x)); }
inline float sigmoid(float x) { return 1.0f / (1.0f + std::exp(-x)); }

// ============================================================================
// TENSOR — Lightweight multi-dimensional array
// ============================================================================

class Tensor {
public:
    Tensor() = default;
    explicit Tensor(size_t size) : data_(size, 0.0f) {}
    Tensor(size_t size, float val) : data_(size, val) {}
    Tensor(std::vector<float>&& d) : data_(std::move(d)) {}

    float* data() { return data_.data(); }
    const float* data() const { return data_.data(); }
    size_t size() const { return data_.size(); }
    float& operator[](size_t i) { return data_[i]; }
    float operator[](size_t i) const { return data_[i]; }

    void resize(size_t n) { data_.resize(n, 0.0f); }
    void fill(float v) { std::fill(data_.begin(), data_.end(), v); }
    void zero() { fill(0.0f); }
    bool empty() const { return data_.empty(); }

    Tensor operator+(const Tensor& o) const {
        Tensor r(data_.size());
        for (size_t i = 0; i < data_.size(); ++i) r[i] = data_[i] + o[i];
        return r;
    }
    Tensor operator*(float s) const {
        Tensor r(data_.size());
        for (size_t i = 0; i < data_.size(); ++i) r[i] = data_[i] * s;
        return r;
    }
    Tensor& operator+=(const Tensor& o) {
        for (size_t i = 0; i < data_.size(); ++i) data_[i] += o[i];
        return *this;
    }

    float dot(const Tensor& o) const {
        float s = 0.0f;
        size_t n = std::min(data_.size(), o.size());
        for (size_t i = 0; i < n; ++i) s += data_[i] * o[i];
        return s;
    }

private:
    std::vector<float> data_;
};

// ============================================================================
// RMS NORMALIZATION
// ============================================================================

class RMSNorm {
public:
    RMSNorm() = default;
    explicit RMSNorm(int dim, float eps = 1e-6f) : dim_(dim), eps_(eps), weight_(dim, 1.0f) {}

    Tensor forward(const Tensor& x) const {
        float ss = 0.0f;
        for (int i = 0; i < dim_; ++i) ss += x[i] * x[i];
        float rms = 1.0f / std::sqrt(ss / static_cast<float>(dim_) + eps_);

        Tensor out(dim_);
        for (int i = 0; i < dim_; ++i) {
            out[i] = x[i] * rms * weight_[i];
        }
        return out;
    }

    int dim() const { return dim_; }
    Tensor& weight() { return weight_; }

private:
    int dim_ = 0;
    float eps_ = 1e-6f;
    Tensor weight_;
};

// ============================================================================
// LINEAR PROJECTION
// ============================================================================

class LinearProjection {
public:
    LinearProjection() = default;
    LinearProjection(int in_feat, int out_feat)
        : in_(in_feat), out_(out_feat), weight_(in_feat * out_feat) {
        float scale = std::sqrt(2.0f / static_cast<float>(in_feat + out_feat));
        std::mt19937 rng(42);
        std::normal_distribution<float> dist(0.0f, scale);
        for (size_t i = 0; i < weight_.size(); ++i) {
            weight_[i] = dist(rng);
        }
    }

    Tensor forward(const Tensor& x) const {
        Tensor out(out_);
        for (int o = 0; o < out_; ++o) {
            float sum = 0.0f;
            for (int i = 0; i < in_; ++i) {
                sum += x[i] * weight_[o * in_ + i];
            }
            out[o] = sum;
        }
        return out;
    }

    int in_features() const { return in_; }
    int out_features() const { return out_; }
    Tensor& weight() { return weight_; }

private:
    int in_ = 0, out_ = 0;
    Tensor weight_;
};

// ============================================================================
// EXPERT — SwiGLU Feed-Forward Network
// ============================================================================

class Expert {
public:
    Expert() = default;
    Expert(int dim, int inter_dim)
        : dim_(dim), inter_dim_(inter_dim)
        , w1_(dim, inter_dim)
        , w2_(inter_dim, dim)
        , w3_(dim, inter_dim)
    {}

    Tensor forward(const Tensor& x) const {
        auto h1 = w1_.forward(x);
        auto h3 = w3_.forward(x);
        Tensor gated(inter_dim_);
        for (int i = 0; i < inter_dim_; ++i) {
            gated[i] = silu(h1[i]) * h3[i];
        }
        return w2_.forward(gated);
    }

    int dim() const { return dim_; }
    int inter_dim() const { return inter_dim_; }

private:
    int dim_ = 0, inter_dim_ = 0;
    LinearProjection w1_, w2_, w3_;
};

// ============================================================================
// MLP — Dense Feed-Forward
// ============================================================================

class MLP {
public:
    MLP() = default;
    MLP(int dim, int inter_dim) : expert_(dim, inter_dim) {}
    Tensor forward(const Tensor& x) const { return expert_.forward(x); }
private:
    Expert expert_;
};

} // namespace neural
} // namespace sovereign

#endif // SOVEREIGN_NEURAL_CORE_LAYERS_HPP
