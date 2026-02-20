// Device detail page with status control, department assignment, hardware/activity history, and remote access section.

import { useState } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getDevice, updateDeviceStatus, updateDeviceDepartment, deleteDevice, getDeviceActivity, getHardwareHistory } from '../api/devices';
import { getDepartments } from '../api/departments';
import { useAuth } from '../hooks/useAuth';
import { useToast } from '../hooks/useToast';
import { Button, Badge, Select, Card, CardHeader, CardContent, Modal } from '../components/ui';
import type { RemoteTool, DeviceActivityLog, HardwareHistory } from '../types';

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

type Tab = 'overview' | 'storage' | 'network' | 'software' | 'remote' | 'history';

const TABS: { key: Tab; label: string; icon: React.ReactNode }[] = [
  {
    key: 'overview',
    label: 'Visão Geral',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M11.25 11.25l.041-.02a.75.75 0 011.063.852l-.708 2.836a.75.75 0 001.063.853l.041-.021M21 12a9 9 0 11-18 0 9 9 0 0118 0zm-9-3.75h.008v.008H12V8.25z" /></svg>,
  },
  {
    key: 'storage',
    label: 'Armazenamento',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375m16.5 0v3.75m-16.5-3.75v3.75m16.5 0v3.75C20.25 16.153 16.556 18 12 18s-8.25-1.847-8.25-4.125v-3.75m16.5 0c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125" /></svg>,
  },
  {
    key: 'network',
    label: 'Rede',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3m0 0a8.997 8.997 0 017.843 4.582M12 3a8.997 8.997 0 00-7.843 4.582m15.686 0A11.953 11.953 0 0112 10.5c-2.998 0-5.74-1.1-7.843-2.918m15.686 0A8.959 8.959 0 0121 12c0 .778-.099 1.533-.284 2.253m0 0A17.919 17.919 0 0112 16.5c-3.162 0-6.133-.815-8.716-2.247m0 0A9.015 9.015 0 013 12c0-1.605.42-3.113 1.157-4.418" /></svg>,
  },
  {
    key: 'software',
    label: 'Software',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M6.429 9.75L2.25 12l4.179 2.25m0-4.5l5.571 3 5.571-3m-11.142 0L2.25 7.5 12 2.25l9.75 5.25-4.179 2.25m0 0L21.75 12l-4.179 2.25m0 0l4.179 2.25L12 21.75 2.25 16.5l4.179-2.25m11.142 0l-5.571 3-5.571-3" /></svg>,
  },
  {
    key: 'remote',
    label: 'Acesso Remoto',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M13.19 8.688a4.5 4.5 0 011.242 7.244l-4.5 4.5a4.5 4.5 0 01-6.364-6.364l1.757-1.757m9.07-9.07l-1.757 1.757a4.5 4.5 0 010 6.364L8.93 15.84a4.5 4.5 0 01-1.242-7.244l4.5-4.5a4.5 4.5 0 016.364 0z" /></svg>,
  },
  {
    key: 'history',
    label: 'Histórico',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>,
  },
];

