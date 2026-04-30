#include "context_manager.cpp"
#include "symbolic_engine.cpp"
#include "neural_engine.cpp"
#include <string>
#include <functional>
#include <iostream>

namespace titan {

/**
 * Main TitanEngine orchestrator.
 */
class TitanEngine {
public:
    TitanEngine(const std::string& model_path) 
        : ctx_manager(4096), neural_engine(model_path) {}

    void set_audit_callback(std::function<void(const std::string&)> callback) {
        audit_callback = callback;
    }

    std::string process(const std::string& prompt) {
        if (audit_callback) audit_callback("Processing query: " + prompt);

        // 1. Try Symbolic Engine (Fast Path)
        auto symbolic_res = symbolic_engine.deduce(prompt);
        if (symbolic_res.has_value() && symbolic_engine.confidence_score(prompt) > 0.95f) {
            if (audit_callback) audit_callback("Path: Symbolic Fast-Path");
            return *symbolic_res;
        }

        // 2. Fall through to Neural Engine
        if (audit_callback) audit_callback("Path: Neural Pass");
        
        GenerateParams params;
        std::string response = neural_engine.generate(prompt, params, [this](const std::string& /*token*/) {
            // Update context window with generated tokens
            // (In real impl, we'd convert text to token IDs first)
            uint16_t mock_token = 42; 
            ctx_manager.add_tokens({&mock_token, 1});
        });

        return response;
    }

private:
    ContextManager ctx_manager;
    SymbolicEngine symbolic_engine;
    NeuralEngine   neural_engine;
    std::function<void(const std::string&)> audit_callback;
};

} // namespace titan
