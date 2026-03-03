import { useState } from 'react';

const PipelineDiagram: React.FC = () => {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="bg-white/90 dark:bg-gray-800/90 backdrop-blur-sm rounded-lg shadow-sm transition-colors border border-cyan-500/10 dark:border-cyan-400/10 bg-gradient-to-br from-white to-cyan-50/30 dark:from-gray-800 dark:to-cyan-950/20">
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
            <div className="px-4 py-2 rounded-lg bg-gradient-to-r from-gray-100 to-gray-200 dark:from-gray-700 dark:to-gray-600 text-sm font-medium text-gray-900 dark:text-white">
              EpochReleased
            </div>
            <span className="text-cyan-500 dark:text-cyan-400">→</span>
            <div className="px-4 py-2 rounded-lg bg-gradient-to-r from-cyan-100 to-blue-100 dark:from-cyan-900/40 dark:to-blue-900/40 text-sm font-medium text-gray-900 dark:text-white border border-cyan-300/30 dark:border-cyan-500/30">
              Submission Window 45s
            </div>
            <span className="text-cyan-500 dark:text-cyan-400">→</span>
            <div className="px-4 py-2 rounded-lg bg-gradient-to-r from-amber-100 to-orange-100 dark:from-amber-900/40 dark:to-orange-900/40 text-sm font-medium text-gray-900 dark:text-white border border-amber-300/30 dark:border-amber-500/30">
              L1 Finalization
            </div>
            <span className="text-cyan-500 dark:text-cyan-400">→</span>
            <div className="px-4 py-2 rounded-lg bg-gradient-to-r from-green-100 to-emerald-100 dark:from-green-900/40 dark:to-emerald-900/40 text-sm font-medium text-gray-900 dark:text-white border border-green-300/30 dark:border-green-500/30">
              L2 Consensus 30s
            </div>
            <span className="text-cyan-500 dark:text-cyan-400">→</span>
            <div className="px-4 py-2 rounded-lg bg-gradient-to-r from-fuchsia-100 to-purple-100 dark:from-fuchsia-900/40 dark:to-purple-900/40 text-sm font-medium text-gray-900 dark:text-white border border-fuchsia-300/30 dark:border-fuchsia-500/30">
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
