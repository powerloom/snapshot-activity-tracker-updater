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

      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8 space-y-6">
        {/* Summary Cards */}
        {!summaryError && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400">Total Epochs</div>
              <div className="mt-2 text-3xl font-semibold text-gray-900 dark:text-white">
                {summaryLoading ? '...' : summary?.total_epochs || 0}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400 flex items-center gap-1">
                DSV Validators
                <span title="Validators that produced L1 batches" className="cursor-help text-gray-400 dark:text-gray-500">ⓘ</span>
              </div>
              <div className="mt-2 text-3xl font-semibold text-green-600 dark:text-green-400">
                {summaryLoading ? '...' : summary?.total_validators || 0}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
              <div className="text-sm font-medium text-gray-500 dark:text-gray-400 flex items-center gap-1">
                Unique Slots
                <span title="Snapshotter (lite) nodes that have contributed to consensus" className="cursor-help text-gray-400 dark:text-gray-500">ⓘ</span>
              </div>
              <div className="mt-2 text-3xl font-semibold text-blue-600 dark:text-blue-400">
                {summaryLoading ? '...' : summary?.total_slots || 0}
              </div>
            </div>
            <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
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
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
          <h2 className="text-lg font-medium text-gray-900 dark:text-white mb-4">Slots & Projects per Epoch</h2>
          <EpochCharts />
        </div>

        {/* Epoch Activity Table */}
        <div className="bg-white dark:bg-gray-800 rounded-lg shadow transition-colors">
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
          <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6 transition-colors">
            <EpochDetailView epochId={selectedEpochId} />
          </div>
        )}
      </main>
    </div>
  );
};

export default Dashboard;
