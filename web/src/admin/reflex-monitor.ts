/**
 * Sovereign Reflex Monitor - Autonomous Self-Healing Visualization
 */

export interface ReflexEvent {
  id: string;
  timestamp: string;
  derivationId: string;
  anomalyType: string;
  actionTaken: string;
  status: 'triggered' | 'correcting' | 'resolved' | 'failed';
}

export class ReflexMonitor {
  private container: HTMLElement;
  private events: ReflexEvent[] = [];

  constructor(containerId: string) {
    const el = document.getElementById(containerId);
    if (!el) throw new Error(`Container ${containerId} not found`);
    this.container = el;
    this.container.style.background = '#0b0e11';
    this.container.style.color = '#fff';
    this.container.style.fontFamily = 'monospace';
  }

  public addEvent(event: ReflexEvent): void {
    this.events.unshift(event);
    if (this.events.length > 100) this.events.pop();
    this.render();
  }

  private render(): void {
    const statusColors: Record<string, string> = {
      triggered: '#ffbf00',
      correcting: '#ff8533',
      resolved: '#4caf50',
      failed: '#f44336'
    };

    const rows = this.events.map(e => `
      <div style="border-bottom: 1px solid #1f2226; padding: 10px; display: flex; gap: 20px;">
        <span style="color: #808b96">${new Date(e.timestamp).toLocaleTimeString()}</span>
        <span style="color: #5dade2">[${e.derivationId.substring(0, 8)}]</span>
        <span style="color: ${statusColors[e.status]}; font-weight: bold;">${e.status.toUpperCase()}</span>
        <span>${e.anomalyType} -> ${e.actionTaken}</span>
      </div>
    `).join('');

    this.container.innerHTML = `
      <div style="padding: 15px; border-bottom: 2px solid #2c313a; font-weight: bold;">
        Live Reflex Activity
      </div>
      <div style="overflow-y: auto; max-height: 500px;">
        ${rows}
      </div>
    `;
  }
}
