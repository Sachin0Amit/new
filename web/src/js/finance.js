import SovereignCursor from './Cursor.js';

class FinanceDashboard {
    constructor() {
        this.currentSymbol = 'AAPL';
        this.backendUrl = 'http://localhost:8081';
        this.tvChart = null;
        this.candleSeries = null;
        this.volumeSeries = null;
        this.lastLogCount = 0;
        this.currentPrice = 150;
        this.init();
    }

    init() {
        new SovereignCursor();
        this.bindAssetSelector();
        this.bindTraderEvents();
        this.bindTerminalTabs();
        this.bindOrderButtons();
        this.fetchDashboardData();

        // Polling
        setInterval(() => this.fetchTraderLogs(), 5000);
        setInterval(() => this.updateOrderBook(), 800);
    }

    // ──────────────────────────────────────────────────────────────────────────
    // EVENT BINDING
    // ──────────────────────────────────────────────────────────────────────────

    bindAssetSelector() {
        document.querySelectorAll('.asset-item').forEach(item => {
            item.addEventListener('click', e => {
                document.querySelectorAll('.asset-item').forEach(i => i.classList.remove('active'));
                e.currentTarget.classList.add('active');
                this.currentSymbol = e.currentTarget.dataset.symbol;
                document.getElementById('ticker-symbol').textContent = this.currentSymbol;
                this.fetchDashboardData();
            });
        });

        const refreshBtn = document.querySelector('.win-btn[title="Refresh Model"]');
        if (refreshBtn) refreshBtn.addEventListener('click', () => this.fetchDashboardData());
    }

    bindTerminalTabs() {
        document.querySelectorAll('.tab-btn').forEach(btn => {
            btn.addEventListener('click', e => {
                document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
                document.querySelectorAll('.tab-content').forEach(c => {
                    c.style.display = 'none';
                    c.classList.remove('active');
                });
                e.currentTarget.classList.add('active');
                const target = document.getElementById(e.currentTarget.dataset.target);
                if (target) {
                    target.style.display = 'flex';
                    target.classList.add('active');
                }
            });
        });
    }

    bindOrderButtons() {
        document.getElementById('btn-buy').addEventListener('click', () => this.placeOrder('BUY'));
        document.getElementById('btn-sell').addEventListener('click', () => this.placeOrder('SELL'));
    }

    bindTraderEvents() {
        const modal    = document.getElementById('broker-modal');
        const loginBtn = document.getElementById('broker-login-btn');
        const cancelBtn= document.getElementById('broker-cancel-btn');
        const connectBtn=document.getElementById('broker-connect-btn');
        const toggle   = document.getElementById('auto-trade-switch');

        if (loginBtn)  loginBtn.addEventListener('click',  () => modal.style.display = 'flex');
        if (cancelBtn) cancelBtn.addEventListener('click', () => modal.style.display = 'none');

        if (connectBtn) {
            connectBtn.addEventListener('click', async () => {
                const apiKey    = document.getElementById('broker-api-key').value;
                const apiSecret = document.getElementById('broker-api-secret').value;
                const broker    = document.getElementById('broker-select').value;

                const body = new URLSearchParams({ api_key: apiKey, api_secret: apiSecret, broker });
                try {
                    const res = await fetch(`${this.backendUrl}/api/v1/finance/trader/login`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                        body
                    });
                    if (res.ok) {
                        modal.style.display = 'none';
                        loginBtn.textContent = '✓ Broker Connected';
                        loginBtn.style.color = '#00e676';
                        loginBtn.style.borderColor = 'rgba(0,230,118,0.3)';
                        this.appendLog('[SYSTEM] Authenticated with broker.');
                    } else {
                        const err = await res.json();
                        alert('Login failed: ' + (err.message || 'Unknown error'));
                    }
                } catch (e) {
                    alert('Connection error: ' + e.message);
                }
            });
        }

