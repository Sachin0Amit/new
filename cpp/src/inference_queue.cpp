#include "titan.h"
#include "inference_queue.hpp"
#include "titan.cpp"
#include <iostream>
#include <cstring>
#include <memory>

struct titan_handle_t {
    std::unique_ptr<titan::InferenceQueue> queue;
    std::unique_ptr<titan::TitanEngine> engine;
};

extern "C" {

titan_handle_t* titan_init(const char* config_json) {
    try {
        auto handle = new titan_handle_t();
        handle->queue = std::make_unique<titan::InferenceQueue>();
        handle->engine = std::make_unique<titan::TitanEngine>("model.gguf");
        // Mock init success
        std::cout << "Titan Engine Initialized with config: " << (config_json ? config_json : "{}") << std::endl;
        return handle;
    } catch (...) {
        return nullptr;
    }
}

char* titan_derive(titan_handle_t* handle, const char* prompt, int max_tokens) {
    if (!handle || !prompt) return nullptr;

    // Enqueue request
    if (!handle->queue->enqueue(prompt, max_tokens)) {
        return nullptr;
    }

    // Mock immediate synchronous drain for the C API demonstration
    auto req = handle->queue->dequeue();
    if (req) {
        std::string res = "Sovereign derivation for: " + req->prompt;
        char* c_res = (char*)malloc(res.size() + 1);
        strcpy(c_res, res.c_str());
        return c_res;
    }

    return nullptr;
}

void titan_free_result(char* result) {
    if (result) free(result);
}

void titan_destroy(titan_handle_t* handle) {
    if (handle) {
        delete handle;
        std::cout << "Titan Engine Destroyed." << std::endl;
    }
}

}
