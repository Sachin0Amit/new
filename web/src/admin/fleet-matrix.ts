/**
 * Sovereign Fleet Matrix - Node Resource Visualization
 */

export interface NodeStats {
  nodeId: string;
  cpuLoad: number; // 0-100
  memUsage: number; // 0-100
  lastSeen: number;
}

export class FleetMatrix {
  private container: HTMLElement;
  private nodes: Map<string, NodeStats> = new Map();

  constructor(containerId: string) {
    const el = document.getElementById(containerId);
    if (!el) throw new Error(`Container ${containerId} not found`);
    this.container = el;
    this.container.style.display = 'grid';
    this.container.style.gridTemplateColumns = 'repeat(auto-fill, minmax(150px, 1fr))';
    this.container.style.gap = '20px';
    this.container.style.padding = '20px';
  }

  public updateNode(stats: NodeStats): void {
    this.nodes.set(stats.nodeId, stats);
    this.render();
  }

  private render(): void {
    this.container.innerHTML = '';
    this.nodes.forEach(node => {
      const card = this.createNodeCard(node);
      this.container.appendChild(card);
    });
  }

  private createNodeCard(node: NodeStats): HTMLElement {
    const card = document.createElement('div');
    card.className = 'node-card';
    card.style.textAlign = 'center';
    card.style.padding = '15px';
    card.style.background = '#1a1d21';
    card.style.borderRadius = '12px';
    card.style.cursor = 'pointer';

    const gravity = (node.cpuLoad + node.memUsage) / 2;
    const isCritical = gravity > 80;

    const svg = `
      <svg width="100" height="100" viewBox="0 0 100 100">
        <!-- Outer Ring (CPU) -->
        <circle cx="50" cy="50" r="45" fill="none" stroke="#2c313a" stroke-width="8" />
        <circle cx="50" cy="50" r="45" fill="none" stroke="${isCritical ? '#ff4d4d' : '#4caf50'}" 
                stroke-width="8" stroke-dasharray="${node.cpuLoad * 2.82} 282" 
                transform="rotate(-90 50 50)" style="transition: stroke-dasharray 0.5s ease" />
        
        <!-- Inner Ring (MEM) -->
        <circle cx="50" cy="50" r="35" fill="none" stroke="#2c313a" stroke-width="8" />
        <circle cx="50" cy="50" r="35" fill="none" stroke="${isCritical ? '#ff8533' : '#2196f3'}" 
                stroke-width="8" stroke-dasharray="${node.memUsage * 2.2} 220" 
                transform="rotate(-90 50 50)" style="transition: stroke-dasharray 0.5s ease" />
        
        <text x="50" y="55" text-anchor="middle" fill="#fff" font-size="12px" font-family="monospace">
          ${node.nodeId.substring(0, 4)}
        </text>

        ${isCritical ? `
          <style>
            @keyframes pulse { 0% { opacity: 1; } 50% { opacity: 0.5; } 100% { opacity: 1; } }
            circle { animation: pulse 1s infinite; }
          </style>
        ` : ''}
      </svg>
    `;

    card.innerHTML = svg;
    card.onclick = () => this.showNodeDetails(node);
    return card;
  }

  private showNodeDetails(node: NodeStats): void {
    console.log(`[FLEET] Node Details:`, node);
    // Side panel logic would be implemented here
  }
}
