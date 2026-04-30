/**
 * Sovereign Intelligence Core — Premium Chat Frontend
 * Real-time streaming, ReAct visualization, particle background
 */

// ── Particle Background ──
class ParticleField {
    constructor(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
        this.particles = [];
        this.resize();
        window.addEventListener('resize', () => this.resize());
        for (let i = 0; i < 60; i++) this.particles.push(this.createParticle());
        this.animate();
    }
    resize() {
        this.canvas.width = window.innerWidth;
        this.canvas.height = window.innerHeight;
    }
    createParticle() {
        return {
            x: Math.random() * this.canvas.width,
            y: Math.random() * this.canvas.height,
            vx: (Math.random() - 0.5) * 0.3,
            vy: (Math.random() - 0.5) * 0.3,
            r: Math.random() * 1.5 + 0.5,
            a: Math.random() * 0.4 + 0.1
        };
    }
    animate() {
        this.ctx.clearRect(0, 0, this.canvas.width, this.canvas.height);
        for (const p of this.particles) {
            p.x += p.vx; p.y += p.vy;
            if (p.x < 0 || p.x > this.canvas.width) p.vx *= -1;
            if (p.y < 0 || p.y > this.canvas.height) p.vy *= -1;
            this.ctx.beginPath();
            this.ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2);
            this.ctx.fillStyle = `rgba(129,140,248,${p.a})`;
            this.ctx.fill();
        }
        // Draw connections
        for (let i = 0; i < this.particles.length; i++) {
            for (let j = i + 1; j < this.particles.length; j++) {
                const dx = this.particles[i].x - this.particles[j].x;
                const dy = this.particles[i].y - this.particles[j].y;
                const d = Math.sqrt(dx * dx + dy * dy);
                if (d < 150) {
                    this.ctx.beginPath();
                    this.ctx.moveTo(this.particles[i].x, this.particles[i].y);
                    this.ctx.lineTo(this.particles[j].x, this.particles[j].y);
                    this.ctx.strokeStyle = `rgba(129,140,248,${0.06 * (1 - d / 150)})`;
                    this.ctx.stroke();
                }
            }
        }
        requestAnimationFrame(() => this.animate());
    }
}

// ── Main Chat Application ──
class SovereignChat {
    constructor() {
        this.ws = null;
        this.sessionId = null;
        this.currentMessageId = null;
        this.messageCount = 0;
        this.settings = this.loadSettings();
        this.conversations = new Map();
        this.setupEventListeners();
        this.loadSessions();
        this.connect();
    }

    // ── Connection ──
    connect() {
        const protocol = location.protocol === 'https:' ? 'wss:' : 'ws:';
        this.ws = new WebSocket(`${protocol}//${location.host}/ws/chat`);
        this.ws.onopen = () => {
            this.updateConnection(true);
        };
        this.ws.onmessage = (e) => {
            const msg = JSON.parse(e.data);
            this.handleMessage(msg);
        };
        this.ws.onerror = () => this.updateConnection(false);
        this.ws.onclose = () => {
            this.updateConnection(false);
            setTimeout(() => this.connect(), 3000);
        };
    }

    updateConnection(on) {
        const dot = document.getElementById('connectionDot');
        const label = document.getElementById('connectionLabel');
        dot.className = 'conn-dot ' + (on ? 'connected' : 'disconnected');
        label.textContent = on ? 'Connected' : 'Reconnecting...';
    }

    // ── Message Dispatch ──
    handleMessage(msg) {
        switch (msg.type) {
            case 'session_started':
                this.sessionId = msg.session_id;
                break;
            case 'chunk':
                this.showWelcome(false);
                this.hideTyping();
                this.appendChunk(msg.delta);
                break;
            case 'message':
                this.showWelcome(false);
                this.hideTyping();
                this.displayMessage('assistant', msg.content);
                if (this.settings.voiceOutput) this.speak(msg.content);
                break;
            case 'step':
                if (this.settings.showSteps !== false) this.displayStep(msg.step);
                break;
            case 'done':
                this.hideTyping();
                this.markComplete();
                break;
            case 'error':
                this.hideTyping();
                this.displayError(msg.error);
                break;
        }
    }

    // ── Display Messages ──
    displayMessage(role, content) {
        const container = document.getElementById('messagesContainer');
        const el = document.createElement('div');
        el.className = `message message-${role}`;
        const id = `msg-${Date.now()}-${Math.random().toString(36).substr(2,6)}`;
        el.id = id;
        if (role === 'assistant') this.currentMessageId = id;
        this.messageCount++;

        const avatar = document.createElement('div');
        avatar.className = 'msg-avatar';
        avatar.textContent = role === 'user' ? 'U' : 'S';

        const body = document.createElement('div');
        body.className = 'msg-body';

        const bubble = document.createElement('div');
        bubble.className = 'msg-bubble';
        bubble.innerHTML = role === 'assistant' ? this.renderMd(content) : this.esc(content);

        const time = document.createElement('div');
        time.className = 'msg-time';
        time.textContent = new Date().toLocaleTimeString([], {hour:'2-digit',minute:'2-digit'});

        body.appendChild(bubble);
        body.appendChild(time);
        el.appendChild(avatar);
        el.appendChild(body);
        container.appendChild(el);
        container.scrollTop = container.scrollHeight;
        document.getElementById('sessionMeta').textContent = `${this.messageCount} messages`;
    }

