import { useEffect, useState } from 'react';
import { Search, Download, Trash2, XCircle, CheckCircle2 } from 'lucide-react';
import { api, DEPARTMENTS } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';

export function UsersPanel() {
  const { showToast } = useToast();
  const [items, setItems] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [department, setDepartment] = useState('');
  const [incomplete, setIncomplete] = useState(false);
  const [resetTarget, setResetTarget] = useState<any>(null);
  const [newPwd, setNewPwd] = useState('');
  const [selected, setSelected] = useState<number[]>([]);
  const [batchReset, setBatchReset] = useState(false);

  const load = () => {
    setSelected([]);
    setLoading(true);
    const params: Record<string, string> = {};
    if (keyword) params.keyword = keyword;
    if (department) params.department = department;
    if (incomplete) params.incomplete = '1';
    api.adminListUsers(params).then((res) => setItems(res.data || [])).catch(() => setItems([])).finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

  const handleExport = async () => {
    try {
      await api.adminExportUsers(department || undefined);
      showToast('导出成功', 'success');
    } catch (err: any) {
      showToast(err.message || '导出失败', 'error');
    }
  };

  const handleToggleStatus = async (u: any) => {
    const newStatus = u.status === 1 ? 0 : 1;
    try {
      await api.adminSetUserStatus(u.id, newStatus as 0 | 1);
      showToast(newStatus === 1 ? '已启用' : '已禁用', 'success');
      load();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const toggleSelect = (id: number) => {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]));
  };

  const toggleSelectAll = () => {
    setSelected((prev) => (prev.length === items.length ? [] : items.map((u) => u.id)));
  };

  const handleBatchStatus = async (status: 0 | 1) => {
    try {
      const res = await api.adminBatchSetUserStatus(selected, status);
      showToast(`已${status === 1 ? '启用' : '禁用'} ${res.data?.affected ?? selected.length} 个用户`, 'success');
      load();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const handleBatchDelete = async () => {
    if (!confirm(`确定要删除选中的 ${selected.length} 个用户吗？`)) return;
    try {
      const res = await api.adminBatchDeleteUsers(selected);
      showToast(`已删除 ${res.data?.affected ?? selected.length} 个用户`, 'success');
      load();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const handleDelete = async (u: any) => {
    if (!confirm(`确定要删除用户 ${u.username} 吗？`)) return;
    try {
      await api.adminDeleteUser(u.id);
      showToast('已删除', 'success');
      load();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const handleResetPwd = async () => {
    if (!newPwd || newPwd.length < 6 || !/(?=.*[A-Za-z])(?=.*\d)/.test(newPwd)) {
      showToast('密码至少 6 位，且必须同时包含字母和数字', 'error');
      return;
    }
    try {
      if (batchReset) {
        const res = await api.adminBatchResetUserPassword(selected, newPwd);
        showToast(`已重置 ${res.data?.affected ?? selected.length} 个用户的密码`, 'success');
        setBatchReset(false);
      } else if (resetTarget) {
        await api.adminResetUserPassword(resetTarget.id, newPwd);
        showToast(`已重置 ${resetTarget.username} 的密码`, 'success');
      }
      setResetTarget(null);
      setNewPwd('');
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="sidebar-panel rounded-[2rem] p-6">
        <div className="flex gap-3 items-center mb-4">
          <div className="relative flex-1">
            <input
              type="text"
              placeholder="搜索用户名/邮箱/昵称/姓名..."
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && load()}
              className="app-input w-full rounded-full py-2 pl-9 pr-4"
            />
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
          </div>
          <select value={department} onChange={(e) => { setDepartment(e.target.value); setTimeout(load, 0); }} className="app-input rounded-xl px-3 py-2 text-sm">
            <option value="">全部部门</option>
            {DEPARTMENTS.map((d) => (
              <option key={d} value={d}>{d}</option>
            ))}
          </select>
          <button
            onClick={() => { setIncomplete(!incomplete); setTimeout(load, 0); }}
            title="筛选姓名/班级/部门未完善的会员"
            className={`px-3 py-2 rounded-xl text-sm whitespace-nowrap ${incomplete ? 'bg-primary text-white' : 'bg-white text-zinc-600 border border-border'}`}
          >
            未完善会员
          </button>
          <button onClick={load} className="btn-secondary !py-2 !text-xs">查询</button>
          <button onClick={handleExport} className="btn-primary !py-2 !text-xs flex items-center gap-1.5">
            <Download className="w-3.5 h-3.5" /> 导出 Excel
          </button>
        </div>

        {selected.length > 0 && (
          <div className="flex items-center gap-2 mb-3 px-4 py-2 rounded-xl bg-primary/5 ring-1 ring-primary/20 text-sm">
            <span className="font-medium text-zinc-700">已选 {selected.length} 人</span>
            <div className="flex-1" />
            <button onClick={() => handleBatchStatus(1)} className="px-3 py-1 rounded-lg text-xs text-emerald-600 hover:bg-emerald-50">批量启用</button>
            <button onClick={() => handleBatchStatus(0)} className="px-3 py-1 rounded-lg text-xs text-amber-600 hover:bg-amber-50">批量禁用</button>
            <button onClick={() => { setBatchReset(true); setNewPwd(''); }} className="px-3 py-1 rounded-lg text-xs text-zinc-600 hover:bg-white">重置密码</button>
            <button onClick={handleBatchDelete} className="px-3 py-1 rounded-lg text-xs text-rose-500 hover:bg-rose-50 flex items-center gap-1">
              <Trash2 className="w-3.5 h-3.5" /> 批量删除
            </button>
            <button onClick={() => setSelected([])} className="px-2 py-1 rounded-lg text-xs text-zinc-400 hover:bg-white">取消选择</button>
          </div>
        )}

        {loading ? (
          <div className="text-center py-12 text-zinc-500">加载中...</div>
        ) : items.length === 0 ? (
          <div className="text-center py-12 text-zinc-500">暂无用户</div>
        ) : (
          <div className="space-y-2">
            <div className="flex items-center gap-3 px-3 pb-1 text-xs text-zinc-400">
              <input
                type="checkbox"
                checked={selected.length > 0 && selected.length === items.length}
                onChange={toggleSelectAll}
                className="w-4 h-4 accent-primary cursor-pointer"
              />
              <span>全选本页（{items.length}）</span>
            </div>
            {items.map((u) => (
              <div key={u.id} className={`flex items-center gap-4 p-3 rounded-xl ${selected.includes(u.id) ? 'bg-primary/5 ring-1 ring-primary/20' : 'hover:bg-white/60'}`}>
                <input
                  type="checkbox"
                  checked={selected.includes(u.id)}
                  onChange={() => toggleSelect(u.id)}
                  className="w-4 h-4 accent-primary cursor-pointer shrink-0"
                />
                <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary text-sm font-bold">
                  {(u.real_name || u.username)?.[0]?.toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-zinc-800 truncate">
                    {u.real_name || '未填写'} <span className="text-xs text-zinc-400">({u.username})</span>
                  </div>
                  <div className="text-xs text-zinc-500 mt-0.5">
                    {u.department || '未分配'} · {u.class_name || '未填写'} · {u.status === 1 ? '✓ 正常' : '✕ 禁用'}
                  </div>
                </div>
                <button onClick={() => handleToggleStatus(u)} className={`px-3 py-1.5 rounded-lg text-xs flex items-center gap-1 ${u.status === 1 ? 'text-amber-600 hover:bg-amber-50' : 'text-emerald-600 hover:bg-emerald-50'}`}>
                  {u.status === 1 ? <XCircle className="w-3.5 h-3.5" /> : <CheckCircle2 className="w-3.5 h-3.5" />}
                  {u.status === 1 ? '禁用' : '启用'}
                </button>
                <button onClick={() => setResetTarget(u)} className="px-3 py-1.5 rounded-lg text-xs text-zinc-600 hover:bg-white">重置密码</button>
                <button onClick={() => handleDelete(u)} className="px-3 py-1.5 rounded-lg text-xs text-rose-500 hover:bg-rose-50 flex items-center gap-1">
                  <Trash2 className="w-3.5 h-3.5" /> 删除
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {(resetTarget || batchReset) && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => { setResetTarget(null); setBatchReset(false); }}>
          <div className="glass-card rounded-3xl p-6 max-w-md w-full" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-serif text-primary mb-3">重置密码</h3>
            {batchReset ? (
              <p className="text-sm text-zinc-600 mb-4">为选中的 <b>{selected.length}</b> 个用户设置新密码：</p>
            ) : (
              <p className="text-sm text-zinc-600 mb-4">为用户 <b>{resetTarget?.username}</b> 设置新密码：</p>
            )}
            <input
              type="text"
              placeholder="新密码（至少 6 位，含字母和数字）"
              value={newPwd}
              onChange={(e) => setNewPwd(e.target.value)}
              className="app-input w-full rounded-xl py-2.5 px-4 mb-4"
            />
            <div className="flex gap-2">
              <button onClick={() => { setResetTarget(null); setBatchReset(false); }} className="btn-secondary flex-1">取消</button>
              <button onClick={handleResetPwd} className="btn-primary flex-1">确认重置</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
