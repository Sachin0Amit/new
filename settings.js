/**
 * Sovereign Intelligence: Unified Settings System
 * Logic & State Management
 */

class SovereignSettings {
    constructor() {
        this.defaults = {
            hardware: {
                device: 'auto',
                threads: 4
            },
            visuals: {
                palette: 'spectral', // spectral, gray, obsidian
                bloomScale: 1.5,
                visualizerActive: true,
                orbitSpeed: 1.0
            },
            storage: {
                path: './data/sovereign',
                autoClear: false
            }
        };

        this.state = this.loadState();
        this.isOpen = false;
        this.init();
    }

    init() {
        this.injectHTML();
        this.applyState();
        this.attachListeners();
    }

    loadState() {
        const saved = localStorage.getItem('sovereign_settings');
        return saved ? JSON.parse(saved) : JSON.parse(JSON.stringify(this.defaults));
    }

    saveState() {
        localStorage.setItem('sovereign_settings', JSON.stringify(this.state));
        this.applyState();
        
        // Dispatch global event for other components (e.g. Tower.js)
        window.dispatchEvent(new CustomEvent('sovereign_settings_updated', { detail: this.state }));
    }

    applyState() {
        // Apply visual settings to CSS variables
        const root = document.documentElement;
        if (this.state.visuals.palette === 'spectral') {
            root.style.setProperty('--accent-glow', 'rgba(138, 43, 226, 0.4)');
        } else if (this.state.visuals.palette === 'gray') {
            root.style.setProperty('--accent-glow', 'rgba(200, 200, 200, 0.2)');
        } else {
            root.style.setProperty('--accent-glow', 'rgba(0, 0, 0, 0.8)');
        }

        // Handle visualizer visibility immediately
        const vis = document.getElementById('pro-visualizer') || document.getElementById('canvas-container');
        if (vis) {
            vis.style.opacity = this.state.visuals.visualizerActive ? '1' : '0.1';
            vis.style.pointerEvents = this.state.visuals.visualizerActive ? 'auto' : 'none';
        }
    }

    injectHTML() {
        const overlay = document.createElement('div');
        overlay.className = 'settings-overlay';
        overlay.id = 'sovereign-settings-system';
        
        overlay.innerHTML = `
            <div class="settings-panel">
                <div class="settings-header">
                    <h2>System Configuration</h2>
                    <button class="close-settings" id="close-sovereign-settings">✕</button>
                </div>
                <div class="settings-body">
                    <div class="settings-section">
                        <h3>Cognitive Core (Hardware)</h3>
                        <div class="setting-row">
                            <div class="setting-info">
                                <span class="setting-label">Inference Device</span>
                                <span class="setting-desc">Primary hardware for local derivation</span>
                            </div>
                            <select class="sovereign-select" id="set-hw-device">
                                <option value="auto" ${this.state.hardware.device === 'auto' ? 'selected' : ''}>Auto Optimize</option>
                                <option value="cpu" ${this.state.hardware.device === 'cpu' ? 'selected' : ''}>Standard CPU</option>
                                <option value="gpu" ${this.state.hardware.device === 'gpu' ? 'selected' : ''}>Chimera GPU</option>
                                <option value="vulkan" ${this.state.hardware.device === 'vulkan' ? 'selected' : ''}>Vulkan Mesh</option>
                            </select>
                        </div>
                        <div class="setting-row">
                            <div class="setting-info">
                                <span class="setting-label">Parallel Threads</span>
                                <span class="setting-desc">Neural thread allocation (${this.state.hardware.threads})</span>
                            </div>
                            <input type="range" class="sovereign-slider" id="set-hw-threads" min="1" max="16" value="${this.state.hardware.threads}">
                        </div>
                    </div>

                    <div class="settings-section">
                        <h3>Sensory Array (Visuals)</h3>
                        <div class="setting-row">
                            <div class="setting-info">
                                <span class="setting-label">Orbit Palette</span>
                                <span class="setting-desc">Theme alignment for the Mobius core</span>
                            </div>
                            <select class="sovereign-select" id="set-vis-palette">
                                <option value="spectral" ${this.state.visuals.palette === 'spectral' ? 'selected' : ''}>Spectral Purple</option>
                                <option value="gray" ${this.state.visuals.palette === 'gray' ? 'selected' : ''}>Obsidian Gray</option>
                                <option value="obsidian" ${this.state.visuals.palette === 'obsidian' ? 'selected' : ''}>Absolute Zero</option>
                            </select>
                        </div>
                        <div class="setting-row">
                            <div class="setting-info">
                                <span class="setting-label">Visualizer Status</span>
                                <span class="setting-desc">Toggle 3D immersion layer</span>
                            </div>
                            <div class="sovereign-toggle ${this.state.visuals.visualizerActive ? 'active' : ''}" id="toggle-vis-active"></div>
                        </div>
                        <div class="setting-row">
                            <div class="setting-info">
                                <span class="setting-label">Bloom Intensity</span>
                                <span class="setting-desc">Glow scale for biomechanical nodes</span>
                            </div>
                            <input type="range" class="sovereign-slider" id="set-vis-bloom" min="0" max="3" step="0.1" value="${this.state.visuals.bloomScale}">
                        </div>
                    </div>

                    <div class="settings-section">
                        <h3>Sovereign Vault (Storage)</h3>
                        <div class="setting-row">
                            <div class="setting-info">
                                <span class="setting-label">Persistence Root</span>
                                <span class="setting-desc">Local path for the LSM-tree storage</span>
                            </div>
                            <input type="text" class="sovereign-select" id="set-store-path" value="${this.state.storage.path}" style="width: 200px; font-size: 0.8rem;">
                        </div>
                    </div>
                </div>
                <div class="settings-footer">
                    <button class="btn-save" id="save-sovereign-settings">Commit Changes</button>
                </div>
            </div>
        `;

        document.body.appendChild(overlay);
        this.overlay = overlay;
    }

    attachListeners() {
        document.getElementById('close-sovereign-settings').addEventListener('click', () => this.toggle(false));
        document.getElementById('save-sovereign-settings').addEventListener('click', () => {
            this.commitChanges();
            this.toggle(false);
        });

        // Toggle handler
        document.getElementById('toggle-vis-active').addEventListener('click', (e) => {
            e.currentTarget.classList.toggle('active');
            this.state.visuals.visualizerActive = e.currentTarget.classList.contains('active');
        });

        // Slider real-time feedback
        document.getElementById('set-hw-threads').addEventListener('input', (e) => {
            e.target.previousElementSibling.querySelector('.setting-desc').innerText = `Neural thread allocation (${e.target.value})`;
        });

        // Close on background click
        this.overlay.addEventListener('click', (e) => {
            if (e.target === this.overlay) this.toggle(false);
        });
    }

    commitChanges() {
        this.state.hardware.device = document.getElementById('set-hw-device').value;
        this.state.hardware.threads = parseInt(document.getElementById('set-hw-threads').value);
        this.state.visuals.palette = document.getElementById('set-vis-palette').value;
        this.state.visuals.bloomScale = parseFloat(document.getElementById('set-vis-bloom').value);
        this.state.storage.path = document.getElementById('set-store-path').value;
        
        this.saveState();
    }

    toggle(force) {
        this.isOpen = (force !== undefined) ? force : !this.isOpen;
        if (this.isOpen) {
            this.overlay.classList.add('active');
            document.body.style.overflow = 'hidden';
        } else {
            this.overlay.classList.remove('active');
            document.body.style.overflow = '';
        }
    }
}

// Global Export
export const Settings = new SovereignSettings();
window.toggleSovereignSettings = () => Settings.toggle();
