import React, { useState, useEffect } from 'react';
import { 
  Activity, 
  Cpu, 
  Database, 
  ShieldCheck, 
  Radio, 
  Terminal, 
  Layers, 
  AlertCircle 
} from 'lucide-react';
import { 
  AreaChart, 
  Area, 
  XAxis, 
  YAxis, 
  CartesianGrid, 
  Tooltip, 
  ResponsiveContainer 
} from 'recharts';
import { motion, AnimatePresence } from 'framer-motion';

/**
 * Sovereign Intelligence - Command Center
 * Principal Admin Interface
 */

function App() {
  const [telemetry, setTelemetry] = useState([]);
  const [current, setCurrent] = useState({
    cpu_usage: 0,
    memory_used: 0,
    task_throughput: 0,
    active_inference: false
  });
  const [logs, setLogs] = useState([
    { id: 1, time: '14:20:05', tag: 'CORE', msg: 'Neural orchestrator synchronized.' },
    { id: 2, time: '14:20:08', tag: 'SECURITY', msg: 'LSM-tree vault integrity verified.' }
  ]);

  useEffect(() => {
    const socket = new WebSocket('ws://localhost:8081/ws/telemetry');

    socket.onmessage = (event) => {
      const data = JSON.parse(event.data);
      setCurrent(data);
      setTelemetry(prev => {
        const next = [...prev, { time: new Date().toLocaleTimeString(), val: data.cpu_usage }];
        return next.slice(-20); // Keep last 20 frames
      });

      // Periodic log simulation
      if (Math.random() > 0.8) {
        setLogs(prev => [
          { id: Date.now(), time: new Date().toLocaleTimeString(), tag: 'INTEL', msg: `Task node ${Math.floor(Math.random() * 1000)} processed.` },
          ...prev.slice(0, 10)
        ]);
      }
    };

    return () => socket.close();
  }, []);

  const formatMB = (bytes) => (bytes / (1024 * 1024)).toFixed(2);

  return (
    <div className="admin-layout">
      {/* Sidebar Navigation */}
      <aside class="admin-sidebar">
        <div className="brand">
          <Layers size={28} />
          Sovereign <span>Core</span>
        </div>

        <div className="nav-group">
          <div className="nav-label">Global Fleet</div>
          <div className="nav-item active">
            <Radio size={18} /> Node Alpha
          </div>
          <div className="nav-item">
            <Database size={18} /> Storage Clusters
          </div>
        </div>

        <div className="nav-group">
          <div className="nav-label">Intelligence</div>
          <div className="nav-item">
            <Activity size={18} /> Derivation Flow
          </div>
          <div className="nav-item">
            <ShieldCheck size={18} /> Security Guard
          </div>
        </div>

        <div style={{ marginTop: 'auto' }}>
          <div className="status-badge">
            <div className="status-dot-pulse"></div>
            Core Active
          </div>
        </div>
      </aside>

      {/* Main Dashboard */}
      <main className="admin-main">
        <header style={{ marginBottom: '2.5rem' }}>
          <h1 style={{ fontSize: '1.8rem', fontWeight: 700 }}>Node Alpha: Command Center</h1>
          <p style={{ color: 'var(--text-dim)', fontSize: '0.9rem' }}>Real-time hardware telemetry and task orchestration.</p>
        </header>

        {/* Stats Summary */}
        <section className="stats-grid">
          <StatCard 
            icon={<Cpu color="var(--accent-purple)" />} 
            label="Neural Threads" 
            value={`${current.cpu_usage.toFixed(0)} active`}
            trend="+2.1%"
          />
          <StatCard 
            icon={<Database color="#60A5FA" />} 
            label="Memory Allocation" 
            value={`${formatMB(current.memory_used)} MB`}
            trend="Stable"
          />
          <StatCard 
            icon={<Activity color="#4ADE80" />} 
            label="Inference Velocity" 
            value={`${current.task_throughput}/s`}
            trend="Peak"
          />
          <StatCard 
            icon={<ShieldCheck color="#F87171" />} 
            label="Security Barrier" 
            value="Active"
            trend="Level 4"
          />
        </section>

        {/* Neural Pulse Chart */}
        <section className="chart-section">
          <div className="chart-header">
            <h3>Neural Pulse (Load Variance)</h3>
            <Terminal size={18} color="var(--text-dim)" />
          </div>
          <div style={{ width: '100%', height: 300 }}>
            <ResponsiveContainer>
              <AreaChart data={telemetry}>
                <defs>
                  <linearGradient id="colorVal" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#8A2BE2" stopOpacity={0.3}/>
                    <stop offset="95%" stopColor="#8A2BE2" stopOpacity={0}/>
                  </linearGradient>
                </defs>
                <CartesianGrid vertical={false} stroke="rgba(255,255,255,0.03)" />
                <XAxis hide dataKey="time" />
                <YAxis hide domain={[0, 100]} />
                <Tooltip 
                  contentStyle={{ background: '#111', border: 'none', borderRadius: '8px', fontSize: '12px' }}
                  itemStyle={{ color: '#fff' }}
                />
                <Area 
                  type="monotone" 
                  dataKey="val" 
                  stroke="#8A2BE2" 
                  strokeWidth={3}
                  fillOpacity={1} 
                  fill="url(#colorVal)" 
                  animationDuration={300}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </section>

        {/* Cognitive Log Stream */}
        <section>
          <div className="chart-header">
            <h3>Cognitive Logs</h3>
            <AlertCircle size={18} color="var(--text-dim)" />
          </div>
          <div className="log-stream">
            {logs.map(log => (
              <motion.div 
                key={log.id} 
                initial={{ opacity: 0, x: -10 }} 
                animate={{ opacity: 1, x: 0 }}
                className="log-entry"
              >
                <span className="log-time">[{log.time}]</span>
                <span className="log-tag">_{log.tag}</span>
                <span className="log-msg">{log.msg}</span>
              </motion.div>
            ))}
          </div>
        </section>
      </main>
    </div>
  );
}

function StatCard({ icon, label, value, trend }) {
  return (
    <motion.div 
      whileHover={{ translateY: -5 }}
      className="stat-card"
    >
      <div style={{ marginBottom: '1rem' }}>{icon}</div>
      <div className="stat-label">{label}</div>
      <div className="stat-value">{value}</div>
      <div className="stat-trend">{trend}</div>
    </motion.div>
  );
}

export default App;
