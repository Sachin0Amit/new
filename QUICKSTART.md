# Quick Start & Testing Guide

## Prerequisites

```bash
# Install Docker & Docker Compose
sudo apt-get install docker.io docker-compose

# Verify Go 1.21+
go version

# Verify Python 3.10+
python3 --version
```

## One-Command Setup

```bash
# From project root
make docker-up
```

This starts all 6 services:
- Sovereign Core (8081)
- Ollama (11434) 
- Python Sidecar (5000)
- Prometheus (9090)
- Grafana (3000)
- BadgerDB Explorer (8002)

## Verify Services

```bash
# Health check all endpoints
make health-check

# Expected output:
# ✓ Sovereign Core (http://localhost:8081) - OK
# ✓ Ollama (http://localhost:11434) - OK
# ✓ Python Sidecar (http://localhost:5000) - OK
# ✓ Prometheus (http://localhost:9090) - OK
# ✓ Grafana (http://localhost:3000) - OK
```

## Test Chat Interface

1. Open browser: **http://localhost:8081/web/chat.html**
2. Type a message: "What is the capital of France?"
3. Observe:
   - Message sent to Ollama via WebSocket
   - Real-time streaming response with chunks
   - Markdown rendering
   - Session saved to localStorage

## Test API Endpoints

```bash
# Health check
curl http://localhost:8081/health

# Chat via REST (non-streaming)
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello","tier":"local"}'

# WebSocket (via wscat or browser console)
# Connection: ws://localhost:8081/ws/chat
# Message format: {"type":"chat","text":"Your message"}
```

## Training Pipeline

```bash
# Download datasets (Alpaca + Dolly)
make download-data
# Creates: data/training/train.jsonl, val.jsonl

# Fine-tune with LoRA (requires GPU)
make finetune
# Output: models/checkpoint-*/

# Evaluate model
make eval
# Output: results/eval_metrics.json
```

## View Metrics

- **Prometheus:** http://localhost:9090/graph
- **Grafana:** http://localhost:3000/ (user: admin, pass: admin)
- **BadgerDB:** http://localhost:8002/

## Common Issues & Solutions

### Ollama Not Available
```bash
# Check if Ollama is running
docker-compose ps

# Restart Ollama
docker-compose restart ollama

# View Ollama logs
docker-compose logs ollama
```

### Port Already in Use
```bash
# Find process on port 8081
lsof -i :8081

# Or change docker-compose port mapping
# Edit docker-compose.yml, change "8081:8081" to "8080:8081"
```

### Memory Issues
```bash
# Reduce model size in docker-compose.yml
# Change OLLAMA_MODEL from "mistral" to "orca-mini"
# Or set OLLAMA_NUM_GPU=0 for CPU-only mode
```

## Stop Services

```bash
make docker-down

# Or manual cleanup
docker-compose down -v
```

## Logs & Debugging

```bash
# View all logs
docker-compose logs -f

# View specific service
docker-compose logs -f sovereign-core

# Enter container
docker-compose exec sovereign-core bash
```

## Performance Benchmarks

```bash
# Test LLM response time
time curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Explain quantum computing in 100 words.","tier":"local"}'

# Expected: <5 seconds (GPU), <30 seconds (CPU)
```

## Integration Test Checklist

- [ ] Services start without errors (`docker-compose logs`)
- [ ] Health check passes all 5 endpoints (`make health-check`)
- [ ] Chat UI loads (http://localhost:8081/web/chat.html)
- [ ] Message sends and receives streaming response
- [ ] WebSocket connection shows "connected" status
- [ ] Metrics endpoint returns JSON (http://localhost:8081/metrics)
- [ ] Can interact with agent (ask a question, get tool usage)
- [ ] Training data downloads successfully (`make download-data`)
- [ ] Model fine-tunes without errors (`make finetune`)
- [ ] Evaluation metrics generate (`make eval`)

## Success Indicators

✅ **System Ready When:**
1. All docker-compose services are healthy
2. Ollama model loads (check logs: "successfully loaded model")
3. Chat messages round-trip with responses
4. WebSocket streaming shows real-time chunks
5. ReAct agent executes tool calls
6. Training pipeline completes without errors

## Next Steps After Setup

1. **Customize Training Data:** Add custom examples to `data/training/`
2. **Fine-tune Custom Model:** Run `make finetune` with your data
3. **Deploy to Production:** Use `./scripts/launch.sh`
4. **Monitor Performance:** Check Grafana dashboards
5. **Scale with P2P:** Configure libp2p node for distributed inference

## Support & Documentation

- Implementation Details: [IMPLEMENTATION_GUIDE.md](IMPLEMENTATION_GUIDE.md)
- Integration Status: [INTEGRATION_STATUS.md](INTEGRATION_STATUS.md)
- Architecture: [docs/README.md](docs/README.md)
- Dependencies: [DEPENDENCIES.md](DEPENDENCIES.md)

---

**Status:** ✅ Ready for Testing
