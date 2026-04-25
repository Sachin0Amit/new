#include <benchmark/benchmark.h>
#include "../src/inference_queue.hpp"
#include <string>

using namespace titan;

static void BM_SingleDerivation(benchmark::State& state) {
    InferenceQueue q;
    for (auto _ : state) {
        q.enqueue("Test prompt for cold cache", 50);
        auto req = q.dequeue();
        benchmark::DoNotOptimize(req);
    }
    state.SetItemsProcessed(state.iterations());
    state.counters["tokens/s"] = benchmark::Counter(state.iterations() * 50, benchmark::Counter::kIsRate);
}
BENCHMARK(BM_SingleDerivation);

static void BM_BatchDerivation(benchmark::State& state) {
    InferenceQueue q;
    for (auto _ : state) {
        for (int i = 0; i < 100; ++i) {
            q.enqueue("Batch prompt", 20);
        }
        for (int i = 0; i < 100; ++i) {
            auto req = q.dequeue();
            benchmark::DoNotOptimize(req);
        }
    }
    state.SetItemsProcessed(state.iterations() * 100);
    state.counters["tokens/s"] = benchmark::Counter(state.iterations() * 100 * 20, benchmark::Counter::kIsRate);
}
BENCHMARK(BM_BatchDerivation);

static void BM_QueueThroughput(benchmark::State& state) {
    InferenceQueue q;
    for (auto _ : state) {
        q.enqueue("Throughput test", 1);
        auto req = q.dequeue();
        benchmark::DoNotOptimize(req);
    }
    state.SetItemsProcessed(state.iterations());
}
BENCHMARK(BM_QueueThroughput)->OperationsPerSecond();

BENCHMARK_MAIN();
