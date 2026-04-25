/**
 * Sovereign Intelligence Core - Chat Frontend
 * Handles real-time streaming, markdown rendering, tool visualization
 */

class SovereignChat {
    constructor() {
        this.ws = null;
        this.sessionId = null;
        this.currentMessageId = null;
        this.settings = this.loadSettings();
        this.conversations = new Map();
        
        this.setupEventListeners();
        this.loadSessions();
        this.connect();
    }

    // ==================== CONNECTION ====================
    
    connect() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const url = `${protocol}//${window.location.host}/ws/chat`;
        
        this.ws = new WebSocket(url);
        
        this.ws.onopen = () => {
            console.log('Connected to Sovereign Core');
            this.updateConnectionStatus(true);
        };
        
        this.ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            this.handleMessage(msg);
        };
        
        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
            this.updateConnectionStatus(false);
        };
        
        this.ws.onclose = () => {
            console.log('Disconnected from Sovereign Core');
            this.updateConnectionStatus(false);
            // Reconnect after 3 seconds
            setTimeout(() => this.connect(), 3000);
        };
    }

    updateConnectionStatus(connected) {
        const status = document.getElementById('connectionStatus');
        if (connected) {
            status.classList.remove('disconnected');
            status.classList.add('connected');
        } else {
            status.classList.add('disconnected');
            status.classList.remove('connected');
        }
    }

    // ==================== MESSAGE HANDLING ====================
    
    handleMessage(msg) {
        switch (msg.type) {
            case 'session_started':
                this.sessionId = msg.session_id;
                this.newConversation();
                break;
            case 'chunk':
                this.appendChunk(msg.delta);
                break;
            case 'message':
                this.displayMessage('assistant', msg.content);
                break;
            case 'step':
                this.displayStep(msg.step);
                break;
            case 'done':
                this.markMessageComplete();
                break;
            case 'error':
                this.displayError(msg.error);
                break;
        }
    }

    // ==================== UI UPDATES ====================
    
    displayMessage(role, content, isStreaming = false) {
        const container = document.getElementById('messagesContainer');
        
        if (role === 'assistant' && isStreaming) {
            // Update existing streaming message
            if (this.currentMessageId) {
                const msgEl = document.getElementById(this.currentMessageId);
                if (msgEl) {
                    msgEl.innerHTML = this.renderMarkdown(content);
                    return;
                }
            }
        }
        
        const messageEl = document.createElement('div');
        messageEl.className = `message message-${role}`;
        
        const id = `msg-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
        messageEl.id = id;
        
        if (role === 'assistant') {
            this.currentMessageId = id;
        }
        
        const contentEl = document.createElement('div');
        contentEl.className = 'message-content';
        
        if (role === 'assistant') {
            contentEl.innerHTML = this.renderMarkdown(content);
        } else {
            contentEl.textContent = content;
        }
        
        messageEl.appendChild(contentEl);
        
        // Add timestamp
        const timeEl = document.createElement('div');
        timeEl.className = 'message-time';
        timeEl.textContent = new Date().toLocaleTimeString();
        messageEl.appendChild(timeEl);
        
        container.appendChild(messageEl);
        container.scrollTop = container.scrollHeight;
    }

    appendChunk(delta) {
        if (!this.currentMessageId) {
            this.displayMessage('assistant', delta, true);
            return;
        }
        
        const msgEl = document.getElementById(this.currentMessageId);
        if (msgEl) {
            const contentEl = msgEl.querySelector('.message-content');
            if (!contentEl.dataset.fullText) {
                contentEl.dataset.fullText = '';
            }
            contentEl.dataset.fullText += delta;
            contentEl.innerHTML = this.renderMarkdown(contentEl.dataset.fullText);
        }
        
        // Scroll to bottom
        const container = document.getElementById('messagesContainer');
        container.scrollTop = container.scrollHeight;
    }

    displayStep(step) {
        const container = document.getElementById('messagesContainer');
        
        const stepEl = document.createElement('div');
        stepEl.className = 'step-display';
        
        stepEl.innerHTML = `
            <div class="step-header">
                <span class="step-number">Step ${step.number}</span>
                <span class="step-duration">${step.duration_ms}ms</span>
            </div>
            <details class="step-details">
                <summary>🤔 Thought</summary>
                <div class="step-content">${this.escapeHtml(step.thought)}</div>
            </details>
            ${step.action ? `
                <details class="step-details">
                    <summary>⚙️ Action</summary>
                    <div class="step-content"><code>${this.escapeHtml(step.action)}</code></div>
                </details>
            ` : ''}
            ${step.observation ? `
                <details class="step-details">
                    <summary>👁️ Observation</summary>
                    <div class="step-content">${this.renderMarkdown(step.observation)}</div>
                </details>
            ` : ''}
            ${step.error ? `
                <div class="step-error">❌ Error: ${this.escapeHtml(step.error)}</div>
            ` : ''}
        `;
        
        container.appendChild(stepEl);
        container.scrollTop = container.scrollHeight;
    }

    displayError(error) {
        const container = document.getElementById('messagesContainer');
        const errorEl = document.createElement('div');
        errorEl.className = 'message message-error';
        errorEl.textContent = `Error: ${error}`;
        container.appendChild(errorEl);
        container.scrollTop = container.scrollHeight;
    }

    markMessageComplete() {
        if (this.currentMessageId) {
            const msgEl = document.getElementById(this.currentMessageId);
            if (msgEl) {
                msgEl.classList.add('complete');
            }
        }
        this.currentMessageId = null;
    }

    // ==================== EVENT LISTENERS ====================
    
    setupEventListeners() {
        // Chat form
        document.getElementById('chatForm').addEventListener('submit', (e) => {
            e.preventDefault();
            this.sendMessage();
        });
        
        // Auto-resize textarea
        const input = document.getElementById('messageInput');
        input.addEventListener('input', () => this.resizeInput());
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });
        
        // Sidebar buttons
        document.getElementById('newChatBtn').addEventListener('click', () => this.newConversation());
        document.getElementById('settingsBtn').addEventListener('click', () => this.showSettings());
        document.getElementById('voiceBtn').addEventListener('click', () => this.startVoiceInput());
        
        // Modal
        document.querySelector('.close').addEventListener('click', () => this.hideSettings());
        
        // Settings
        document.getElementById('temperatureSlider').addEventListener('input', (e) => {
            document.getElementById('tempValue').textContent = e.target.value;
            this.settings.temperature = parseFloat(e.target.value);
            this.saveSettings();
        });
        
        document.getElementById('enableVoiceOutput').addEventListener('change', (e) => {
            this.settings.voiceOutput = e.target.checked;
            this.saveSettings();
        });
    }

    sendMessage() {
        const input = document.getElementById('messageInput');
        const message = input.value.trim();
        
        if (!message || !this.ws || this.ws.readyState !== WebSocket.OPEN) {
            return;
        }
        
        // Display user message
        this.displayMessage('user', message);
        
        // Send to server
        this.ws.send(JSON.stringify({
            message: message,
            tier: this.settings.model || 'local'
        }));
        
        input.value = '';
        this.resizeInput();
    }

    resizeInput() {
        const input = document.getElementById('messageInput');
        input.style.height = 'auto';
        input.style.height = Math.min(input.scrollHeight, 150) + 'px';
    }

    // ==================== VOICE INPUT/OUTPUT ====================
    
    startVoiceInput() {
        if (!('webkitSpeechRecognition' in window || 'SpeechRecognition' in window)) {
            alert('Speech Recognition not supported in this browser');
            return;
        }
        
        const SpeechRecognition = window.SpeechRecognition || window.webkitSpeechRecognition;
        const recognition = new SpeechRecognition();
        
        recognition.onstart = () => {
            console.log('Listening...');
            document.getElementById('voiceBtn').classList.add('active');
        };
        
        recognition.onresult = (event) => {
            let transcript = '';
            for (let i = event.resultIndex; i < event.results.length; i++) {
                transcript += event.results[i][0].transcript;
            }
            document.getElementById('messageInput').value = transcript;
            this.resizeInput();
        };
        
        recognition.onend = () => {
            console.log('Stopped listening');
            document.getElementById('voiceBtn').classList.remove('active');
        };
        
        recognition.start();
    }

    speakOutput(text) {
        if (!this.settings.voiceOutput || !('speechSynthesis' in window)) {
            return;
        }
        
        const utterance = new SpeechSynthesisUtterance(text);
        utterance.rate = 1.0;
        utterance.pitch = 1.0;
        
        speechSynthesis.speak(utterance);
    }

    // ==================== RENDERING ====================
    
    renderMarkdown(text) {
        // Configure marked
        marked.setOptions({
            breaks: true,
            gfm: true,
        });
        
        // Render markdown
        let html = marked.parse(text);
        
        // Highlight code blocks
        html = html.replace(/<code class="language-(\w+)">([^<]+)<\/code>/g, (match, lang, code) => {
            try {
                const highlighted = hljs.highlight(code, { language: lang }).value;
                return `<code class="hljs language-${lang}">${highlighted}</code>`;
            } catch {
                return match;
            }
        });
        
        return html;
    }

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // ==================== CONVERSATIONS ====================
    
    newConversation() {
        this.currentMessageId = null;
        document.getElementById('messagesContainer').innerHTML = '';
        document.getElementById('sessionTitle').textContent = 'New Conversation';
    }

    loadSessions() {
        // Load from localStorage (or could load from BadgerDB via API)
        const sessionsData = localStorage.getItem('sovereignSessions');
        if (sessionsData) {
            try {
                const sessions = JSON.parse(sessionsData);
                this.displaySessions(sessions);
            } catch (e) {
                console.error('Failed to load sessions:', e);
            }
        }
    }

    displaySessions(sessions) {
        const list = document.getElementById('sessionsList');
        list.innerHTML = '';
        
        sessions.forEach(session => {
            const item = document.createElement('div');
            item.className = 'session-item';
            item.innerHTML = `
                <div class="session-title">${session.title}</div>
                <div class="session-time">${new Date(session.timestamp).toLocaleDateString()}</div>
            `;
            item.addEventListener('click', () => this.loadSession(session));
            list.appendChild(item);
        });
    }

    loadSession(session) {
        // Load conversation from storage
        const conv = this.conversations.get(session.id);
        if (conv) {
            document.getElementById('messagesContainer').innerHTML = '';
            document.getElementById('sessionTitle').textContent = session.title;
            conv.messages.forEach(msg => {
                this.displayMessage(msg.role, msg.content);
            });
        }
    }

    // ==================== SETTINGS ====================
    
    showSettings() {
        document.getElementById('settingsModal').style.display = 'block';
        document.getElementById('temperatureSlider').value = this.settings.temperature;
        document.getElementById('enableVoiceOutput').checked = this.settings.voiceOutput;
        document.getElementById('enableStreaming').checked = this.settings.streaming;
    }

    hideSettings() {
        document.getElementById('settingsModal').style.display = 'none';
    }

    loadSettings() {
        const saved = localStorage.getItem('sovereignSettings');
        return saved ? JSON.parse(saved) : {
            temperature: 0.7,
            model: 'local',
            voiceOutput: false,
            streaming: true
        };
    }

    saveSettings() {
        localStorage.setItem('sovereignSettings', JSON.stringify(this.settings));
    }
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    new SovereignChat();
});
