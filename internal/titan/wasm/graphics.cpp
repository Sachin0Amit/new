#include <emscripten/bind.h>
#include <vector>
#include <cmath>
#include <iostream>

using namespace emscripten;

/**
 * Biomechanical Graphics Engine (Titan Core)
 * Procedural generation of Giger-esque alien geometry.
 */

struct Vertex {
    float x, y, z;
};

class BiomechanicalGen {
public:
    BiomechanicalGen() {}

    /**
     * Generates vertex positions for a rib-like structural element.
     * Higher performance than JS loops for complex trigonometry.
     */
    val generateRibs(int segments, float radius, float height) {
        std::vector<float> vertices;
        for (int i = 0; i <= segments; ++i) {
            float theta = (i * 2.0f * M_PI) / segments;
            float r = radius + (std::sin(theta * 5.0f) * 0.2f); // Organic variation
            
            vertices.push_back(r * std::cos(theta));
            vertices.push_back(height);
            vertices.push_back(r * std::sin(theta));
        }
        
        // Convert to typed array for JS consumption
        return val(typed_memory_view(vertices.size(), vertices.data()));
    }

    /**
     * Generates a sinuous pipe curve path.
     */
    val generatePipes(int points, int offset) {
        std::vector<float> path;
        for (int i = 0; i < points; ++i) {
            float y = (i * 0.2f) - 5.0f;
            float angle = (i * 0.2f) + offset;
            float radius = 2.0f + (std::sin(i * 0.1f) * 0.2f);

            path.push_back(std::cos(angle) * radius);
            path.push_back(y);
            path.push_back(std::sin(angle) * radius);
        }
        return val(typed_memory_view(path.size(), path.data()));
    }
};

// --- Emscripten Bindings ---
EMSCRIPTEN_BINDINGS(titan_module) {
    class_<BiomechanicalGen>("BiomechanicalGen")
        .constructor<>()
        .function("generateRibs", &BiomechanicalGen::generateRibs)
        .function("generatePipes", &BiomechanicalGen::generatePipes);
}
