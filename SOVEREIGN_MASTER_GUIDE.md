# Sovereign Intelligence Core: The Master Guide

Welcome to the **Sovereign Intelligence Core**, a local-first, autonomous AI ecosystem built from first principles. This guide provides a unified overview of all sub-systems, their architectures, and operational procedures.

---

## 1. System Architecture Overview

The Sovereign Core is a polyglot system designed for maximum privacy and performance.

| Component | Language/Stack | Role |
| :--- | :--- | :--- |
| **Orchestrator** | Go | WebSocket server, P2P fleet management, API routing. |
| **Neural Engine** | C++ (Titan) | High-performance inference core using shared libraries. |
| **Foundation Model** | Python (PyTorch) | From-scratch Transformer architecture for local training. |
| **Finance Engine** | Go + C++ | Real-time market analysis and algorithmic trading. |
| **Portal Hub** | THREE.js + GSAP | Premium 3D-visualized interface for system access. |

---

## 2. Core Modules

### 2.1 The Hub (Portal)
Accessible at `http://localhost:8081/`.
The central entry point. Features a Mobius-strip 3D visualization.
- **Neural Chat**: Standard LLM interaction with smooth streaming.
- **Finance Engine**: Professional-grade trading terminal with glassmorphic UI.
- **Command Center**: Real-time metrics and system health monitoring.

### 2.2 PapiAi Foundation Model (From Scratch)
Located in `papiai_foundation_model/`.
Build your own AI from zero to high-level without external dependencies.
- **Generate Data**: `make foundation-gen` (Synthetic logic/math/code data).
- **Train Matrix**: `make foundation-train` (Neural compilation from scratch).
- **Run Inference**: `make foundation-chat` (Direct interaction with your weights).

---

## 3. Operational Commands

### 3.1 Development & Execution
```bash
make dev             # Launch the Sovereign Core backend (Go)
make build           # Compile Go, C++, and Frontend assets
make clean           # Wipe all build artifacts
```

### 3.2 AI Training & Inference
```bash
make foundation-gen     # Generate synthetic intelligence data
make foundation-train   # Compile the neural engine from scratch
make foundation-chat    # Chat with the custom-built AI
```

---

## 4. Design Aesthetics
The UI follows a **Glassmorphic Premium Dark** theme.
- **CSS Tokens**: Located in `web/css/admin_premium.css` and `web/src/css/finance_premium.css`.
- **Animations**: Powered by GSAP for smooth micro-interactions.

---

## 5. Security & Isolation
- **Purged External AI**: All connections to external providers (OpenAI, Anthropic, etc.) have been completely removed from the codebase.
- **100% Local Intelligence**: The system relies exclusively on the **Titan C++ Engine** and your custom **PapiAi Foundation weights**.
- **Sovereign Encryption**: All internal communications are mathematically verified.
- **Auditable**: Every inference step is recorded in the local audit trail.

---

*Architect: Antigravity*
*Sovereign Intelligence Core v1.0.0*