    appendChunk(delta) {
        if (!this.currentMessageId) {
            this.displayMessage('assistant', delta);
            return;
        }
        const el = document.getElementById(this.currentMessageId);
        if (!el) return;
        const bubble = el.querySelector('.msg-bubble');
        if (!bubble.dataset.raw) bubble.dataset.raw = '';
        bubble.dataset.raw += delta;
        bubble.innerHTML = this.renderMd(bubble.dataset.raw);
        const container = document.getElementById('messagesContainer');
        container.scrollTop = container.scrollHeight;
    }

    displayStep(step) {
        const container = document.getElementById('messagesContainer');
        const el = document.createElement('div');
        el.className = 'step-display';
        el.innerHTML = `
            <div class="step-header">
                <span class="step-number">Step ${step.number}</span>
                <span class="step-duration">${step.duration_ms}ms</span>
            </div>
            <details class="step-details">
                <summary>🤔 Thought</summary>
                <div class="step-content">${this.esc(step.thought || '')}</div>
            </details>
            ${step.action ? `<details class="step-details"><summary>⚙️ Action</summary><div class="step-content"><code>${this.esc(step.action)}</code></div></details>` : ''}
            ${step.observation ? `<details class="step-details"><summary>👁️ Observation</summary><div class="step-content">${this.renderMd(step.observation)}</div></details>` : ''}
            ${step.error ? `<div class="step-error">❌ ${this.esc(step.error)}</div>` : ''}`;
        container.appendChild(el);
        container.scrollTop = container.scrollHeight;
    }

    displayError(error) {
        const container = document.getElementById('messagesContainer');
        const el = document.createElement('div');
        el.className = 'message message-error';
        el.innerHTML = `<div class="msg-avatar" style="background:rgba(248,113,113,.15);color:var(--danger)">!</div>
            <div class="msg-body"><div class="msg-bubble">${this.esc(error)}</div></div>`;
        container.appendChild(el);
        container.scrollTop = container.scrollHeight;
    }

    markComplete() { this.currentMessageId = null; }

    showTyping() {
        document.getElementById('typingIndicator').style.display = 'flex';
    }
    hideTyping() {
        document.getElementById('typingIndicator').style.display = 'none';
    }
    showWelcome(show) {
        document.getElementById('welcomeScreen').style.display = show ? 'flex' : 'none';
        document.getElementById('messagesContainer').style.display = show ? 'none' : 'flex';
    }

    // ── Events ──
    setupEventListeners() {
        document.getElementById('chatForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.send();
        });

