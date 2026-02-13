// Device list page with hostname and OS search filters.

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
      <h1 className="text-xl font-semibold text-gray-900 mb-6">Devices</h1>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          placeholder="Search hostname..."
          value={hostname}
          onChange={(e) => setHostname(e.target.value)}
          className="border border-gray-300 rounded px-3 py-2 text-sm w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        <input
          type="text"
          placeholder="Filter by OS..."
          value={os}
          onChange={(e) => setOs(e.target.value)}
          className="border border-gray-300 rounded px-3 py-2 text-sm w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
      </div>

      {isLoading && <p className="text-gray-500">Loading...</p>}
      {error && <p className="text-red-600">Failed to load devices.</p>}

      {!isLoading && !error && (
        <div className="bg-white rounded-lg shadow overflow-hidden">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Hostname
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  OS
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Agent
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Last Seen
                </th>
                <th className="px-4 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Status
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-100">
              {devices.map((d) => {
                const online =
                  Date.now() - new Date(d.last_seen).getTime() < 60 * 60 * 1000;
                return (
                  <tr key={d.id} className="hover:bg-gray-50 transition-colors">
                    <td className="px-4 py-3 text-sm">
                      <Link
                        to={`/devices/${d.id}`}
                        className="text-blue-600 hover:underline font-medium"
                      >
                        {d.hostname}
                      </Link>
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-700">
                      {d.os_name} {d.os_version}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {d.agent_version || '—'}
                    </td>
                    <td className="px-4 py-3 text-sm text-gray-500">
                      {new Date(d.last_seen).toLocaleString()}
                    </td>
                    <td className="px-4 py-3 text-sm">
                      <span
                        className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${
                          online
                            ? 'bg-green-100 text-green-700'
                            : 'bg-gray-100 text-gray-600'
                        }`}
                      >
                        <span
                          className={`w-1.5 h-1.5 rounded-full ${
                            online ? 'bg-green-500' : 'bg-gray-400'
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
                  <td colSpan={5} className="px-4 py-8 text-center text-sm text-gray-400">
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
