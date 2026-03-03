import { useEpochs } from '../hooks/useDashboardData';
import type { EpochSummary } from '../api/types';

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
  const epochs = data?.epochs ?? [];
  const displayEpochs = epochs.slice(0, 50);

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
              onClick={() => onSelectEpoch(epoch.epoch_id)}
              className={`
                cursor-pointer transition-colors
                ${selectedEpochId === epoch.epoch_id
                  ? 'bg-cyan-500/15 dark:bg-cyan-500/25 border-l-2 border-cyan-400'
                  : 'hover:bg-cyan-500/5 dark:hover:bg-cyan-500/10'}
              `}
            >
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
      {epochs.length > 50 && (
        <p className="mt-2 text-xs text-gray-500 dark:text-gray-400 px-4">
          Showing 50 of {epochs.length} epochs
        </p>
      )}
    </div>
  );
};

export default EpochActivityTable;
