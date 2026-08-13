import { useEffect, useState } from 'react';
import { Bell, Send } from 'lucide-react';
import { api, DEPARTMENTS } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';

/**
 * 通知管理：单发（按用户名）/ 按部门 / 全部会员，附发送记录。
 */
export function NotificationsPanel() {
  const { showToast } = useToast();
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [ntype, setNtype] = useState('system');
  const [mode, setMode] = useState<'all' | 'department' | 'users'>('all');
  const [department, setDepartment] = useState('');
  const [usernames, setUsernames] = useState('');
  const [sending, setSending] = useState(false);
  const [batches, setBatches] = useState<any[]>([]);
  const [batchesTotal, setBatchesTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);

  const loadBatches = (p: number) => {
    setLoading(true);
    api.adminListNotificationBatches(p, 20)
      .then((res) => {
        setBatches((prev) => (p === 1 ? res.data || [] : [...prev, ...(res.data || [])]));
        setBatchesTotal(res.total || 0);
        setPage(p);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    loadBatches(1);
  }, []);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      showToast('请填写通知标题', 'error');
      return;
    }
    if (!content.trim()) {
      showToast('请填写通知内容', 'error');
      return;
    }
    if (mode === 'department' && !department) {
      showToast('请选择部门', 'error');
      return;
    }
    if (mode === 'users' && !usernames.trim()) {
      showToast('请填写用户名（每行一个）', 'error');
      return;
    }
    setSending(true);
    try {
      const res = await api.adminSendNotification({
        title: title.trim(),
        content: content.trim(),
        type: ntype,
        target:
          mode === 'users'
            ? { mode, usernames: usernames.split(/[\n,，、\s]+/).filter(Boolean) }
            : mode === 'department'
              ? { mode, department }
              : { mode },
      });
      showToast(`已发送给 ${res.data?.count ?? 0} 位会员`, 'success');
      setTitle('');
      setContent('');
      setUsernames('');
      loadBatches(1);
    } catch (err: any) {
      showToast(err.message || '发送失败', 'error');
    } finally {
      setSending(false);
    }
  };

  return (
    <div className="space-y-6">
      <form onSubmit={submit} className="sidebar-panel rounded-[2rem] p-6 space-y-4">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center">
            <Bell className="w-5 h-5 text-primary" />
          </div>
          <div>
            <h3 className="text-lg font-semibold text-charcoal">发送通知</h3>
            <p className="text-xs text-zinc-500">会员将在个人中心「消息通知」中收到</p>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-1.5 block">标题</label>
            <input
              value={title}
              onChange={(e) => setTitle(e.target.value)}
              className="app-input w-full rounded-xl py-2.5 px-3"
              placeholder="如：数学文化节中奖通知"
              maxLength={150}
            />
          </div>
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-1.5 block">类型</label>
            <select value={ntype} onChange={(e) => setNtype(e.target.value)} className="app-input w-full rounded-xl py-2.5 px-3">
              <option value="system">系统通知</option>
              <option value="activity">活动通知</option>
              <option value="reward">获奖通知</option>
            </select>
          </div>
        </div>

        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-1.5 block">内容</label>
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            rows={4}
            className="app-input w-full rounded-xl py-2.5 px-3"
            placeholder="通知正文，支持换行"
            maxLength={5000}
          />
        </div>

        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-1.5 block">发送目标</label>
          <div className="flex gap-2 mb-3">
            {[
              { v: 'all' as const, label: '全部会员' },
              { v: 'department' as const, label: '按部门' },
              { v: 'users' as const, label: '指定用户' },
            ].map((t) => (
              <button
                key={t.v}
                type="button"
                onClick={() => setMode(t.v)}
                className={`px-4 py-2 rounded-full text-xs font-semibold ${
                  mode === t.v ? 'bg-primary text-white shadow-md' : 'bg-white/60 text-zinc-600 hover:bg-white'
                }`}
              >
                {t.label}
              </button>
            ))}
          </div>
          {mode === 'department' && (
            <select value={department} onChange={(e) => setDepartment(e.target.value)} className="app-input rounded-xl py-2.5 px-3 w-full md:w-64">
              <option value="">请选择部门</option>
              {DEPARTMENTS.map((d) => (
                <option key={d} value={d}>{d}</option>
              ))}
            </select>
          )}
          {mode === 'users' && (
            <textarea
              value={usernames}
              onChange={(e) => setUsernames(e.target.value)}
              rows={3}
              className="app-input w-full rounded-xl py-2.5 px-3"
              placeholder={'每行一个用户名，例如：\nzhangsan\nlisi'}
            />
          )}
        </div>

        <button type="submit" disabled={sending} className="btn-primary flex items-center gap-2">
          <Send className="w-4 h-4" /> {sending ? '发送中...' : '发送通知'}
        </button>
      </form>

      <div className="sidebar-panel rounded-[2rem] p-6">
        <h3 className="text-lg font-semibold text-charcoal mb-4">发送记录</h3>
        {loading && batches.length === 0 ? (
          <div className="text-center py-8 text-zinc-500">加载中...</div>
        ) : batches.length === 0 ? (
          <div className="text-center py-8 text-zinc-500">暂无发送记录</div>
        ) : (
          <div className="space-y-2">
            {batches.map((b) => (
              <div key={b.id} className="flex items-center gap-4 p-3 rounded-xl hover:bg-white/60 transition-colors">
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-zinc-800 truncate">{b.title}</div>
                  <div className="text-xs text-zinc-500 mt-0.5 truncate">{b.content}</div>
                </div>
                <div className="shrink-0 text-right">
                  <div className="text-xs font-bold text-zinc-600">{b.target} · {b.count} 人</div>
                  <div className="text-xs text-zinc-400 mt-0.5">{new Date(b.created_at).toLocaleString('zh-CN')}</div>
                </div>
              </div>
            ))}
          </div>
        )}
        {batches.length < batchesTotal && (
          <div className="flex justify-center pt-4">
            <button onClick={() => loadBatches(page + 1)} className="btn-secondary !px-6 !py-2 !text-xs">
              加载更多（{batches.length}/{batchesTotal}）
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
