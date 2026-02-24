import { useDashboardSummary, useNetworkTopology } from '../hooks/useDashboardData';
import NetworkTopology from '../components/NetworkTopology';

const Dashboard = () => {
  const { data: summary, isLoading: summaryLoading, error: summaryError } = useDashboardSummary();
  const { data: topology, isLoading: topologyLoading, error: topologyError } = useNetworkTopology();

  return (
    <div className="min-h-screen bg-gray-50">
      <header className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 py-4 sm:px-6 lg:px-8">
          <h1 className="text-3xl font-bold text-gray-900">DSV Network Activity Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">Real-time visualization of decentralized sequencer validator activity</p>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 py-6 sm:px-6 lg:px-8">
        {/* Summary Cards */}
        {!summaryError && (
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-6">
            <div className="bg-white rounded-lg shadow p-6">
              <div className="text-sm font-medium text-gray-500">Total Epochs</div>
              <div className="mt-2 text-3xl font-semibold text-gray-900">
                {summaryLoading ? '...' : summary?.total_epochs || 0}
              </div>
            </div>
            <div className="bg-white rounded-lg shadow p-6">
              <div className="text-sm font-medium text-gray-500">Total Validators</div>
              <div className="mt-2 text-3xl font-semibold text-green-600">
                {summaryLoading ? '...' : summary?.total_validators || 0}
              </div>
            </div>
            <div className="bg-white rounded-lg shadow p-6">
              <div className="text-sm font-medium text-gray-500">Active Slots</div>
              <div className="mt-2 text-3xl font-semibold text-blue-600">
                {summaryLoading ? '...' : summary?.total_slots || 0}
              </div>
            </div>
            <div className="bg-white rounded-lg shadow p-6">
              <div className="text-sm font-medium text-gray-500">Current Day</div>
              <div className="mt-2 text-lg font-semibold text-gray-900">
                {summaryLoading ? '...' : summary?.current_day || 'N/A'}
              </div>
            </div>
          </div>
        )}

        {/* Network Topology */}
        <div className="bg-white rounded-lg shadow">
          <div className="px-6 py-4 border-b border-gray-200">
            <h2 className="text-lg font-medium text-gray-900">Network Topology</h2>
            <p className="mt-1 text-sm text-gray-500">
              Interactive visualization of validators, slots, and projects
            </p>
          </div>
          <div className="p-6">
            <NetworkTopology data={topology} isLoading={topologyLoading} error={topologyError} />
          </div>
        </div>
      </main>
    </div>
  );
};

export default Dashboard;
