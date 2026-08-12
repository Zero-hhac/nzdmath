import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { X, LogIn, Eye, EyeOff } from 'lucide-react';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { api } from '@/src/lib/api';

type LoginMode = 'user' | 'admin';

type Props = {
  open: boolean;
  onClose: () => void;
  defaultMode?: LoginMode;
};

export const LoginModal: React.FC<Props> = ({ open, onClose, defaultMode = 'user' }) => {
  const [mode, setMode] = useState<LoginMode>(defaultMode);
  const [isRegistering, setIsRegistering] = useState(false);
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [nickname, setNickname] = useState('');
  const [showPwd, setShowPwd] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const { loginUser, loginAdmin } = useAuth();
  const { showToast } = useToast();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!username.trim() || !password) {
      setError('请填写必填项');
      return;
    }
    setError('');
    setLoading(true);
    try {
      if (isRegistering) {
        await api.userRegister({ username, password, nickname });
        showToast('注册成功，请登录！', 'success');
        setIsRegistering(false);
        setPassword('');
      } else {
        if (mode === 'user') {
          await loginUser(username, password);
          showToast('登录成功', 'success');
        } else {
          await loginAdmin(username, password);
          showToast('管理员登录成功', 'success');
        }
        onClose();
        setUsername('');
        setPassword('');
      }
    } catch (err: any) {
      setError(err.message || (isRegistering ? '注册失败' : '登录失败'));
    } finally {
      setLoading(false);
    }
  };

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4"
          onClick={onClose}
        >
          <motion.div
            initial={{ scale: 0.95, opacity: 0, y: 20 }}
            animate={{ scale: 1, opacity: 1, y: 0 }}
            exit={{ scale: 0.95, opacity: 0, y: 20 }}
            className="glass-card rounded-3xl p-8 max-w-md w-full relative"
            onClick={(e) => e.stopPropagation()}
          >
            <button
              onClick={onClose}
              className="absolute top-4 right-4 w-8 h-8 flex items-center justify-center rounded-full hover:bg-white/40 transition-colors"
            >
              <X className="w-4 h-4 text-zinc-500" />
            </button>

            <div className="flex items-center gap-3 mb-6">
              <div className="w-12 h-12 rounded-2xl bg-primary/10 flex items-center justify-center">
                <LogIn className="w-5 h-5 text-primary" />
              </div>
              <div>
                <h3 className="text-xl font-serif text-primary">
                  {isRegistering ? '会员注册' : mode === 'user' ? '会员登录' : '管理员登录'}
                </h3>
                <p className="text-xs text-zinc-500">
                  {isRegistering ? '创建新会员账号' : mode === 'user' ? '登录后可使用会员专区' : '仅限管理员访问后台'}
                </p>
              </div>
            </div>

            {!isRegistering && (
              <div className="flex gap-2 mb-6 p-1 bg-zinc-100 rounded-2xl">
                <button
                  type="button"
                  onClick={() => { setMode('user'); setError(''); }}
                  className={`flex-1 py-2 text-sm font-medium rounded-xl transition-all ${
                    mode === 'user' ? 'bg-white shadow text-primary' : 'text-zinc-500'
                  }`}
                >
                  会员
                </button>
                <button
                  type="button"
                  onClick={() => { setMode('admin'); setError(''); }}
                  className={`flex-1 py-2 text-sm font-medium rounded-xl transition-all ${
                    mode === 'admin' ? 'bg-white shadow text-primary' : 'text-zinc-500'
                  }`}
                >
                  管理员
                </button>
              </div>
            )}

            <form onSubmit={submit} className="space-y-4">
              <div>
                <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                  用户名
                </label>
                <input
                  type="text"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  className="app-input w-full rounded-xl py-3 px-4"
                  placeholder={mode === 'admin' ? 'admin' : '请输入用户名'}
                  autoComplete="username"
                  autoFocus
                />
              </div>

              <div>
                <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                  密码
                </label>
                <div className="relative">
                  <input
                    type={showPwd ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="app-input w-full rounded-xl py-3 px-4 pr-12"
                    placeholder="请输入密码"
                    autoComplete={isRegistering ? 'new-password' : 'current-password'}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPwd(!showPwd)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-zinc-400 hover:text-zinc-600"
                  >
                    {showPwd ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </button>
                </div>
              </div>

              {isRegistering && (
                <div>
                  <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                    昵称 (选填)
                  </label>
                  <input
                    type="text"
                    value={nickname}
                    onChange={(e) => setNickname(e.target.value)}
                    className="app-input w-full rounded-xl py-3 px-4"
                    placeholder="请输入昵称"
                    autoComplete="nickname"
                  />
                </div>
              )}

              {error && (
                <div className="text-sm text-rose-500 bg-rose-50 px-3 py-2 rounded-xl">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="btn-primary w-full !py-3 disabled:opacity-50"
              >
                {loading ? '请稍候...' : isRegistering ? '注册账号' : '登录'}
              </button>

              {mode === 'user' && !isRegistering && (
                <p className="text-xs text-zinc-400 text-center">
                  还没有账号？
                  <button
                    type="button"
                    onClick={() => { setIsRegistering(true); setError(''); }}
                    className="text-primary ml-1"
                  >
                    立即注册
                  </button>
                </p>
              )}
              {isRegistering && (
                <p className="text-xs text-zinc-400 text-center">
                  已有账号？
                  <button
                    type="button"
                    onClick={() => { setIsRegistering(false); setError(''); }}
                    className="text-primary ml-1"
                  >
                    返回登录
                  </button>
                </p>
              )}
            </form>
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};
