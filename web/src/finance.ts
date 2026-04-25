import { createChart, ISeriesApi, CandlestickData } from 'lightweight-charts';

export interface PriceTick {
  time: number;
  open: number;
  high: number;
  low: number;
  close: number;
}

export class FinanceTerminal {
  private chart: ReturnType<typeof createChart> | null = null;
  private series: ISeriesApi<'Candlestick'> | null = null;
  private container: HTMLElement;
  private isPolling = false;

  constructor(containerId: string) {
    const el = document.getElementById(containerId);
    if (!el) throw new Error(`Container ${containerId} not found`);
    this.container = el;
    this.initChart();
  }

  private initChart(): void {
    this.chart = createChart(this.container, {
      width: this.container.clientWidth,
      height: 400,
      layout: {
        background: { color: '#0b0e11' },
        textColor: '#d1d4dc',
      },
      grid: {
        vertLines: { color: '#1f2226' },
        horzLines: { color: '#1f2226' },
      },
    });

    this.series = this.chart.addCandlestickSeries({
      upColor: '#26a69a',
      downColor: '#ef5350',
      borderVisible: false,
      wickUpColor: '#26a69a',
      wickDownColor: '#ef5350',
    });
  }

  public updatePrice(tick: PriceTick): void {
    if (this.series) {
      this.series.update(tick as CandlestickData);
    }
  }

  public switchToPolling(symbol: string): void {
    if (this.isPolling) return;
    this.isPolling = true;
    console.warn(`[FINANCE] Switching to polling mode for ${symbol}`);
    
    const poll = async () => {
      if (!this.isPolling) return;
      try {
        const res = await fetch(`/api/finance/price?symbol=${symbol}`);
        const data: PriceTick = await res.json();
        this.updatePrice(data);
      } catch (err) {
        console.error('[FINANCE] Polling failed:', err);
      }
      setTimeout(poll, 5000);
    };
    poll();
  }

  public stopPolling(): void {
    this.isPolling = false;
  }

  public resize(width: number, height: number): void {
    if (this.chart) {
      this.chart.applyOptions({ width, height });
    }
  }
}
