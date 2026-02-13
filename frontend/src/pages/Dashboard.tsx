// Dashboard page showing total devices and online/offline counts.

import { useQuery } from '@tanstack/react-query';
import { getDevices } from '../api/devices';
import { Link } from 'react-router-dom';

function isOnline(lastSeen: string): boolean {
  const diff = Date.now() - new Date(lastSeen).getTime();
  return diff < 60 * 60 * 1000; // online if seen within the last hour
}

export default function Dashboard() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['devices'],
    queryFn: () => getDevices(),
  });

  if (isLoading) {
    return <p className="text-gray-500">Loading...</p>;
  }

  if (error) {
    return <p className="text-red-600">Failed to load dashboard data.</p>;
  }

  const devices = data?.devices ?? [];
  const total = devices.length;
  const online = devices.filter((d) => isOnline(d.last_seen)).length;
  const offline = total - online;

  return (
    <div>
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Dashboard</h1>

      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
        <StatCard label="Total Devices" value={total} color="blue" />
        <StatCard label="Online" value={online} color="green" />
        <StatCard label="Offline" value={offline} color="gray" />
      </div>

      {total > 0 && (
        <Link to="/devices" className="text-sm text-blue-600 hover:underline">
          View all devices &rarr;
        </Link>
      )}
    </div>
  );
}

function StatCard({ label, value, color }: { label: string; value: number; color: string }) {
  const colors: Record<string, string> = {
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    green: 'bg-green-50 text-green-700 border-green-200',
    gray: 'bg-gray-50 text-gray-700 border-gray-200',
  };
  return (
    <div className={`rounded-lg border p-5 ${colors[color] ?? colors.gray}`}>
      <p className="text-sm font-medium opacity-80">{label}</p>
      <p className="text-3xl font-bold mt-1">{value}</p>
    </div>
  );
}
