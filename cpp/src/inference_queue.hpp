#ifndef INFERENCE_QUEUE_HPP
#define INFERENCE_QUEUE_HPP

#include <atomic>
#include <string>
#include <vector>
#include <optional>

namespace titan {

/**
 * A thread-safe MPSC (Multi-Producer Single-Consumer) Lock-Free Ring Buffer.
 * Capacity is fixed at 4096.
 */
class InferenceQueue {
public:
    static constexpr size_t Capacity = 4096;
    static constexpr size_t Mask = Capacity - 1;

    struct Request {
        std::string prompt;
        int max_tokens;
        std::atomic<bool> ready{false};

        Request() = default;
        Request(const Request& other) : prompt(other.prompt), max_tokens(other.max_tokens), ready(other.ready.load()) {}
        Request& operator=(const Request& other) {
            prompt = other.prompt;
            max_tokens = other.max_tokens;
            ready.store(other.ready.load());
            return *this;
        }
    };

    InferenceQueue() : head(0), tail(0) {
        buffer.resize(Capacity);
    }

    bool enqueue(const std::string& prompt, int max_tokens) {
        size_t current_head = head.load(std::memory_order_relaxed);
        size_t current_tail = tail.load(std::memory_order_acquire);

        if (current_head - current_tail >= Capacity) {
            return false;
        }

        size_t pos = head.fetch_add(1, std::memory_order_relaxed);
        auto& slot = buffer[pos & Mask];
        slot.prompt = prompt;
        slot.max_tokens = max_tokens;
        slot.ready.store(true, std::memory_order_release);
        
        return true;
    }

    std::optional<Request> dequeue() {
        size_t current_tail = tail.load(std::memory_order_relaxed);
        auto& slot = buffer[current_tail & Mask];

        if (!slot.ready.load(std::memory_order_acquire)) {
            return std::nullopt;
        }

        Request req;
        req.prompt = std::move(slot.prompt);
        req.max_tokens = slot.max_tokens;
        
        slot.ready.store(false, std::memory_order_relaxed);
        tail.store(current_tail + 1, std::memory_order_release);
        
        return req;
    }

private:
    std::atomic<size_t> head;
    std::atomic<size_t> tail;
    std::vector<Request> buffer;
};

} // namespace titan

#endif // INFERENCE_QUEUE_HPP
