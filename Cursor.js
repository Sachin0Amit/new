/**
 * Sovereign Spectral Cursor
 * High-performance cursor tracking with GSAP QuickSetter and reactive states.
 */
export default class SovereignCursor {
    constructor() {
        this.createElements();
        this.init();
        this.addEventListeners();
    }

    createElements() {
        this.cursor = document.createElement('div');
        this.cursor.className = 'sov-cursor';
        this.cursor.innerHTML = `
            <div class="sov-cursor--outer"></div>
            <div class="sov-cursor--inner"></div>
        `;
        document.body.appendChild(this.cursor);
        
        this.inner = this.cursor.querySelector('.sov-cursor--inner');
        this.outer = this.cursor.querySelector('.sov-cursor--outer');
    }

    init() {
        // Use GSAP QuickSetter for ultra-fast performance
        this.xSetterInner = gsap.quickSetter(this.inner, "left", "px");
        this.ySetterInner = gsap.quickSetter(this.inner, "top", "px");
        
        this.xSetterOuter = gsap.quickSetter(this.outer, "left", "px");
        this.ySetterOuter = gsap.quickSetter(this.outer, "top", "px");
        
        this.mouse = { x: 0, y: 0 };
        this.pos = { x: 0, y: 0 };
        this.ratio = 0.15; // Smoothness factor for the outer follower
    }

    addEventListeners() {
        window.addEventListener("mousemove", (e) => {
            this.mouse.x = e.clientX;
            this.mouse.y = e.clientY;
            
            this.xSetterInner(this.mouse.x);
            this.ySetterInner(this.mouse.y);
        });

        // Loop for the outer follower smoothing
        gsap.ticker.add(() => {
            this.pos.x += (this.mouse.x - this.pos.x) * this.ratio;
            this.pos.y += (this.mouse.y - this.pos.y) * this.ratio;
            
            this.xSetterOuter(this.pos.x);
            this.ySetterOuter(this.pos.y);
        });

        // Interaction States
        window.addEventListener("mousedown", () => this.cursor.classList.add('is-clicking'));
        window.addEventListener("mouseup", () => this.cursor.classList.remove('is-clicking'));

        // Handle hover on interactive elements
        const hoverTargets = 'a, button, .taskbar-item, .win-btn, .new-chat-btn, .btn';
        
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
