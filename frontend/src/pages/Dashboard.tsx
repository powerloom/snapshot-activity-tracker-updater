import { useState } from 'react';
import { useDashboardSummary } from '../hooks/useDashboardData';
import { useTheme } from '../contexts/ThemeContext';
import PipelineDiagram from '../components/PipelineDiagram';
import EpochCharts from '../components/EpochCharts';
import EpochActivityTable from '../components/EpochActivityTable';
import EpochDetailView from '../components/EpochDetailView';

const Dashboard = () => {
  const { data: summary, isLoading: summaryLoading, error: summaryError } = useDashboardSummary();
  const { toggleTheme, theme } = useTheme();
  const [selectedEpochId, setSelectedEpochId] = useState<number | null>(null);

  return (
    <div className="min-h-screen futuristic-bg bg-gradient-to-br from-slate-50 via-cyan-50/20 to-slate-100 dark:from-gray-950 dark:via-slate-900 dark:to-cyan-950/30 transition-colors">
      <header className="relative z-10 bg-white/70 dark:bg-gray-900/70 backdrop-blur-xl border-b border-cyan-500/30 dark:border-cyan-400/20 shadow-[0_0_30px_rgba(34,211,238,0.1)] dark:shadow-[0_0_40px_rgba(34,211,238,0.08)]">
        <div className="max-w-7xl mx-auto px-4 py-5 sm:px-6 lg:px-8 flex justify-between items-start">
          <div>
            <h1 className="font-orbitron text-3xl font-bold tracking-tight text-gray-900 dark:text-white dark:neon-text">
              DSV Network Activity
            </h1>
            <p className="mt-1.5 font-mono text-xs uppercase tracking-widest text-cyan-600 dark:text-cyan-400 opacity-90">
              Real-time decentralized sequencer validator activity
            </p>
          </div>
          <button
            onClick={toggleTheme}
            className="p-2.5 rounded-lg bg-white/80 dark:bg-gray-800/80 border border-cyan-500/20 dark:border-cyan-400/20 hover:border-cyan-500/40 hover:shadow-glow-cyan transition-all duration-300"
            title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {theme === 'dark' ? '☀️' : '🌙'}
          </button>
        </div>
      </header>

      <main className="relative z-10 max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8 space-y-8">
        {/* Summary Cards */}
        {!summaryError && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            <div className="futuristic-card rounded-xl p-6 group">
              <div className="font-mono text-xs uppercase tracking-wider text-cyan-600 dark:text-cyan-400 mb-2">Total Epochs</div>
              <div className="font-orbitron text-3xl font-bold text-gray-900 dark:text-white tabular-nums">
                {summaryLoading ? '...' : summary?.total_epochs?.toLocaleString() || 0}
              </div>
            </div>
            <div className="futuristic-card rounded-xl p-6 group">
              <div className="font-mono text-xs uppercase tracking-wider text-cyan-600 dark:text-cyan-400 flex items-center gap-1 mb-2">
                DSV Validators
                <span title="Validators that produced L1 batches" className="cursor-help opacity-60">ⓘ</span>
              </div>
              <div className="font-orbitron text-3xl font-bold text-emerald-500 dark:text-emerald-400 tabular-nums">
                {summaryLoading ? '...' : summary?.total_validators || 0}
              </div>
            </div>
            <div className="futuristic-card rounded-xl p-6 group">
              <div className="font-mono text-xs uppercase tracking-wider text-cyan-600 dark:text-cyan-400 flex items-center gap-1 mb-2">
                Unique Slots
                <span title="Snapshotter (lite) nodes that have contributed to consensus" className="cursor-help opacity-60">ⓘ</span>
              </div>
              <div className="font-orbitron text-3xl font-bold text-blue-500 dark:text-blue-400 tabular-nums">
                {summaryLoading ? '...' : summary?.total_slots?.toLocaleString() || 0}
              </div>
            </div>
            <div className="futuristic-card rounded-xl p-6 group">
              <div className="font-mono text-xs uppercase tracking-wider text-cyan-600 dark:text-cyan-400 mb-2">Current Day</div>
              <div className="font-orbitron text-2xl font-bold text-gray-900 dark:text-white tabular-nums">
                {summaryLoading ? '...' : summary?.current_day || '—'}
              </div>
            </div>
          </div>
        )}

        {/* Pipeline Diagram (collapsible) */}
        <PipelineDiagram />

        {/* Epoch Charts */}
        <div className="futuristic-card rounded-xl p-6">
          <h2 className="font-orbitron text-lg font-semibold text-gray-900 dark:text-white mb-4 flex items-center gap-2">
            <span className="w-1 h-6 bg-gradient-to-b from-cyan-400 to-blue-500 rounded-full" />
            Slots & Projects per Epoch
          </h2>
          <EpochCharts />
        </div>

        {/* Epoch Activity Table */}
        <div className="futuristic-card rounded-xl overflow-hidden">
          <div className="px-6 py-4 border-b border-cyan-500/20 dark:border-cyan-400/10">
            <h2 className="font-orbitron text-lg font-semibold text-gray-900 dark:text-white flex items-center gap-2">
              <span className="w-1 h-6 bg-gradient-to-b from-cyan-400 to-blue-500 rounded-full" />
              Epoch Activity
            </h2>
            <p className="mt-1 font-mono text-xs text-gray-500 dark:text-gray-400">
              Click a row to view epoch detail
            </p>
          </div>
          <div className="p-6">
            <EpochActivityTable
              selectedEpochId={selectedEpochId}
              onSelectEpoch={setSelectedEpochId}
            />
          </div>
        </div>

        {/* Epoch Detail */}
        {selectedEpochId && (
          <div className="futuristic-card rounded-xl p-6">
            <EpochDetailView epochId={selectedEpochId} />
          </div>
        )}
      </main>
    </div>
  );
};

export default Dashboard;
