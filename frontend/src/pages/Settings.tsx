// Settings page with user management and audit logs.

import { useState, type FormEvent } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useAuth } from '../hooks/useAuth';
import { getUsers, createUser, updateUser, deleteUser } from '../api/users';
import type { User } from '../types';
import { PageHeader, Button, Input, Card, CardContent, Modal } from '../components/ui';
import { useToast } from '../hooks/useToast';
import AuditLogs from './AuditLogs';

export default function Settings() {
  const [activeTab, setActiveTab] = useState<'users' | 'audit'>('users');

  return (
    <div className="animate-fade-in">
      <PageHeader title="Settings" />

      {/* Tabs */}
      <div className="flex gap-1 mb-6 border-b border-border-primary">
        <button
          onClick={() => setActiveTab('users')}
          className={`px-4 py-2.5 text-sm font-medium transition-colors relative cursor-pointer ${
            activeTab === 'users'
              ? 'text-accent'
              : 'text-text-muted hover:text-text-primary'
          }`}
        >
          User Management
          {activeTab === 'users' && (
            <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-accent rounded-full" />
          )}
        </button>
        <button
          onClick={() => setActiveTab('audit')}
          className={`px-4 py-2.5 text-sm font-medium transition-colors relative cursor-pointer ${
            activeTab === 'audit'
              ? 'text-accent'
              : 'text-text-muted hover:text-text-primary'
          }`}
        >
          Audit Logs
          {activeTab === 'audit' && (
            <span className="absolute bottom-0 left-0 right-0 h-0.5 bg-accent rounded-full" />
          )}
        </button>
      </div>

      <div className="space-y-6">
        {activeTab === 'users' ? <UserManagementSection /> : <AuditLogs embedded />}
      </div>
    </div>
  );
}

// ── User Management ─────────────────────────────────────────

