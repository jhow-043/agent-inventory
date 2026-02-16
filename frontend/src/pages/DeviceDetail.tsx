// Device detail page with dark theme and remote access section.

import { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getDevice } from '../api/devices';
import type { RemoteTool } from '../types';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export default function DeviceDetail() {
  const { id } = useParams<{ id: string }>();
  const { data, isLoading, error } = useQuery({
    queryKey: ['device', id],
    queryFn: () => getDevice(id!),
    enabled: !!id,
  });

  if (isLoading) return <p className="text-text-muted">Loading...</p>;
  if (error) return <p className="text-danger">Failed to load device details.</p>;
  if (!data) return null;

  const { device, hardware, disks, network_interfaces, installed_software, remote_tools } = data;

  return (
    <div>
      <Link to="/devices" className="text-sm text-accent hover:underline mb-4 inline-block">
        &larr; Back to devices
      </Link>

      <h1 className="text-xl font-semibold text-text-primary mb-6">{device.hostname}</h1>

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
      {disks.length > 0 && (
        <Section title="Disks">
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
                {disks.map((d) => (
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
// Reusable sub-components (dark themed)
// ---------------------------------------------------------------------------

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-bg-secondary rounded-lg border border-border-primary mb-6 overflow-hidden">
      <div className="px-4 py-3 bg-bg-tertiary border-b border-border-primary">
        <h2 className="text-sm font-semibold text-text-secondary">{title}</h2>
      </div>
      <div className="p-4">{children}</div>
    </div>
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
