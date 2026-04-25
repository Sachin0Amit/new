/**
 * Sovereign Window Manager
 * Logic for managing draggable, resizable, and prioritizable operative windows.
 */
export default class WindowManager {
    constructor() {
        this.windows = [];
        this.zIndex = 1000;
        this.desktop = document.getElementById('desktop-workspace');
    }

    createWindow(id, title, content, options = {}) {
        const win = document.createElement('div');
        win.className = 'sov-window glass-panel';
        win.id = id;
        win.style.width = options.width || '400px';
        win.style.height = options.height || '300px';
        win.style.left = options.left || '100px';
        win.style.top = options.top || '100px';
        win.style.zIndex = this.zIndex++;

        win.innerHTML = `
            <div class="window-header">
                <span class="window-title">${title}</span>
                <div class="window-controls">
                    <button class="win-btn minimize"><i data-lucide="minus"></i></button>
                    <button class="win-btn close"><i data-lucide="x"></i></button>
                </div>
            </div>
            <div class="window-content">${content}</div>
            <div class="resize-handle"></div>
        `;

        this.desktop.appendChild(win);
        this.makeDraggable(win);
        this.makeResizable(win);
        
        // --- MOTION: WINDOW ENTRANCE ---
        if (window.animateMotion) {
            window.animateMotion(win, 
                { opacity: [0, 1], scale: [0.9, 1], filter: ["blur(10px)", "blur(0px)"] },
                { duration: 0.6, easing: window.springMotion({ stiffness: 200, damping: 20 }) }
            );
        }

        win.addEventListener('mousedown', () => this.focus(win));
        
        win.querySelector('.close').addEventListener('click', () => {
            if (window.animateMotion) {
                window.animateMotion(win, 
                    { opacity: 0, scale: 0.9, filter: "blur(10px)" },
                    { duration: 0.3 }
                ).finished.then(() => win.remove());
            } else {
                win.remove();
            }
        });

        // Initialize Lucide icons
        if (window.lucide) window.lucide.createIcons();

        return win;
    }

    focus(win) {
        win.style.zIndex = this.zIndex++;
        document.querySelectorAll('.sov-window').forEach(w => w.classList.remove('active'));
        win.classList.add('active');
    }

    makeDraggable(win) {
        const header = win.querySelector('.window-header');
        let pos1 = 0, pos2 = 0, pos3 = 0, pos4 = 0;

        header.onmousedown = (e) => {
            e.preventDefault();
            pos3 = e.clientX;
            pos4 = e.clientY;
            document.onmouseup = () => {
                document.onmouseup = null;
                document.onmousemove = null;
            };
            document.onmousemove = (e) => {
                pos1 = pos3 - e.clientX;
                pos2 = pos4 - e.clientY;
                pos3 = e.clientX;
                pos4 = e.clientY;
                win.style.top = (win.offsetTop - pos2) + "px";
                win.style.left = (win.offsetLeft - pos1) + "px";
            };
        };
    }

    makeResizable(win) {
        const handle = win.querySelector('.resize-handle');
        handle.onmousedown = (e) => {
            e.preventDefault();
            const startX = e.clientX;
            const startY = e.clientY;
            const startWidth = parseInt(win.style.width);
            const startHeight = parseInt(win.style.height);

            document.onmouseup = () => {
                document.onmouseup = null;
                document.onmousemove = null;
            };

            document.onmousemove = (e) => {
                const width = startWidth + (e.clientX - startX);
                const height = startHeight + (e.clientY - startY);
                win.style.width = width + "px";
                win.style.height = height + "px";
            };
        };
    }
}
