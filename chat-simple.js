/**
 * Simple Chat Experience - PapiAi
 * Focus: Scholarly minimalism & high readability
 */

class SimpleChat {
    constructor() {
        this.input = document.getElementById('chat-input');
        this.sendBtn = document.getElementById('send-btn');
        this.messagesArea = document.getElementById('messages-area');
        this.welcome = document.getElementById('welcome-msg');
        
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.setupTextarea();
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

    handleSend() {
        const text = this.input.value.trim();
        if (!text) return;

        // Hide welcome message on first interaction
        if (this.welcome) {
            this.welcome.style.display = 'none';
            this.welcome = null;
        }

        this.addMessage('user', text);
        this.input.value = '';
        this.input.style.height = 'auto';

        // Simulate AI Response
        setTimeout(() => {
            this.addMessage('ai', "I have processed your query locally via the Chimera Core. As a sovereign intelligence, I am analyzing the symbolic implications of your request with verified mathematical safety. [LOCAL_EXECUTION_VERIFIED]");
        }, 800);
    }

    addMessage(type, text) {
        const messageDiv = document.createElement('div');
        messageDiv.className = `message ${type}`;
        
        let content = '';
        if (type === 'ai') {
            content += `<div class="ai-title">Sovereign Intel</div>`;
        }
        content += `<div class="bubble">${text}</div>`;
        
        messageDiv.innerHTML = content;
        this.messagesArea.appendChild(messageDiv);
        this.messagesArea.scrollTop = this.messagesArea.scrollHeight;

        // Animation
        gsap.from(messageDiv, {
            opacity: 0,
            y: 20,
            duration: 0.6,
            ease: "power2.out"
        });
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new SimpleChat();
});
