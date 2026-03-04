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
import { useEpochs } from '../hooks/useDashboardData';
import type { EpochSummary } from '../api/types';

const formatEpochLabel = (ts: number) => {
  const d = new Date(ts * 1000);
  return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit' });
};

const EpochCharts: React.FC = () => {
  const { data, isLoading, error } = useEpochs();
  const epochs = data?.epochs ?? [];

  const chartData = epochs.map((e: EpochSummary) => ({
    epoch: e.epoch_id,
    label: formatEpochLabel(e.timestamp),
    slots: e.slot_count,
    projects: e.aggregated_projects,
  }));

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-pl-accent" />
      </div>
    );
  }

  if (error || chartData.length === 0) {
    return (
      <div className="h-64 flex items-center justify-center text-pl-text-muted">
        No chart data available
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="h-64">
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
      </div>
    </div>
  );
};

export default EpochCharts;
