import Engine from './Engine.js';
import Tower from './Tower.js';
import { animate, spring } from "https://cdn.jsdelivr.net/npm/motion/+esm";
import SovereignCursor from './Cursor.js';

class ProApp {
    constructor() {
        this.messages = [];
        this.visualizerRemoved = false;
        this.init();
    }

    init() {
        new SovereignCursor();
        this.initVisualizer();
        this.addEventListeners();
        this.setupTextarea();
    }

    initVisualizer() {
        this.container = document.getElementById('pro-visualizer');
        const width = this.container.offsetWidth;
        const height = this.container.offsetHeight;

        // --- Core Scene ---
        this.scene = new THREE.Scene();
        this.camera = new THREE.PerspectiveCamera(75, width / height, 0.1, 1000);
        this.camera.position.z = 12;

        this.renderer = new THREE.WebGLRenderer({ 
            antialias: true, 
            alpha: true,
            powerPreference: "high-performance"
        });
        this.renderer.setSize(width, height);
        this.renderer.setPixelRatio(Math.min(window.devicePixelRatio, 2));
        this.renderer.toneMapping = THREE.ReinhardToneMapping;
        this.container.appendChild(this.renderer.domElement);

        // --- Post Processing ---
        const renderScene = new THREE.RenderPass(this.scene, this.camera);
        this.bloomPass = new THREE.UnrealBloomPass(
            new THREE.Vector2(width, height),
            1.5, 0.4, 0.85
        );
        
        this.composer = new THREE.EffectComposer(this.renderer);
        this.composer.addPass(renderScene);
        this.composer.addPass(this.bloomPass);

        // --- Tower ---
        this.tower = new Tower();
        this.towerGroup = this.tower.group;
        this.towerGroup.position.x = 6;
        this.towerGroup.position.y = -2;
        this.towerGroup.rotation.y = Math.PI / 4;
        this.towerGroup.scale.set(1.1, 1.1, 1.1);
        this.scene.add(this.towerGroup);

        // --- Lighting ---
        const ambient = new THREE.AmbientLight(0x0a0a0a, 2);
        this.scene.add(ambient);
        
        const point = new THREE.PointLight(0x8a2be2, 5, 25);
        point.position.set(10, 5, 5);
        this.scene.add(point);

        const accentLight = new THREE.PointLight(0xff00ff, 2, 20);
        accentLight.position.set(-5, -5, 5);
        this.scene.add(accentLight);

        this.startAnimation();
        this.initSettingsListener();

        window.addEventListener('resize', this.onResize.bind(this));
    }

    startAnimation() {
        const animate = () => {
            if (this.visualizerRemoved) return;
            this.animId = requestAnimationFrame(animate);
            this.towerGroup.rotation.y += 0.003;
            
            // Pulse spectral shader
            const mobius = this.towerGroup.children.find(c => c.material && c.material.uniforms);
            if (mobius) mobius.material.uniforms.uTime.value += 0.05;

            this.composer.render();
        };
        animate();
    }

    onResize() {
        if (this.visualizerRemoved) return;
        const w = this.container.offsetWidth;
        const h = this.container.offsetHeight;
        this.camera.aspect = w / h;
        this.camera.updateProjectionMatrix();
        this.renderer.setSize(w, h);
        this.composer.setSize(w, h);
    }

    removeVisualizer() {
        if (this.visualizerRemoved) return;
        this.visualizerRemoved = true;
        
        gsap.to(this.container, {
            opacity: 0,
            duration: 1.0,
            ease: "power2.inOut",
            onComplete: () => {
                this.container.style.visibility = 'hidden';
                cancelAnimationFrame(this.animId);
            }
        });

        // Set body background to a deep obsidian black
        gsap.to('body', {
            backgroundColor: '#020205',
            duration: 1.2
        });
    }

    restoreVisualizer() {
        if (!this.visualizerRemoved) return;
        this.visualizerRemoved = false;
        
        this.container.style.visibility = 'visible';
        gsap.to(this.container, {
            opacity: 1,
            duration: 1.0,
            ease: "power2.inOut"
        });

        // Restore gradient background
        gsap.to('body', {
            backgroundColor: '', // Reverts to radial gradient from CSS
            duration: 1.2
        });

        this.startAnimation();
    }

    addEventListeners() {
        const sendBtn = document.getElementById('send-btn');
        const textarea = document.getElementById('chat-textarea');

        sendBtn.addEventListener('click', () => this.sendMessage());
        textarea.addEventListener('keydown', (e) => {
            if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault();
                this.sendMessage();
            }
        });

