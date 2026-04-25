#include <gtest/gtest.h>
#include "../src/inference_queue.hpp"
#include <thread>
#include <vector>

using namespace titan;

TEST(InferenceQueueTest, BasicEnqueueDequeue) {
    InferenceQueue q;
    EXPECT_TRUE(q.enqueue("hello", 10));
    
    auto req = q.dequeue();
    ASSERT_TRUE(req.has_value());
    EXPECT_EQ(req->prompt, "hello");
    EXPECT_EQ(req->max_tokens, 10);
}

TEST(InferenceQueueTest, QueueFull) {
    InferenceQueue q;
    for (size_t i = 0; i < InferenceQueue::Capacity; ++i) {
        EXPECT_TRUE(q.enqueue("test", 1));
    }
    EXPECT_FALSE(q.enqueue("overflow", 1));
}

TEST(InferenceQueueTest, ConcurrentProducers) {
    InferenceQueue q;
    const int num_producers = 4;
    const int reqs_per_producer = 500;
    
    std::vector<std::thread> producers;
    for (int i = 0; i < num_producers; ++i) {
        producers.emplace_back([&q, i, reqs_per_producer]() {
            for (int j = 0; j < reqs_per_producer; ++j) {
                while (!q.enqueue("p" + std::to_string(i), j)) {
                    std::this_thread::yield();
                }
            }
        });
    }

    int total_received = 0;
    for (int i = 0; i < num_producers * reqs_per_producer; ++i) {
        while (true) {
            auto req = q.dequeue();
            if (req) {
                total_received++;
                break;
            }
            std::this_thread::yield();
        }
    }

    for (auto& t : producers) t.join();
    EXPECT_EQ(total_received, num_producers * reqs_per_producer);
}
