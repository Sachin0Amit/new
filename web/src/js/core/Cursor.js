/**
 * Sovereign Spectral Cursor
 * High-performance cursor tracking with GSAP QuickSetter, reactive states, and trail effect.
 * Auto-disables on touch devices.
 */
export default class SovereignCursor {
    constructor() {
        // Don't initialize on touch devices
        if (this.isTouchDevice()) return;

        this.createElements();
        this.init();
        this.addEventListeners();
    }

    isTouchDevice() {
        return (
            'ontouchstart' in window ||
            navigator.maxTouchPoints > 0 ||
            window.matchMedia('(pointer: coarse)').matches
        );
    }

    createElements() {
        this.cursor = document.createElement('div');
        this.cursor.className = 'sov-cursor';
        this.cursor.innerHTML = `
            <div class="sov-cursor--trail"></div>
            <div class="sov-cursor--outer"></div>
            <div class="sov-cursor--inner"></div>
        `;
        document.body.appendChild(this.cursor);

        this.inner = this.cursor.querySelector('.sov-cursor--inner');
        this.outer = this.cursor.querySelector('.sov-cursor--outer');
        this.trail = this.cursor.querySelector('.sov-cursor--trail');
    }

    init() {
        // Use GSAP QuickSetter for ultra-fast performance
        this.xSetterInner = gsap.quickSetter(this.inner, "left", "px");
        this.ySetterInner = gsap.quickSetter(this.inner, "top", "px");

        this.xSetterOuter = gsap.quickSetter(this.outer, "left", "px");
        this.ySetterOuter = gsap.quickSetter(this.outer, "top", "px");

        this.xSetterTrail = gsap.quickSetter(this.trail, "left", "px");
        this.ySetterTrail = gsap.quickSetter(this.trail, "top", "px");

        this.mouse = { x: 0, y: 0 };
        this.pos = { x: 0, y: 0 };
        this.trailPos = { x: 0, y: 0 };
        this.ratio = 0.15;
        this.trailRatio = 0.08;
    }

    addEventListeners() {
        window.addEventListener("mousemove", (e) => {
            this.mouse.x = e.clientX;
            this.mouse.y = e.clientY;

            this.xSetterInner(this.mouse.x);
            this.ySetterInner(this.mouse.y);
        });

        // Smooth outer follower + trail
        gsap.ticker.add(() => {
            this.pos.x += (this.mouse.x - this.pos.x) * this.ratio;
            this.pos.y += (this.mouse.y - this.pos.y) * this.ratio;
            this.xSetterOuter(this.pos.x);
            this.ySetterOuter(this.pos.y);

            this.trailPos.x += (this.mouse.x - this.trailPos.x) * this.trailRatio;
            this.trailPos.y += (this.mouse.y - this.trailPos.y) * this.trailRatio;
            this.xSetterTrail(this.trailPos.x);
            this.ySetterTrail(this.trailPos.y);
        });

        // Interaction States
        window.addEventListener("mousedown", () => this.cursor.classList.add('is-clicking'));
        window.addEventListener("mouseup", () => this.cursor.classList.remove('is-clicking'));

        // Handle hover on interactive elements
        const hoverTargets = 'a, button, .taskbar-item, .win-btn, .new-chat-btn, .btn, .feature-card, .nav-item, .history-item, .quick-prompt, .arch-node';

        document.body.addEventListener('mouseover', (e) => {
            if (e.target.closest(hoverTargets)) {
                this.cursor.classList.add('is-hovering');
            }
        });

        document.body.addEventListener('mouseout', (e) => {
            if (e.target.closest(hoverTargets)) {
                this.cursor.classList.remove('is-hovering');
            }
        });

        // Typing / Caret States
        const inputTargets = 'input, textarea, [contenteditable="true"]';

        document.addEventListener('focusin', (e) => {
            if (e.target.matches(inputTargets)) {
                this.cursor.classList.add('is-typing');
            }
        });

        document.addEventListener('focusout', (e) => {
            if (e.target.matches(inputTargets)) {
                this.cursor.classList.remove('is-typing');
            }
        });
    }
}