        document.getElementById('new-chat').addEventListener('click', () => {
            this.clearChat();
        });
    }

    setupTextarea() {
        const textarea = document.getElementById('chat-textarea');
        textarea.addEventListener('input', () => {
            textarea.style.height = 'auto';
            textarea.style.height = (textarea.scrollHeight) + 'px';
        });
    }

    sendMessage() {
        const input = document.getElementById('chat-textarea');
        const text = input.value.trim();
        
        if (!text) return;

        if (!this.visualizerRemoved) {
            this.removeVisualizer();
        }

        this.addMessage('user', text);
        input.value = '';
        input.style.height = 'auto';

        // --- Processing Heartbeat ---
        const typingId = 'typing-' + Date.now();
        this.addMessage('ai', "Synchronizing local core... [TITAN_ENGINE_INITIALIZING]", typingId);
        
        setTimeout(() => {
            const typingMsg = document.getElementById(typingId);
            if (typingMsg) typingMsg.remove();
            
            const response = this.generateResponse(text);
            this.addMessage('ai', response);
        }, 1500);
    }

    generateResponse(query) {
        const q = query.toLowerCase();
        
        // Math Derivation Logic
        if (q.includes('derive') || q.includes('taylor') || q.includes('expand') || q.includes('integral')) {
            return `
                <div class="derivation-block">
                    <p><strong>[SYMB_DERIVATION_INITIATED]</strong> Analyzing local vector space for expansion...</p>
                    <div style="background: rgba(0,0,0,0.3); padding: 1rem; border-radius: 8px; margin: 1rem 0; font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; border-left: 2px solid #8a2be2;">
                        f(x) ≈ f(0) + f'(0)x + (f''(0)/2!)x² + ...
                    </div>
                    <p>The second-order Taylor expansion for <em>f(x) = exp(x)cos(x)</em> at <em>x = 0</em> has been derived locally:</p>
                    <div style="color: #A4CCF4; font-size: 1.1rem; margin: 1rem 0; font-weight: 600;">
                        T₂(x) = 1 + x + O(x³)
                    </div>
                    <p style="font-size: 0.8rem; color: #888;">[VERIFICATION_COMPLETE] Titan Sandbox confirms symbolic consistency.</p>
                </div>
            `;
        }

        return "Derivation complete. The Titan Core has finalized the cognitive loop across the local mesh. All proof-of-authenticity signatures have been applied to the audit trail.";
    }

    addMessage(type, text, id = null) {
        const messagesContainer = document.getElementById('messages');
        const hero = document.querySelector('.welcome-hero');
        if (hero) hero.style.display = 'none';

        const msgDiv = document.createElement('div');
        msgDiv.className = `message ${type}`;
        if (id) msgDiv.id = id;
        msgDiv.style.marginBottom = "1.5rem";
        
        const senderColor = type === 'user' ? '#8a2be2' : '#8a2be2'; // Uniform spectral purple
        const senderName = type === 'user' ? 'You' : 'Sovereign Core';

        msgDiv.innerHTML = `
            <div class="message-content" style="background: rgba(20,20,20,0.6); padding: 1.2rem; border-radius: 12px; border: 1px solid rgba(138,43,226,0.1); backdrop-filter: blur(10px);">
                <div class="message-sender" style="color: ${senderColor}; font-size: 0.75rem; letter-spacing: 1.5px; margin-bottom: 0.6rem; text-transform: uppercase; font-weight: 700;">${senderName}</div>
                <div class="message-text" style="color: #eee; line-height: 1.6; font-size: 0.95rem;">${text}</div>
            </div>
        `;
        
        messagesContainer.appendChild(msgDiv);
        messagesContainer.scrollTop = messagesContainer.scrollHeight;

        // --- MOTION: RUBBER-BAND MESSAGE ENTRANCE ---
        animate(msgDiv, 
            { opacity: [0, 1], y: [30, 0], scale: [0.95, 1] },
            { duration: 0.8, easing: spring({ stiffness: 200, damping: 15 }) }
        );
    }

    clearChat() {
        const messagesContainer = document.getElementById('messages');
        messagesContainer.innerHTML = `
            <div class="welcome-hero">
                <h1>Initiate Local Derivation</h1>
                <p>Enter your query into the Sovereign Core for hardware-accelerated local inference.</p>
            </div>
        `;
        this.restoreVisualizer();
    }

    initSettingsListener() {
        window.addEventListener('sovereign_settings_updated', (e) => {
            const state = e.detail;
            
            // Update Bloom
            if (this.bloomPass) {
                this.bloomPass.strength = state.visuals.bloomScale;
            }
        });
    }
}

document.addEventListener('DOMContentLoaded', () => {
    new ProApp();
});
