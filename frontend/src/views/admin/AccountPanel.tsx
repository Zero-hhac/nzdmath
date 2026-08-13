import { useState } from 'react';
import { KeyRound } from 'lucide-react';
import { api } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';

/**
 * 账号设置：管理员自助修改登录密码。
 * 规则与会员一致：至少 6 位，必须同时包含字母和数字（后端共用校验）。
 */
export function AccountPanel() {
  const { showToast } = useToast();
  const [oldPwd, setOldPwd] = useState('');
  const [newPwd, setNewPwd] = useState('');
  const [confirmPwd, setConfirmPwd] = useState('');
  const [saving, setSaving] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!oldPwd || !newPwd) {
      showToast('请填写完整', 'error');
      return;
    }
    if (newPwd.length < 6 || !/(?=.*[A-Za-z])(?=.*\d)/.test(newPwd)) {
      showToast('新密码至少 6 位，且必须同时包含字母和数字', 'error');
      return;
    }
    if (newPwd !== confirmPwd) {
      showToast('两次输入的新密码不一致', 'error');
      return;
    }
    setSaving(true);
    try {
      await api.adminChangePassword(oldPwd, newPwd);
      showToast('密码修改成功', 'success');
      setOldPwd('');
      setNewPwd('');
      setConfirmPwd('');
    } catch (err: any) {
      showToast(err.message || '修改失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="sidebar-panel rounded-[2rem] p-6 md:p-8 max-w-lg">
      <div className="flex items-center gap-3 mb-6">
        <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
          <KeyRound className="w-5 h-5 text-primary" />
        </div>
        <div>
          <h3 className="text-lg font-semibold text-charcoal">账号设置</h3>
          <p className="text-xs text-zinc-500">修改管理员登录密码</p>
        </div>
      </div>

      <form onSubmit={submit} className="space-y-4">
        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">当前密码</label>
          <input
            type="password"
            value={oldPwd}
            onChange={(e) => setOldPwd(e.target.value)}
            className="app-input w-full rounded-xl py-2.5 px-4"
            autoComplete="current-password"
          />
        </div>
        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">
            新密码（至少 6 位，需包含字母和数字）
          </label>
          <input
            type="password"
            value={newPwd}
            onChange={(e) => setNewPwd(e.target.value)}
            className="app-input w-full rounded-xl py-2.5 px-4"
            autoComplete="new-password"
          />
        </div>
        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">确认新密码</label>
          <input
            type="password"
            value={confirmPwd}
            onChange={(e) => setConfirmPwd(e.target.value)}
            className="app-input w-full rounded-xl py-2.5 px-4"
            autoComplete="new-password"
          />
        </div>
        <button type="submit" disabled={saving} className="btn-primary flex items-center gap-2">
          <KeyRound className="w-4 h-4" />
          {saving ? '提交中...' : '确认修改'}
        </button>
      </form>
    </div>
  );
}
