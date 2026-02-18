// Departments CRUD page.

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router-dom';
import { getDepartments, createDepartment, updateDepartment, deleteDepartment } from '../api/departments';
import { getDevices } from '../api/devices';
import { PageHeader, Button, Input, Modal, Badge } from '../components/ui';

export default function Departments() {
  const queryClient = useQueryClient();
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; name: string } | null>(null);

  const { data, isLoading, error } = useQuery({
    queryKey: ['departments'],
    queryFn: getDepartments,
  });

  // Fetch all devices to count per department
  const { data: devicesData } = useQuery({
    queryKey: ['devices', '', '', '', '', 'hostname', 'asc', 1],
    queryFn: () => getDevices({ page: 1, limit: 10000 }),
  });
  const allDevices = devicesData?.devices ?? [];

  const createMut = useMutation({
    mutationFn: (name: string) => createDepartment(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['departments'] });
      setNewName('');
    },
  });

  const updateMut = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => updateDepartment(id, name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['departments'] });
      setEditingId(null);
    },
  });

  const deleteMut = useMutation({
    mutationFn: (id: string) => deleteDepartment(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['departments'] });
      setDeleteTarget(null);
    },
  });

  const departments = data?.departments ?? [];

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = newName.trim();
    if (trimmed) createMut.mutate(trimmed);
  };

  const startEdit = (id: string, name: string) => {
    setEditingId(id);
    setEditName(name);
  };

  const handleUpdate = (id: string) => {
    const trimmed = editName.trim();
    if (trimmed) updateMut.mutate({ id, name: trimmed });
  };

  return (
    <div className="animate-fade-in">
      <PageHeader title="Departments" subtitle={`${departments.length} departments`} />

      {/* Create form */}
      <form onSubmit={handleCreate} className="flex gap-3 mb-6">
        <div className="w-full sm:w-64">
          <Input
            placeholder="New department name..."
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
          />
        </div>
        <Button
          type="submit"
          disabled={createMut.isPending || !newName.trim()}
          loading={createMut.isPending}
        >
          Add
        </Button>
      </form>

      {error && <p className="text-danger mb-4">Failed to load departments.</p>}

      {isLoading ? (
        <div className="space-y-3">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-14 bg-bg-secondary rounded-xl border border-border-primary animate-pulse" />
          ))}
        </div>
      ) : departments.length === 0 ? (
        <div className="text-center py-12 text-text-muted">
          <svg className="w-10 h-10 mx-auto mb-3 text-text-muted/40" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1}>
            <path strokeLinecap="round" strokeLinejoin="round" d="M2.25 21h19.5m-18-18v18m10.5-18v18m6-13.5V21M6.75 6.75h.75m-.75 3h.75m-.75 3h.75m3-6h.75m-.75 3h.75m-.75 3h.75M6.75 21v-3.375c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125V21M3 3h12m-.75 4.5H21m-3.75 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008zm0 3h.008v.008h-.008v-.008z" />
          </svg>
          <p className="text-sm">No departments yet.</p>
        </div>
      ) : (
        <div className="bg-bg-secondary rounded-xl border border-border-primary overflow-hidden shadow-sm animate-slide-up">
          <table className="min-w-full divide-y divide-border-primary">
            <thead className="bg-bg-tertiary/50">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">Devices</th>
                <th className="px-4 py-3 text-left text-xs font-semibold text-text-muted uppercase tracking-wider">Created</th>
                <th className="px-4 py-3 text-right text-xs font-semibold text-text-muted uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-primary">
              {departments.map((d) => (
                <tr key={d.id} className="hover:bg-bg-tertiary/50 transition-colors">
                  <td className="px-4 py-3 text-sm text-text-primary">
                    {editingId === d.id ? (
                      <Input
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleUpdate(d.id)}
                        className="w-48"
                        autoFocus
                      />
                    ) : (
                      <Link to={`/departments/${d.id}`} className="font-medium text-accent hover:text-accent-hover transition-colors">
                        {d.name}
                      </Link>
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm">
                    <Link to={`/departments/${d.id}`}>
                      <Badge variant={allDevices.filter((dev) => dev.department_id === d.id).length > 0 ? 'accent' : 'neutral'}>
                        {allDevices.filter((dev) => dev.department_id === d.id).length}
                      </Badge>
                    </Link>
                  </td>
                  <td className="px-4 py-3 text-sm text-text-muted">
                    {new Date(d.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-sm text-right">
                    {editingId === d.id ? (
                      <div className="flex justify-end gap-2">
                        <Button variant="success" size="sm" onClick={() => handleUpdate(d.id)} loading={updateMut.isPending}>
                          Save
                        </Button>
                        <Button variant="ghost" size="sm" onClick={() => setEditingId(null)}>
                          Cancel
                        </Button>
                      </div>
                    ) : (
                      <div className="flex justify-end gap-2">
                        <Button variant="ghost" size="sm" onClick={() => startEdit(d.id, d.name)}>
                          Edit
                        </Button>
                        <Button variant="danger" size="sm" onClick={() => setDeleteTarget({ id: d.id, name: d.name })}>
                          Delete
                        </Button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {/* Delete Confirmation Modal */}
      <Modal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title="Delete Department"
        actions={
          <>
            <Button variant="ghost" size="sm" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="danger"
              size="sm"
              onClick={() => deleteTarget && deleteMut.mutate(deleteTarget.id)}
              loading={deleteMut.isPending}
            >
              Delete
            </Button>
          </>
        }
      >
        <p className="text-sm text-text-secondary">
          Are you sure you want to delete <strong className="text-text-primary">{deleteTarget?.name}</strong>? This action cannot be undone.
        </p>
      </Modal>
    </div>
  );
}
