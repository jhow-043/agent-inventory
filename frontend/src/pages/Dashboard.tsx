// Dashboard page with dark theme showing total devices and online/offline counts.

import { useQuery } from '@tanstack/react-query';
import { getStats } from '../api/dashboard';
import { Link } from 'react-router-dom';

export default function Dashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['dashboard-stats'],
    queryFn: getStats,
  });

  if (isLoading) {
    return <p className="text-text-muted">Loading...</p>;
  }

  if (error) {
    return <p className="text-danger">Failed to load dashboard data.</p>;
  }

  const total = data?.total ?? 0;
  const online = data?.online ?? 0;
  const offline = data?.offline ?? 0;
  const inactive = data?.inactive ?? 0;

  return (
    <div>
      <h1 className="text-xl font-semibold text-text-primary mb-6">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 mb-8">
        <StatCard label="Total Devices" value={total} color="blue" />
        <StatCard label="Online" value={online} color="green" />
        <StatCard label="Offline" value={offline} color="gray" />
        <StatCard label="Inactive" value={inactive} color="yellow" />
      </div>

      {total > 0 && (
        <Link to="/devices" className="text-sm text-accent hover:underline">
          View all devices &rarr;
        </Link>
      )}
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  const colors: Record<string, string> = {
    blue: 'bg-accent/10 text-accent border-accent/20',
    green: 'bg-success/10 text-success border-success/20',
    gray: 'bg-bg-tertiary text-text-secondary border-border-primary',
    yellow: 'bg-warning/10 text-warning border-warning/20',
  };
  return (
    <div className={`rounded-lg border p-5 ${colors[color] ?? colors.gray}`}>
      <p className="text-sm font-medium opacity-80">{label}</p>
      <p className="text-3xl font-bold mt-1">{value}</p>
    </div>
  );
}
