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
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors bg-gradient-to-br from-gray-50 via-gray-50 to-cyan-50/30 dark:from-gray-900 dark:via-gray-900 dark:to-cyan-950/20">
      <header className="bg-white/80 dark:bg-gray-800/80 backdrop-blur-sm shadow-sm transition-colors border-b border-cyan-500/10 dark:border-cyan-400/10">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8 flex justify-between items-start">
          <div>
            <h1 className="text-3xl font-bold text-gray-900 dark:text-white">DSV Network Activity Dashboard</h1>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">Real-time visualization of decentralized sequencer validator activity</p>
          </div>
          <button
            onClick={toggleTheme}
            className="p-2 rounded-lg bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 transition-colors"
            title={theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          >
            {theme === 'dark' ? '☀️' : '🌙'}
          </button>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8 space-y-6">
        {/* Summary Cards */}
        {!summaryError && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm p-6 transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-cyan-50/30 dark:from-gray-800 dark:to-cyan-950/20">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Epochs</div>
              <div className="mt-2 text-3xl font-semibold text-gray-900 dark:text-white">
                {summaryLoading ? '...' : summary?.total_epochs || 0}
              </div>
            </div>
            <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm p-6 transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-cyan-50/30 dark:from-gray-800 dark:to-cyan-950/20">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400 flex items-center gap-1">
                DSV Validators
                <span title="Validators that produced L1 batches" className="cursor-help text-gray-400 dark:text-gray-500">ⓘ</span>
              </div>
              <div className="mt-2 text-3xl font-semibold text-green-600 dark:text-green-400">
                {summaryLoading ? '...' : summary?.total_validators || 0}
              </div>
            </div>
            <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm p-6 transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-fuchsia-50/30 dark:from-gray-800 dark:to-fuchsia-950/20">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400 flex items-center gap-1">
                Unique Slots
                <span title="Snapshotter (lite) nodes that have contributed to consensus" className="cursor-help text-gray-400 dark:text-gray-500">ⓘ</span>
              </div>
              <div className="mt-2 text-3xl font-semibold text-blue-600 dark:text-blue-400">
                {summaryLoading ? '...' : summary?.total_slots || 0}
              </div>
            </div>
            <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm p-6 transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-cyan-50/30 dark:from-gray-800 dark:to-cyan-950/20">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Current Day</div>
              <div className="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
                {summaryLoading ? '...' : summary?.current_day || 'N/A'}
              </div>
            </div>
          </div>
        )}

        {/* Pipeline Diagram (collapsible) */}
        <PipelineDiagram />

        {/* Epoch Charts */}
        <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm p-6 transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-cyan-50/20 dark:from-gray-800 dark:to-cyan-950/20">
          <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">Slots & Projects per Epoch</h2>
          <EpochCharts />
        </div>

        {/* Epoch Activity Table */}
        <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-cyan-50/20 dark:from-gray-800 dark:to-cyan-950/20">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-medium text-gray-900 dark:text-white">Epoch Activity</h2>
            <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
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
          <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm p-6 transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-cyan-50/20 dark:from-gray-800 dark:to-cyan-950/20">
            <EpochDetailView epochId={selectedEpochId} />
          </div>
        )}
      </main>
    </div>
  );
};

export default Dashboard;
