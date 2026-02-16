// Device list page with pagination, sorting, debounced filters, status/department filter, and CSV export.

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { getDevices, exportDevicesCSV } from '../api/devices';
import { getDepartments } from '../api/departments';
import { useDebounce } from '../hooks/useDebounce';

type SortCol = 'hostname' | 'os' | 'last_seen' | 'status';
type SortOrder = 'asc' | 'desc';

export default function DeviceList() {
  const [hostname, setHostname] = useState('');
  const [os, setOs] = useState('');
  const [status, setStatus] = useState('');
  const [departmentId, setDepartmentId] = useState('');
  const [sort, setSort] = useState<SortCol>('hostname');
  const [order, setOrder] = useState<SortOrder>('asc');
  const [page, setPage] = useState(1);
  const limit = 50;

  const debouncedHostname = useDebounce(hostname, 300);
  const debouncedOs = useDebounce(os, 300);

  const { data: deptData } = useQuery({
    queryKey: ['departments'],
    queryFn: getDepartments,
  });

  const { data, isLoading, error } = useQuery({
    queryKey: ['devices', debouncedHostname, debouncedOs, status, departmentId, sort, order, page],
    queryFn: () =>
      getDevices({
        hostname: debouncedHostname || undefined,
        os: debouncedOs || undefined,
        status: status || undefined,
        department_id: departmentId || undefined,
        sort,
        order,
        page,
        limit,
      }),
  });

  const devices = data?.devices ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / limit));
  const departments = deptData?.departments ?? [];

  const toggleSort = (col: SortCol) => {
    if (sort === col) {
      setOrder(order === 'asc' ? 'desc' : 'asc');
    } else {
      setSort(col);
      setOrder('asc');
    }
    setPage(1);
  };

  const SortIcon = ({ col }: { col: SortCol }) => {
    if (sort !== col) return <span className="text-text-muted/40 ml-1">↕</span>;
    return <span className="text-accent ml-1">{order === 'asc' ? '↑' : '↓'}</span>;
  };

  // Reset page when filters change.
  const handleHostname = (v: string) => { setHostname(v); setPage(1); };
  const handleOs = (v: string) => { setOs(v); setPage(1); };
  const handleStatus = (v: string) => { setStatus(v); setPage(1); };
  const handleDepartment = (v: string) => { setDepartmentId(v); setPage(1); };

  const handleExportCSV = () => {
    exportDevicesCSV({
      hostname: debouncedHostname || undefined,
      os: debouncedOs || undefined,
      status: status || undefined,
      department_id: departmentId || undefined,
      sort,
      order,
    });
  };

  return (
    <div>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-text-primary">Devices</h1>
        <div className="flex items-center gap-3">
          <button
            onClick={handleExportCSV}
            className="px-3 py-1.5 text-sm bg-bg-secondary border border-border-primary rounded hover:bg-bg-tertiary text-text-primary transition-colors cursor-pointer flex items-center gap-1.5"
          >
            <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
            </svg>
            Export CSV
          </button>
          <span className="text-sm text-text-muted">{total} total</span>
        </div>
      </div>

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <input
          type="text"
          placeholder="Search hostname..."
          value={hostname}
          onChange={(e) => handleHostname(e.target.value)}
          className="bg-bg-secondary border border-border-primary rounded px-3 py-2 text-sm text-text-primary placeholder-text-muted w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-accent"
        />
        <input
          type="text"
          placeholder="Filter by OS..."
          value={os}
          onChange={(e) => handleOs(e.target.value)}
          className="bg-bg-secondary border border-border-primary rounded px-3 py-2 text-sm text-text-primary placeholder-text-muted w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-accent"
        />
        <select
          value={status}
          onChange={(e) => handleStatus(e.target.value)}
          className="bg-bg-secondary border border-border-primary rounded px-3 py-2 text-sm text-text-primary w-full sm:w-40 focus:outline-none focus:ring-2 focus:ring-accent cursor-pointer"
        >
          <option value="">All Active</option>
          <option value="online">Online</option>
          <option value="offline">Offline</option>
          <option value="inactive">Inactive</option>
        </select>
        <select
          value={departmentId}
          onChange={(e) => handleDepartment(e.target.value)}
          className="bg-bg-secondary border border-border-primary rounded px-3 py-2 text-sm text-text-primary w-full sm:w-48 focus:outline-none focus:ring-2 focus:ring-accent cursor-pointer"
        >
          <option value="">All Departments</option>
          {departments.map((d) => (
            <option key={d.id} value={d.id}>{d.name}</option>
          ))}
        </select>
      </div>

      {error && <p className="text-danger mb-4">Failed to load devices.</p>}

      <div className="bg-bg-secondary rounded-lg border border-border-primary overflow-hidden">
        <table className="min-w-full divide-y divide-border-primary">
          <thead className="bg-bg-tertiary">
            <tr>
              {([
                ['hostname', 'Hostname'],
                ['os', 'OS'],
              ] as [SortCol, string][]).map(([col, label]) => (
                <th
                  key={col}
                  onClick={() => toggleSort(col)}
                  className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider cursor-pointer select-none hover:text-text-primary transition-colors"
                >
                  {label}<SortIcon col={col} />
                </th>
              ))}
              <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                Department
              </th>
              <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
                Agent
              </th>
              <th
                onClick={() => toggleSort('last_seen')}
                className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider cursor-pointer select-none hover:text-text-primary transition-colors"
              >
                Last Seen<SortIcon col="last_seen" />
              </th>
              <th
                onClick={() => toggleSort('status')}
                className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider cursor-pointer select-none hover:text-text-primary transition-colors"
              >
                Status<SortIcon col="status" />
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-primary">
            {isLoading
              ? Array.from({ length: 8 }).map((_, i) => (
                  <tr key={i}>
                    {Array.from({ length: 6 }).map((_, j) => (
                      <td key={j} className="px-4 py-3">
                        <div className="h-4 bg-bg-tertiary rounded animate-pulse" style={{ width: `${50 + Math.random() * 40}%` }} />
                      </td>
                    ))}
                  </tr>
                ))
              : devices.map((d) => {
                  const isInactive = d.status === 'inactive';
                  const online = !isInactive && Date.now() - new Date(d.last_seen).getTime() < 60 * 60 * 1000;
                  const statusLabel = isInactive ? 'Inactive' : online ? 'Online' : 'Offline';
                  const statusClass = isInactive
                    ? 'bg-warning/15 text-warning'
                    : online
                      ? 'bg-success/15 text-success'
                      : 'bg-bg-tertiary text-text-muted';
                  const dotClass = isInactive ? 'bg-warning' : online ? 'bg-success' : 'bg-text-muted';

                  return (
                    <tr key={d.id} className="hover:bg-bg-tertiary transition-colors">
                      <td className="px-4 py-3 text-sm">
                        <Link to={`/devices/${d.id}`} className="text-accent hover:underline font-medium">
                          {d.hostname}
                        </Link>
                      </td>
                      <td className="px-4 py-3 text-sm text-text-secondary">
                        {d.os_name} {d.os_version}
                      </td>
                      <td className="px-4 py-3 text-sm text-text-muted">{d.department_name || '—'}</td>
                      <td className="px-4 py-3 text-sm text-text-muted">{d.agent_version || '—'}</td>
                      <td className="px-4 py-3 text-sm text-text-muted">
                        {new Date(d.last_seen).toLocaleString()}
                      </td>
                      <td className="px-4 py-3 text-sm">
                        <span
                          className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium ${statusClass}`}
                        >
                          <span className={`w-1.5 h-1.5 rounded-full ${dotClass}`} />
                          {statusLabel}
                        </span>
                      </td>
                    </tr>
                  );
                })}
            {!isLoading && devices.length === 0 && (
              <tr>
                <td colSpan={6} className="px-4 py-8 text-center text-sm text-text-muted">
                  No devices found.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <p className="text-sm text-text-muted">
            Page {page} of {totalPages} ({total} devices)
          </p>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
              className="px-3 py-1.5 text-sm bg-bg-secondary border border-border-primary rounded hover:bg-bg-tertiary disabled:opacity-40 disabled:cursor-not-allowed text-text-primary transition-colors cursor-pointer"
            >
              Previous
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
              className="px-3 py-1.5 text-sm bg-bg-secondary border border-border-primary rounded hover:bg-bg-tertiary disabled:opacity-40 disabled:cursor-not-allowed text-text-primary transition-colors cursor-pointer"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
