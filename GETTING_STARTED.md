# 🚀 Sovereign Intelligence Core — Getting Started Guide

## ✅ System Status

Your Sovereign Intelligence Core is **fully operational** and ready for use!

### Current Status
- ✅ **Backend Services**: All running and healthy
- ✅ **API Server**: http://localhost:8081 
- ✅ **WebSocket Chat**: ws://localhost:8081/ws/chat
- ✅ **LLM Engine**: Ollama running on port 11434
- ✅ **Database**: BadgerDB persistent storage
- ✅ **Frontend**: Beautiful modern UIs ready

---

## 🎯 Quick Start (Choose Your Interface)

### 1. **📊 Main Dashboard** (Recommended)
Access the unified command center with all features:
```
http://localhost:8081/web/index-pro.html
```
Shows:
- System status overview
- Quick access to all interfaces
- Service health monitoring
- Live connectivity status

### 2. **💬 AI Chat Interface**
Talk to the ReAct agent with real-time streaming:
```
http://localhost:8081/web/chat.html
```
Features:
- Real-time message streaming
- ReAct reasoning visualization (Thought → Action → Observation)
- Voice input/output (Web Speech API)
- Conversation history
- Settings panel (temperature, model selection)
- Session management

### 3. **📈 Finance Pro Trading System**
Advanced trading with AI signals and market analysis:
```
http://localhost:8081/web/src/finance-pro.html
```
Features:
- Live candlestick charts (TradingView Lightweight Charts)
- Real-time market data
- AI buy/sell signals with confidence scoring
- Order book depth (DOM)
- Position management
- Risk analysis
- Paper trading + broker integration

### 4. **⚙️ Command Center**
Complete system monitoring and control:
```
http://localhost:8081/web/command-center.html
```
Features:
- System health dashboard
- Microservices status (8 services)
- Performance metrics
- Resource allocation (CPU, Memory, Disk, Network)
- AI agent statistics
- Quick control buttons

### 5. **🔧 Status & Health Check**
Verify all systems are operational:
```
http://localhost:8081/web/status.html
```
Performs:
- Service connectivity tests
- API availability checks
- WebSocket connection testing
- Real-time system logging

---

## 🔧 Backend Services (Optional Direct Access)

Access backend services directly:

| Service | URL | Purpose |
|---------|-----|---------|
| Sovereign Core API | http://localhost:8081 | Main REST API |
| Ollama LLM | http://localhost:11434 | Local language model |
| Prometheus | http://localhost:9090 | Metrics collection |
| Grafana | http://localhost:3000 | Dashboards (admin/sovereign) |

---

## 💬 Using the Chat Interface

### Basic Chat
1. Open the **Chat** interface
2. Type your message (e.g., "What is the capital of France?")
3. Watch real-time streaming response
4. See ReAct reasoning steps

### Voice Features
- Click **🎤 Voice Input** to speak your question
- Enable **🔊 Voice Output** to hear responses
- Works in most modern browsers

### Advanced Features
- View conversation history in sidebar
- Access settings panel for model tuning
- Export conversations as JSON/PDF
- Create new chat sessions

### Example Queries
```
"What is machine learning?"
"Solve: 2x^2 + 3x - 5 = 0"
"Find information about quantum computing"
"Analyze this text for sentiment"
"Generate a creative story about robots"
```

---

## 📈 Using Finance Pro

### Getting Started
1. Open **Finance Pro**
2. Select a market from the left sidebar (AAPL, MSFT, BTC, etc.)
3. View real-time chart and market data

### Placing Orders
1. Enter **Quantity** in the trading panel
2. Optionally set **Price** for limit orders
3. Select **Risk Level** (Conservative/Moderate/Aggressive)
4. Click **BUY** or **SELL**

### Understanding AI Signals
- 🟢 **STRONG BUY**: High confidence (>80%)
- 🟡 **HOLD**: Neutral or uncertain signals
- 🔴 **STRONG SELL**: Short opportunity

### Monitoring Positions
- View active positions in the right panel
- See P&L (profit/loss) in real-time
- Track return percentages
- Manage position sizes

---

## ⚙️ System Administration (Command Center)

### Viewing System Health
1. Open **Command Center**
2. View core engine status, latency, memory usage
3. Check microservices status
4. Monitor performance metrics

### Resource Monitoring
- CPU Usage: Real-time percentage
- Memory (RAM): Allocation breakdown
- Disk Space: Storage utilization
- Network I/O: Data throughput

### Quick Controls
- **Clear Cache**: Free up memory
- **Rebuild Indices**: Optimize database
- **Scale Up**: Increase resources
- **Maintenance Mode**: System updates

---

## 🔌 API Reference

### Health Check
```bash
curl http://localhost:8081/health
```
Response:
```json
{"status":"OK","service":"sovereign-core"}
```

### Chat via REST
```bash
curl -X POST http://localhost:8081/api/v1/chat \
  -H "Content-Type: application/json" \
  -d '{"message":"Hello!"}'
```

