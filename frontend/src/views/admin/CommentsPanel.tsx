import { useEffect, useState } from 'react';
import { api } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';

export function CommentsPanel() {
  const { showToast } = useToast();
  const [items, setItems] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);

  const load = () => {
    setLoading(true);
    api.adminListComments().then((res) => setItems(res.data || [])).catch(() => setItems([])).finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除该评论吗？')) return;
    try {
      await api.adminDeleteComment(id);
      showToast('已删除', 'success');
      load();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  const handleToggleStatus = async (c: any) => {
    const newStatus = c.status === 1 ? 0 : 1;
    try {
      await api.adminSetCommentStatus(c.id, newStatus as 0 | 1);
      showToast(newStatus === 1 ? '已显示' : '已隐藏', 'success');
      load();
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="sidebar-panel rounded-[2rem] p-6">
        {loading ? (
          <div className="text-center py-12 text-zinc-500">加载中...</div>
        ) : items.length === 0 ? (
          <div className="text-center py-12 text-zinc-500">暂无评论</div>
        ) : (
          <div className="space-y-3">
            {items.map((c) => (
              <div key={c.id} className="p-4 rounded-2xl bg-white/60">
                <div className="flex items-center gap-3 mb-2">
                  <div className="w-8 h-8 rounded-full bg-primary/10 flex items-center justify-center text-primary text-xs font-bold">
                    {c.user_name?.[0] || c.user_id}
                  </div>
                  <div className="text-sm font-bold">{c.user_name || `用户${c.user_id}`}</div>
                  <div className="text-xs text-zinc-500 ml-auto">{c.created_at ? new Date(c.created_at).toLocaleString('zh-CN') : ''}</div>
                </div>
                <div className="text-sm text-zinc-700 ml-11 mb-2">{c.content}</div>
                <div className="ml-11 text-xs text-zinc-500 flex items-center gap-3">
                  <span>{c.target_type} #{c.target_id}</span>
                  <span>{c.status === 1 ? '✓ 显示' : '✕ 隐藏'}</span>
                  <span>♥ {c.like_count || 0}</span>
                </div>
                <div className="ml-11 mt-2 flex gap-2">
                  <button onClick={() => handleToggleStatus(c)} className="px-3 py-1 rounded-lg text-xs text-zinc-600 hover:bg-white">
                    {c.status === 1 ? '隐藏' : '显示'}
                  </button>
                  <button onClick={() => handleDelete(c.id)} className="px-3 py-1 rounded-lg text-xs text-rose-500 hover:bg-rose-50">
                    删除
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}
