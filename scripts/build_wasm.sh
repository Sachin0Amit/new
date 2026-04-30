#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────
# Sovereign Intelligence Core — WASM Expression Parser Builder
# Compiles the C++ expression parser to WebAssembly via Emscripten
# ──────────────────────────────────────────────────────────────
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
CPP_SRC="$PROJECT_ROOT/cpp/src/expression_parser.cpp"
OUTPUT_DIR="$PROJECT_ROOT/web/wasm"
OUTPUT_NAME="expression_parser"

echo "╔══════════════════════════════════════════════╗"
echo "║  Sovereign WASM Builder — Expression Parser  ║"
echo "╚══════════════════════════════════════════════╝"

# ── Prerequisite check ──
if ! command -v emcc &>/dev/null; then
    echo "❌ Emscripten (emcc) not found."
    echo "   Install: https://emscripten.org/docs/getting_started/downloads.html"
    echo ""
    echo "   Quick setup:"
    echo "     git clone https://github.com/emscripten-core/emsdk.git"
    echo "     cd emsdk && ./emsdk install latest && ./emsdk activate latest"
    echo "     source ./emsdk_env.sh"
    exit 1
fi

if [ ! -f "$CPP_SRC" ]; then
    echo "❌ Source file not found: $CPP_SRC"
    exit 1
fi

# ── Create output directory ──
mkdir -p "$OUTPUT_DIR"

echo "📦 Source:  $CPP_SRC"
echo "📂 Output:  $OUTPUT_DIR/"
echo ""

# ── Compile to WASM ──
echo "🔨 Compiling C++ → WebAssembly..."
emcc "$CPP_SRC" \
    -O3 \
    -std=c++17 \
    -s WASM=1 \
    -s ALLOW_MEMORY_GROWTH=1 \
    -s MODULARIZE=1 \
    -s EXPORT_NAME="ExpressionParser" \
    -s EXPORTED_FUNCTIONS='["_parse_and_differentiate","_evaluate_expression","_malloc","_free"]' \
    -s EXPORTED_RUNTIME_METHODS='["cwrap","ccall","UTF8ToString","stringToUTF8","lengthBytesUTF8"]' \
    -s NO_EXIT_RUNTIME=1 \
    -s ENVIRONMENT=web \
    --no-entry \
    -o "$OUTPUT_DIR/${OUTPUT_NAME}.js"

echo ""
echo "✅ WASM build complete:"
ls -lh "$OUTPUT_DIR/${OUTPUT_NAME}".{js,wasm} 2>/dev/null || true

# ── Generate JS wrapper ──
cat > "$OUTPUT_DIR/parser_wrapper.js" << 'WRAPPER_EOF'
/**
 * Sovereign Expression Parser — WASM Wrapper
 * Provides a clean API over the raw Emscripten bindings.
 */
class SovereignExpressionParser {
    constructor() {
        this.module = null;
        this.ready = false;
    }

    async init() {
        if (this.ready) return;
        this.module = await ExpressionParser();
        this._differentiate = this.module.cwrap('parse_and_differentiate', 'string', ['string', 'string']);
        this._evaluate = this.module.cwrap('evaluate_expression', 'number', ['string']);
        this.ready = true;
    }

    /**
     * Compute the symbolic derivative of an expression.
     * @param {string} expr - e.g. "x^2 + sin(x)"
     * @param {string} variable - e.g. "x"
     * @returns {string} The derivative expression
     */
    differentiate(expr, variable = 'x') {
        if (!this.ready) throw new Error('Parser not initialized. Call init() first.');
        return this._differentiate(expr, variable);
    }

    /**
     * Numerically evaluate an expression (variables must be substituted first).
     * @param {string} expr - e.g. "3.14 * 2"
     * @returns {number} The result
     */
    evaluate(expr) {
        if (!this.ready) throw new Error('Parser not initialized. Call init() first.');
        return this._evaluate(expr);
    }
}

// Export for browser and Node.js
if (typeof window !== 'undefined') window.SovereignExpressionParser = SovereignExpressionParser;
if (typeof module !== 'undefined') module.exports = SovereignExpressionParser;
WRAPPER_EOF

echo "📝 Generated wrapper: $OUTPUT_DIR/parser_wrapper.js"
echo ""
echo "🎯 Usage in browser:"
echo '   <script src="wasm/expression_parser.js"></script>'
echo '   <script src="wasm/parser_wrapper.js"></script>'
echo '   <script>'
echo '     const parser = new SovereignExpressionParser();'
echo '     await parser.init();'
echo '     console.log(parser.differentiate("x^2 + sin(x)", "x"));'
echo '   </script>'
