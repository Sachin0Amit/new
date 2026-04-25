import Engine from './Engine.js';
import Tower from './Tower.js';

document.addEventListener('DOMContentLoaded', () => {
    // 1. Initialize Icons
    lucide.createIcons();

    // 2. Initialize 3D Engine
    const engine = new Engine();
    const tower = new Tower();
    engine.add(tower.getMesh());

    // 3. UI Animations (Non-3D)
    gsap.registerPlugin(ScrollTrigger);

    // Fade in sections as we scroll
    document.querySelectorAll('.feature-section').forEach(section => {
        gsap.from(section.querySelector('.content-wrapper'), {
            opacity: 0,
            y: 50,
            duration: 1.5,
            scrollTrigger: {
                trigger: section,
                start: "top 80%",
                end: "top 40%",
                scrub: true
            }
        });
    });

    // Header scroll background
    const sidebar = document.querySelector('.sidebar');
    window.addEventListener('scroll', () => {
        if (window.scrollY > 100) {
            sidebar.style.background = '#F5F2ED'; // Solid on scroll
        } else {
            sidebar.style.background = '#E8E4DB'; // Default
        }
    });

    // 4. Interactive Elements
    const chatBtn = document.querySelector('.new-chat-btn');
    chatBtn.addEventListener('click', () => {
        alert("Initializing new Sovereign session...");
    });

    // 5. Hero parallax for tower (Optional refinement)
    window.addEventListener('mousemove', (e) => {
        const x = (e.clientX / window.innerWidth - 0.5) * 2;
        const y = (e.clientY / window.innerHeight - 0.5) * 2;
        
        gsap.to(tower.getMesh().position, {
            x: x * 0.5,
            y: -y * 0.5,
            duration: 1,
            ease: "power2.out"
        });
    });
});
