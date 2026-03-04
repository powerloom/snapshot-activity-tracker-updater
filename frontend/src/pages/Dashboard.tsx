import { useState, useRef, useEffect } from 'react';
import { useDashboardSummary } from '../hooks/useDashboardData';
import PipelineDiagram from '../components/PipelineDiagram';
import EpochCharts from '../components/EpochCharts';
import EpochActivityTable from '../components/EpochActivityTable';
import EpochDetailView from '../components/EpochDetailView';

const Dashboard = () => {
  const { data: summary, isLoading: summaryLoading, error: summaryError } = useDashboardSummary();
  const [selectedEpochId, setSelectedEpochId] = useState<number | null>(null);
  const detailRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (selectedEpochId && detailRef.current) {
      detailRef.current.scrollIntoView({ behavior: 'smooth', block: 'nearest' });
    }
  }, [selectedEpochId]);

  return (
    <div className="min-h-screen bg-pl-bg">
      <header className="sticky top-0 z-50 bg-pl-bg-nav bg-opacity-95 border-b-2 border-pl-border">
        <div className="max-w-7xl mx-auto px-4 py-5 sm:px-6 lg:px-8 flex flex-wrap items-center gap-6">
          <img src="/logo.png" alt="Powerloom" className="w-[200px] max-md:w-[140px] h-auto" />
          <div>
            <h1 className="font-orbitron text-3xl font-bold tracking-tight text-white">
              BDS Data Market Activity
            </h1>
            <p className="mt-1.5 font-mono text-xs uppercase tracking-widest text-pl-text-muted">
              DSV validators & snapshotters on Powerloom Mainnet
            </p>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-8 sm:px-6 lg:px-8 space-y-8">
        {/* Summary Cards */}
        {!summaryError && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
            <div className="pl-card p-6">
              <div className="font-mono text-xs uppercase tracking-wider text-pl-text-muted mb-2">Total Epochs</div>
              <div className="font-orbitron text-3xl font-bold text-white tabular-nums">
                {summaryLoading ? '...' : summary?.total_epochs?.toLocaleString() || 0}
              </div>
            </div>
            <div className="pl-card p-6">
              <div className="font-mono text-xs uppercase tracking-wider text-pl-text-muted flex items-center gap-1 mb-2">
                DSV Validators
                <span title="Validators that produced L1 batches" className="cursor-help opacity-60">ⓘ</span>
              </div>
              <div className="font-orbitron text-3xl font-bold text-pl-accent tabular-nums">
                {summaryLoading ? '...' : summary?.total_validators || 0}
              </div>
            </div>
            <div className="pl-card p-6">
              <div className="font-mono text-xs uppercase tracking-wider text-pl-text-muted flex items-center gap-1 mb-2">
                Unique Slots
                <span title="Snapshotter (lite) nodes that have contributed to consensus" className="cursor-help opacity-60">ⓘ</span>
              </div>
              <div className="font-orbitron text-3xl font-bold text-pl-accent tabular-nums">
                {summaryLoading ? '...' : summary?.total_slots?.toLocaleString() || 0}
              </div>
            </div>
            <div className="pl-card p-6">
              <div className="font-mono text-xs uppercase tracking-wider text-pl-text-muted mb-2">Current Day</div>
              <div className="font-orbitron text-2xl font-bold text-white tabular-nums">
                {summaryLoading ? '...' : summary?.current_day || '—'}
              </div>
            </div>
          </div>
        )}

        {/* Pipeline Diagram (collapsible) */}
        <PipelineDiagram />

        {/* Epoch Charts */}
        <div className="pl-card p-6">
          <h2 className="font-orbitron text-lg font-semibold text-white mb-4 flex items-center gap-2">
            <span className="w-1 h-6 bg-pl-accent rounded-full" />
            Slots & Projects per Epoch
          </h2>
          <EpochCharts />
        </div>

        {/* Epoch Activity Table */}
        <div className="pl-card overflow-hidden">
          <div className="px-6 py-4 border-b-2 border-pl-border">
            <h2 className="font-orbitron text-lg font-semibold text-white flex items-center gap-2">
              <span className="w-1 h-6 bg-pl-accent rounded-full" />
              Epoch Activity
            </h2>
            <p className="mt-1 font-mono text-xs text-pl-text-muted">
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
          <div ref={detailRef} className="pl-card p-6">
            <EpochDetailView epochId={selectedEpochId} />
          </div>
        )}
      </main>
    </div>
  );
};

export default Dashboard;