        const input = document.getElementById('messageInput');
        input.addEventListener('input', () => this.resizeInput());
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); this.send(); }
        });

        document.getElementById('newChatBtn').addEventListener('click', () => this.newChat());
        document.getElementById('settingsBtn').addEventListener('click', () => this.toggleModal(true));
        document.getElementById('closeSettings').addEventListener('click', () => this.toggleModal(false));
        document.getElementById('voiceBtn').addEventListener('click', () => this.voiceInput());
        document.getElementById('exportBtn').addEventListener('click', () => this.exportChat());

        document.getElementById('temperatureSlider').addEventListener('input', (e) => {
            document.getElementById('tempValue').textContent = parseFloat(e.target.value).toFixed(2);
            this.settings.temperature = parseFloat(e.target.value);
            this.saveSettings();
        });
        document.getElementById('enableVoiceOutput').addEventListener('change', (e) => {
            this.settings.voiceOutput = e.target.checked;
            this.saveSettings();
        });
        document.getElementById('enableSteps').addEventListener('change', (e) => {
            this.settings.showSteps = e.target.checked;
            this.saveSettings();
        });

        // Prompt suggestions
        document.querySelectorAll('.prompt-suggestion').forEach(btn => {
            btn.addEventListener('click', () => {
                document.getElementById('messageInput').value = btn.dataset.prompt;
                this.send();
            });
        });

        // Sidebar toggle (mobile)
        const toggle = document.getElementById('sidebarToggle');
        if (toggle) toggle.addEventListener('click', () => {
            document.getElementById('sidebar').classList.toggle('open');
        });

        // Close modal on overlay click
        document.getElementById('settingsModal').addEventListener('click', (e) => {
            if (e.target.id === 'settingsModal') this.toggleModal(false);
        });
    }

    send() {
        const input = document.getElementById('messageInput');
        const text = input.value.trim();
        if (!text || !this.ws || this.ws.readyState !== WebSocket.OPEN) return;

        this.showWelcome(false);
        this.displayMessage('user', text);
        this.showTyping();

        this.ws.send(JSON.stringify({ message: text, tier: this.settings.model || 'local' }));
        input.value = '';
        this.resizeInput();
    }

    resizeInput() {
        const el = document.getElementById('messageInput');
        el.style.height = 'auto';
        el.style.height = Math.min(el.scrollHeight, 140) + 'px';
    }

    newChat() {
        this.currentMessageId = null;
        this.messageCount = 0;
        document.getElementById('messagesContainer').innerHTML = '';
        document.getElementById('sessionTitle').textContent = 'New Conversation';
        document.getElementById('sessionMeta').textContent = '';
        this.showWelcome(true);
    }

    // ── Voice ──
    voiceInput() {
        const SR = window.SpeechRecognition || window.webkitSpeechRecognition;
        if (!SR) return alert('Speech Recognition not supported');
        const rec = new SR();
        rec.onstart = () => document.getElementById('voiceBtn').classList.add('active');
        rec.onresult = (e) => {
            let t = '';
            for (let i = e.resultIndex; i < e.results.length; i++) t += e.results[i][0].transcript;
            document.getElementById('messageInput').value = t;
            this.resizeInput();
        };
        rec.onend = () => document.getElementById('voiceBtn').classList.remove('active');
        rec.start();
    }

    speak(text) {
        if (!this.settings.voiceOutput || !('speechSynthesis' in window)) return;
        const u = new SpeechSynthesisUtterance(text.replace(/[#*`_\[\]]/g, ''));
        u.rate = 1.0;
        speechSynthesis.speak(u);
    }

    // ── Export ──
    exportChat() {
        const msgs = document.querySelectorAll('.message');
        let md = `# Sovereign Intelligence Core — Chat Export\n_Exported: ${new Date().toLocaleString()}_\n\n---\n\n`;
        msgs.forEach(m => {
            const role = m.classList.contains('message-user') ? '**You**' : '**Sovereign**';
            const text = m.querySelector('.msg-bubble')?.textContent || '';
            md += `### ${role}\n${text}\n\n`;
        });
        const blob = new Blob([md], { type: 'text/markdown' });
        const a = document.createElement('a');
        a.href = URL.createObjectURL(blob);
        a.download = `sovereign-chat-${Date.now()}.md`;
        a.click();
    }

    // ── Rendering ──
    renderMd(text) {
        if (!text) return '';
        marked.setOptions({ breaks: true, gfm: true });
        let html = marked.parse(text);
        html = html.replace(/<code class="language-(\w+)">([^<]+)<\/code>/g, (m, lang, code) => {
            try {
                return `<code class="hljs language-${lang}">${hljs.highlight(code, {language:lang}).value}</code>`;
            } catch { return m; }
        });
        return html;
    }

    esc(t) {
        const d = document.createElement('div');
        d.textContent = t || '';
        return d.innerHTML;
    }

    // ── Sessions ──
    loadSessions() {
        const data = localStorage.getItem('sovereignSessions');
        if (data) {
            try { this.displaySessions(JSON.parse(data)); } catch {}
        }
    }
    displaySessions(sessions) {
        const list = document.getElementById('sessionsList');
        list.innerHTML = '';
        sessions.forEach(s => {
            const el = document.createElement('div');
            el.className = 'session-item';
            el.innerHTML = `<div class="session-title">${s.title}</div><div class="session-time">${new Date(s.timestamp).toLocaleDateString()}</div>`;
            el.addEventListener('click', () => this.loadSession(s));
            list.appendChild(el);
        });
    }
    loadSession(s) {
        const conv = this.conversations.get(s.id);
        if (conv) {
            document.getElementById('messagesContainer').innerHTML = '';
            document.getElementById('sessionTitle').textContent = s.title;
            this.showWelcome(false);
            conv.messages.forEach(m => this.displayMessage(m.role, m.content));
        }
    }

    // ── Settings ──
    toggleModal(show) { document.getElementById('settingsModal').style.display = show ? 'flex' : 'none'; }
    loadSettings() {
        const s = localStorage.getItem('sovereignSettings');
        return s ? JSON.parse(s) : { temperature: 0.7, model: 'local', voiceOutput: false, streaming: true, showSteps: true };
    }
    saveSettings() { localStorage.setItem('sovereignSettings', JSON.stringify(this.settings)); }
}

// ── Init ──
document.addEventListener('DOMContentLoaded', () => {
    const canvas = document.getElementById('particleCanvas');
    if (canvas) new ParticleField(canvas);
    new SovereignChat();
});
