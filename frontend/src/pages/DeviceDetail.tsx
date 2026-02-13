// Device detail page with sections for system, hardware, disks, network, and software.

import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { getDevice } from '../api/devices';

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

  if (isLoading) return <p className="text-gray-500">Loading...</p>;
  if (error) return <p className="text-red-600">Failed to load device details.</p>;
  if (!data) return null;

  const { device, hardware, disks, network_interfaces, installed_software } = data;

  return (
    <div>
      <Link to="/devices" className="text-sm text-blue-600 hover:underline mb-4 inline-block">
        &larr; Back to devices
      </Link>

      <h1 className="text-xl font-semibold text-gray-900 mb-6">{device.hostname}</h1>

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
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <Th>Model</Th>
                  <Th>Size</Th>
                  <Th>Type</Th>
                  <Th>Interface</Th>
                  <Th>Serial</Th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
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
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50">
                <tr>
                  <Th>Name</Th>
                  <Th>MAC</Th>
                  <Th>IPv4</Th>
                  <Th>IPv6</Th>
                  <Th>Speed</Th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
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
            <table className="min-w-full divide-y divide-gray-200 text-sm">
              <thead className="bg-gray-50 sticky top-0">
                <tr>
                  <Th>Name</Th>
                  <Th>Version</Th>
                  <Th>Vendor</Th>
                  <Th>Install Date</Th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-100">
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
// Reusable sub-components
// ---------------------------------------------------------------------------

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="bg-white rounded-lg shadow mb-6 overflow-hidden">
      <div className="px-4 py-3 bg-gray-50 border-b border-gray-200">
        <h2 className="text-sm font-semibold text-gray-700">{title}</h2>
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
      <span className="text-xs text-gray-500 uppercase">{label}</span>
      <p className="text-sm text-gray-800 mt-0.5">{value || '—'}</p>
    </div>
  );
}

function Th({ children }: { children: React.ReactNode }) {
  return (
    <th className="px-4 py-2 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
      {children}
    </th>
  );
}

function Td({ children }: { children: React.ReactNode }) {
  return <td className="px-4 py-2 text-gray-700">{children}</td>;
}
