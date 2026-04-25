# Diff Details

Date : 2026-04-25 16:15:57

Directory /home/sachin-kumar/Desktop/coding/1

Total : 87 files,  2971 codes, 18 comments, -151 blanks, all 2838 lines

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details

## Files
| filename | language | code | comment | blank | total |
| :--- | :--- | ---: | ---: | ---: | ---: |
| [cpp/src/context_manager.cpp](/cpp/src/context_manager.cpp) | C++ | 58 | 7 | 11 | 76 |
| [cpp/src/inference_queue.cpp](/cpp/src/inference_queue.cpp) | C++ | 3 | -1 | 0 | 2 |
| [cpp/src/neural_engine.cpp](/cpp/src/neural_engine.cpp) | C++ | 42 | 8 | 13 | 63 |
| [cpp/src/simd_utils.h](/cpp/src/simd_utils.h) | C++ | 52 | 1 | 13 | 66 |
| [cpp/src/symbolic_engine.cpp](/cpp/src/symbolic_engine.cpp) | C++ | 35 | 3 | 9 | 47 |
| [cpp/src/titan.cpp](/cpp/src/titan.cpp) | C++ | 39 | 4 | 11 | 54 |
| [deploy/grafana/dashboards/sovereign.json](/deploy/grafana/dashboards/sovereign.json) | JSON | 32 | 0 | 1 | 33 |
| [deploy/prometheus/alerts.yml](/deploy/prometheus/alerts.yml) | YAML | 27 | 0 | 3 | 30 |
| [deploy/prometheus/prometheus.yml](/deploy/prometheus/prometheus.yml) | YAML | 9 | 0 | 3 | 12 |
| [go.mod](/go.mod) | Go Module File | 12 | 0 | 0 | 12 |
| [internal/auditor/auditor_test.go](/internal/auditor/auditor_test.go) | Go | 27 | 5 | 9 | 41 |
| [internal/auditor/keystore.go](/internal/auditor/keystore.go) | Go | 88 | 2 | 23 | 113 |
| [internal/auditor/store.go](/internal/auditor/store.go) | Go | 46 | 1 | 11 | 58 |
| [internal/auditor/trail.go](/internal/auditor/trail.go) | Go | 63 | 1 | 10 | 74 |
| [internal/auditor/verifier.go](/internal/auditor/verifier.go) | Go | 29 | 4 | 8 | 41 |
| [internal/auth/auth_test.go](/internal/auth/auth_test.go) | Go | 58 | 7 | 15 | 80 |
| [internal/auth/jwt.go](/internal/auth/jwt.go) | Go | 76 | 6 | 17 | 99 |
| [internal/auth/middleware.go](/internal/auth/middleware.go) | Go | 97 | 5 | 20 | 122 |
| [internal/core/health.go](/internal/core/health.go) | Go | 30 | 2 | 8 | 40 |
| [internal/guard/audit_hook.go](/internal/guard/audit_hook.go) | Go | 16 | 3 | 7 | 26 |
| [internal/guard/capabilities.go](/internal/guard/capabilities.go) | Go | 33 | 0 | 6 | 39 |
| [internal/guard/enforcer.go](/internal/guard/enforcer.go) | Go | 53 | 3 | 13 | 69 |
| [internal/guard/guard_test.go](/internal/guard/guard_test.go) | Go | -4 | 4 | 8 | 8 |
| [internal/guard/sandbox.go](/internal/guard/sandbox.go) | Go | 52 | 2 | 14 | 68 |
| [internal/mathsolver/client.go](/internal/mathsolver/client.go) | Go | 93 | 2 | 19 | 114 |
| [internal/metrics/metrics.go](/internal/metrics/metrics.go) | Go | 45 | 1 | 12 | 58 |
| [internal/reflex/budget.go](/internal/reflex/budget.go) | Go | 29 | 0 | 9 | 38 |
| [internal/reflex/corrector.go](/internal/reflex/corrector.go) | Go | 48 | 5 | 12 | 65 |
| [internal/reflex/detector.go](/internal/reflex/detector.go) | Go | 73 | 5 | 12 | 90 |
| [internal/reflex/learner.go](/internal/reflex/learner.go) | Go | 27 | 2 | 9 | 38 |
| [internal/reflex/reflex_test.go](/internal/reflex/reflex_test.go) | Go | 55 | 3 | 13 | 71 |
| [internal/scheduler/gravity.go](/internal/scheduler/gravity.go) | Go | 61 | 2 | 11 | 74 |
| [internal/scheduler/migrator.go](/internal/scheduler/migrator.go) | Go | 35 | 4 | 8 | 47 |
| [internal/scheduler/receiver.go](/internal/scheduler/receiver.go) | Go | 23 | 3 | 9 | 35 |
| [internal/scheduler/scheduler.go](/internal/scheduler/scheduler.go) | Go | 38 | 6 | 10 | 54 |
| [internal/scheduler/scheduler_test.go](/internal/scheduler/scheduler_test.go) | Go | 48 | 5 | 15 | 68 |
| [internal/telemetry/otel.go](/internal/telemetry/otel.go) | Go | 41 | 1 | 9 | 51 |
| [internal/titan/bridge.go](/internal/titan/bridge.go) | Go | 51 | 5 | 13 | 69 |
| [internal/titan/bridge_test.go](/internal/titan/bridge_test.go) | Go | 41 | 5 | 13 | 59 |
| [internal/titan/fallback.go](/internal/titan/fallback.go) | Go | 63 | 3 | 15 | 81 |
| [internal/titan/pool.go](/internal/titan/pool.go) | Go | 60 | 1 | 14 | 75 |
| [internal/titan/titan.go](/internal/titan/titan.go) | Go | -3 | -4 | -4 | -11 |
| [math_solver/Dockerfile](/math_solver/Dockerfile) | Docker | 16 | 8 | 10 | 34 |
| [math_solver/server.py](/math_solver/server.py) | Python | 36 | 0 | 7 | 43 |
| [math_solver/solver.py](/math_solver/solver.py) | Python | 55 | 0 | 11 | 66 |
| [math_solver/tests/test_solver.py](/math_solver/tests/test_solver.py) | Python | 26 | 1 | 8 | 35 |
| [web/DESIGN_SYSTEM.md](/web/DESIGN_SYSTEM.md) | Markdown | 38 | 0 | 12 | 50 |
| [web/README.md](/web/README.md) | Markdown | 32 | 0 | 11 | 43 |
| [web/dist/admin/WindowManager.js](/web/dist/admin/WindowManager.js) | JavaScript | -100 | -2 | -15 | -117 |
| [web/dist/admin/desktop.css](/web/dist/admin/desktop.css) | CSS | -141 | 0 | -20 | -161 |
| [web/dist/admin/desktop.html](/web/dist/admin/desktop.html) | HTML | -247 | 0 | -20 | -267 |
| [web/dist/assets/index-CH0pkqpO.js](/web/dist/assets/index-CH0pkqpO.js) | JavaScript | 65 | 1 | 6 | 72 |
| [web/dist/chat-simple.html](/web/dist/chat-simple.html) | HTML | -359 | 0 | -46 | -405 |
| [web/dist/chat.html](/web/dist/chat.html) | HTML | -106 | 0 | -10 | -116 |
| [web/dist/css/chat.css](/web/dist/css/chat.css) | CSS | -638 | 0 | -86 | -724 |
| [web/dist/css/cursor.css](/web/dist/css/cursor.css) | CSS | -116 | 0 | -17 | -133 |
| [web/dist/css/finance.css](/web/dist/css/finance.css) | CSS | -283 | 0 | -49 | -332 |
| [web/dist/css/login.css](/web/dist/css/login.css) | CSS | -126 | 0 | -18 | -144 |
| [web/dist/css/settings.css](/web/dist/css/settings.css) | CSS | -198 | 0 | -26 | -224 |
| [web/dist/css/style.css](/web/dist/css/style.css) | CSS | -775 | 0 | -102 | -877 |
| [web/dist/finance.html](/web/dist/finance.html) | HTML | -164 | 0 | -24 | -188 |
| [web/dist/index.html](/web/dist/index.html) | HTML | -480 | 0 | -38 | -518 |
| [web/dist/js/chat-simple.js](/web/dist/js/chat-simple.js) | JavaScript | -148 | -2 | -26 | -176 |
| [web/dist/js/chat.js](/web/dist/js/chat.js) | JavaScript | -565 | -43 | -121 | -729 |
| [web/dist/js/core/Cursor.js](/web/dist/js/core/Cursor.js) | JavaScript | -88 | -6 | -20 | -114 |
| [web/dist/js/core/Engine.js](/web/dist/js/core/Engine.js) | JavaScript | -93 | -5 | -26 | -124 |
| [web/dist/js/core/Tower.js](/web/dist/js/core/Tower.js) | JavaScript | -257 | -12 | -36 | -305 |
| [web/dist/js/finance.js](/web/dist/js/finance.js) | JavaScript | -368 | -30 | -64 | -462 |
| [web/dist/js/login.js](/web/dist/js/login.js) | JavaScript | -98 | -4 | -19 | -121 |
| [web/dist/js/script.js](/web/dist/js/script.js) | JavaScript | -44 | -7 | -9 | -60 |
| [web/dist/js/settings.js](/web/dist/js/settings.js) | JavaScript | -184 | -8 | -24 | -216 |
| [web/index.html](/web/index.html) | HTML | 104 | 0 | 15 | 119 |
| [web/package-lock.json](/web/package-lock.json) | JSON | 5,542 | 0 | 1 | 5,543 |
| [web/package.json](/web/package.json) | JSON | 28 | 0 | 1 | 29 |
| [web/src/admin/dashboard.ts](/web/src/admin/dashboard.ts) | TypeScript | 57 | 0 | 8 | 65 |
| [web/src/admin/export.ts](/web/src/admin/export.ts) | TypeScript | 31 | 0 | 7 | 38 |
| [web/src/admin/fleet-matrix.ts](/web/src/admin/fleet-matrix.ts) | TypeScript | 73 | 1 | 14 | 88 |
| [web/src/admin/reasoning-visualizer.ts](/web/src/admin/reasoning-visualizer.ts) | TypeScript | 92 | 3 | 19 | 114 |
| [web/src/admin/reflex-monitor.ts](/web/src/admin/reflex-monitor.ts) | TypeScript | 52 | 0 | 8 | 60 |
| [web/src/chat.ts](/web/src/chat.ts) | TypeScript | 90 | 0 | 17 | 107 |
| [web/src/finance.ts](/web/src/finance.ts) | TypeScript | 71 | 0 | 11 | 82 |
| [web/src/login.ts](/web/src/login.ts) | TypeScript | 76 | 4 | 18 | 98 |
| [web/src/settings.ts](/web/src/settings.ts) | TypeScript | 54 | 3 | 14 | 71 |
| [web/src/styles/components.css](/web/src/styles/components.css) | CSS | 113 | 0 | 17 | 130 |
| [web/src/styles/tokens.css](/web/src/styles/tokens.css) | CSS | 47 | 0 | 13 | 60 |
| [web/tsconfig.json](/web/tsconfig.json) | JSON with Comments | 23 | 0 | 3 | 26 |
| [web/vite.config.ts](/web/vite.config.ts) | TypeScript | 27 | 0 | 2 | 29 |

[Summary](results.md) / [Details](details.md) / [Diff Summary](diff.md) / Diff Details