export default function DeviceDetail() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { role } = useAuth();
  const toast = useToast();
  const [showDeleteModal, setShowDeleteModal] = useState(false);
  const [activeTab, setActiveTab] = useState<Tab>('overview');
  const [activityPage, setActivityPage] = useState(1);
  const [hwHistoryPage, setHwHistoryPage] = useState(1);
  const [hwHistoryFilter, setHwHistoryFilter] = useState('');
  const { data, isLoading, error } = useQuery({
    queryKey: ['device', id],
    queryFn: () => getDevice(id!),
    enabled: !!id,
  });

  const { data: activityData, isLoading: activityLoading } = useQuery({
    queryKey: ['device-activity', id, activityPage],
    queryFn: () => getDeviceActivity(id!, activityPage),
    enabled: !!id && activeTab === 'history',
  });

  const { data: hwHistoryData, isLoading: hwHistoryLoading } = useQuery({
    queryKey: ['device-hw-history', id, hwHistoryPage, hwHistoryFilter],
    queryFn: () => getHardwareHistory(id!, hwHistoryPage, 20, hwHistoryFilter),
    enabled: !!id && activeTab === 'history',
  });

  const { data: deptData } = useQuery({
    queryKey: ['departments'],
    queryFn: getDepartments,
  });

  const statusMutation = useMutation({
    mutationFn: (newStatus: 'active' | 'inactive') => updateDeviceStatus(id!, newStatus),
    onSuccess: (_data, newStatus) => {
      queryClient.invalidateQueries({ queryKey: ['device', id] });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard-stats'] });
      toast.success(newStatus === 'active' ? 'Dispositivo ativado' : 'Dispositivo desativado');
    },
    onError: () => toast.error('Falha ao atualizar status'),
  });

  const departmentMutation = useMutation({
    mutationFn: (deptId: string | null) => updateDeviceDepartment(id!, deptId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['device', id] });
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      toast.success('Departamento atualizado');
    },
    onError: () => toast.error('Falha ao atualizar departamento'),
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteDevice(id!),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['devices'] });
      queryClient.invalidateQueries({ queryKey: ['dashboard-stats'] });
      toast.success('Dispositivo excluído');
      navigate('/devices');
    },
    onError: () => toast.error('Falha ao excluir dispositivo'),
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

  const { device, hardware, disks, network_interfaces, installed_software, remote_tools } = data;
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

      {/* Tabs */}
      <div className="flex gap-1 border-b border-border-primary mb-6 overflow-x-auto">
        {TABS.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 transition-all whitespace-nowrap cursor-pointer ${
              activeTab === tab.key
                ? 'border-accent text-accent'
                : 'border-transparent text-text-muted hover:text-text-primary hover:border-border-secondary'
            }`}
          >
            {tab.icon}
            {tab.label}
          </button>
        ))}
      </div>

      {/* Tab Content */}
      <div className="animate-fade-in">
        {activeTab === 'overview' && (
          <>
            {/* Remote Access Tools */}
            <Section title="Acesso Remoto">
              {remote_tools && remote_tools.length > 0 ? (
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-xs text-text-muted uppercase">
                      <th className="pb-2 pr-4 font-medium w-8"></th>
                      <th className="pb-2 pr-4 font-medium">Ferramenta</th>
                      <th className="pb-2 pr-4 font-medium">Versão</th>
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
                <p className="text-sm text-text-muted">Nenhuma ferramenta de acesso remoto detectada.</p>
              )}
            </Section>

            {/* System Info */}
            <Section title="Sistema">
              <Grid>
                <Field label="Hostname" value={device.hostname} />
                <Field label="Número de Série" value={device.serial_number} />
                <Field label="SO" value={`${device.os_name} ${device.os_version}`} />
                <Field label="Build" value={device.os_build} />
                <Field label="Arquitetura" value={device.os_arch} />
                <Field label="Último Boot" value={device.last_boot_time ? new Date(device.last_boot_time).toLocaleString('pt-BR') : '—'} />
                <Field label="Usuário Logado" value={device.logged_in_user} />
                <Field label="Versão do Agente" value={device.agent_version} />
                <Field label="Licença" value={device.license_status} />
                <Field label="Última Atividade" value={new Date(device.last_seen).toLocaleString('pt-BR')} />
              </Grid>
            </Section>

            {/* Hardware */}
            {hardware && (
              <Section title="Hardware">
                <Grid>
                  <Field label="CPU" value={hardware.cpu_model} />
                  <Field label="Cores / Threads" value={`${hardware.cpu_cores} / ${hardware.cpu_threads}`} />
                  <Field label="RAM" value={formatBytes(hardware.ram_total_bytes)} />
                  <Field label="Placa-mãe" value={`${hardware.motherboard_manufacturer} ${hardware.motherboard_product}`.trim()} />
                  <Field label="Serial da Placa-mãe" value={hardware.motherboard_serial} />
                  <Field label="BIOS" value={`${hardware.bios_vendor} ${hardware.bios_version}`.trim()} />
                </Grid>
              </Section>
            )}
          </>
        )}

        {activeTab === 'storage' && (
          <>
            {disks.length > 0 ? (() => {
              const physicalDisks = disks.filter((d) => d.media_type !== 'Partition');
              const partitions = disks.filter((d) => d.media_type === 'Partition');
              return (
                <>
                  {physicalDisks.length > 0 && (
                    <Section title="Discos Físicos">
                      <div className="overflow-x-auto">
                        <table className="min-w-full divide-y divide-border-primary text-sm">
                          <thead className="bg-bg-tertiary">
                            <tr>
                              <Th>Modelo</Th>
                              <Th>Tamanho</Th>
                              <Th>Tipo</Th>
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
                    <Section title="Partições">
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
                                <span className="text-sm font-semibold text-text-primary">{p.drive_letter || '—'}</span>
                                <span className="text-xs text-text-muted">{pct}% usado</span>
                              </div>
                              <div className="w-full h-2 bg-bg-primary rounded-full overflow-hidden mb-2">
                                <div className={`h-full ${barColor} rounded-full transition-all`} style={{ width: `${pct}%` }} />
                              </div>
                              <div className="flex justify-between text-xs text-text-muted">
                                <span>{formatBytes(used)} usado</span>
                                <span>{formatBytes(free)} livre</span>
                              </div>
                              <div className="text-xs text-text-muted mt-1">Total: {formatBytes(total)}</div>
                            </div>
                          );
                        })}
                      </div>
                    </Section>
                  )}
                </>
              );
            })() : (
              <EmptyState message="Nenhum disco encontrado." />
            )}
          </>
        )}

        {activeTab === 'network' && (
          <>
            {network_interfaces.length > 0 ? (
              <Section title="Interfaces de Rede">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-border-primary text-sm">
                    <thead className="bg-bg-tertiary">
                      <tr>
                        <Th>Nome</Th>
                        <Th>MAC</Th>
                        <Th>IPv4</Th>
                        <Th>IPv6</Th>
                        <Th>Velocidade</Th>
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
            ) : (
              <EmptyState message="Nenhuma interface de rede encontrada." />
            )}
          </>
        )}

        {activeTab === 'software' && (
          <>
            {installed_software.length > 0 ? (
              <Section title={`Software Instalado (${installed_software.length})`}>
                <div className="overflow-x-auto max-h-[600px] overflow-y-auto">
                  <table className="min-w-full divide-y divide-border-primary text-sm">
                    <thead className="bg-bg-tertiary sticky top-0">
                      <tr>
                        <Th>Nome</Th>
                        <Th>Versão</Th>
                        <Th>Fabricante</Th>
                        <Th>Data de Instalação</Th>
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
            ) : (
              <EmptyState message="Nenhum software instalado encontrado." />
            )}
          </>
        )}

        {activeTab === 'remote' && (
          <>
            {remote_tools && remote_tools.length > 0 ? (
              <Section title="Ferramentas de Acesso Remoto">
                <table className="w-full text-sm">
                  <thead>
                    <tr className="text-left text-xs text-text-muted uppercase">
                      <th className="pb-2 pr-4 font-medium w-8"></th>
                      <th className="pb-2 pr-4 font-medium">Ferramenta</th>
                      <th className="pb-2 pr-4 font-medium">Versão</th>
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
              </Section>
            ) : (
              <EmptyState message="Nenhuma ferramenta de acesso remoto detectada." />
            )}
          </>
        )}

        {activeTab === 'history' && (
          <>
            {/* ── Activity Log ──────────────────────────────────────────── */}
            <Section title="Histórico de Atividades">
              {activityLoading ? (
                <p className="text-sm text-text-muted py-4">Carregando...</p>
              ) : activityData && activityData.activities && activityData.activities.length > 0 ? (
                <>
                  <div className="space-y-2">
                    {activityData.activities.map((a: DeviceActivityLog) => (
                      <ActivityRow key={a.id} activity={a} />
                    ))}
                  </div>
                  {activityData.total > activityData.limit && (
                    <div className="flex items-center justify-between mt-4 pt-3 border-t border-border-primary">
                      <span className="text-xs text-text-muted">
                        Página {activityData.page} de {Math.ceil(activityData.total / activityData.limit)}
                        {' '}({activityData.total} registros)
                      </span>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={activityPage <= 1}
                          onClick={() => setActivityPage((p) => p - 1)}
                        >
                          Anterior
                        </Button>
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={activityPage >= Math.ceil(activityData.total / activityData.limit)}
                          onClick={() => setActivityPage((p) => p + 1)}
                        >
                          Próxima
                        </Button>
                      </div>
                    </div>
                  )}
                </>
              ) : (
                <EmptyState message="Nenhuma atividade registrada para este dispositivo." />
              )}
            </Section>

            {/* ── Hardware History ──────────────────────────────────────── */}
            <Section title="Alterações de Hardware">
              {/* Component filter */}
              <div className="flex items-center gap-2 mb-4">
                <span className="text-xs text-text-muted">Filtrar por:</span>
                {['', 'cpu', 'ram', 'motherboard', 'bios', 'disk', 'network'].map((comp) => (
                  <button
                    key={comp}
                    onClick={() => { setHwHistoryFilter(comp); setHwHistoryPage(1); }}
                    className={`text-xs px-2.5 py-1 rounded-full border transition-colors cursor-pointer ${
                      hwHistoryFilter === comp
                        ? 'bg-accent text-white border-accent'
                        : 'bg-bg-tertiary text-text-secondary border-border-primary hover:border-accent hover:text-accent'
                    }`}
                  >
                    {comp === '' ? 'Todos' : COMPONENT_LABELS[comp] ?? comp}
                  </button>
                ))}
              </div>

              {hwHistoryLoading ? (
                <p className="text-sm text-text-muted py-4">Carregando...</p>
              ) : hwHistoryData && hwHistoryData.changes && hwHistoryData.changes.length > 0 ? (
                <>
                  <div className="space-y-2">
                    {hwHistoryData.changes.map((h: HardwareHistory) => (
                      <HardwareChangeRow key={h.id} change={h} />
                    ))}
                  </div>
                  {hwHistoryData.total > hwHistoryData.limit && (
                    <div className="flex items-center justify-between mt-4 pt-3 border-t border-border-primary">
                      <span className="text-xs text-text-muted">
                        Página {hwHistoryData.page} de {Math.ceil(hwHistoryData.total / hwHistoryData.limit)}
                        {' '}({hwHistoryData.total} registros)
                      </span>
                      <div className="flex gap-2">
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={hwHistoryPage <= 1}
                          onClick={() => setHwHistoryPage((p) => p - 1)}
                        >
                          Anterior
                        </Button>
                        <Button
                          size="sm"
                          variant="secondary"
                          disabled={hwHistoryPage >= Math.ceil(hwHistoryData.total / hwHistoryData.limit)}
                          onClick={() => setHwHistoryPage((p) => p + 1)}
                        >
                          Próxima
                        </Button>
                      </div>
                    </div>
                  )}
                </>
              ) : (
                <EmptyState message="Nenhuma alteração de hardware registrada para este dispositivo." />
              )}
            </Section>
          </>
        )}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Activity Row — renders a single activity log entry with icon and colors
// ---------------------------------------------------------------------------

const ACTIVITY_CONFIG: Record<string, { icon: React.ReactNode; color: string; label: string }> = {
  software_installed: {
    label: 'Instalação',
    color: 'text-green-500 bg-green-500/10',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 4.5v15m7.5-7.5h-15" /></svg>,
  },
  software_removed: {
    label: 'Remoção',
    color: 'text-red-500 bg-red-500/10',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M19.5 12h-15" /></svg>,
  },
};

function ActivityRow({ activity }: { activity: DeviceActivityLog }) {
  const cfg = ACTIVITY_CONFIG[activity.activity_type] ?? {
    label: activity.activity_type,
    color: 'text-text-muted bg-bg-tertiary',
    icon: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 6v6h4.5m4.5 0a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>,
  };

  return (
    <div className="flex items-start gap-3 py-2.5 px-3 rounded-lg hover:bg-bg-tertiary/50 transition-colors">
      <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${cfg.color}`}>
        {cfg.icon}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2">
          <span className={`text-xs font-medium px-1.5 py-0.5 rounded ${cfg.color}`}>
            {cfg.label}
          </span>
          <span className="text-xs text-text-muted">
            {new Date(activity.detected_at).toLocaleString('pt-BR')}
          </span>
        </div>
        <p className="text-sm text-text-primary mt-0.5">{activity.description}</p>

      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// Hardware Change Row — renders a single hardware change entry with timeline style
