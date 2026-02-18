// Device detail page with status control, department assignment, hardware history, and remote access section.

import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getDevice, updateDeviceStatus, updateDeviceDepartment, deleteDevice } from '../api/devices';
import { getDepartments } from '../api/departments';
import { useAuth } from '../hooks/useAuth';
import { Button, Badge, Select, Card, CardHeader, CardContent, Modal } from '../components/ui';
import type { RemoteTool, Hardware } from '../types';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function DeviceDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { role } = useAuth();
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const { data, isLoading, error } = useQuery({
    queryKey: ['device', id],
    queryFn: () => getDevice(id!),
    enabled: !!id,
  });

  const { data: deptData } = useQuery({
    queryKey: ['departments'],
    queryFn: getDepartments,
  });

  const statusMutation = useMutation({
    mutationFn: (newStatus: 'active' | 'inactive') => updateDeviceStatus(id!, newStatus),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['device', id] });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard-stats'] });
    },
  });

  const departmentMutation = useMutation({
    mutationFn: (deptId: string | null) => updateDeviceDepartment(id!, deptId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['device', id] });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteDevice(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard-stats'] });
      navigate('/devices');
    },
  });

  if (isLoading) return (
    <div className="animate-fade-in space-y-4">
      <div className="h-4 w-32 bg-bg-tertiary rounded animate-pulse" />
      <div className="h-8 w-48 bg-bg-tertiary rounded animate-pulse" />
      <div className="h-40 bg-bg-secondary rounded-xl border border-border-primary animate-pulse" />
    </div>
  );
  if (error) return <p className="text-danger">Failed to load device details.</p>;
  if (!data) return null;

  const { device, hardware, disks, network_interfaces, installed_software, remote_tools, hardware_history } = data;
  const departments = deptData?.departments ?? [];
  const isInactive = device.status === 'inactive';

  return (
    <div className="animate-fade-in">
      <Link to="/devices" className="inline-flex items-center gap-1.5 text-sm text-accent hover:text-accent-hover font-medium transition-colors mb-4">
        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
          <path strokeLinecap="round" strokeLinejoin="round" d="M10.5 19.5L3 12m0 0l7.5-7.5M3 12h18" />
        </svg>
        Back to devices
      </Link>

      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-bold text-text-primary">{device.hostname}</h1>
          <Badge variant={isInactive ? 'warning' : 'success'} dot>
            {isInactive ? 'Inactive' : 'Active'}
          </Badge>
        </div>
        <div className="flex items-center gap-3">
          {/* Department selector — admin only */}
          {role === 'admin' && (
            <div className="w-48">
              <Select
                value={device.department_id ?? ''}
                onChange={(e) => departmentMutation.mutate(e.target.value || null)}
                disabled={departmentMutation.isPending}
              >
                <option value="">No Department</option>
                {departments.map((d) => (
                  <option key={d.id} value={d.id}>{d.name}</option>
                ))}
              </Select>
            </div>
          )}

          {/* Status toggle — admin only */}
          {role === 'admin' && (
            <Button
              variant={isInactive ? 'success' : 'secondary'}
              size="sm"
              onClick={() => statusMutation.mutate(isInactive ? 'active' : 'inactive')}
              loading={statusMutation.isPending}
              icon={
                <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5.636 5.636a9 9 0 1012.728 0M12 3v9" />
                </svg>
              }
            >
              {isInactive ? 'Reactivate' : 'Deactivate'}
            </Button>
          )}

          {/* Delete device — admin only */}
          {role === 'admin' && (
            <Button
              variant="danger"
              size="sm"
              onClick={() => setShowDeleteModal(true)}
              icon={
                <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                </svg>
              }
            >
              Delete
            </Button>
          )}
        </div>
      </div>

      {/* Delete Confirmation Modal */}
      <Modal
        open={showDeleteModal}
        onClose={() => setShowDeleteModal(false)}
        title="Delete Device"
        actions={
          <>
            <Button variant="ghost" size="sm" onClick={() => setShowDeleteModal(false)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              size="sm"
              onClick={() => deleteMutation.mutate()}
              loading={deleteMutation.isPending}
            >
              Delete
            </Button>
          </>
        }
      >
        <p className="text-sm text-text-secondary">
          Are you sure you want to permanently delete <strong className="text-text-primary">{device.hostname}</strong>?
          All related data (hardware, disks, network interfaces, software, remote tools) will be removed.
          This action cannot be undone.
        </p>
      </Modal>

      {/* Remote Access Tools */}
      <Section title="Remote Access">
        {remote_tools && remote_tools.length > 0 ? (
          <table className="w-full text-sm">
            <thead>
              <tr className="text-left text-xs text-text-muted uppercase">
                <th className="pb-2 pr-4 font-medium w-8"></th>
                <th className="pb-2 pr-4 font-medium">Tool</th>
                <th className="pb-2 pr-4 font-medium">Version</th>
                <th className="pb-2 pr-4 font-medium">ID</th>
                <th className="pb-2 font-medium w-10"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-primary">
              {remote_tools.map((tool) => (
                <RemoteToolRow key={tool.id} tool={tool} />
              ))}
            </tbody>
          </table>
        ) : (
          <p className="text-sm text-text-muted">No remote access tools detected.</p>
        )}
      </Section>

      {/* System Info */}
      <Section title="System">
        <Grid>
          <Field label="Hostname" value={device.hostname} />
          <Field label="Serial Number" value={device.serial_number} />
          <Field label="OS" value={`${device.os_name} ${device.os_version}`} />
          <Field label="Build" value={device.os_build} />
          <Field label="Architecture" value={device.os_arch} />
          <Field
            label="Last Boot"
            value={device.last_boot_time ? new Date(device.last_boot_time).toLocaleString() : '—'}
          />
          <Field label="Logged-in User" value={device.logged_in_user} />
          <Field label="Agent Version" value={device.agent_version} />
          <Field label="License Status" value={device.license_status} />
          <Field label="Last Seen" value={new Date(device.last_seen).toLocaleString()} />
        </Grid>
      </Section>

      {/* Hardware */}
      {hardware && (
        <Section title="Hardware">
          <Grid>
            <Field label="CPU" value={hardware.cpu_model} />
            <Field label="Cores / Threads" value={`${hardware.cpu_cores} / ${hardware.cpu_threads}`} />
            <Field label="RAM" value={formatBytes(hardware.ram_total_bytes)} />
            <Field
              label="Motherboard"
              value={`${hardware.motherboard_manufacturer} ${hardware.motherboard_product}`.trim()}
            />
            <Field label="Motherboard Serial" value={hardware.motherboard_serial} />
            <Field label="BIOS" value={`${hardware.bios_vendor} ${hardware.bios_version}`.trim()} />
          </Grid>
        </Section>
      )}

      {/* Disks */}
      {disks.length > 0 && (() => {
        const physicalDisks = disks.filter((d) => d.media_type !== 'Partition');
        const partitions = disks.filter((d) => d.media_type === 'Partition');
        return (
          <>
            {physicalDisks.length > 0 && (
              <Section title="Physical Disks">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-border-primary text-sm">
                    <thead className="bg-bg-tertiary">
                      <tr>
                        <Th>Model</Th>
                        <Th>Size</Th>
                        <Th>Type</Th>
                        <Th>Interface</Th>
                        <Th>Serial</Th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-border-primary">
                      {physicalDisks.map((d) => (
                        <tr key={d.id}>
                          <Td>{d.model}</Td>
                          <Td>{formatBytes(d.size_bytes)}</Td>
                          <Td>{d.media_type}</Td>
                          <Td>{d.interface_type}</Td>
                          <Td>{d.serial_number}</Td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </Section>
            )}

            {partitions.length > 0 && (
              <Section title="Partitions">
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                  {partitions.map((p) => {
                    const total = p.partition_size_bytes || 0;
                    const free = p.free_space_bytes || 0;
                    const used = total - free;
                    const pct = total > 0 ? Math.round((used / total) * 100) : 0;
                    const barColor = pct >= 90 ? 'bg-danger' : pct >= 70 ? 'bg-warning' : 'bg-accent';
                    return (
                      <div key={p.id} className="bg-bg-tertiary rounded-lg border border-border-primary p-4">
                        <div className="flex items-center justify-between mb-2">
                          <span className="text-sm font-semibold text-text-primary">
                            {p.drive_letter || '—'}
                          </span>
                          <span className="text-xs text-text-muted">{pct}% used</span>
                        </div>
                        <div className="w-full h-2 bg-bg-primary rounded-full overflow-hidden mb-2">
                          <div className={`h-full ${barColor} rounded-full transition-all`} style={{ width: `${pct}%` }} />
                        </div>
                        <div className="flex justify-between text-xs text-text-muted">
                          <span>{formatBytes(used)} used</span>
                          <span>{formatBytes(free)} free</span>
                        </div>
                        <div className="text-xs text-text-muted mt-1">
                          Total: {formatBytes(total)}
                        </div>
                      </div>
                    );
                  })}
                </div>
              </Section>
            )}
          </>
        );
      })()}

      {/* Network */}
      {network_interfaces.length > 0 && (
        <Section title="Network Interfaces">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-border-primary text-sm">
              <thead className="bg-bg-tertiary">
                <tr>
                  <Th>Name</Th>
                  <Th>MAC</Th>
                  <Th>IPv4</Th>
                  <Th>IPv6</Th>
                  <Th>Speed</Th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-primary">
                {network_interfaces.map((n) => (
                  <tr key={n.id}>
                    <Td>{n.name}</Td>
                    <Td>{n.mac_address}</Td>
                    <Td>{n.ipv4_address || '—'}</Td>
                    <Td>{n.ipv6_address || '—'}</Td>
                    <Td>{n.speed_mbps ? `${n.speed_mbps} Mbps` : '—'}</Td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Section>
      )}

      {/* Software */}
      {installed_software.length > 0 && (
        <Section title={`Installed Software (${installed_software.length})`}>
          <div className="overflow-x-auto max-h-96 overflow-y-auto">
            <table className="min-w-full divide-y divide-border-primary text-sm">
              <thead className="bg-bg-tertiary sticky top-0">
                <tr>
                  <Th>Name</Th>
                  <Th>Version</Th>
                  <Th>Vendor</Th>
                  <Th>Install Date</Th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-primary">
                {installed_software.map((s) => (
                  <tr key={s.id}>
                    <Td>{s.name}</Td>
                    <Td>{s.version || '—'}</Td>
                    <Td>{s.vendor || '—'}</Td>
                    <Td>{s.install_date || '—'}</Td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </Section>
      )}

      {/* Hardware History */}
      {hardware_history && hardware_history.length > 0 && (
        <Section title="Hardware Changes">
          <div className="space-y-3">
            {hardware_history.map((h) => {
              let snapshot: Partial<Hardware> = {};
              try { snapshot = JSON.parse(h.snapshot); } catch { /* ignore */ }
              return (
                <div key={h.id} className="bg-bg-tertiary rounded-lg border border-border-primary p-4">
                  <p className="text-xs text-text-muted mb-2">
                    Changed at: {new Date(h.changed_at).toLocaleString()}
                  </p>
                  <Grid>
                    {snapshot.cpu_model && <Field label="CPU" value={snapshot.cpu_model} />}
                    {snapshot.cpu_cores != null && snapshot.cpu_threads != null && (
                      <Field label="Cores / Threads" value={`${snapshot.cpu_cores} / ${snapshot.cpu_threads}`} />
                    )}
                    {snapshot.ram_total_bytes != null && <Field label="RAM" value={formatBytes(snapshot.ram_total_bytes)} />}
                    {(snapshot.motherboard_manufacturer || snapshot.motherboard_product) && (
                      <Field label="Motherboard" value={`${snapshot.motherboard_manufacturer ?? ''} ${snapshot.motherboard_product ?? ''}`.trim()} />
                    )}
                    {snapshot.bios_vendor && <Field label="BIOS" value={`${snapshot.bios_vendor} ${snapshot.bios_version ?? ''}`.trim()} />}
                  </Grid>
                </div>
              );
            })}
          </div>
        </Section>
      )}
    </div>
  );
}

// ---------------------------------------------------------------------------
// Remote Tool Row — compact list item with copy-to-clipboard
// ---------------------------------------------------------------------------

function RemoteToolRow({ tool }: { tool: RemoteTool }) {
  const [copied, setCopied] = useState(false);
  const dotColor: Record<string, string> = {
    TeamViewer: 'bg-blue-500',
    AnyDesk: 'bg-red-500',
    RustDesk: 'bg-orange-500',
  };

  const handleCopy = async () => {
    if (!tool.remote_id) return;
    await navigator.clipboard.writeText(tool.remote_id);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <tr>
      <td className="py-2.5 pr-4">
        <span className={`inline-block w-2.5 h-2.5 rounded-full ${dotColor[tool.tool_name] ?? 'bg-text-muted'}`} />
      </td>
      <td className="py-2.5 pr-4 font-medium text-text-primary">{tool.tool_name}</td>
      <td className="py-2.5 pr-4 text-xs text-text-muted">{tool.version ? `v${tool.version}` : '—'}</td>
      <td className="py-2.5 pr-4">
        {tool.remote_id ? (
          <code className="text-sm font-mono bg-bg-primary border border-border-primary rounded px-2 py-0.5 text-text-primary">
            {tool.remote_id}
          </code>
        ) : (
          <span className="text-xs text-text-muted">ID not available</span>
        )}
      </td>
      <td className="py-2.5">
        {tool.remote_id && (
          <button
            onClick={handleCopy}
            className="text-xs px-2 py-1 rounded bg-bg-tertiary border border-border-primary text-text-secondary hover:text-text-primary hover:bg-border-primary transition-colors cursor-pointer"
            title="Copy ID"
          >
            {copied ? (
              <svg className="w-3.5 h-3.5 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M4.5 12.75l6 6 9-13.5" />
              </svg>
            ) : (
              <svg className="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M15.666 3.888A2.25 2.25 0 0013.5 2.25h-3c-1.03 0-1.9.693-2.166 1.638m7.332 0c.055.194.084.4.084.612v0a.75.75 0 01-.75.75H9.75a.75.75 0 01-.75-.75v0c0-.212.03-.418.084-.612m7.332 0c.646.049 1.288.11 1.927.184 1.1.128 1.907 1.077 1.907 2.185V19.5a2.25 2.25 0 01-2.25 2.25H6.75A2.25 2.25 0 014.5 19.5V6.257c0-1.108.806-2.057 1.907-2.185a48.208 48.208 0 011.927-.184" />
              </svg>
            )}
          </button>
        )}
      </td>
    </tr>
  );
}

// ---------------------------------------------------------------------------
// Reusable sub-components
// ---------------------------------------------------------------------------

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Card className="mb-6">
      <CardHeader>
        <h2 className="text-sm font-semibold text-text-secondary">{title}</h2>
      </CardHeader>
      <CardContent>{children}</CardContent>
    </Card>
  );
}

function Grid({ children }: { children: React.ReactNode }) {
  return <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-x-6 gap-y-3">{children}</div>;
}

function Field({ label, value }: { label: string; value: string | undefined | null }) {
  return (
    <div>
      <span className="text-xs text-text-muted uppercase">{label}</span>
      <p className="text-sm text-text-primary mt-0.5">{value || '—'}</p>
    </div>
  );
}

function Th({ children }: { children: React.ReactNode }) {
  return (
    <th className="px-4 py-2 text-left text-xs font-medium text-text-muted uppercase tracking-wider">
      {children}
    </th>
  );
}

function Td({ children }: { children: React.ReactNode }) {
  return <td className="px-4 py-2 text-text-secondary">{children}</td>;
}
