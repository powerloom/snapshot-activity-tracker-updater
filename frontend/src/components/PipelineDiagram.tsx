import { useState } from 'react';

const PipelineDiagram: React.FC = () => {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow transition-colors">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-gray-50 dark:hover:bg-gray-700/50 rounded-lg transition-colors"
      >
        <h2 className="text-lg font-medium text-gray-900 dark:text-white">
          How DSV works
        </h2>
        <span className="text-gray-500 dark:text-gray-400">
          {isExpanded ? '▼' : '▶'}
        </span>
      </button>
      {isExpanded && (
        <div className="px-6 pb-6 pt-2 border-t border-gray-200 dark:border-gray-700">
          <div className="flex flex-wrap items-center gap-2 flex-col sm:flex-row">
            <div className="px-4 py-2 rounded-lg bg-gray-100 dark:bg-gray-700 text-sm font-medium text-gray-900 dark:text-white">
              EpochReleased
            </div>
            <span className="text-gray-400 dark:text-gray-500">→</span>
            <div className="px-4 py-2 rounded-lg bg-blue-100 dark:bg-blue-900/30 text-sm font-medium text-gray-900 dark:text-white">
              Submission Window 45s
            </div>
            <span className="text-gray-400 dark:text-gray-500">→</span>
            <div className="px-4 py-2 rounded-lg bg-amber-100 dark:bg-amber-900/30 text-sm font-medium text-gray-900 dark:text-white">
              L1 Finalization
            </div>
            <span className="text-gray-400 dark:text-gray-500">→</span>
            <div className="px-4 py-2 rounded-lg bg-green-100 dark:bg-green-900/30 text-sm font-medium text-gray-900 dark:text-white">
              L2 Consensus 30s
            </div>
            <span className="text-gray-400 dark:text-gray-500">→</span>
            <div className="px-4 py-2 rounded-lg bg-purple-100 dark:bg-purple-900/30 text-sm font-medium text-gray-900 dark:text-white">
              On-Chain
            </div>
          </div>
          <p className="mt-4 text-sm text-gray-500 dark:text-gray-400">
            Snapshotters submit CIDs during the 45s window. Validators aggregate L1 batches, then reach consensus in L2 (30s window). Final aggregated batch is committed on-chain.
          </p>
        </div>
      )}
    </div>
  );
};

export default PipelineDiagram;
