// Departments CRUD page.

import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getDepartments, createDepartment, updateDepartment, deleteDepartment } from '../api/departments';

export default function Departments() {
  const queryClient = useQueryClient();
  const [newName, setNewName] = useState('');
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editName, setEditName] = useState('');

  const { data, isLoading, error } = useQuery({
    queryKey: ['departments'],
    queryFn: getDepartments,
  });

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
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['departments'] }),
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
    <div>
      <h1 className="text-xl font-semibold text-text-primary mb-6">Departments</h1>

      {/* Create form */}
      <form onSubmit={handleCreate} className="flex gap-3 mb-6">
        <input
          type="text"
          placeholder="New department name..."
          value={newName}
          onChange={(e) => setNewName(e.target.value)}
          className="bg-bg-secondary border border-border-primary rounded px-3 py-2 text-sm text-text-primary placeholder-text-muted w-full sm:w-64 focus:outline-none focus:ring-2 focus:ring-accent"
        />
        <button
          type="submit"
          disabled={createMut.isPending || !newName.trim()}
          className="px-4 py-2 text-sm bg-accent text-white rounded hover:bg-accent/80 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
        >
          {createMut.isPending ? 'Creating...' : 'Add'}
        </button>
      </form>

      {error && <p className="text-danger mb-4">Failed to load departments.</p>}

      {isLoading ? (
        <p className="text-text-muted">Loading...</p>
      ) : departments.length === 0 ? (
        <p className="text-text-muted">No departments yet.</p>
      ) : (
        <div className="bg-bg-secondary rounded-lg border border-border-primary overflow-hidden">
          <table className="min-w-full divide-y divide-border-primary">
            <thead className="bg-bg-tertiary">
              <tr>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Name</th>
                <th className="px-4 py-3 text-left text-xs font-medium text-text-muted uppercase tracking-wider">Created</th>
                <th className="px-4 py-3 text-right text-xs font-medium text-text-muted uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-primary">
              {departments.map((d) => (
                <tr key={d.id} className="hover:bg-bg-tertiary transition-colors">
                  <td className="px-4 py-3 text-sm text-text-primary">
                    {editingId === d.id ? (
                      <input
                        type="text"
                        value={editName}
                        onChange={(e) => setEditName(e.target.value)}
                        onKeyDown={(e) => e.key === 'Enter' && handleUpdate(d.id)}
                        className="bg-bg-primary border border-border-primary rounded px-2 py-1 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-accent w-48"
                        autoFocus
                      />
                    ) : (
                      d.name
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm text-text-muted">
                    {new Date(d.created_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3 text-sm text-right">
                    {editingId === d.id ? (
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => handleUpdate(d.id)}
                          disabled={updateMut.isPending}
                          className="px-2 py-1 text-xs bg-success/10 text-success rounded hover:bg-success/20 transition-colors cursor-pointer"
                        >
                          Save
                        </button>
                        <button
                          onClick={() => setEditingId(null)}
                          className="px-2 py-1 text-xs bg-bg-tertiary text-text-muted rounded hover:bg-border-primary transition-colors cursor-pointer"
                        >
                          Cancel
                        </button>
                      </div>
                    ) : (
                      <div className="flex justify-end gap-2">
                        <button
                          onClick={() => startEdit(d.id, d.name)}
                          className="px-2 py-1 text-xs bg-bg-tertiary text-text-secondary rounded hover:bg-border-primary transition-colors cursor-pointer"
                        >
                          Edit
                        </button>
                        <button
                          onClick={() => { if (confirm('Delete this department?')) deleteMut.mutate(d.id); }}
                          disabled={deleteMut.isPending}
                          className="px-2 py-1 text-xs bg-danger/10 text-danger rounded hover:bg-danger/20 transition-colors cursor-pointer"
                        >
                          Delete
                        </button>
                      </div>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
