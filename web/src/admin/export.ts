/**
 * Sovereign Audit Export - JSON/Markdown Export Logic
 */

import { ReasoningStep } from './reasoning-visualizer';

export class AuditExporter {
  public static exportJSON(derivationId: string, steps: ReasoningStep[]): void {
    const data = JSON.stringify(steps, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    this.download(blob, `audit-${derivationId}.json`);
  }

  public static exportMarkdown(derivationId: string, steps: ReasoningStep[]): void {
    let md = `# Sovereign Audit Trail: ${derivationId}\n\n`;
    md += `Generated: ${new Date().toISOString()}\n\n`;
    md += `| Timestamp | Step ID | Status | Signature (HEX) |\n`;
    md += `|-----------|---------|--------|-----------------|\n`;

    steps.forEach(s => {
      const sigHex = s.signature.substring(0, 16) + '...';
      md += `| ${s.timestamp} | ${s.id} | ${s.status} | ${sigHex} |\n`;
    });

    const blob = new Blob([md], { type: 'text/markdown' });
    this.download(blob, `audit-${derivationId}.md`);
  }

  private static download(blob: Blob, filename: string): void {
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
  }
}
