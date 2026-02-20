// Device list page with pagination, sorting, debounced filters, status/department filter, bulk actions, and CSV export.

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { getDevices, exportDevicesCSV, bulkUpdateStatus, bulkUpdateDepartment, bulkDeleteDevices } from '../api/devices';
import { getDepartments } from '../api/departments';
import { useDebounce } from '../hooks/useDebounce';
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../hooks/useToast';
import { PageHeader, Button, Badge, Input, Select, Modal } from '../components/ui';

type SortCol = 'hostname' | 'os' | 'last_seen' | 'status';
type SortOrder = 'asc' | 'desc';

export default function DeviceList() {
  const { role } = useAuth();
  const toast = useToast();
  const queryClient = useQueryClient();
  const [hostname, setHostname] = useState('');
  const [os, setOs] = useState('');
  const [status, setStatus] = useState('');
  const [departmentId, setDepartmentId] = useState('');
  const [sort, setSort] = useState<SortCol>('hostname');
  const [order, setOrder] = useState<SortOrder>('asc');
  const [page, setPage] = useState(1);
  const limit = 50;

  // Selection state
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());
  const [bulkModal, setBulkModal] = useState<'delete' | 'deactivate' | 'activate' | 'department' | null>(null);
  const [bulkDeptId, setBulkDeptId] = useState('');

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

  // Bulk mutations
  const invalidateAll = () => {
    queryClient.invalidateQueries({ queryKey: ['devices'] });
    queryClient.invalidateQueries({ queryKey: ['dashboard-stats'] });
    setSelectedIds(new Set());
    setBulkModal(null);
  };

  const bulkStatusMutation = useMutation({
    mutationFn: (status: 'active' | 'inactive') => bulkUpdateStatus([...selectedIds], status),
    onSuccess: (data) => { invalidateAll(); toast.success(data.message || `${data.affected} dispositivos atualizados`); },
    onError: () => toast.error('Falha ao atualizar status dos dispositivos'),
  });

  const bulkDeptMutation = useMutation({
    mutationFn: (deptId: string | null) => bulkUpdateDepartment([...selectedIds], deptId),
    onSuccess: (data) => { invalidateAll(); toast.success(data.message || `${data.affected} dispositivos atualizados`); },
    onError: () => toast.error('Falha ao atualizar departamento'),
  });

  const bulkDeleteMutation = useMutation({
    mutationFn: () => bulkDeleteDevices([...selectedIds]),
    onSuccess: (data) => { invalidateAll(); toast.success(data.message || `${data.affected} dispositivos excluídos`); },
    onError: () => toast.error('Falha ao excluir dispositivos'),
  });

  // Selection helpers
  const toggleSelect = (id: string) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  };

  const toggleSelectAll = () => {
    if (selectedIds.size === devices.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(devices.map((d) => d.id)));
    }
  };

  const selectionCount = selectedIds.size;
  const allSelected = devices.length > 0 && selectionCount === devices.length;

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
    <div className="animate-fade-in">
      <PageHeader
        title="Devices"
        actions={
          <>
            <Button
              variant="secondary"
              size="sm"
              onClick={handleExportCSV}
              icon={
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3" />
                </svg>
              }
            >
              Export CSV
            </Button>
            <span className="text-sm text-text-muted">{total} total</span>
          </>
        }
      />

      {/* Filters */}
      <div className="flex flex-col sm:flex-row gap-3 mb-6">
        <div className="w-full sm:w-64">
          <Input
            placeholder="Search hostname..."
            value={hostname}
            onChange={(e) => handleHostname(e.target.value)}
            icon={
              <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M21 21l-5.197-5.197m0 0A7.5 7.5 0 105.196 5.196a7.5 7.5 0 0010.607 10.607z" />
              </svg>
            }
          />
        </div>
        <div className="w-full sm:w-64">
          <Input
            placeholder="Filter by OS..."
            value={os}
            onChange={(e) => handleOs(e.target.value)}
          />
        </div>
        <div className="w-full sm:w-40">
          <Select value={status} onChange={(e) => handleStatus(e.target.value)}>
            <option value="">All Active</option>
            <option value="online">Online</option>
            <option value="offline">Offline</option>
            <option value="inactive">Inactive</option>
          </Select>
        </div>
        <div className="w-full sm:w-48">
          <Select value={departmentId} onChange={(e) => handleDepartment(e.target.value)}>
            <option value="">All Departments</option>
            {departments.map((d) => (
              <option key={d.id} value={d.id}>{d.name}</option>
            ))}
          </Select>
        </div>
      </div>

      {error && <p className="text-danger mb-4">Failed to load devices.</p>}

      {/* Bulk action toolbar — admin only */}
      {role === 'admin' && selectionCount > 0 && (
        <div className="flex items-center gap-3 mb-4 p-3 bg-accent/5 border border-accent/20 rounded-xl animate-fade-in">
          <span className="text-sm font-medium text-text-primary">
            {selectionCount} selected
          </span>
          <div className="h-4 w-px bg-border-primary" />
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setBulkModal('department')}
            icon={
              <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008z" />
              </svg>
            }
          >
            Set Department
          </Button>
          <Button
            variant="success"
            size="sm"
            onClick={() => setBulkModal('activate')}
            icon={
              <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            }
          >
            Activate
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => setBulkModal('deactivate')}
            icon={
              <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M5.636 5.636a9 9 0 1012.728 0M12 3v9" />
              </svg>
            }
          >
            Deactivate
          </Button>
          <Button
            variant="danger"
            size="sm"
            onClick={() => setBulkModal('delete')}
            icon={
              <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
              </svg>
            }
          >
            Delete
          </Button>
          <div className="flex-1" />
          <Button variant="ghost" size="sm" onClick={() => setSelectedIds(new Set())}>
            Clear selection
          </Button>
        </div>
      )}

      {/* Bulk modals */}
      <Modal
        open={bulkModal === 'delete'}
        onClose={() => setBulkModal(null)}
        title="Delete Devices"
        actions={
          <>
            <Button variant="ghost" size="sm" onClick={() => setBulkModal(null)}>Cancel</Button>
            <Button
              variant="danger"
              size="sm"
              onClick={() => bulkDeleteMutation.mutate()}
              loading={bulkDeleteMutation.isPending}
              icon={
                <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                </svg>
              }
            >
              Delete {selectionCount} device(s)
            </Button>
          </>
        }
      >
        <p className="text-sm text-text-secondary">
          Are you sure you want to permanently delete <strong className="text-text-primary">{selectionCount} device(s)</strong>?
          All related data will be removed. This action cannot be undone.
        </p>
      </Modal>

      <Modal
        open={bulkModal === 'activate'}
        onClose={() => setBulkModal(null)}
        title="Activate Devices"
        actions={
          <>
            <Button variant="ghost" size="sm" onClick={() => setBulkModal(null)}>Cancel</Button>
            <Button
              variant="success"
              size="sm"
              onClick={() => bulkStatusMutation.mutate('active')}
              loading={bulkStatusMutation.isPending}
              icon={
                <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 12.75L11.25 15 15 9.75M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              }
            >
              Activate {selectionCount} device(s)
            </Button>
          </>
        }
      >
        <p className="text-sm text-text-secondary">
          This will mark <strong className="text-text-primary">{selectionCount} device(s)</strong> as active.
          They will appear in the default device list.
        </p>
      </Modal>

      <Modal
        open={bulkModal === 'deactivate'}
        onClose={() => setBulkModal(null)}
        title="Deactivate Devices"
        actions={
          <>
            <Button variant="ghost" size="sm" onClick={() => setBulkModal(null)}>Cancel</Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => bulkStatusMutation.mutate('inactive')}
              loading={bulkStatusMutation.isPending}
              icon={
                <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5.636 5.636a9 9 0 1012.728 0M12 3v9" />
                </svg>
              }
            >
              Deactivate {selectionCount} device(s)
            </Button>
          </>
        }
      >
        <p className="text-sm text-text-secondary">
          This will mark <strong className="text-text-primary">{selectionCount} device(s)</strong> as inactive.
          They will no longer appear in the default list. You can reactivate them later.
        </p>
      </Modal>

      <Modal
        open={bulkModal === 'department'}
        onClose={() => setBulkModal(null)}
        title="Set Department"
        actions={
          <>
            <Button variant="ghost" size="sm" onClick={() => setBulkModal(null)}>Cancel</Button>
            <Button
              variant="primary"
              size="sm"
              onClick={() => bulkDeptMutation.mutate(bulkDeptId || null)}
              loading={bulkDeptMutation.isPending}
            >
              Apply to {selectionCount} device(s)
            </Button>
          </>
        }
      >
        <div className="space-y-3">
          <p className="text-sm text-text-secondary">
            Choose a department to assign to <strong className="text-text-primary">{selectionCount} device(s)</strong>.
          </p>
          <Select value={bulkDeptId} onChange={(e) => setBulkDeptId(e.target.value)}>
            <option value="">No Department</option>
            {departments.map((d) => (
              <option key={d.id} value={d.id}>{d.name}</option>
            ))}
          </Select>
        </div>
      </Modal>

      <div className="bg-bg-secondary rounded-xl border border-border-primary overflow-hidden shadow-sm animate-slide-up">
        <table className="min-w-full divide-y divide-border-primary">
          <thead className="bg-bg-tertiary/50">
            <tr>
              {/* Checkbox column — admin only */}
              {role === 'admin' && (
                <th className="w-10 px-3 py-3">
                  <input
                    type="checkbox"
                    checked={allSelected}
                    onChange={toggleSelectAll}
                    className="rounded border-border-primary text-accent focus:ring-accent/30 cursor-pointer"
                  />
                </th>
              )}
              {([
                ['hostname', 'Hostname'],
                ['os', 'OS'],
              ] as [SortCol, string][]).map(([col, label]) => (
                <th
                  key={col}
                  onClick={() => toggleSort(col)}
                  className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider cursor-pointer select-none hover:text-text-primary transition-colors"
                >
                  {label}<SortIcon col={col} />
                </th>
              ))}
              <th className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                Department
              </th>
              <th className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">
                Agent
              </th>
              <th
                onClick={() => toggleSort('last_seen')}
                className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider cursor-pointer select-none hover:text-text-primary transition-colors"
              >
                Last Seen<SortIcon col="last_seen" />
              </th>
              <th
                onClick={() => toggleSort('status')}
                className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider cursor-pointer select-none hover:text-text-primary transition-colors"
              >
                Status<SortIcon col="status" />
              </th>
            </tr>
          </thead>
          <tbody className="divide-y divide-border-primary">
            {isLoading
              ? Array.from({ length: 8 }).map((_, i) => (
                  <tr key={i}>
                    {role === 'admin' && <td className="px-3 py-3"><div className="h-4 w-4 bg-bg-tertiary rounded animate-pulse" /></td>}
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
                  const badgeVariant = isInactive ? 'warning' as const : online ? 'success' as const : 'neutral' as const;

                  return (
                    <tr key={d.id} className={`hover:bg-bg-tertiary/50 transition-colors ${selectedIds.has(d.id) ? 'bg-accent/5' : ''}`}>
                      {/* Checkbox — admin only */}
                      {role === 'admin' && (
                        <td className="w-10 px-3 py-3">
                          <input
                            type="checkbox"
                            checked={selectedIds.has(d.id)}
                            onChange={() => toggleSelect(d.id)}
                            className="rounded border-border-primary text-accent focus:ring-accent/30 cursor-pointer"
                          />
                        </td>
                      )}
                      <td className="px-4 py-3 text-sm">
                        <Link to={`/devices/${encodeURIComponent(d.hostname)}`} className="text-accent hover:text-accent-hover font-medium transition-colors">
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
                        <Badge variant={badgeVariant} dot pulseDot={online}>
                          {statusLabel}
                        </Badge>
                      </td>
                    </tr>
                  );
                })}
            {!isLoading && devices.length === 0 && (
              <tr>
                <td colSpan={role === 'admin' ? 7 : 6} className="px-4 py-12 text-center text-sm text-text-muted">
                  <svg className="w-10 h-10 mx-auto mb-3 text-text-muted/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25A2.25 2.25 0 015.25 3h13.5A2.25 2.25 0 0121 5.25z" />
                  </svg>
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
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page <= 1}
            >
              Previous
            </Button>
            <Button
              variant="secondary"
              size="sm"
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page >= totalPages}
            >
              Next
            </Button>
          </div>
        </div>
      )}
    </div>
  );
}
