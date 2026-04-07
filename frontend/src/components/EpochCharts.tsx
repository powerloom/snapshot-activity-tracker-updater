import { useState } from 'react';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { useChartEpochs } from '../hooks/useDashboardData';
import type { EpochSummary } from '../api/types';

const formatEpochLabel = (ts: number) => {
  const d = new Date(ts * 1000);
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit' });
};

/** Most recent N epochs; times assume ~12s/epoch (~7200/day). limit 0 = server “all retained”. */
const CHART_RANGE_OPTIONS: { limit: number; label: string }[] = [
  { limit: 500, label: 'Last 500 (~1 h)' },
  { limit: 2000, label: 'Last 2k (~3 h)' },
  { limit: 7200, label: 'Last ~1 day' },
  { limit: 14400, label: 'Last ~2 days (max page)' },
  { limit: 0, label: 'All retained' },
];

const DEFAULT_CHART_LIMIT = 2000;

const EpochCharts: React.FC = () => {
  const [chartLimit, setChartLimit] = useState(DEFAULT_CHART_LIMIT);
  const { data, isLoading, error } = useChartEpochs(chartLimit);
  const epochs = data?.epochs ?? [];

  const chartData = epochs.map((e: EpochSummary) => ({
    epoch: e.epoch_id,
    label: formatEpochLabel(e.timestamp),
    slots: e.slot_count,
    projects: e.aggregated_projects,
  }));

  const header = (
    <div className="flex flex-wrap items-center justify-between gap-3">
      <p className="font-mono text-xs text-pl-text-muted">
        {!isLoading && !error && chartData.length > 0
          ? chartLimit === 0
            ? `Showing all ${chartData.length.toLocaleString()} epochs in retention`
            : `Showing the most recent ${chartData.length.toLocaleString()} epochs`
          : isLoading
            ? 'Loading…'
            : 'No data'}
      </p>
      <label className="flex items-center gap-2 font-mono text-xs text-pl-text-muted">
        <span className="shrink-0">Range</span>
        <select
          value={chartLimit}
          onChange={(e) => setChartLimit(Number(e.target.value))}
          className="rounded-lg border-2 border-pl-border bg-[var(--pl-bg-input)] px-3 py-1.5 text-sm text-white focus:border-pl-accent focus:outline-none"
        >
          {CHART_RANGE_OPTIONS.map((o) => (
            <option key={o.limit} value={o.limit}>
              {o.label}
            </option>
          ))}
        </select>
      </label>
    </div>
  );

  if (error) {
    return (
      <div className="space-y-4">
        {header}
        <div className="h-64 flex items-center justify-center text-pl-text-muted">
          No chart data available
        </div>
      </div>
    );
  }

  if (!isLoading && chartData.length === 0) {
    return (
      <div className="space-y-4">
        {header}
        <div className="h-64 flex items-center justify-center text-pl-text-muted">
          No chart data available
        </div>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {header}
      <div className="h-64">
        {isLoading ? (
          <div className="flex h-full items-center justify-center">
            <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-pl-accent" />
          </div>
        ) : (
          <ResponsiveContainer width="100%" height="100%">
          <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
            <defs>
              <linearGradient id="slotsGradient" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stopColor="#54e794" />
                <stop offset="100%" stopColor="#3b82f6" />
              </linearGradient>
              <linearGradient id="projectsGradient" x1="0" y1="0" x2="1" y2="0">
                <stop offset="0%" stopColor="#54e794" />
                <stop offset="100%" stopColor="#b4ccc5" />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#384949" strokeOpacity={0.6} />
            <XAxis
              dataKey="label"
              tick={{ fontSize: 11, fill: '#b4ccc5' }}
            />
            <YAxis
              tick={{ fontSize: 11, fill: '#b4ccc5' }}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: '#151818',
                border: '1px solid #384949',
                borderRadius: '0.5rem',
                color: '#fff',
              }}
              labelStyle={{ color: '#b4ccc5' }}
              formatter={(value: number) => [value, '']}
            />
            <Legend wrapperStyle={{ color: '#b4ccc5' }} />
            <Line
              type="monotone"
              dataKey="slots"
              name="Slots"
              stroke="url(#slotsGradient)"
              strokeWidth={2.5}
              dot={false}
            />
            <Line
              type="monotone"
              dataKey="projects"
              name="Projects"
              stroke="url(#projectsGradient)"
              strokeWidth={2.5}
              dot={false}
            />
          </LineChart>
        </ResponsiveContainer>
        )}
      </div>
    </div>
  );
};

export default EpochCharts;