### WebSocket Chat
```javascript
const ws = new WebSocket('ws://localhost:8081/ws/chat');
ws.send(JSON.stringify({type: 'message', content: 'Hello!'}));
```

---

## 🚨 Troubleshooting

### Services Not Starting?
```bash
cd /home/sachin-kumar/Desktop/coding/1
docker-compose up -d
docker-compose ps
```

### Chat Not Connecting?
1. Check if API is running: `curl http://localhost:8081/health`
2. Check WebSocket: Open Status page and run tests
3. Check browser console for errors (F12)
4. Verify Docker containers: `docker ps`

### High Latency?
1. Open **Command Center**
2. Check CPU/Memory usage
3. Click **Clear Cache** to optimize
4. Monitor request latency metric

### Ollama Not Responding?
```bash
docker-compose logs ollama
docker-compose restart ollama
```

---

## 📚 Feature Overview

### 20 Implemented Features

| # | Feature | Location |
|---|---------|----------|
| 1 | LLM Integration | Chat Interface |
| 2 | WebSocket Streaming | All real-time updates |
| 3 | RAG Pipeline | Knowledge base search |
| 4 | Context Manager | Conversation tracking |
| 5 | Tool System | Web search, math, code |
| 6 | ReAct Agent | Reasoning loop (8 steps) |
| 7 | Memory System | Episodic + semantic |
| 8 | Self-Correction | Auto-retry on error |
| 9 | Expression Parser | Math symbolic computation |
| 10 | WASM Config | Browser-based inference |
| 11 | Dataset Pipeline | Training data management |
| 12 | LoRA Fine-tuning | Model customization |
| 13 | Evaluation Metrics | Perplexity, BLEU, ROUGE |
| 14 | Capability Enforcer | Access control |
| 15 | Rate Limiting | 60 req/min per IP |
| 16 | Chat UI | Real-time interface |
| 17 | Sidebar | Session management |
| 18 | Voice I/O | Web Speech API |
| 19 | Docker Setup | 6-service orchestration |
| 20 | Build Automation | 15+ Makefile targets |

---

## 🎨 UI Customization

### Chat Interface
- Dark theme optimized
- Responsive design (mobile-friendly)
- Customizable text size
- Theme toggle (light/dark)

### Finance Pro
- Real-time chart updates
- Customizable timeframes (1M, 3M, 1Y)
- Draggable panels
- Mobile-responsive layout

### Command Center
- Auto-refreshing metrics
- Sidebar navigation
- Responsive grid layout
- Real-time status updates

---

## 🔐 Security

### Built-in Security Features
- ✅ Rate limiting (60 req/min)
- ✅ Token bucket algorithm
- ✅ Capability-based access control
- ✅ Audit trail with ED25519 signatures
- ✅ HTTPS-ready (use SSL proxy)

### Recommended for Production
1. Set up HTTPS/SSL
2. Change default passwords
3. Configure firewall rules
4. Enable authentication
5. Monitor audit logs

---

## 📊 Monitoring Dashboard

Access Grafana for advanced monitoring:
```
http://localhost:3000
Username: admin
Password: sovereign
```

Includes:
- Request rate metrics
- Error tracking
- Latency graphs
- Service uptime
- Custom alerts

---

## 🧪 Testing

### Run Health Check
Visit: http://localhost:8081/web/status.html

### Test API
```bash
curl http://localhost:8081/health
curl http://localhost:8081/api/v1/chat \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"message":"test"}'
```

### Run Unit Tests
```bash
cd /home/sachin-kumar/Desktop/coding/1
make test
```

---

## 📖 Additional Resources

- **Main Documentation**: IMPLEMENTATION_GUIDE.md
- **Architecture**: INTEGRATION_STATUS.md  
- **Validation Report**: FINAL_VALIDATION_REPORT.md
- **Quick Reference**: QUICKSTART.md

---

## 🆘 Support & Issues

### Common Issues

**Chat not responding?**
- Check: http://localhost:8081/web/status.html
- Verify WebSocket is connected
- Check browser console (F12)

**Finance charts not loading?**
- Refresh the page
- Clear browser cache
- Check network tab in DevTools

**Command center shows errors?**
- Ensure Docker services are running
- Run: `docker-compose ps`
- Check logs: `docker-compose logs`

---

## 🎉 You're All Set!

Your Sovereign Intelligence Core is ready to use!

### Next Steps
1. **Visit Dashboard**: http://localhost:8081/web/index-pro.html
2. **Try Chat**: Ask any question in the AI Chat
3. **Explore Trading**: Check Finance Pro for market analysis
4. **Monitor System**: Use Command Center for oversight

---

**Created**: April 30, 2026  
**Version**: 1.2.0 Production-Grade  
**Status**: ✅ Fully Operational  

Enjoy your Sovereign Intelligence Core! 🚀
