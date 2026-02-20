// Audit Logs page — admin only.

import { useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { getAuditLogs } from '../api/audit';
import { PageHeader, Card, CardContent } from '../components/ui';

const PAGE_SIZE = 20;

const ACTION_LABELS: Record<string, { label: string; color: string }> = {
  create: { label: 'Criar', color: 'bg-success/10 text-success' },
  update: { label: 'Atualizar', color: 'bg-info/10 text-info' },
  delete: { label: 'Excluir', color: 'bg-danger/10 text-danger' },
  login: { label: 'Login', color: 'bg-accent/10 text-accent' },
  logout: { label: 'Logout', color: 'bg-bg-tertiary text-text-muted' },
  activate: { label: 'Ativar', color: 'bg-success/10 text-success' },
  deactivate: { label: 'Desativar', color: 'bg-warning/10 text-warning' },
  bulk_update: { label: 'Atualização em Massa', color: 'bg-info/10 text-info' },
  bulk_delete: { label: 'Exclusão em Massa', color: 'bg-danger/10 text-danger' },
};

const RESOURCE_LABELS: Record<string, string> = {
  device: 'Dispositivo',
  user: 'Usuário',
  department: 'Departamento',
  enrollment_key: 'Chave de Registro',
};

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'agora';
  if (mins < 60) return `${mins}m atrás`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours}h atrás`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days}d atrás`;
  return new Date(dateStr).toLocaleDateString('pt-BR');
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('pt-BR', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
}

export default function AuditLogs({ embedded }: { embedded?: boolean }) {
  const [page, setPage] = useState(1);
  const [actionFilter, setActionFilter] = useState('');
  const [resourceFilter, setResourceFilter] = useState('');

  const offset = (page - 1) * PAGE_SIZE;

  const { data, isLoading, error } = useQuery({
    queryKey: ['audit-logs', actionFilter, resourceFilter, page],
    queryFn: () =>
      getAuditLogs({
        action: actionFilter || undefined,
        resource_type: resourceFilter || undefined,
        limit: PAGE_SIZE,
        offset,
      }),
  });

  const logs = data?.logs ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.ceil(total / PAGE_SIZE);

  return (
    <div className={embedded ? '' : 'animate-fade-in'}>
      {!embedded && <PageHeader title="Audit Logs" subtitle={`${total} registros`} />}

      {/* Filters */}
      <div className="flex flex-wrap gap-3 mb-6">
        <select
          value={actionFilter}
          onChange={(e) => { setActionFilter(e.target.value); setPage(1); }}
          className="px-3 py-2 rounded-lg border border-border-primary bg-bg-secondary text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50"
        >
          <option value="">Todas as ações</option>
          {Object.entries(ACTION_LABELS).map(([key, { label }]) => (
            <option key={key} value={key}>{label}</option>
          ))}
        </select>
        <select
          value={resourceFilter}
          onChange={(e) => { setResourceFilter(e.target.value); setPage(1); }}
          className="px-3 py-2 rounded-lg border border-border-primary bg-bg-secondary text-text-primary text-sm focus:outline-none focus:ring-2 focus:ring-accent/50"
        >
          <option value="">Todos os recursos</option>
          {Object.entries(RESOURCE_LABELS).map(([key, label]) => (
            <option key={key} value={key}>{label}</option>
          ))}
        </select>
      </div>

      {/* Table */}
      <Card>
        <CardContent className="p-0">
          {isLoading ? (
            <div className="p-8 space-y-3">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="h-12 bg-bg-tertiary rounded-lg animate-pulse" />
              ))}
            </div>
          ) : error ? (
            <div className="p-8 text-center text-danger">Erro ao carregar audit logs.</div>
          ) : logs.length === 0 ? (
            <div className="p-8 text-center text-text-muted">Nenhum registro encontrado.</div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-border-primary text-left text-text-secondary">
                    <th className="px-4 py-3 font-medium">Quando</th>
                    <th className="px-4 py-3 font-medium">Usuário</th>
                    <th className="px-4 py-3 font-medium">Ação</th>
                    <th className="px-4 py-3 font-medium">Recurso</th>
                    <th className="px-4 py-3 font-medium">Detalhes</th>
                    <th className="px-4 py-3 font-medium">IP</th>
                  </tr>
                </thead>
                <tbody>
                  {logs.map((log) => {
                    const actionInfo = ACTION_LABELS[log.action] ?? { label: log.action, color: 'bg-bg-tertiary text-text-muted' };
                    const resourceLabel = RESOURCE_LABELS[log.resource_type] ?? log.resource_type;

                    let parsedDetails: string | null = null;
                    if (log.details) {
                      try {
                        const obj = JSON.parse(log.details);
                        parsedDetails = Object.entries(obj)
                          .map(([k, v]) => `${k}: ${v}`)
                          .join(', ');
                      } catch {
                        parsedDetails = log.details;
                      }
                    }

                    return (
                      <tr key={log.id} className="border-b border-border-primary/50 last:border-0 hover:bg-bg-tertiary/50 transition-colors">
                        <td className="px-4 py-3 whitespace-nowrap">
                          <div className="text-text-primary">{timeAgo(log.created_at)}</div>
                          <div className="text-xs text-text-muted">{formatDate(log.created_at)}</div>
                        </td>
                        <td className="px-4 py-3 text-text-primary font-medium">{log.username}</td>
                        <td className="px-4 py-3">
                          <span className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium ${actionInfo.color}`}>
                            {actionInfo.label}
                          </span>
                        </td>
                        <td className="px-4 py-3">
                          <span className="text-text-secondary">{resourceLabel}</span>
                          {log.resource_id && (
                            <span className="text-xs text-text-muted ml-1">
                              ({log.resource_id.substring(0, 8)}…)
                            </span>
                          )}
                        </td>
                        <td className="px-4 py-3 text-text-muted text-xs max-w-xs truncate" title={parsedDetails ?? ''}>
                          {parsedDetails || '—'}
                        </td>
                        <td className="px-4 py-3 text-text-muted text-xs whitespace-nowrap">{log.ip_address || '—'}</td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}
        </CardContent>
      </Card>

      {/* Pagination */}
      {totalPages > 1 && (
        <div className="flex items-center justify-between mt-4">
          <span className="text-sm text-text-muted">
            Página {page} de {totalPages} ({total} registros)
          </span>
          <div className="flex gap-2">
            <button
              onClick={() => setPage((p) => Math.max(1, p - 1))}
              disabled={page === 1}
              className="px-3 py-1.5 text-sm rounded-lg border border-border-primary bg-bg-secondary text-text-primary hover:bg-bg-tertiary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Anterior
            </button>
            <button
              onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
              disabled={page === totalPages}
              className="px-3 py-1.5 text-sm rounded-lg border border-border-primary bg-bg-secondary text-text-primary hover:bg-bg-tertiary disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
            >
              Próxima
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
