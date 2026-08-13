import React, { useState } from 'react';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';

export function AdminLoginModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { loginAdmin } = useAuth();
  const { showToast } = useToast();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await loginAdmin(username, password);
      showToast('管理员登录成功', 'success');
      onClose();
    } catch (err: any) {
      setError(err.message || '登录失败');
    } finally {
      setLoading(false);
    }
  };

  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4" onClick={onClose}>
      <form onSubmit={submit} className="glass-card rounded-3xl p-8 max-w-md w-full" onClick={(e) => e.stopPropagation()}>
        <h3 className="text-xl font-serif text-primary mb-4">管理员登录</h3>
        <input type="text" placeholder="用户名" value={username} onChange={(e) => setUsername(e.target.value)} className="app-input w-full rounded-xl py-2.5 px-4 mb-3" />
        <input type="password" placeholder="密码" value={password} onChange={(e) => setPassword(e.target.value)} className="app-input w-full rounded-xl py-2.5 px-4 mb-3" />
        {error && <div className="text-sm text-rose-500 mb-3">{error}</div>}
        <button type="submit" disabled={loading} className="btn-primary w-full">{loading ? '登录中...' : '登录'}</button>
      </form>
    </div>
  );
}
