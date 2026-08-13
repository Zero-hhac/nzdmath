import React, { useState } from 'react';
import { motion, AnimatePresence } from 'motion/react';
import { X, LogIn, Eye, EyeOff } from 'lucide-react';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { api, DEPARTMENTS } from '@/src/lib/api';

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
  const [realName, setRealName] = useState('');
  const [className, setClassName] = useState('');
  const [department, setDepartment] = useState('');
  const [showPwd, setShowPwd] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  // 找回密码流程：request=第一步(用户名+邮箱)，reset=第二步(验证码+新密码)
  const [forgotStep, setForgotStep] = useState<'request' | 'reset' | null>(null);
  const [email, setEmail] = useState('');
  const [resetCode, setResetCode] = useState('');
  const [resetPwd, setResetPwd] = useState('');
  const [resetConfirm, setResetConfirm] = useState('');

  const { loginUser, loginAdmin } = useAuth();
  const { showToast } = useToast();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (forgotStep === 'request') {
      if (!username.trim() || !email.trim()) {
        setError('请填写用户名和邮箱');
        return;
      }
      setError('');
      setNotice('');
      setLoading(true);
      try {
        const res = await api.forgotPassword(username.trim(), email.trim());
        setNotice(res.data?.message || '验证码已发送，请查收');
        if (res.data?.dev_code) setResetCode(res.data.dev_code);
        setForgotStep('reset');
      } catch (err: any) {
        setError(err.message || '请求失败');
      } finally {
        setLoading(false);
      }
      return;
    }
    if (forgotStep === 'reset') {
      if (!resetCode.trim()) {
        setError('请输入验证码');
        return;
      }
      if (resetPwd.length < 6 || !/(?=.*[A-Za-z])(?=.*\d)/.test(resetPwd)) {
        setError('新密码至少 6 位，且必须同时包含字母和数字');
        return;
      }
      if (resetPwd !== resetConfirm) {
        setError('两次输入的新密码不一致');
        return;
      }
      setError('');
      setLoading(true);
      try {
        await api.resetPassword(username.trim(), email.trim(), resetCode.trim(), resetPwd);
        showToast('密码重置成功，请使用新密码登录', 'success');
        setForgotStep(null);
        setResetCode('');
        setResetPwd('');
        setResetConfirm('');
        setPassword('');
      } catch (err: any) {
        setError(err.message || '重置失败');
      } finally {
        setLoading(false);
      }
      return;
    }
    if (!username.trim() || !password) {
      setError('请填写必填项');
      return;
    }
    if (isRegistering && (!realName.trim() || !className.trim() || !department)) {
      setError('请填写姓名、班级并选择部门');
      return;
    }
    if (isRegistering && (password.length < 6 || !/(?=.*[A-Za-z])(?=.*\d)/.test(password))) {
      setError('密码至少 6 位，且必须同时包含字母和数字');
      return;
    }
    setError('');
    setLoading(true);
    try {
      if (isRegistering) {
        await api.userRegister({ username, password, nickname, email: email.trim(), real_name: realName.trim(), class_name: className.trim(), department });
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
                  {forgotStep ? '找回密码' : isRegistering ? '会员注册' : mode === 'user' ? '会员登录' : '管理员登录'}
                </h3>
                <p className="text-xs text-zinc-500">
                  {forgotStep === 'request'
                    ? '输入用户名和注册邮箱获取验证码'
                    : forgotStep === 'reset'
                      ? '输入邮箱收到的验证码并设置新密码'
                      : isRegistering
                        ? '创建新会员账号'
                        : mode === 'user'
                          ? '登录后可使用会员专区'
                          : '仅限管理员访问后台'}
                </p>
              </div>
            </div>

            {!isRegistering && !forgotStep && (
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
                  readOnly={forgotStep === 'reset'}
                />
              </div>

              {forgotStep === 'request' && (
                <div>
                  <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                    注册邮箱
                  </label>
                  <input
                    type="email"
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="app-input w-full rounded-xl py-3 px-4"
                    placeholder="请输入注册时填写的邮箱"
                    autoComplete="email"
                  />
                  <p className="text-xs text-zinc-400 mt-1.5">
                    未绑定邮箱的账号无法自助找回，请联系管理员重置密码。
                  </p>
                </div>
              )}

              {forgotStep === 'reset' && (
                <>
                  <div>
                    <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                      验证码
                    </label>
                    <input
                      type="text"
                      value={resetCode}
                      onChange={(e) => setResetCode(e.target.value)}
                      className="app-input w-full rounded-xl py-3 px-4 font-mono tracking-widest"
                      placeholder="6 位验证码"
                      maxLength={6}
                      autoFocus
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                      新密码（至少 6 位，需包含字母和数字）
                    </label>
                    <input
                      type={showPwd ? 'text' : 'password'}
                      value={resetPwd}
                      onChange={(e) => setResetPwd(e.target.value)}
                      className="app-input w-full rounded-xl py-3 px-4"
                      placeholder="请输入新密码"
                      autoComplete="new-password"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                      确认新密码
                    </label>
                    <input
                      type={showPwd ? 'text' : 'password'}
                      value={resetConfirm}
                      onChange={(e) => setResetConfirm(e.target.value)}
                      className="app-input w-full rounded-xl py-3 px-4"
                      placeholder="请再次输入新密码"
                      autoComplete="new-password"
                    />
                  </div>
                </>
              )}

              {!forgotStep && (
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
              )}

              {isRegistering && (
                <>
                  <div>
                    <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                      邮箱 (选填，用于找回密码)
                    </label>
                    <input
                      type="email"
                      value={email}
                      onChange={(e) => setEmail(e.target.value)}
                      className="app-input w-full rounded-xl py-3 px-4"
                      placeholder="建议填写，忘记密码时可自助找回"
                      autoComplete="email"
                    />
                  </div>
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
                  <div>
                    <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                      姓名 <span className="text-rose-500">*</span>
                    </label>
                    <input
                      type="text"
                      value={realName}
                      onChange={(e) => setRealName(e.target.value)}
                      className="app-input w-full rounded-xl py-3 px-4"
                      placeholder="请输入真实姓名"
                      autoComplete="name"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                      班级 <span className="text-rose-500">*</span>
                    </label>
                    <input
                      type="text"
                      value={className}
                      onChange={(e) => setClassName(e.target.value)}
                      className="app-input w-full rounded-xl py-3 px-4"
                      placeholder="请输入所在班级"
                    />
                  </div>
                  <div>
                    <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
                      注册部门 <span className="text-rose-500">*</span>
                    </label>
                    <select
                      value={department}
                      onChange={(e) => setDepartment(e.target.value)}
                      className="app-input w-full rounded-xl py-3 px-4"
                    >
                      <option value="">请选择部门</option>
                      {DEPARTMENTS.map((d) => (
                        <option key={d} value={d}>{d}</option>
                      ))}
                    </select>
                  </div>
                  <p className="text-xs text-zinc-400">
                    姓名、部门用于内部统计与会员名片展示（仅登录会员可见），班级仅后台可见。
                  </p>
                </>
              )}

              {error && (
                <div className="text-sm text-rose-500 bg-rose-50 px-3 py-2 rounded-xl">
                  {error}
                </div>
              )}
              {notice && (
                <div className="text-sm text-emerald-600 bg-emerald-50 px-3 py-2 rounded-xl">
                  {notice}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className="btn-primary w-full !py-3 disabled:opacity-50"
              >
                {loading
                  ? '请稍候...'
                  : forgotStep === 'request'
                    ? '发送验证码'
                    : forgotStep === 'reset'
                      ? '重置密码'
                      : isRegistering
                        ? '注册账号'
                        : '登录'}
              </button>

              {mode === 'user' && !isRegistering && !forgotStep && (
                <div className="flex items-center justify-between text-xs text-zinc-400">
                  <span>
                    还没有账号？
                    <button
                      type="button"
                      onClick={() => { setIsRegistering(true); setError(''); }}
                      className="text-primary ml-1"
                    >
                      立即注册
                    </button>
                  </span>
                  <button
                    type="button"
                    onClick={() => { setForgotStep('request'); setError(''); setNotice(''); }}
                    className="text-primary"
                  >
                    忘记密码？
                  </button>
                </div>
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
              {forgotStep && (
                <p className="text-xs text-zinc-400 text-center">
                  想起密码了？
                  <button
                    type="button"
                    onClick={() => { setForgotStep(null); setError(''); setNotice(''); setResetCode(''); setResetPwd(''); setResetConfirm(''); }}
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
