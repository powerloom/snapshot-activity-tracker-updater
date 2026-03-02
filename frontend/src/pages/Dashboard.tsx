import { useState } from 'react';
import { useDashboardSummary, useNetworkTopology } from '../hooks/useDashboardData';
import { useTheme } from '../contexts/ThemeContext';
import NetworkTopology from '../components/NetworkTopology';

const Dashboard = () => {
  const { data: summary, isLoading: summaryLoading, error: summaryError } = useDashboardSummary();
  const { data: topology, isLoading: topologyLoading, error: topologyError } = useNetworkTopology();
  const { theme, toggleTheme } = useTheme();
  const [showTopologyHelp, setShowTopologyHelp] = useState(false);

  return (
    <div className="min-h-screen bg-gray-50 dark:bg-gray-900 transition-colors">
      <header className="bg-white dark:bg-gray-800 shadow-sm transition-colors">
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

      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
        {/* Summary Cards */}
        {!summaryError && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Epochs</div>
              <div className="mt-2 text-3xl font-semibold text-gray-900 dark:text-white">
                {summaryLoading ? '...' : summary?.total_epochs || 0}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Validators</div>
              <div className="mt-2 text-3xl font-semibold text-green-600 dark:text-green-400">
                {summaryLoading ? '...' : summary?.total_validators || 0}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Active Slots</div>
              <div className="mt-2 text-3xl font-semibold text-blue-600 dark:text-blue-400">
                {summaryLoading ? '...' : summary?.total_slots || 0}
              </div>
              <p className="mt-1 text-xs text-gray-400 dark:text-gray-500">Unique snapshotter nodes</p>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Current Day</div>
              <div className="mt-2 text-lg font-semibold text-gray-900 dark:text-white">
                {summaryLoading ? '...' : summary?.current_day || 'N/A'}
              </div>
            </div>
          </div>
        )}

        {/* Network Topology */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow transition-colors">
          <div className="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
            <div className="flex items-start justify-between gap-4">
              <div>
                <h2 className="text-lg font-medium text-gray-900 dark:text-white">Network Topology</h2>
                <p className="mt-1 text-sm text-gray-500 dark:text-gray-400">
                  Who submits data to what — snapshotter slots, data projects, and validators (last 30 epochs)
                </p>
              </div>
              <button
                onClick={() => setShowTopologyHelp(!showTopologyHelp)}
                className="text-sm text-blue-600 dark:text-blue-400 hover:underline shrink-0"
              >
                {showTopologyHelp ? 'Hide' : 'What does this show?'}
              </button>
            </div>
            {showTopologyHelp && (
              <div className="mt-4 p-4 rounded-lg bg-gray-50 dark:bg-gray-700/50 text-sm text-gray-600 dark:text-gray-300 space-y-2">
                <p><strong className="text-green-600 dark:text-green-400">Validators</strong> — Aggregate batches from snapshotter submissions and reach consensus on winning CIDs.</p>
                <p><strong className="text-blue-500 dark:text-blue-400">Slots</strong> — Snapshotter nodes (unique IDs). Each slot submits snapshots to one or more projects.</p>
                <p><strong className="text-amber-500 dark:text-amber-400">Projects</strong> — Data types (e.g. baseSnapshot, activePools, metadata). Slots submit to projects; validators validate them.</p>
                <p className="text-gray-500 dark:text-gray-400">Lines show: green = validator validates project, blue = slot submits to project. Drag nodes to rearrange; scroll to zoom.</p>
              </div>
            )}
          </div>
          <div className="p-6">
            <NetworkTopology data={topology} isLoading={topologyLoading} error={topologyError} />
          </div>
        </div>
      </main>
    </div>
  );
};

export default Dashboard;
