import { useState } from 'react';
import { useEpochs } from '../hooks/useDashboardData';
import type { EpochSummary } from '../api/types';

const PAGE_SIZE = 20;

interface EpochActivityTableProps {
  selectedEpochId: number | null;
  onSelectEpoch: (epochId: number) => void;
}

const formatTimestamp = (ts: number) => {
  const d = new Date(ts * 1000);
  return d.toLocaleString(undefined, {
    dateStyle: 'short',
    timeStyle: 'short',
  });
};

const EpochActivityTable: React.FC<EpochActivityTableProps> = ({ selectedEpochId, onSelectEpoch }) => {
  const { data, isLoading, error } = useEpochs();
  const [page, setPage] = useState(0);
  const epochs = data?.epochs ?? [];
  // API returns oldest first; reverse to show newest first
  const sortedEpochs = [...epochs].reverse();
  const totalPages = Math.ceil(sortedEpochs.length / PAGE_SIZE) || 1;
  const displayEpochs = sortedEpochs.slice(page * PAGE_SIZE, page * PAGE_SIZE + PAGE_SIZE);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="py-16 text-center text-red-500 dark:text-red-400">
        Error loading epochs
      </div>
    );
  }

  if (epochs.length === 0) {
    return (
      <div className="py-16 text-center text-gray-500 dark:text-gray-400">
        No epoch data available
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-cyan-500/10 dark:divide-cyan-400/10">
        <thead>
          <tr>
            <th className="w-10 px-2 py-3" aria-label="Expand" />
            <th className="px-4 py-3 text-left font-mono text-xs font-medium text-cyan-600 dark:text-cyan-400 uppercase tracking-wider">
              Epoch ID
            </th>
            <th className="px-4 py-3 text-left font-mono text-xs font-medium text-cyan-600 dark:text-cyan-400 uppercase tracking-wider">
              Timestamp
            </th>
            <th className="px-4 py-3 text-right font-mono text-xs font-medium text-cyan-600 dark:text-cyan-400 uppercase tracking-wider">
              Slots
            </th>
            <th className="px-4 py-3 text-right font-mono text-xs font-medium text-cyan-600 dark:text-cyan-400 uppercase tracking-wider">
              Projects
            </th>
            <th className="px-4 py-3 text-right font-mono text-xs font-medium text-cyan-600 dark:text-cyan-400 uppercase tracking-wider">
              Eligible Nodes
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-cyan-500/10 dark:divide-cyan-400/10">
          {displayEpochs.map((epoch: EpochSummary) => (
            <tr
              key={epoch.epoch_id}
              role="button"
              tabIndex={0}
              onClick={() => onSelectEpoch(epoch.epoch_id)}
              onKeyDown={(e) => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  onSelectEpoch(epoch.epoch_id);
                }
              }}
              className={`
                cursor-pointer transition-colors select-none
                ${selectedEpochId === epoch.epoch_id
                  ? 'bg-cyan-500/15 dark:bg-cyan-500/25 border-l-4 border-cyan-400'
                  : 'hover:bg-cyan-500/5 dark:hover:bg-cyan-500/10'}
              `}
            >
              <td className="w-10 px-2 py-3 text-cyan-500 dark:text-cyan-400">
                {selectedEpochId === epoch.epoch_id ? '▼' : '▶'}
              </td>
              <td className="px-4 py-3 font-mono text-sm font-medium text-gray-900 dark:text-white">
                {epoch.epoch_id}
              </td>
              <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300">
                {formatTimestamp(epoch.timestamp)}
              </td>
              <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300 text-right">
                {epoch.slot_count}
              </td>
              <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300 text-right">
                {epoch.aggregated_projects}
              </td>
              <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-300 text-right">
                {epoch.eligible_nodes_count}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {/* Pagination */}
      {sortedEpochs.length > PAGE_SIZE && (
        <div className="mt-4 px-4 flex items-center justify-between border-t border-cyan-500/10 dark:border-cyan-400/10 pt-4">
          <p className="font-mono text-sm text-gray-500 dark:text-gray-400">
            {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, sortedEpochs.length)} of {sortedEpochs.length}
          </p>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className="px-3 py-1.5 rounded font-mono text-sm border border-cyan-500/30 dark:border-cyan-400/30 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-cyan-500/10 dark:hover:bg-cyan-500/20 transition-colors"
            >
              ← Prev
            </button>
            <span className="px-3 py-1.5 font-mono text-sm text-cyan-600 dark:text-cyan-400">
              {page + 1} / {totalPages}
            </span>
            <button
              type="button"
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1}
              className="px-3 py-1.5 rounded font-mono text-sm border border-cyan-500/30 dark:border-cyan-400/30 disabled:opacity-40 disabled:cursor-not-allowed hover:bg-cyan-500/10 dark:hover:bg-cyan-500/20 transition-colors"
            >
              Next →
            </button>
          </div>
        </div>
      )}
    </div>
  );
};

export default EpochActivityTable;