// ---------------------------------------------------------------------------

const COMPONENT_LABELS: Record<string, string> = {
  cpu: 'CPU',
  ram: 'RAM',
  motherboard: 'Placa-mãe',
  bios: 'BIOS',
  disk: 'Disco',
  network: 'Rede',
};

const COMPONENT_COLORS: Record<string, string> = {
  cpu: 'text-blue-500 bg-blue-500/10',
  ram: 'text-purple-500 bg-purple-500/10',
  motherboard: 'text-orange-500 bg-orange-500/10',
  bios: 'text-cyan-500 bg-cyan-500/10',
  disk: 'text-amber-500 bg-amber-500/10',
  network: 'text-teal-500 bg-teal-500/10',
};

const CHANGE_TYPE_LABELS: Record<string, { label: string; color: string }> = {
  changed: { label: 'Alterado', color: 'text-yellow-600 bg-yellow-500/10' },
  added: { label: 'Adicionado', color: 'text-green-500 bg-green-500/10' },
  removed: { label: 'Removido', color: 'text-red-500 bg-red-500/10' },
};

const COMPONENT_ICONS: Record<string, React.ReactNode> = {
  cpu: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M8.25 3v1.5M4.5 8.25H3m18 0h-1.5M4.5 12H3m18 0h-1.5m-15 3.75H3m18 0h-1.5M8.25 19.5V21M12 3v1.5m0 15V21m3.75-18v1.5m0 15V21m-9-1.5h10.5a2.25 2.25 0 002.25-2.25V6.75a2.25 2.25 0 00-2.25-2.25H6.75A2.25 2.25 0 004.5 6.75v10.5a2.25 2.25 0 002.25 2.25zm.75-12h9v9h-9v-9z" /></svg>,
  ram: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M6 6h12M6 6v12m0-12L4.5 4.5M18 6v12m0-12l1.5-1.5M6 18h12M6 18l-1.5 1.5M18 18l1.5 1.5M9 9h6v6H9V9z" /></svg>,
  motherboard: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9 17.25v1.007a3 3 0 01-.879 2.122L7.5 21h9l-.621-.621A3 3 0 0115 18.257V17.25m6-12V15a2.25 2.25 0 01-2.25 2.25H5.25A2.25 2.25 0 013 15V5.25A2.25 2.25 0 015.25 3h13.5A2.25 2.25 0 0121 5.25z" /></svg>,
  bios: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M11.42 15.17l-5.1-5.1m0 0L11.42 4.97m-5.1 5.1H21" /></svg>,
  disk: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M20.25 6.375c0 2.278-3.694 4.125-8.25 4.125S3.75 8.653 3.75 6.375m16.5 0c0-2.278-3.694-4.125-8.25-4.125S3.75 4.097 3.75 6.375m16.5 0v11.25c0 2.278-3.694 4.125-8.25 4.125s-8.25-1.847-8.25-4.125V6.375" /></svg>,
  network: <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M12 21a9.004 9.004 0 008.716-6.747M12 21a9.004 9.004 0 01-8.716-6.747M12 21c2.485 0 4.5-4.03 4.5-9S14.485 3 12 3m0 18c-2.485 0-4.5-4.03-4.5-9S9.515 3 12 3" /></svg>,
};

