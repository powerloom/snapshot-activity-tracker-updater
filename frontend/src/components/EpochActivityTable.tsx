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
  const totalPages = Math.ceil(epochs.length / PAGE_SIZE) || 1;
  const displayEpochs = epochs.slice(page * PAGE_SIZE, page * PAGE_SIZE + PAGE_SIZE);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-16">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-pl-accent" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="py-16 text-center text-red-400">
        Error loading epochs
      </div>
    );
  }

  if (epochs.length === 0) {
    return (
      <div className="py-16 text-center text-pl-text-muted">
        No epoch data available
      </div>
    );
  }

  return (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-pl-border-subtle">
        <thead>
          <tr>
            <th className="w-10 px-2 py-3" aria-label="Expand" />
            <th className="px-4 py-3 text-left font-mono text-xs font-medium text-pl-text-muted uppercase tracking-wider">
              Epoch ID
            </th>
            <th className="px-4 py-3 text-left font-mono text-xs font-medium text-pl-text-muted uppercase tracking-wider">
              Timestamp
            </th>
            <th className="px-4 py-3 text-right font-mono text-xs font-medium text-pl-text-muted uppercase tracking-wider">
              Slots
            </th>
            <th className="px-4 py-3 text-right font-mono text-xs font-medium text-pl-text-muted uppercase tracking-wider">
              Projects
            </th>
            <th className="px-4 py-3 text-right font-mono text-xs font-medium text-pl-text-muted uppercase tracking-wider">
              Eligible Nodes
            </th>
          </tr>
        </thead>
        <tbody className="divide-y divide-pl-border-subtle">
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
                  ? 'bg-pl-accent/15 border-l-4 border-pl-accent'
                  : 'hover:bg-pl-bg-elevated'}
              `}
            >
              <td className="w-10 px-2 py-3 text-pl-accent">
                {selectedEpochId === epoch.epoch_id ? '▼' : '▶'}
              </td>
              <td className="px-4 py-3 font-mono text-sm font-medium text-white">
                {epoch.epoch_id}
              </td>
              <td className="px-4 py-3 text-sm text-pl-text-muted">
                {formatTimestamp(epoch.timestamp)}
              </td>
              <td className="px-4 py-3 text-sm text-pl-text-muted text-right">
                {epoch.slot_count}
              </td>
              <td className="px-4 py-3 text-sm text-pl-text-muted text-right">
                {epoch.aggregated_projects}
              </td>
              <td className="px-4 py-3 text-sm text-pl-accent text-right">
                {epoch.eligible_nodes_count}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {epochs.length > PAGE_SIZE && (
        <div className="mt-4 px-4 flex items-center justify-between border-t border-pl-border-subtle pt-4">
          <p className="font-mono text-sm text-pl-text-muted">
            {page * PAGE_SIZE + 1}–{Math.min((page + 1) * PAGE_SIZE, epochs.length)} of {epochs.length}
          </p>
          <div className="flex gap-2">
            <button
              type="button"
              onClick={() => setPage((p) => Math.max(0, p - 1))}
              disabled={page === 0}
              className="px-3 py-1.5 rounded font-mono text-sm border-2 border-pl-border bg-pl-bg-elevated text-white disabled:opacity-40 disabled:cursor-not-allowed hover:border-pl-accent/50 hover:bg-pl-accent/10 transition-colors"
            >
              ← Prev
            </button>
            <span className="px-3 py-1.5 font-mono text-sm text-pl-accent">
              {page + 1} / {totalPages}
            </span>
            <button
              type="button"
              onClick={() => setPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={page >= totalPages - 1}
              className="px-3 py-1.5 rounded font-mono text-sm border-2 border-pl-border bg-pl-bg-elevated text-white disabled:opacity-40 disabled:cursor-not-allowed hover:border-pl-accent/50 hover:bg-pl-accent/10 transition-colors"
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
