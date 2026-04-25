# Sovereign Core Dependency Policy

To maintain a maintainable and decoupled architecture, the following dependency rules are enforced:

## 1. Import Hierarchy
The project follows a strict layered architecture:
- `cmd/` -> `internal/core`
- `internal/core` -> `internal/titan`, `internal/guard`, `internal/mesh`, `internal/auditor`, `internal/reflex`
- `internal/*` -> `internal/models`, `pkg/*`

## 2. Forbidden Imports
- **No Circular Dependencies**: Packages must never import each other. Use interfaces to break cycles.
- **No Reverse Imports**: `internal/` or `pkg/` must NEVER import `cmd/`.
- **No Concrete Dependency**: Other packages must depend on the interfaces defined in the module's root file (e.g., `titan.Engine`), never the concrete implementation (e.g., `*titan.sovereignEngine`).

## 3. Module Boundaries
- **Titan**: Only handles C++ bridge and inference logic.
- **Guard**: Only handles authentication and authorization.
- **Mesh**: Only handles storage and retrieval.
- **Auditor**: Only handles audit logging.
- **Reflex**: Only handles monitoring and self-healing.

Any cross-module orchestration must happen in `internal/core` or a dedicated orchestrator service.
