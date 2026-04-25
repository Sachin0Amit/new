/**
 * Unified Command Center - Admin Dashboard Orchestrator
 */

import { FleetMatrix } from './fleet-matrix';
import { ReasoningVisualizer } from './reasoning-visualizer';
import { ReflexMonitor } from './reflex-monitor';

export class AdminDashboard {
  public fleet: FleetMatrix;
  public visualizer: ReasoningVisualizer;
  public monitor: ReflexMonitor;

  constructor() {
    this.fleet = new FleetMatrix('fleet-container');
    this.visualizer = new ReasoningVisualizer('graph-container');
    this.monitor = new ReflexMonitor('reflex-container');
    this.initLayout();
  }

  private initLayout(): void {
    const layout = `
      <div id="admin-dashboard" style="display: flex; flex-direction: column; height: 100vh; background: #0b0e11; color: #fff;">
        <!-- Header -->
        <header style="height: 60px; border-bottom: 1px solid #1f2226; display: flex; align-items: center; padding: 0 20px; justify-content: space-between;">
          <div style="font-weight: bold; font-size: 1.2rem;">Sovereign Core // Command Center</div>
          <div id="global-stats" style="display: flex; gap: 30px; font-size: 0.9rem;">
            <span>ACTIVE NODES: <b id="stat-nodes">0</b></span>
            <span>DERIVATIONS: <b id="stat-derivations">0</b></span>
            <span>ANOMALY RATE: <b id="stat-anomalies" style="color: #4caf50">0.0%</b></span>
          </div>
        </header>

        <div style="display: flex; flex: 1;">
          <!-- Sidebar -->
          <nav style="width: 250px; border-right: 1px solid #1f2226; padding: 20px;">
            <ul style="list-style: none; padding: 0;">
              <li class="tab-btn active" onclick="showTab('fleet')">Fleet Matrix</li>
              <li class="tab-btn" onclick="showTab('visualizer')">Reasoning DAG</li>
              <li class="tab-btn" onclick="showTab('reflex')">Reflex Monitor</li>
            </ul>
          </nav>

          <!-- Main Area -->
          <main style="flex: 1; position: relative;">
            <div id="fleet-container" class="tab-content"></div>
            <div id="graph-container" class="tab-content" style="display: none;"></div>
            <div id="reflex-container" class="tab-content" style="display: none;"></div>
          </main>
        </div>
      </div>
    `;
    document.body.innerHTML = layout;
  }

  public updateGlobalStats(nodes: number, derivations: number, anomalies: number): void {
    document.getElementById('stat-nodes')!.innerText = nodes.toString();
    document.getElementById('stat-derivations')!.innerText = derivations.toString();
    const rate = (anomalies / derivations * 100).toFixed(1);
    const el = document.getElementById('stat-anomalies')!;
    el.innerText = `${rate}%`;
    el.style.color = anomalies > 0 ? '#f44336' : '#4caf50';
  }
}