        if (toggle) {
            toggle.addEventListener('change', async e => {
                const action = e.target.checked ? 'start' : 'stop';
                try {
                    const res  = await fetch(`${this.backendUrl}/api/v1/finance/trader/toggle?action=${action}`, { method: 'POST' });
                    const data = await res.json();
                    if (!res.ok) { alert(data.message || 'Failed'); e.target.checked = !e.target.checked; return; }
                    this.appendLog(data.is_running === 'true'
                        ? '[SYSTEM] Autonomous Engine ARMED.'
                        : '[SYSTEM] Autonomous Engine DISARMED.');
                } catch {
                    e.target.checked = !e.target.checked;
                }
            });
        }
    }

    // ──────────────────────────────────────────────────────────────────────────
    // DATA FETCHING
    // ──────────────────────────────────────────────────────────────────────────

    async fetchDashboardData() {
        try {
            const [mktRes, indRes, predRes, riskRes, intelRes] = await Promise.all([
                fetch(`${this.backendUrl}/api/v1/finance/market?symbol=${this.currentSymbol}`),
                fetch(`${this.backendUrl}/api/v1/finance/indicators?symbol=${this.currentSymbol}`),
                fetch(`${this.backendUrl}/api/v1/finance/predict?symbol=${this.currentSymbol}&horizon=30`),
                fetch(`${this.backendUrl}/api/v1/finance/risk?symbol=${this.currentSymbol}`),
                fetch(`${this.backendUrl}/api/v1/finance/intelligence?symbol=${this.currentSymbol}`)
            ]);

            const [market, indicators, predict, risk, intel] = await Promise.all([
                mktRes.json(), indRes.json(), predRes.json(), riskRes.json(), intelRes.json()
            ]);

            this.renderAll(market, indicators, predict, risk, intel);
        } catch (err) {
            console.error('Fetch failed:', err);
        }
    }

    async fetchTraderLogs() {
        try {
            const res = await fetch(`${this.backendUrl}/api/v1/finance/trader/status`);
            if (!res.ok) return;
            const data = await res.json();
            this.syncTraderStatus(data);
        } catch { /* silent */ }
    }

    // ──────────────────────────────────────────────────────────────────────────
    // RENDERING
    // ──────────────────────────────────────────────────────────────────────────

    renderAll(market, indicators, predict, risk, intel) {
        this.renderPriceHeader(market);
        this.renderCandlestickChart(market);
        this.renderSignal(predict);
        this.renderRisk(risk);
        this.renderIntelligence(intel);
    }

    renderPriceHeader(market) {
        if (!market || market.length < 2) return;
        const last = market[market.length - 1];
        const prev = market[market.length - 2];
        this.currentPrice = last.close;

        const decimals = last.close < 10 ? 4 : 2;
        document.getElementById('current-price').textContent = `$${last.close.toFixed(decimals)}`;

        const pct = ((last.close - prev.close) / prev.close) * 100;
        const changeEl = document.querySelector('.price-change');
        changeEl.textContent = `${pct >= 0 ? '+' : ''}${pct.toFixed(2)}%`;
        changeEl.className = `price-change ${pct >= 0 ? 'up' : 'down'}`;
    }

    renderCandlestickChart(market) {
        const container = document.getElementById('main-chart');
        if (!container || !market || market.length === 0) return;
        if (!window.LightweightCharts) { console.error('LightweightCharts not loaded'); return; }

        // Destroy old chart
        container.innerHTML = '';
        if (this.tvChart) { this.tvChart.remove(); this.tvChart = null; }

        this.tvChart = LightweightCharts.createChart(container, {
            width:  container.clientWidth  || 800,
            height: container.clientHeight || 380,
            layout: { background: { type: 'solid', color: '#0a0a0a' }, textColor: '#c9d1d9' },
            grid:   { vertLines: { color: '#161b22' }, horzLines: { color: '#161b22' } },
            crosshair: { mode: LightweightCharts.CrosshairMode.Normal },
            rightPriceScale: { borderColor: '#30363d', scaleMargins: { top: 0.1, bottom: 0.25 } },
            timeScale: { borderColor: '#30363d', timeVisible: true, secondsVisible: false },
            watermark: { visible: false }
        });

        this.candleSeries = this.tvChart.addCandlestickSeries({
            upColor:         '#00e676',
            downColor:       '#ff5252',
            borderUpColor:   '#00e676',
            borderDownColor: '#ff5252',
            wickUpColor:     '#00e676',
            wickDownColor:   '#ff5252',
        });

        this.volumeSeries = this.tvChart.addHistogramSeries({
            priceFormat: { type: 'volume' },
            priceScaleId: 'vol',
        });
        this.tvChart.priceScale('vol').applyOptions({ scaleMargins: { top: 0.8, bottom: 0 } });

        // Build OHLCV data from API response
        const seen = new Set();
        const candles = [];
        const volumes = [];

        for (const bar of market) {
            const ts = Math.floor(new Date(bar.timestamp).getTime() / 1000);
            if (seen.has(ts)) continue;
            seen.add(ts);
            candles.push({ time: ts, open: bar.open, high: bar.high, low: bar.low, close: bar.close });
            volumes.push({ time: ts, value: bar.volume, color: bar.close >= bar.open ? 'rgba(0,230,118,0.4)' : 'rgba(255,82,82,0.4)' });
        }

        candles.sort((a, b) => a.time - b.time);
        volumes.sort((a, b) => a.time - b.time);

        this.candleSeries.setData(candles);
        this.volumeSeries.setData(volumes);
        this.tvChart.timeScale().fitContent();

        // Responsive resize
        new ResizeObserver(() => {
            if (this.tvChart && container.clientWidth > 0) {
                this.tvChart.applyOptions({ width: container.clientWidth, height: container.clientHeight });
            }
        }).observe(container);
    }

    renderSignal(predict) {
        if (!predict || !predict.prediction) return;
        const p = predict.prediction;
        const sigBox = document.getElementById('ai-signal');
        sigBox.textContent  = p.direction;
        sigBox.className = `signal-box ${p.direction.toLowerCase()}`;

        // Update header metrics if elements exist
        this.setEl('pred-target', `$${p.predicted_price.toFixed(2)}`);
        this.setEl('pred-conf',   `${(p.confidence * 100).toFixed(1)}%`);
        this.setEl('pred-lower',  `$${p.lower_bound.toFixed(2)}`);
        this.setEl('pred-upper',  `$${p.upper_bound.toFixed(2)}`);
    }

    renderRisk(risk) {
        if (!risk || !risk.risk) return;
        const r = risk.risk;
        this.setEl('risk-sharpe', r.sharpe.toFixed(2));
        this.setEl('risk-vol',    `${(r.volatility_annual * 100).toFixed(1)}%`);
        this.setEl('risk-dd',     `${(r.max_drawdown * 100).toFixed(1)}%`);
        this.setEl('risk-var',    `${(r.var_95 * 100).toFixed(2)}%`);
    }

    renderIntelligence(intel) {
        if (!intel) return;

        // Fundamentals
        if (intel.fundamentals) {
            const f = intel.fundamentals;
            const ratingEl = document.getElementById('fund-rating');
            this.setEl('fund-score',    f.total_score.toFixed(0));
            this.setEl('fund-reasoning',f.reasoning);
            if (ratingEl) {
                ratingEl.textContent = f.rating;
                ratingEl.style.color = f.total_score >= 80 ? '#00e676' : f.total_score >= 60 ? '#00d4ff' : f.total_score >= 40 ? '#ffb300' : '#ff5252';
            }
        }

        // News
        if (intel.news) {
            const container = document.getElementById('news-container');
            if (!container) return;
            if (!intel.news.length) {
                container.innerHTML = '<div style="color:#666;font-style:italic;padding:1rem;">No live news available.</div>';
                return;
            }
            container.innerHTML = '';
            let totalScore = 0;
            intel.news.forEach(article => {
                totalScore += article.score;
                const color = article.score > 0.2 ? '#00e676' : article.score < -0.2 ? '#ff5252' : '#888';
                const label = article.score > 0.2 ? 'BULLISH' : article.score < -0.2 ? 'BEARISH' : 'NEUTRAL';
                const div = document.createElement('div');
                div.style.cssText = 'border-bottom:1px solid #1a1a1a;padding:0.5rem 0;';
                div.innerHTML = `
                    <div style="display:flex;justify-content:space-between;align-items:flex-start;gap:1rem;">
                        <span style="color:#fff;font-size:0.88rem;line-height:1.4;">${article.title}</span>
                        <span style="color:${color};font-size:0.7rem;font-weight:700;white-space:nowrap;padding:2px 6px;background:rgba(255,255,255,0.04);border-radius:3px;">${label}</span>
                    </div>
                    <div style="color:#555;font-size:0.75rem;margin-top:3px;">${new Date(article.pub_date).toLocaleTimeString()}</div>`;
                container.appendChild(div);
            });

            // Update badge
            const avg = totalScore / intel.news.length;
            const badge = document.getElementById('news-sentiment-badge');
            if (badge) {
                if (avg > 0.2)       { badge.textContent = 'BULLISH'; badge.style.cssText = 'background:rgba(0,230,118,0.15);color:#00e676;padding:2px 8px;border-radius:3px;font-size:0.8rem;'; }
                else if (avg < -0.2) { badge.textContent = 'BEARISH'; badge.style.cssText = 'background:rgba(255,82,82,0.15);color:#ff5252;padding:2px 8px;border-radius:3px;font-size:0.8rem;'; }
                else                 { badge.textContent = 'NEUTRAL'; badge.style.cssText = 'background:rgba(255,255,255,0.05);color:#888;padding:2px 8px;border-radius:3px;font-size:0.8rem;'; }
            }
        }
    }

    // ──────────────────────────────────────────────────────────────────────────
    // LIVE ORDER BOOK (DOM)
    // ──────────────────────────────────────────────────────────────────────────

    updateOrderBook() {
        const p = this.currentPrice;
        if (!p) return;

        const asksEl   = document.getElementById('dom-asks');
        const bidsEl   = document.getElementById('dom-bids');
        const spreadEl = document.getElementById('dom-spread-price');
        if (!asksEl || !bidsEl) return;

        if (spreadEl) spreadEl.textContent = `$${p.toFixed(2)}`;

        const decimals = p < 10 ? 4 : 2;
        const spread   = p * 0.0003;

        let asks = '', bids = '';
        let maxQty = 0;
        const rows = [];

        for (let i = 5; i >= 1; i--) {
            const price = p + spread * i;
            const qty   = Math.floor(Math.random() * 800 + 50);
            maxQty = Math.max(maxQty, qty);
            rows.push({ price, qty, side: 'ask' });
        }
        for (let i = 1; i <= 5; i++) {
            const price = p - spread * i;
            const qty   = Math.floor(Math.random() * 800 + 50);
            maxQty = Math.max(maxQty, qty);
            rows.push({ price, qty, side: 'bid' });
        }

        for (const r of rows) {
            const pct = Math.min(95, (r.qty / (maxQty || 1)) * 100);
            const row = `<div style="display:flex;justify-content:space-between;padding:2px 6px;position:relative;overflow:hidden;">
                <div style="position:absolute;top:0;bottom:0;right:0;width:${pct}%;background:${r.side==='ask'?'rgba(255,82,82,0.08)':'rgba(0,230,118,0.08)'};z-index:0;"></div>
                <span style="z-index:1;color:${r.side==='ask'?'#ff5252':'#00e676'};">$${r.price.toFixed(decimals)}</span>
                <span style="z-index:1;color:#888;">${r.qty}</span>
            </div>`;
            if (r.side === 'ask') asks = row + asks;
            else                   bids += row;
        }

        asksEl.innerHTML = asks;
        bidsEl.innerHTML = bids;
    }

    // ──────────────────────────────────────────────────────────────────────────
    // MANUAL ORDER
    // ──────────────────────────────────────────────────────────────────────────

    async placeOrder(action) {
        const qty = parseFloat(document.getElementById('manual-qty').value) || 1;
        const p   = this.currentPrice;
        const decimals = p < 10 ? 4 : 2;

        const btn = document.getElementById(action === 'BUY' ? 'btn-buy' : 'btn-sell');
        const orig = btn.textContent;
        btn.textContent = '...';
        btn.disabled = true;

        // Switch to log tab
        document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
        document.querySelectorAll('.tab-content').forEach(c => { c.style.display='none'; c.classList.remove('active'); });
        const logTab = document.querySelector('.tab-btn[data-target="tab-log"]');
        const logContent = document.getElementById('tab-log');
        if (logTab)     logTab.classList.add('active');
        if (logContent) { logContent.style.display='flex'; logContent.classList.add('active'); }

        await new Promise(r => setTimeout(r, 300)); // micro delay for UX

        const body = new URLSearchParams({ symbol: this.currentSymbol, action: action, qty: qty });
        try {
            const res = await fetch(`${this.backendUrl}/api/v1/finance/trader/order`, {
                method: 'POST',
                headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
                body
            });
            if (res.ok) {
                const data = await res.json();
                const now   = new Date().toLocaleTimeString();
                const color = action === 'BUY' ? '#00e676' : '#ff5252';
                const total = (qty * p).toFixed(2);
                this.appendLog(`[${now}] <span style="color:${color};font-weight:700;">${action}</span> ${qty} × ${this.currentSymbol} @ $${p.toFixed(decimals)} = $${total} <span style="color:#666;">(ID: ${data.receipt.OrderID})</span>`);
            } else {
                const err = await res.json();
                alert('Order failed: ' + (err.message || 'Unknown error'));
            }
        } catch (e) {
            alert('Connection error: ' + e.message);
        }

        btn.textContent = orig;
        btn.disabled = false;
    }

    // ──────────────────────────────────────────────────────────────────────────
    // TRADER STATUS SYNC
    // ──────────────────────────────────────────────────────────────────────────

    syncTraderStatus(data) {
        const toggle = document.getElementById('auto-trade-switch');
        if (toggle) toggle.checked = !!data.is_running;

        if (data.logs && data.logs.length > this.lastLogCount) {
            const newEntries = data.logs.slice(this.lastLogCount);
            newEntries.forEach(log => {
                const time  = new Date(log.Timestamp * 1000).toLocaleTimeString();
                const color = log.Action === 'BUY' ? '#00e676' : '#ff5252';
                this.appendLog(`[${time}] <span style="color:${color};font-weight:700;">${log.Action}</span> ${log.Quantity.toFixed(3)} ${log.Symbol} @ $${log.Price.toFixed(2)} <span style="color:#666;">(${log.Status})</span>`);
            });
            this.lastLogCount = data.logs.length;
        }
    }

    // ──────────────────────────────────────────────────────────────────────────
    // UTILS
    // ──────────────────────────────────────────────────────────────────────────

    appendLog(html) {
        const container = document.getElementById('trade-log');
        if (!container) return;
        const div = document.createElement('div');
        div.style.cssText = 'padding:2px 0;border-bottom:1px solid #111;line-height:1.5;';
        div.innerHTML = html;
        container.appendChild(div);
        container.scrollTop = container.scrollHeight;
    }

    setEl(id, val) {
        const el = document.getElementById(id);
        if (el) el.textContent = val;
    }
}

document.addEventListener('DOMContentLoaded', () => new FinanceDashboard());
