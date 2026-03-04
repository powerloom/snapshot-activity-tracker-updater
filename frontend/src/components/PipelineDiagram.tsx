import { useState } from 'react';

const PipelineDiagram: React.FC = () => {
  const [isExpanded, setIsExpanded] = useState(false);

  return (
    <div className="pl-card overflow-hidden">
      <button
        onClick={() => setIsExpanded(!isExpanded)}
        className="w-full px-6 py-4 flex items-center justify-between text-left hover:bg-pl-bg-elevated rounded-lg transition-all duration-300"
      >
        <h2 className="font-orbitron text-lg font-semibold text-white flex items-center gap-2">
          <span className="w-1 h-5 bg-pl-accent rounded-full" />
          How DSV works
        </h2>
        <span className="text-pl-text-muted">
          {isExpanded ? '▼' : '▶'}
        </span>
      </button>
      {isExpanded && (
        <div className="px-6 pb-6 pt-2 border-t-2 border-pl-border">
          <div className="flex flex-wrap items-center gap-2 flex-col sm:flex-row">
            <div className="px-4 py-2 rounded-lg bg-pl-bg-elevated text-sm font-medium text-white border-2 border-pl-border">
              EpochReleased
            </div>
            <span className="text-pl-accent">→</span>
            <div className="px-4 py-2 rounded-lg bg-pl-bg-elevated text-sm font-medium text-white border-2 border-pl-accent/30">
              Submission Window 45s
            </div>
            <span className="text-pl-accent">→</span>
            <div className="px-4 py-2 rounded-lg bg-pl-bg-elevated text-sm font-medium text-white border-2 border-orange-500/30">
              L1 Finalization
            </div>
            <span className="text-pl-accent">→</span>
            <div className="px-4 py-2 rounded-lg bg-pl-bg-elevated text-sm font-medium text-white border-2 border-pl-accent/50">
              L2 Consensus 30s
            </div>
            <span className="text-pl-accent">→</span>
            <div className="px-4 py-2 rounded-lg bg-pl-bg-elevated text-sm font-medium text-white border-2 border-pl-accent/30">
              On-Chain
            </div>
          </div>
          <p className="mt-4 text-sm text-pl-text-muted">
            Snapshotters submit CIDs during the 45s window. Validators aggregate L1 batches, then reach consensus in L2 (30s window). Final aggregated batch is committed on-chain.
          </p>
        </div>
      )}
    </div>
  );
};

export default PipelineDiagram;