function UserManagementSection() {
  const queryClient = useQueryClient();
  const { username: currentUsername } = useAuth();
  const toast = useToast();
  const { data, isLoading } = useQuery({ queryKey: ['users'], queryFn: getUsers });
  const [username, setUsername] = useState('');
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [role, setRole] = useState('viewer');
  const [showPassword, setShowPassword] = useState(false);
  const [error, setError] = useState('');
  const [deleteTarget, setDeleteTarget] = useState<{ id: string; username: string } | null>(null);
  const [editTarget, setEditTarget] = useState<User | null>(null);
  const [editUsername, setEditUsername] = useState('');
  const [editName, setEditName] = useState('');
  const [editPassword, setEditPassword] = useState('');
  const [editRole, setEditRole] = useState('viewer');
  const [showEditPassword, setShowEditPassword] = useState(false);

  const createMutation = useMutation({
    mutationFn: () => createUser(username, name, password, role),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setUsername('');
      setName('');
      setPassword('');
      setRole('viewer');
      setShowPassword(false);
      setError('');
      toast.success('Usuário criado');
    },
    onError: (err: Error) => { setError(err.message); toast.error('Falha ao criar usuário'); },
  });

  const updateMutation = useMutation({
    mutationFn: (params: { id: string; data: { username?: string; name?: string; password?: string; role?: string } }) =>
      updateUser(params.id, params.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setEditTarget(null);
      setError('');
      toast.success('Usuário atualizado');
    },
    onError: (err: Error) => { setError(err.message); toast.error('Falha ao atualizar usuário'); },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteUser(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDeleteTarget(null);
      toast.success('Usuário excluído');
    },
    onError: (err: Error) => { setError(err.message); toast.error('Falha ao excluir usuário'); },
  });

  const handleCreate = (e: FormEvent) => {
    e.preventDefault();
    setError('');
    createMutation.mutate();
  };

  const openEdit = (user: User) => {
    setEditTarget(user);
    setEditUsername(user.username);
    setEditName(user.name || '');
    setEditPassword('');
    setEditRole(user.role);
    setShowEditPassword(false);
    setError('');
  };

  const handleEdit = (e: FormEvent) => {
    e.preventDefault();
    if (!editTarget) return;
    const payload: { username?: string; name?: string; password?: string; role?: string } = {};
    if (editUsername !== editTarget.username) payload.username = editUsername;
    if (editName !== (editTarget.name || '')) payload.name = editName;
    if (editPassword) payload.password = editPassword;
    if (editRole !== editTarget.role) payload.role = editRole;
    if (Object.keys(payload).length === 0) {
      setEditTarget(null);
      return;
    }
    updateMutation.mutate({ id: editTarget.id, data: payload });
  };

  const isSelf = (user: User) => user.username === currentUsername;

  const roleBadge = (r: string) => {
    const isAdmin = r === 'admin';
    return (
      <span className={`inline-flex items-center px-2 py-0.5 rounded-md text-xs font-medium ${
        isAdmin
          ? 'bg-accent/10 text-accent border border-accent/20'
          : 'bg-bg-tertiary text-text-secondary border border-border-primary'
      }`}>
        {isAdmin ? 'Admin' : 'Viewer'}
      </span>
    );
  };

  return (
    <Card>
      <CardContent>
        <h2 className="text-base font-semibold text-text-primary mb-4">User Management</h2>

        {/* Create user form */}
        <form onSubmit={handleCreate} className="space-y-3 mb-5">
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1">
              <Input
                placeholder="Full Name"
                required
                value={name}
                onChange={(e) => setName(e.target.value)}
                icon={
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                  </svg>
                }
              />
            </div>
            <div className="flex-1">
              <Input
                placeholder="Username"
                required
                minLength={3}
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                icon={
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M17.982 18.725A7.488 7.488 0 0012 15.75a7.488 7.488 0 00-5.982 2.975m11.963 0a9 9 0 10-11.963 0m11.963 0A8.966 8.966 0 0112 21a8.966 8.966 0 01-5.982-2.275M15 9.75a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                }
              />
            </div>
          </div>
          <div className="flex flex-col sm:flex-row gap-3">
            <div className="flex-1 relative">
              <Input
                type={showPassword ? 'text' : 'password'}
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
              <button
                type="button"
                onClick={() => setShowPassword(!showPassword)}
                className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary transition-colors cursor-pointer"
                tabIndex={-1}
              >
                {showPassword ? (
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                  </svg>
                ) : (
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                )}
              </button>
            </div>
            <div className="w-full sm:w-32">
              <select
                value={role}
                onChange={(e) => setRole(e.target.value)}
                className="w-full h-10 rounded-lg bg-bg-secondary border border-border-primary px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-accent/40 focus:border-accent transition-all cursor-pointer"
              >
                <option value="viewer">Viewer</option>
                <option value="admin">Admin</option>
              </select>
            </div>
            <Button type="submit" loading={createMutation.isPending}>
              Add User
            </Button>
          </div>
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
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Name</th>
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Username</th>
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Role</th>
                  <th className="text-left px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Created</th>
                  <th className="text-right px-4 py-2.5 text-xs font-semibold text-text-muted uppercase tracking-wider">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border-primary">
                {data?.users.map((user) => (
                  <tr key={user.id} className="hover:bg-bg-tertiary/50 transition-colors">
                    <td className="px-4 py-3 text-text-primary font-medium">
                      {user.name || '—'}
                    </td>
                    <td className="px-4 py-3 text-text-secondary">
                      {user.username}
                      {isSelf(user) && (
                        <span className="ml-2 text-xs text-text-muted">(you)</span>
                      )}
                    </td>
                    <td className="px-4 py-3">{roleBadge(user.role)}</td>
                    <td className="px-4 py-3 text-text-secondary">
                      {new Date(user.created_at).toLocaleDateString('pt-BR', { day: '2-digit', month: '2-digit', year: 'numeric' })}
                    </td>
                    <td className="px-4 py-3 text-right">
                      <div className="flex items-center justify-end gap-1">
                        {/* Edit button */}
                        <button
                          onClick={() => openEdit(user)}
                          className="text-text-muted hover:text-accent transition-colors cursor-pointer p-1 rounded-lg hover:bg-accent/5"
                          title="Edit user"
                        >
                          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                          </svg>
                        </button>
                        {/* Delete button — disabled for self */}
                        <button
                          onClick={() => setDeleteTarget({ id: user.id, username: user.username })}
                          disabled={isSelf(user)}
                          className={`p-1 rounded-lg transition-colors ${
                            isSelf(user)
                              ? 'text-text-muted/30 cursor-not-allowed'
                              : 'text-text-muted hover:text-danger cursor-pointer hover:bg-danger/5'
                          }`}
                          title={isSelf(user) ? 'Cannot delete yourself' : 'Delete user'}
                        >
                          <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
                          </svg>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
                {data?.users.length === 0 && (
                  <tr>
                    <td colSpan={5} className="text-center py-6 text-text-muted">No users found</td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        )}

        {/* Edit User Modal */}
        <Modal
          open={!!editTarget}
          onClose={() => setEditTarget(null)}
          title="Edit User"
          actions={
            <>
              <Button variant="ghost" size="sm" onClick={() => setEditTarget(null)}>
                Cancel
              </Button>
              <Button
                size="sm"
                onClick={handleEdit}
                loading={updateMutation.isPending}
              >
                Save Changes
              </Button>
            </>
          }
        >
          <form onSubmit={handleEdit} className="space-y-4">
            <div>
              <label className="block text-xs font-medium text-text-muted mb-1.5">Name</label>
              <Input
                value={editName}
                onChange={(e) => setEditName(e.target.value)}
                placeholder="Full name"
                icon={
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                  </svg>
                }
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-text-muted mb-1.5">Username</label>
              <Input
                value={editUsername}
                onChange={(e) => setEditUsername(e.target.value)}
                minLength={3}
                icon={
                  <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                    <path strokeLinecap="round" strokeLinejoin="round" d="M15.75 6a3.75 3.75 0 11-7.5 0 3.75 3.75 0 017.5 0zM4.501 20.118a7.5 7.5 0 0114.998 0A17.933 17.933 0 0112 21.75c-2.676 0-5.216-.584-7.499-1.632z" />
                  </svg>
                }
              />
            </div>
            <div>
              <label className="block text-xs font-medium text-text-muted mb-1.5">New Password <span className="text-text-muted font-normal">(leave blank to keep current)</span></label>
              <div className="relative">
                <Input
                  type={showEditPassword ? 'text' : 'password'}
                  placeholder="••••••••"
                  value={editPassword}
                  onChange={(e) => setEditPassword(e.target.value)}
                  minLength={8}
                  icon={
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M16.5 10.5V6.75a4.5 4.5 0 10-9 0v3.75m-.75 11.25h10.5a2.25 2.25 0 002.25-2.25v-6.75a2.25 2.25 0 00-2.25-2.25H6.75a2.25 2.25 0 00-2.25 2.25v6.75a2.25 2.25 0 002.25 2.25z" />
                    </svg>
                  }
                />
                <button
                  type="button"
                  onClick={() => setShowEditPassword(!showEditPassword)}
                  className="absolute right-3 top-1/2 -translate-y-1/2 text-text-muted hover:text-text-primary transition-colors cursor-pointer"
                  tabIndex={-1}
                >
                  {showEditPassword ? (
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M3.98 8.223A10.477 10.477 0 001.934 12C3.226 16.338 7.244 19.5 12 19.5c.993 0 1.953-.138 2.863-.395M6.228 6.228A10.45 10.45 0 0112 4.5c4.756 0 8.773 3.162 10.065 7.498a10.523 10.523 0 01-4.293 5.774M6.228 6.228L3 3m3.228 3.228l3.65 3.65m7.894 7.894L21 21m-3.228-3.228l-3.65-3.65m0 0a3 3 0 10-4.243-4.243m4.242 4.242L9.88 9.88" />
                    </svg>
                  ) : (
                    <svg className="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={1.5}>
                      <path strokeLinecap="round" strokeLinejoin="round" d="M2.036 12.322a1.012 1.012 0 010-.639C3.423 7.51 7.36 4.5 12 4.5c4.638 0 8.573 3.007 9.963 7.178.07.207.07.431 0 .639C20.577 16.49 16.64 19.5 12 19.5c-4.638 0-8.573-3.007-9.963-7.178z" />
                      <path strokeLinecap="round" strokeLinejoin="round" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                    </svg>
                  )}
                </button>
              </div>
            </div>
            <div>
              <label className="block text-xs font-medium text-text-muted mb-1.5">Role</label>
              <select
                value={editRole}
                onChange={(e) => setEditRole(e.target.value)}
                disabled={editTarget ? isSelf(editTarget) : false}
                className="w-full h-10 rounded-lg bg-bg-secondary border border-border-primary px-3 text-sm text-text-primary focus:outline-none focus:ring-2 focus:ring-accent/40 focus:border-accent transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <option value="viewer">Viewer</option>
                <option value="admin">Admin</option>
              </select>
              {editTarget && isSelf(editTarget) && (
                <p className="text-xs text-text-muted mt-1">You cannot change your own role</p>
              )}
            </div>
          </form>
        </Modal>

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
