/**
 * Sovereign Login System
 * Masterful session management & UI
 */

class SovereignLogin {
    constructor() {
        this.isOpen = false;
        this.init();
    }

    init() {
        this.injectHTML();
        this.attachListeners();
    }

    injectHTML() {
        const overlay = document.createElement('div');
        overlay.className = 'login-overlay';
        overlay.id = 'sovereign-login-overlay';
        overlay.innerHTML = `
            <div class="login-card">
                <button class="close-btn" id="close-login">✕</button>
                <h2>Sovereign Identity</h2>
                <p>Private authentication for the local Chimera Core. <br>Your keys never leave this hardware.</p>
                
                <form class="login-form" id="login-form">
                    <div class="input-group">
                        <label>Sovereign ID</label>
                        <input type="text" class="login-input" placeholder="Enter your identity..." required>
                    </div>
                    <div class="input-group">
                        <label>Decryption Key</label>
                        <input type="password" class="login-input" placeholder="••••••••" required>
                    </div>
                    <button type="submit" class="login-submit">Authorize Session</button>
                </form>

                <div class="login-footer">
                    New to Sovereign? <a href="#">Generate Identity</a>
                </div>
            </div>
        `;
        document.body.appendChild(overlay);
        this.overlay = overlay;
    }

    attachListeners() {
        document.getElementById('close-login').addEventListener('click', () => this.toggle(false));
        
        const form = document.getElementById('login-form');
        form.addEventListener('submit', (e) => {
            e.preventDefault();
            this.handleLogin();
        });

        // Close on background click
        this.overlay.addEventListener('click', (e) => {
            if (e.target === this.overlay) this.toggle(false);
        });

        // Escape key to close
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Escape' && this.isOpen) this.toggle(false);
        });
    }

    handleLogin() {
        const btn = this.overlay.querySelector('.login-submit');
        const originalText = btn.innerText;
        
        btn.innerText = 'Decrypting...';
        btn.disabled = true;

        setTimeout(() => {
            btn.innerText = 'Authorized';
            btn.style.background = '#B4BFA1'; // Sage accent
            
            setTimeout(() => {
                this.updateUserUI();
                this.toggle(false);
                
                // Reset for next time
                btn.innerText = originalText;
                btn.style.background = '';
                btn.disabled = false;
            }, 800);
        }, 1500);
    }

    updateUserUI() {
        // Update user cards across the application
        const userNameFields = document.querySelectorAll('.user-name');
        const userStatusFields = document.querySelectorAll('.user-status');
        const avatars = document.querySelectorAll('.avatar');

        userNameFields.forEach(f => f.innerText = 'Sovereign Authorized');
        userStatusFields.forEach(f => {
            if (f.innerText.includes('Local Engine')) {
                f.innerText = 'Encrypted Session Active';
            }
        });
        avatars.forEach(a => {
            a.innerText = 'A';
            a.style.background = '#B4BFA1';
        });
    }

    toggle(force) {
        this.isOpen = (force !== undefined) ? force : !this.isOpen;
        if (this.isOpen) {
            this.overlay.classList.add('active');
        } else {
            this.overlay.classList.remove('active');
        }
    }
}

const Login = new SovereignLogin();
window.toggleSovereignLogin = () => Login.toggle();
