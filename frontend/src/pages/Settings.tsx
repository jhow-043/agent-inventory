// Settings page with theme toggle and user management.

import { useState, type FormEvent } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTheme } from '../hooks/useTheme';
import { getUsers, createUser, deleteUser } from '../api/users';
import { PageHeader, Button, Input, Card, CardContent, Modal } from '../components/ui';

export default function Settings() {
  return (
    <div className="animate-fade-in">
      <PageHeader title="Settings" />
      <div className="space-y-6">
        <AppearanceSection />
        <UserManagementSection />
      </div>
    </div>
  );
}

// ── Appearance ──────────────────────────────────────────────

function AppearanceSection() {
  const { theme, toggleTheme } = useTheme();

  return (
    <Card>
      <CardContent>
        <h2 className="text-base font-semibold text-text-primary mb-4">Appearance</h2>
        <div className="flex items-center justify-between">
          <div>
            <p className="text-sm text-text-primary">Theme</p>
            <p className="text-xs text-text-muted mt-0.5">Choose between dark and light mode</p>
          </div>
          <button
            onClick={toggleTheme}
            className="relative inline-flex h-8 w-[120px] items-center rounded-lg bg-bg-tertiary border border-border-primary cursor-pointer transition-colors"
          >
            <span
              className={`absolute h-7 w-[58px] rounded-md bg-gradient-to-r from-accent to-accent-light transition-transform duration-200 shadow-sm ${
                theme === 'light' ? 'translate-x-[59px]' : 'translate-x-0.5'
              }`}
            />
            <span className={`relative z-10 flex-1 text-center text-xs font-medium transition-colors ${theme === 'dark' ? 'text-white' : 'text-text-secondary'}`}>
              Dark
            </span>
            <span className={`relative z-10 flex-1 text-center text-xs font-medium transition-colors ${theme === 'light' ? 'text-white' : 'text-text-secondary'}`}>
              Light
            </span>
          </button>
        </div>
      </CardContent>
    </Card>
  );
}

// ── User Management ─────────────────────────────────────────

function UserManagementSection() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['users'], queryFn: getUsers });
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; username: string } | null>(null);

  const createMutation = useMutation({
    mutationFn: () => createUser(username, password),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setUsername('');
      setPassword('');
      setError('');
    },
    onError: (err: Error) => setError(err.message),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDeleteTarget(null);
    },
    onError: (err: Error) => setError(err.message),
  });

  const handleCreate = (e: FormEvent) => {
    e.preventDefault();
    setError('');
    createMutation.mutate();
  };

  return (
    <Card>
      <CardContent>
        <h2 className="text-base font-semibold text-text-primary mb-4">User Management</h2>

        {/* Create user form */}
        <form onSubmit={handleCreate} className="flex flex-col sm:flex-row gap-3 mb-5">
          <div className="flex-1">
            <Input
              placeholder="Username"
              required
              minLength={3}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              icon={
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                </svg>
              }
            />
          </div>
          <div className="flex-1">
            <Input
              type="password"
              placeholder="Password (min 8 chars)"
              required
              minLength={8}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              icon={
                <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
                </svg>
              }
            />
          </div>
          <Button type="submit" loading={createMutation.isPending}>
            Add User
          </Button>
        </form>

        {error && (
          <div className="bg-danger/10 text-danger text-sm px-4 py-3 rounded-lg border border-danger/20 mb-4 animate-slide-up">
            {error}
          </div>
        )}

        {/* Users table */}
        {isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 2 }).map((_, i) => (
              <div key={i} className="h-12 bg-bg-tertiary rounded-lg animate-pulse" />
            ))}
          </div>
        ) : (
          <div className="overflow-x-auto rounded-lg border border-border-primary">
            <table className="w-full text-sm">
              <thead>
                <tr className="bg-bg-tertiary/50">
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Username</th>
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Created</th>
                  <th className="text-right px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-primary">
                {data?.users.map((user) => (
                  <tr key={user.id} className="hover:bg-bg-tertiary/50 transition-colors">
                    <td className="px-4 py-3 text-text-primary font-medium">{user.username}</td>
                    <td className="px-4 py-3 text-text-secondary">
                      {new Date(user.created_at).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' })}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <button
                        onClick={() => setDeleteTarget({ id: user.id, username: user.username })}
                        className="text-text-muted hover:text-danger transition-colors cursor-pointer p-1 rounded-lg hover:bg-danger/5"
                        title="Delete user"
                      >
                        <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                        </svg>
                      </button>
                    </td>
                  </tr>
                ))}
                {data?.users.length === 0 && (
                  <tr>
                    <td colSpan={3} className="text-center py-6 text-text-muted">No users found</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}

        {/* Delete Confirmation Modal */}
        <Modal
          open={!!deleteTarget}
          onClose={() => setDeleteTarget(null)}
          title="Delete User"
          actions={
            <>
              <Button variant="ghost" size="sm" onClick={() => setDeleteTarget(null)}>
                Cancel
              </Button>
              <Button
                variant="danger"
                size="sm"
                onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget.id)}
                loading={deleteMutation.isPending}
              >
                Delete
              </Button>
            </>
          }
        >
          <p className="text-sm text-text-secondary">
            Are you sure you want to delete user <strong className="text-text-primary">{deleteTarget?.username}</strong>?
          </p>
        </Modal>
      </CardContent>
    </Card>
  );
}