function HardwareChangeRow({ change }: { change: HardwareHistory }) {
  const component = change.component ?? 'unknown';
  const changeType = change.change_type ?? 'changed';
  const compColor = COMPONENT_COLORS[component] ?? 'text-text-muted bg-bg-tertiary';
  const compLabel = COMPONENT_LABELS[component] ?? component;
  const compIcon = COMPONENT_ICONS[component] ?? <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}><path strokeLinecap="round" strokeLinejoin="round" d="M9.594 3.94c.09-.542.56-.94 1.11-.94h2.593c.55 0 1.02.398 1.11.94l.213 1.281c.063.374.313.686.645.87.074.04.147.083.22.127.324.196.72.257 1.075.124l1.217-.456a1.125 1.125 0 011.37.49l1.296 2.247a1.125 1.125 0 01-.26 1.431l-1.003.827c-.293.24-.438.613-.431.992a6.759 6.759 0 010 .255c-.007.378.138.75.43.99l1.005.828c.424.35.534.954.26 1.43l-1.298 2.247a1.125 1.125 0 01-1.369.491l-1.217-.456c-.355-.133-.75-.072-1.076.124a6.57 6.57 0 01-.22.128c-.331.183-.581.495-.644.869l-.213 1.28c-.09.543-.56.941-1.11.941h-2.594c-.55 0-1.02-.398-1.11-.94l-.213-1.281c-.062-.374-.312-.686-.644-.87a6.52 6.52 0 01-.22-.127c-.325-.196-.72-.257-1.076-.124l-1.217.456a1.125 1.125 0 01-1.369-.49l-1.297-2.247a1.125 1.125 0 01.26-1.431l1.004-.827c.292-.24.437-.613.43-.992a6.932 6.932 0 010-.255c.007-.378-.138-.75-.43-.99l-1.004-.828a1.125 1.125 0 01-.26-1.43l1.297-2.247a1.125 1.125 0 011.37-.491l1.216.456c.356.133.751.072 1.076-.124.072-.044.146-.087.22-.128.332-.183.582-.495.644-.869l.214-1.281z" /><path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" /></svg>;
  const ctConfig = CHANGE_TYPE_LABELS[changeType] ?? { label: changeType, color: 'text-text-muted bg-bg-tertiary' };
  const fieldLabel = change.field ? (FIELD_LABELS[change.field] ?? change.field) : '';

  return (
    <div className="flex items-start gap-3 py-2.5 px-3 rounded-lg hover:bg-bg-tertiary/50 transition-colors">
      <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center ${compColor}`}>
        {compIcon}
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-center gap-2 flex-wrap">
          <span className={`text-xs font-medium px-1.5 py-0.5 rounded ${compColor}`}>
            {compLabel}
          </span>
          {fieldLabel && (
            <span className="text-xs text-text-muted">
              {fieldLabel}
            </span>
          )}
          <span className={`text-xs font-medium px-1.5 py-0.5 rounded ${ctConfig.color}`}>
            {ctConfig.label}
          </span>
          <span className="text-xs text-text-muted ml-auto">
            {new Date(change.changed_at).toLocaleString('pt-BR')}
          </span>
        </div>
        {(change.old_value || change.new_value) && (
          <div className="flex items-center gap-1.5 mt-1.5 text-sm">
            {change.old_value && (
              <code className="bg-red-500/10 text-red-400 px-1.5 py-0.5 rounded text-xs">
                {change.old_value}
              </code>
            )}
            {change.old_value && change.new_value && (
              <svg className="w-3.5 h-3.5 text-text-muted flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                <path strokeLinecap="round" strokeLinejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
              </svg>
            )}
            {change.new_value && (
              <code className="bg-green-500/10 text-green-400 px-1.5 py-0.5 rounded text-xs">
                {change.new_value}
              </code>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

const FIELD_LABELS: Record<string, string> = {
  model: 'Modelo',
  cores: 'Cores',
  threads: 'Threads',
  total_bytes: 'Total',
  manufacturer: 'Fabricante',
  product: 'Produto',
  serial: 'Serial',
  vendor: 'Fabricante',
  version: 'Versão',
  disk: 'Disco',
  size_bytes: 'Tamanho',
  media_type: 'Tipo de Mídia',
  interface: 'Interface',
};

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

function EmptyState({ message }: { message: string }) {
  return (
    <div className="text-center py-12">
      <svg className="w-10 h-10 text-text-muted mx-auto mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
        <path strokeLinecap="round" strokeLinejoin="round" d="M20.25 7.5l-.625 10.632a2.25 2.25 0 01-2.247 2.118H6.622a2.25 2.25 0 01-2.247-2.118L3.75 7.5m6 4.125l2.25 2.25m0 0l2.25 2.25M12 13.875l2.25-2.25M12 13.875l-2.25 2.25M3.375 7.5h17.25c.621 0 1.125-.504 1.125-1.125v-1.5c0-.621-.504-1.125-1.125-1.125H3.375c-.621 0-1.125.504-1.125 1.125v1.5c0 .621.504 1.125 1.125 1.125z" />
      </svg>
      <p className="text-sm text-text-muted">{message}</p>
    </div>
  );
}
