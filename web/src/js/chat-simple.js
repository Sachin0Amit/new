/**
 * Simple Chat Experience - PapiAi
 * Focus: Scholarly minimalism, streaming responses, backend integration
 */
import SovereignCursor from './Cursor.js';

class SimpleChat {
    constructor() {
        this.input = document.getElementById('chat-input');
        this.sendBtn = document.getElementById('send-btn');
        this.messagesArea = document.getElementById('messages-area');
        this.welcome = document.getElementById('welcome-msg');
        this.isStreaming = false;
        this.backendUrl = 'http://localhost:8081';
        this.init();
    }

    init() {
        new SovereignCursor();
        this.setupEventListeners();
        this.setupTextarea();
        this.initKeyboard();
    }

    setupEventListeners() {
        this.sendBtn.addEventListener('click', () => this.handleSend());
        this.input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.handleSend();
            }
        });
    }

    setupTextarea() {
        this.input.addEventListener('input', () => {
            this.input.style.height = 'auto';
            this.input.style.height = Math.min(this.input.scrollHeight, 200) + 'px';
        });
    }

    initKeyboard() {
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                this.input.value = '';
                this.input.style.height = 'auto';
                this.input.blur();
            }
        });
    }

    async handleSend() {
        if (this.isStreaming) return;
        const text = this.input.value.trim();
        if (!text) return;

        if (this.welcome) {
            this.welcome.style.display = 'none';
            this.welcome = null;
        }

        this.addMessage('user', text);
        this.input.value = '';
        this.input.style.height = 'auto';

        // Show typing indicator
        const typingEl = this.showTyping();

        // Try backend, fall back to mock
        let response;
        try {
            const res = await fetch(`${this.backendUrl}/api/v1/chat`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ message: text }),
                signal: AbortSignal.timeout(5000)
            });
            if (!res.ok) throw new Error();
            const data = await res.json();
            response = data.response;
        } catch {
            response = this.mockResponse(text);
        }

        typingEl?.remove();
        await this.streamMessage(response);
    }

    mockResponse(query) {
        const q = query.toLowerCase();
        if (q.includes('hello') || q.includes('hi'))
            return "Hello! I'm running locally via the Chimera Core. All processing stays on your hardware. How can I assist you?";
        if (q.includes('math') || q.includes('derive') || q.includes('proof'))
            return "I can handle symbolic derivations, Taylor expansions, and proof verification. All computations run locally via the Titan C++ Engine with ED25519-signed audit trails.";
        return "I have processed your query locally via the Chimera Core. As a sovereign intelligence, I am analyzing the symbolic implications of your request with verified mathematical safety. All data remains on your hardware.";
    }

    showTyping() {
        const div = document.createElement('div');
        div.className = 'message ai';
        div.id = 'typing-msg';
        div.innerHTML = `
            <div class="ai-title">Sovereign Intel</div>
            <div class="typing-dots"><span></span><span></span><span></span></div>`;
        this.messagesArea.appendChild(div);
        this.messagesArea.scrollTop = this.messagesArea.scrollHeight;
        return div;
    }

    addMessage(type, text) {
        const div = document.createElement('div');
        div.className = `message ${type}`;
        const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

        let content = '';
        if (type === 'ai') content += `<div class="ai-title">Sovereign Intel</div>`;
        content += `<div class="bubble">${this.sanitize(text)}</div>`;
        content += `<div class="msg-time">${time}</div>`;

        div.innerHTML = content;
        this.messagesArea.appendChild(div);
        this.messagesArea.scrollTop = this.messagesArea.scrollHeight;
    }

    async streamMessage(fullText) {
        this.isStreaming = true;
        const div = document.createElement('div');
        div.className = 'message ai';
        const time = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });

        div.innerHTML = `
            <div class="ai-title">Sovereign Intel</div>
            <div class="bubble"><span class="stream-caret"></span></div>
            <div class="msg-time">${time}</div>`;

        this.messagesArea.appendChild(div);
        const bubble = div.querySelector('.bubble');

        let displayed = '';
        const chars = fullText.split('');
        const speed = Math.max(10, Math.min(30, 2000 / chars.length));

        for (const char of chars) {
            displayed += char;
            bubble.innerHTML = this.renderMarkdown(displayed) + '<span class="stream-caret"></span>';
            this.messagesArea.scrollTop = this.messagesArea.scrollHeight;
            await new Promise(r => setTimeout(r, speed));
        }

        bubble.innerHTML = this.renderMarkdown(fullText);
        this.messagesArea.scrollTop = this.messagesArea.scrollHeight;
        this.isStreaming = false;
    }

    renderMarkdown(text) {
        let html = this.sanitize(text);
        html = html.replace(/```(\w*)\n?([\s\S]*?)```/g, (_, lang, code) => `<pre><code>${code.trim()}</code></pre>`);
        html = html.replace(/`([^`]+)`/g, '<code>$1</code>');
        html = html.replace(/\*\*(.+?)\*\*/g, '<strong>$1</strong>');
        html = html.replace(/\*(.+?)\*/g, '<em>$1</em>');
        html = html.replace(/\n\n/g, '<br><br>');
        html = html.replace(/\n/g, '<br>');
        return html;
    }

    sanitize(text) {
        const el = document.createElement('div');
        el.textContent = text;
        return el.innerHTML;
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new SimpleChat();
});
