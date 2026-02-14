// Device list page with dark theme, hostname and OS search filters.

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { getDevices } from '../api/devices';

export default function DeviceList() {
  const [hostname, setHostname] = useState('');
  const [os, setOs] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['devices', hostname, os],
    queryFn: () => getDevices(hostname || undefined, os || undefined),
  });

  const devices = data?.devices ?? [];

  return (
    <div>
      <h1 className="text-xl font-semibold text-text-primary mb-6">Devices</h1>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          placeholder="Search hostname..."
          value={hostname}
          onChange={(e) => setHostname(e.target.value)}
          className="bg-bg-secondary border border-border-primary rounded px-3 py-2 text-sm text-text-primary placeholder-text-muted w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-accent"
        />
        <input
          type="text"
          placeholder="Filter by OS..."
          value={os}
          onChange={(e) => setOs(e.target.value)}
          className="bg-bg-secondary border border-border-primary rounded px-3 py-2 text-sm text-text-primary placeholder-text-muted w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-accent"
        />
      </div>

      {isLoading && <p className="text-text-muted">Loading...</p>}
      {error && <p className="text-danger">Failed to load devices.</p>}

      {!isLoading && !error && (
        <div className="bg-bg-secondary rounded-lg border border-border-primary overflow-hidden">
          <table className="min-w-full divide-y divide-border-primary">
            <thead className="bg-bg-tertiary">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Hostname
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  OS
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Agent
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Last Seen
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                  Status
                </th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-primary">
              {devices.map((d) => {
                const online =
                  Date.now() - new Date(d.last_seen).getTime() < 60 * 60 * 1000;
                return (
                  <tr key={d.id} className="hover:bg-bg-tertiary transition-colors">
                    <td className="px-4 py-3 text-sm">
                      <Link
                        to={`/devices/${d.id}`}
                        className="text-accent hover:underline font-medium"
                      >
                        {d.hostname}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm text-text-secondary">
                      {d.os_name} {d.os_version}
                    </td>
                    <td className="px-4 py-3 text-sm text-text-muted">
                      {d.agent_version || '—'}
                    </td>
                    <td className="px-4 py-3 text-sm text-text-muted">
                      {new Date(d.last_seen).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span
                        className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                          online
                            ? 'bg-success/15 text-success'
                            : 'bg-bg-tertiary text-text-muted'
                        }`}
                      >
                        <span
                          className={`w-1.5 h-1.5 rounded-full ${
                            online ? 'bg-success' : 'bg-text-muted'
                          }`}
                        />
                        {online ? 'Online' : 'Offline'}
                      </span>
                    </td>
                  </tr>
                );
              })}
              {devices.length === 0 && (
                <tr>
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-text-muted">
                    No devices found.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
