/**
 * Sovereign Command Center - Minimal Admin Controller
 * Handles tab switching and UI updates for the premium dashboard.
 */

document.addEventListener('DOMContentLoaded', () => {
    // Initial stats
    updateStats();
    setInterval(updateStats, 5000);

    // Initialize Lucide icons
    if (window.lucide) {
        window.lucide.createIcons();
    }
});

function updateStats() {
    const nodes = document.querySelectorAll('.metric-value');
    if (nodes.length >= 2) {
        // Mock some live updates
        const cpuLoad = (20 + Math.random() * 10).toFixed(1) + '%';
        const memLoad = (8 + Math.random() * 0.5).toFixed(1) + 'GB';
        
        nodes[0].innerText = cpuLoad;
        nodes[1].innerText = memLoad;
    }
}

// Global tab switcher
window.showTab = (tab) => {
    console.log('Switching to tab:', tab);
    
    // Hide all sections
    document.querySelectorAll('section.tab-content-root').forEach(el => {
        el.style.display = 'none';
    });

    // Remove active class from all buttons
    document.querySelectorAll('nav .nav-item').forEach(el => {
        el.classList.remove('active');
    });

    // Show target section
    const targetId = tab === 'fleet' ? 'fleet-container' : 
                     tab === 'visualizer' ? 'graph-container' : 
                     tab === 'reflex' ? 'reflex-container' : 'logs-container';
    
    const target = document.getElementById(targetId);
    if (target) {
        target.style.display = 'block';
        // Re-animate cards
        const cards = target.querySelectorAll('.card');
        cards.forEach((card, i) => {
            card.style.animation = 'none';
            card.offsetHeight; // trigger reflow
            card.style.animation = `slideIn 0.4s ease-out forwards ${i * 0.1}s`;
        });
    }

    // Set active button
    const buttons = document.querySelectorAll('nav .nav-item');
    buttons.forEach(btn => {
        if (btn.innerText.toLowerCase().includes(tab.toLowerCase()) || 
            (tab === 'visualizer' && btn.innerText.toLowerCase().includes('dag'))) {
            btn.classList.add('active');
        }
    });
};
