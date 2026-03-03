import { useEpochDetail } from '../hooks/useDashboardData';

interface EpochDetailViewProps {
  epochId: number;
}

const truncateId = (id: string, len = 12) =>
  id.length <= len ? id : `${id.slice(0, 6)}...${id.slice(-4)}`;

const truncateCid = (cid: string, len = 20) =>
  cid.length <= len ? cid : `${cid.slice(0, 10)}...${cid.slice(-6)}`;

const IPFS_GATEWAY = 'https://ipfs.io/ipfs/';

const EpochDetailView: React.FC<EpochDetailViewProps> = ({ epochId }) => {
  const { data, isLoading, error } = useEpochDetail(epochId);

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-12">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-blue-500" />
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="py-12 text-center text-red-500 dark:text-red-400">
        Error loading epoch detail
      </div>
    );
  }

  const slotEntries = Object.entries(data.submission_counts ?? {}).sort(
    ([, a], [, b]) => b - a
  );
  const projectEntries = Object.entries(data.aggregated_projects ?? {}).sort(
    ([a], [b]) => a.localeCompare(b)
  );
  const validators = data.validator_batches ?? [];

  return (
    <div className="space-y-6">
      <div>
        <h3 className="font-orbitron text-lg font-semibold text-gray-900 dark:text-white mb-2">
          Epoch {data.epoch_id} — {new Date(data.timestamp * 1000).toLocaleString()}
        </h3>
        <p className="font-mono text-sm text-gray-500 dark:text-gray-400">
          {data.total_validators} validators · {data.eligible_nodes_count} eligible nodes
        </p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Slots table */}
        <div className="bg-cyan-500/5 dark:bg-cyan-500/10 rounded-lg p-4 border border-cyan-500/10 dark:border-cyan-400/10">
          <h4 className="font-mono text-xs uppercase tracking-wider text-cyan-600 dark:text-cyan-400 mb-3">
            Slots (Submission Count)
          </h4>
          <div className="overflow-x-auto max-h-64 overflow-y-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr>
                  <th className="text-left py-2 text-gray-500 dark:text-gray-400">Slot ID</th>
                  <th className="text-right py-2 text-gray-500 dark:text-gray-400">Count</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {slotEntries.map(([slotId, count]) => (
                  <tr key={slotId}>
                    <td className="py-1.5 font-mono text-gray-900 dark:text-gray-100">
                      {truncateId(slotId)}
                    </td>
                    <td className="py-1.5 text-right text-gray-600 dark:text-gray-300">
                      {count}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Projects table */}
        <div className="bg-cyan-500/5 dark:bg-cyan-500/10 rounded-lg p-4 border border-cyan-500/10 dark:border-cyan-400/10">
          <h4 className="font-mono text-xs uppercase tracking-wider text-cyan-600 dark:text-cyan-400 mb-3">
            Projects (Winning CID)
          </h4>
          <div className="overflow-x-auto max-h-64 overflow-y-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr>
                  <th className="text-left py-2 text-gray-500 dark:text-gray-400">Project ID</th>
                  <th className="text-left py-2 text-gray-500 dark:text-gray-400">CID</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {projectEntries.map(([projectId, cid]) => (
                  <tr key={projectId}>
                    <td className="py-1.5 font-mono text-gray-900 dark:text-gray-100">
                      {truncateId(projectId)}
                    </td>
                    <td className="py-1.5">
                      <a
                        href={`${IPFS_GATEWAY}${cid}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="font-mono text-cyan-600 dark:text-cyan-400 hover:text-cyan-500 dark:hover:text-cyan-300 hover:underline"
                      >
                        {truncateCid(cid)}
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>

        {/* Validators table */}
        <div className="bg-cyan-500/5 dark:bg-cyan-500/10 rounded-lg p-4 border border-cyan-500/10 dark:border-cyan-400/10">
          <h4 className="font-mono text-xs uppercase tracking-wider text-cyan-600 dark:text-cyan-400 mb-3">
            Validators (Batch CID)
          </h4>
          <div className="overflow-x-auto max-h-64 overflow-y-auto">
            <table className="min-w-full text-sm">
              <thead>
                <tr>
                  <th className="text-left py-2 text-gray-500 dark:text-gray-400">Validator ID</th>
                  <th className="text-left py-2 text-gray-500 dark:text-gray-400">Batch CID</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200 dark:divide-gray-700">
                {validators.map((v) => (
                  <tr key={v.validator_id}>
                    <td className="py-1.5 font-mono text-gray-900 dark:text-gray-100">
                      {truncateId(v.validator_id)}
                    </td>
                    <td className="py-1.5">
                      <a
                        href={`${IPFS_GATEWAY}${v.batch_cid}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="font-mono text-cyan-600 dark:text-cyan-400 hover:text-cyan-500 dark:hover:text-cyan-300 hover:underline"
                      >
                        {truncateCid(v.batch_cid)}
                      </a>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </div>
  );
};

export default EpochDetailView;
