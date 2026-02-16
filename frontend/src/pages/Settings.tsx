// Settings page with theme toggle and user management.

import { useState, type FormEvent } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTheme } from '../hooks/useTheme';
import { getUsers, createUser, deleteUser } from '../api/users';

export default function Settings() {
  return (
    <div>
      <h1 className="text-xl font-semibold text-text-primary mb-6">Settings</h1>
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
    <div className="bg-bg-secondary rounded-lg border border-border-primary p-6">
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
            className={`absolute h-7 w-[58px] rounded-md bg-accent transition-transform duration-200 ${
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
    </div>
  );
}

// ── User Management ─────────────────────────────────────────

function UserManagementSection() {
  const queryClient = useQueryClient();
  const { data, isLoading } = useQuery({ queryKey: ['users'], queryFn: getUsers });
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);

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
      setConfirmDeleteId(null);
    },
    onError: (err: Error) => setError(err.message),
  });

  const handleCreate = (e: FormEvent) => {
    e.preventDefault();
    setError('');
    createMutation.mutate();
  };

  return (
    <div className="bg-bg-secondary rounded-lg border border-border-primary p-6">
      <h2 className="text-base font-semibold text-text-primary mb-4">User Management</h2>

      {/* Create user form */}
      <form onSubmit={handleCreate} className="flex flex-col sm:flex-row gap-3 mb-5">
        <input
          type="text"
          placeholder="Username"
          required
          minLength={3}
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          className="flex-1 bg-bg-tertiary border border-border-primary rounded px-3 py-2 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
        />
        <input
          type="password"
          placeholder="Password (min 8 chars)"
          required
          minLength={8}
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          className="flex-1 bg-bg-tertiary border border-border-primary rounded px-3 py-2 text-sm text-text-primary placeholder-text-muted focus:outline-none focus:ring-2 focus:ring-accent focus:border-transparent"
        />
        <button
          type="submit"
          disabled={createMutation.isPending}
          className="px-4 py-2 bg-accent text-white text-sm font-medium rounded hover:bg-accent-hover disabled:opacity-50 transition-colors cursor-pointer whitespace-nowrap"
        >
          {createMutation.isPending ? 'Adding...' : 'Add User'}
        </button>
      </form>

      {error && (
        <div className="bg-danger/10 text-danger text-sm px-3 py-2 rounded border border-danger/20 mb-4">
          {error}
        </div>
      )}

      {/* Users table */}
      {isLoading ? (
        <p className="text-text-muted text-sm">Loading users...</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-bg-tertiary">
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wider">Username</th>
                <th className="text-left px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wider">Created</th>
                <th className="text-right px-4 py-2.5 text-xs font-medium text-text-muted uppercase tracking-wider">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border-primary">
              {data?.users.map((user) => (
                <tr key={user.id} className="hover:bg-bg-tertiary transition-colors">
                  <td className="px-4 py-3 text-text-primary font-medium">{user.username}</td>
                  <td className="px-4 py-3 text-text-secondary">
                    {new Date(user.created_at).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' })}
                  </td>
                  <td className="px-4 py-3 text-right">
                    {confirmDeleteId === user.id ? (
                      <span className="inline-flex gap-2">
                        <button
                          onClick={() => deleteMutation.mutate(user.id)}
                          disabled={deleteMutation.isPending}
                          className="text-xs text-danger hover:text-danger/80 font-medium cursor-pointer"
                        >
                          Confirm
                        </button>
                        <button
                          onClick={() => setConfirmDeleteId(null)}
                          className="text-xs text-text-muted hover:text-text-secondary cursor-pointer"
                        >
                          Cancel
                        </button>
                      </span>
                    ) : (
                      <button
                        onClick={() => setConfirmDeleteId(user.id)}
                        className="text-xs text-text-muted hover:text-danger transition-colors cursor-pointer"
                        title="Delete user"
                      >
                        <svg className="w-4 h-4 inline" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                          <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                        </svg>
                      </button>
                    )}
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
    </div>
  );
}
