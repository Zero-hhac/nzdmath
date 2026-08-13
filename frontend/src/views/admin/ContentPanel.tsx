import React, { useEffect, useState } from 'react';
import { Search, Plus, Edit3, Trash2, X, Upload, AlertTriangle } from 'lucide-react';
import { api } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';
import type { ContentType } from './types';

export function ContentPanel({ type, setType }: { type: ContentType; setType: (t: ContentType) => void }) {
  const { showToast } = useToast();
  const [items, setItems] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState<any>(null);
  const [expireTarget, setExpireTarget] = useState<any>(null);
  const [expiring, setExpiring] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [statusFilter, setStatusFilter] = useState<string>('');

  const load = () => {
    setLoading(true);
    const params: Record<string, string> = {};
    if (keyword) params.keyword = keyword;
    if (statusFilter) params.status = statusFilter;
    const apiMap = {
      events: api.adminListEvents,
      news: api.adminListNews,
      resources: api.adminListResources,
      showcases: api.adminListShowcases,
    };
    apiMap[type](params)
      .then((res) => setItems(res.data || []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, [type]);

  const handleDelete = async (id: number) => {
    if (!confirm('确定要删除吗？')) return;
    const apiMap = {
      events: api.adminDeleteEvent,
      news: api.adminDeleteNews,
      resources: api.adminDeleteResource,
      showcases: api.adminDeleteShowcase,
    } as const;
    try {
      await apiMap[type](id);
      showToast('删除成功', 'success');
      load();
    } catch (err: any) {
      showToast(err.message || '删除失败', 'error');
    }
  };

  // 活动过期切换：设为过期需二次确认；恢复直接生效
  const handleToggleExpired = async (it: any) => {
    if (it.is_expired) {
      // 恢复为进行中
      try {
        await api.adminSetEventExpired(it.id, false);
        showToast('已恢复为进行中活动', 'success');
        load();
      } catch (err: any) {
        showToast(err.message || '操作失败', 'error');
      }
      return;
    }
    setExpireTarget(it);
  };

  const doConfirmExpire = async () => {
    if (!expireTarget) return;
    setExpiring(true);
    try {
      await api.adminSetEventExpired(expireTarget.id, true);
      showToast('已标记为过期', 'success');
      setExpireTarget(null);
      load();
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    } finally {
      setExpiring(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap gap-2">
        {[
          { v: 'events' as const, label: '活动' },
          { v: 'news' as const, label: '资讯' },
          { v: 'resources' as const, label: '资源' },
          { v: 'showcases' as const, label: '作品' },
        ].map((t) => (
          <button
            key={t.v}
            onClick={() => setType(t.v)}
            className={`px-5 py-2 rounded-full text-xs font-semibold ${
              type === t.v ? 'bg-primary text-white shadow-md' : 'bg-white/60 text-zinc-600 hover:bg-white'
            }`}
          >
            {t.label}
          </button>
        ))}
      </div>

      <div className="sidebar-panel rounded-[2rem] p-6">
        <div className="flex gap-3 items-center mb-4">
          <div className="relative flex-1">
            <input
              type="text"
              placeholder="搜索标题..."
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && load()}
              className="app-input w-full rounded-full py-2 pl-9 pr-4"
            />
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
          </div>
          <select value={statusFilter} onChange={(e) => { setStatusFilter(e.target.value); setTimeout(load, 0); }} className="app-input rounded-xl px-3 py-2 text-sm">
            <option value="">全部状态</option>
            <option value="1">已发布</option>
            <option value="0">草稿</option>
          </select>
          <button onClick={load} className="btn-secondary !py-2 !text-xs">查询</button>
          <button onClick={() => setEditing({})} className="btn-primary !py-2 !text-xs flex items-center gap-1.5">
            <Plus className="w-3.5 h-3.5" /> 新建
          </button>
        </div>

        {loading ? (
          <div className="text-center py-12 text-zinc-500">加载中...</div>
        ) : items.length === 0 ? (
          <div className="text-center py-12 text-zinc-500">暂无数据</div>
        ) : (
          <div className="space-y-2">
            {items.map((it) => (
              <div key={it.id} className="flex items-center gap-4 p-3 rounded-xl hover:bg-white/60 transition-colors">
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-zinc-800 truncate">
                    {it.title || it.name || `#${it.id}`}
                    {type === 'events' && it.is_expired && (
                      <span className="ml-2 px-2 py-0.5 rounded-full text-[10px] font-bold bg-zinc-100 text-zinc-500 border border-zinc-200">
                        已过期
                      </span>
                    )}
                  </div>
                  <div className="text-xs text-zinc-500 mt-0.5">
                    ID: {it.id} · {it.status === 1 ? '已发布' : '草稿'} · {it.created_at ? new Date(it.created_at).toLocaleDateString('zh-CN') : ''}
                  </div>
                </div>
                {type === 'events' && (
                  <button
                    onClick={() => handleToggleExpired(it)}
                    className={`px-3 py-1.5 rounded-lg text-xs flex items-center gap-1 ${
                      it.is_expired ? 'text-emerald-600 hover:bg-emerald-50' : 'text-zinc-500 hover:bg-white'
                    }`}
                  >
                    {it.is_expired ? '恢复' : '设为过期'}
                  </button>
                )}
                <button onClick={() => setEditing(it)} className="px-3 py-1.5 rounded-lg text-xs text-zinc-600 hover:bg-white flex items-center gap-1">
                  <Edit3 className="w-3.5 h-3.5" /> 编辑
                </button>
                <button onClick={() => handleDelete(it.id)} className="px-3 py-1.5 rounded-lg text-xs text-rose-500 hover:bg-rose-50 flex items-center gap-1">
                  <Trash2 className="w-3.5 h-3.5" /> 删除
                </button>
              </div>
            ))}
          </div>
        )}
      </div>

      {editing && <ContentEditor type={type} item={editing} onClose={() => { setEditing(null); load(); }} />}

      {expireTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4" onClick={() => !expiring && setExpireTarget(null)}>
          <div onClick={(e) => e.stopPropagation()} className="glass-card rounded-3xl p-6 max-w-md w-full">
            <div className="flex items-start gap-3 mb-4">
              <div className="w-10 h-10 rounded-xl bg-amber-50 flex items-center justify-center shrink-0">
                <AlertTriangle className="w-5 h-5 text-amber-500" />
              </div>
              <div className="flex-1 min-w-0">
                <h3 className="text-lg font-serif text-primary">确认设为过期？</h3>
                <p className="text-sm text-zinc-500 mt-1 leading-relaxed">
                  将活动「{expireTarget.title}」标记为过期后：
                </p>
              </div>
              <button onClick={() => setExpireTarget(null)} className="text-zinc-400 hover:text-zinc-600">
                <X className="w-5 h-5" />
              </button>
            </div>
            <ul className="text-sm text-zinc-600 space-y-1.5 bg-canvas-alt rounded-2xl p-4 mb-5">
              <li>· 前台活动卡片将显示「已过期」标记，会员无法再报名</li>
              <li>· 活动从首页推荐中移除</li>
              <li>· 已有报名与签到记录保留，可随时「恢复」</li>
            </ul>
            <div className="flex gap-2">
              <button
                onClick={() => setExpireTarget(null)}
                disabled={expiring}
                className="btn-secondary flex-1 !py-2.5 !text-xs"
              >
                取消
              </button>
              <button
                onClick={doConfirmExpire}
                disabled={expiring}
                className="flex-1 px-4 py-2.5 rounded-full text-xs font-semibold bg-amber-500 text-white hover:bg-amber-600 transition-colors disabled:opacity-50"
              >
                {expiring ? '处理中...' : '确认设为过期'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function ContentEditor({ type, item, onClose }: { type: ContentType; item: any; onClose: () => void }) {
  const { showToast } = useToast();
  const isNew = !item.id;
  const [form, setForm] = useState<any>(() => {
    if (isNew) {
      if (type === 'events') return { title: '', summary: '', content: '', category: '讲座', location: '', start_time: '', end_time: '', cover_url: '', capacity: 0, status: 1, is_featured: false };
      if (type === 'news') return { title: '', summary: '', content: '', category: '通知', tag: '', cover_url: '', status: 1, is_featured: false };
      if (type === 'resources') return { title: '', summary: '', category: '课程笔记', cover_url: '', status: 1, is_featured: false };
      if (type === 'showcases') return { title: '', author: '', field: '几何学', competition: '', summary: '', cover_url: '', h5_url: '', status: 1 };
    }
    return item;
  });
  const [saving, setSaving] = useState(false);
  const [uploadFile, setUploadFile] = useState<File | null>(null);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSaving(true);
    try {
      // 资源 + 新建：走文件上传接口
      if (type === 'resources' && isNew) {
        if (!uploadFile) {
          showToast('请选择要上传的文件', 'error');
          setSaving(false);
          return;
        }
        const fd = new FormData();
        fd.append('file', uploadFile);
        fd.append('title', form.title || uploadFile.name);
        fd.append('summary', form.summary || '');
        fd.append('category', form.category || '');
        await api.adminCreateResource(fd);
        showToast('上传成功', 'success');
        onClose();
        return;
      }

      const apiMap = {
        events: { create: api.adminCreateEvent, update: api.adminUpdateEvent },
        news: { create: api.adminCreateNews, update: api.adminUpdateNews },
        resources: { create: null as any, update: api.adminUpdateResource },
        showcases: { create: api.adminCreateShowcase, update: api.adminUpdateShowcase },
      } as const;
      if (isNew) {
        await apiMap[type].create(form);
      } else {
        await apiMap[type].update(item.id, form);
      }
      showToast(isNew ? '创建成功' : '更新成功', 'success');
      onClose();
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm p-4" onClick={onClose}>
      <form onSubmit={submit} onClick={(e) => e.stopPropagation()} className="glass-card rounded-3xl p-6 max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        <div className="flex justify-between items-center mb-4">
          <h3 className="text-xl font-serif text-primary">{isNew ? '新建' : '编辑'} {type === 'events' ? '活动' : type === 'news' ? '资讯' : type === 'resources' ? '资源' : '作品'}</h3>
          <button type="button" onClick={onClose} className="text-zinc-500"><X className="w-5 h-5" /></button>
        </div>
        <div className="space-y-3">
          {/* 资源 + 新建：文件上传字段 */}
          {type === 'resources' && isNew && (
            <Field label="资源文件 *">
              <div className="flex items-center gap-3">
                <label className="btn-secondary !py-2 !text-xs cursor-pointer flex items-center gap-2">
                  <Upload className="w-3.5 h-3.5" />
                  选择文件
                  <input
                    type="file"
                    className="hidden"
                    onChange={(e) => setUploadFile(e.target.files?.[0] || null)}
                  />
                </label>
                <span className="text-xs text-zinc-500 truncate flex-1">
                  {uploadFile ? `${uploadFile.name} (${(uploadFile.size / 1024).toFixed(1)} KB)` : '尚未选择'}
                </span>
              </div>
            </Field>
          )}

          <Field label="标题"><input value={form.title || ''} onChange={(e) => setForm({ ...form, title: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" required /></Field>
          {type === 'events' && <>
            <Field label="分类"><input value={form.category || ''} onChange={(e) => setForm({ ...form, category: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="地点"><input value={form.location || ''} onChange={(e) => setForm({ ...form, location: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="开始时间"><input type="datetime-local" value={form.start_time ? form.start_time.slice(0, 16) : ''} onChange={(e) => setForm({ ...form, start_time: e.target.value + ':00+08:00' })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="结束时间"><input type="datetime-local" value={form.end_time ? form.end_time.slice(0, 16) : ''} onChange={(e) => setForm({ ...form, end_time: e.target.value + ':00+08:00' })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="报名名额（0 = 不限）">
              <input type="number" min={0} max={100000} value={form.capacity ?? 0} onChange={(e) => setForm({ ...form, capacity: parseInt(e.target.value) || 0 })} className="app-input w-full rounded-xl py-2 px-3" />
            </Field>
          </>}
          {type === 'news' && <>
            <Field label="分类"><input value={form.category || ''} onChange={(e) => setForm({ ...form, category: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="标签"><input value={form.tag || ''} onChange={(e) => setForm({ ...form, tag: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
          </>}
          {type === 'resources' && <>
            <Field label="分类"><input value={form.category || ''} onChange={(e) => setForm({ ...form, category: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" placeholder="如：课程笔记 / 竞赛讲义" /></Field>
          </>}
          {type === 'showcases' && <>
            <Field label="作者"><input value={form.author || ''} onChange={(e) => setForm({ ...form, author: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="领域"><input value={form.field || ''} onChange={(e) => setForm({ ...form, field: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="赛事"><input value={form.competition || ''} onChange={(e) => setForm({ ...form, competition: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
            <Field label="H5 演示 URL"><input value={form.h5_url || ''} onChange={(e) => setForm({ ...form, h5_url: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" placeholder="例如：/uploads/h5_unified_light/index.html" /></Field>
          </>}
          <Field label="摘要"><textarea value={form.summary || ''} onChange={(e) => setForm({ ...form, summary: e.target.value })} rows={3} className="app-input w-full rounded-xl py-2 px-3" /></Field>
          <Field label="封面图 URL"><input value={form.cover_url || ''} onChange={(e) => setForm({ ...form, cover_url: e.target.value })} className="app-input w-full rounded-xl py-2 px-3" /></Field>
          {/* 资源新建时无需 content 字段（无 API 支持） */}
          {((type === 'events' || type === 'news') || (type === 'resources' && !isNew)) && (
            <Field label="内容（HTML/纯文本）">
              <textarea value={form.content || ''} onChange={(e) => setForm({ ...form, content: e.target.value })} rows={6} className="app-input w-full rounded-xl py-2 px-3 font-mono text-xs" />
            </Field>
          )}
          <div className="grid grid-cols-2 gap-3">
            <Field label="状态">
              <select value={form.status || 1} onChange={(e) => setForm({ ...form, status: parseInt(e.target.value) })} className="app-input w-full rounded-xl py-2 px-3">
                <option value={1}>已发布</option>
                <option value={0}>草稿</option>
              </select>
            </Field>
            {(type === 'events' || type === 'news' || type === 'resources') && (
              <Field label="是否推荐">
                <select value={form.is_featured ? '1' : '0'} onChange={(e) => setForm({ ...form, is_featured: e.target.value === '1' })} className="app-input w-full rounded-xl py-2 px-3">
                  <option value="0">否</option>
                  <option value="1">是</option>
                </select>
              </Field>
            )}
          </div>
        </div>
        <div className="flex gap-2 mt-6">
          <button type="button" onClick={onClose} className="btn-secondary flex-1">取消</button>
          <button type="submit" disabled={saving} className="btn-primary flex-1">{saving ? '保存中...' : '保存'}</button>
        </div>
      </form>
    </div>
  );
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-1.5 block">{label}</label>
      {children}
    </div>
  );
}
