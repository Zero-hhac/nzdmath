import React, { useEffect, useState } from 'react';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  BarChart, Bar, Cell,
} from 'recharts';
import {
  Bell, Settings, Eye, User, Users, BookOpen, Activity, Search, ShieldCheck,
  Settings2,   Plus, Edit3, Trash2, CheckCircle2, XCircle, ChevronRight, Save, Clock,
  LogIn, LogOut, Calendar, FileText, Newspaper, RefreshCw, MessageSquare,
  X, Upload,
} from 'lucide-react';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';

type SubView = 'dashboard' | 'content' | 'users' | 'comments';
type ContentType = 'events' | 'resources' | 'news' | 'showcases';

export const AdminView: React.FC<ViewProps> = () => {
  const { admin, logoutAdmin } = useAuth();
  const { showToast } = useToast();
  const [loginOpen, setLoginOpen] = useState(false);
  const [subView, setSubView] = useState<SubView>('dashboard');
  const [contentType, setContentType] = useState<ContentType>('events');

  if (!admin) {
    return (
      <div className="text-center py-20">
        <ShieldCheck className="w-16 h-16 text-zinc-300 mx-auto mb-4" />
        <h2 className="text-2xl font-serif text-primary mb-2">需要管理员登录</h2>
        <p className="text-zinc-500 mb-6">请登录后访问后台管理</p>
        <button onClick={() => setLoginOpen(true)} className="btn-primary flex items-center gap-2 mx-auto">
          <LogIn className="w-4 h-4" /> 管理员登录
        </button>
        <AdminLoginModal open={loginOpen} onClose={() => setLoginOpen(false)} />
      </div>
    );
  }

  return (
    <div className="space-y-12">
      <header className="page-intro flex justify-between items-center rounded-3xl p-8">
        <div className="flex items-center gap-6">
          <div className="w-16 h-16 rounded-full overflow-hidden border-4 border-white shadow-lg bg-primary/10 flex items-center justify-center">
            <ShieldCheck className="w-7 h-7 text-primary" />
          </div>
          <div>
            <h2 className="text-2xl font-serif text-primary">管理员控制台</h2>
            <p className="section-subtitle !text-sm">已登录为 {admin.username}</p>
          </div>
        </div>
        <div className="flex gap-3">
          <button onClick={() => setSubView('dashboard')} className={subBtnCls(subView === 'dashboard')}>
            <Activity className="w-4 h-4" /> 概览
          </button>
          <button onClick={() => setSubView('content')} className={subBtnCls(subView === 'content')}>
            <FileText className="w-4 h-4" /> 内容
          </button>
          <button onClick={() => setSubView('users')} className={subBtnCls(subView === 'users')}>
            <User className="w-4 h-4" /> 用户
          </button>
          <button onClick={() => setSubView('comments')} className={subBtnCls(subView === 'comments')}>
            <MessageSquare className="w-4 h-4" /> 评论
          </button>
          <button
            onClick={async () => { await logoutAdmin(); showToast('已退出后台', 'info'); }}
            className="ml-2 px-4 py-2 rounded-full text-xs font-semibold bg-rose-50 text-rose-500 border border-rose-200 flex items-center gap-1.5"
          >
            <LogOut className="w-3.5 h-3.5" /> 退出
          </button>
        </div>
      </header>

      {subView === 'dashboard' && <DashboardView onSwitch={setSubView} />}
      {subView === 'content' && <ContentManager type={contentType} setType={setContentType} showToast={showToast} />}
      {subView === 'users' && <UserManager showToast={showToast} />}
      {subView === 'comments' && <CommentManager showToast={showToast} />}
    </div>
  );
};

function subBtnCls(active: boolean) {
  return `px-4 py-2 rounded-full text-xs font-semibold flex items-center gap-1.5 transition-colors ${
    active ? 'bg-primary text-white shadow-md' : 'bg-white/60 text-zinc-600 hover:bg-white'
  }`;
}

function AdminLoginModal({ open, onClose }: { open: boolean; onClose: () => void }) {
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

function DashboardView({ onSwitch }: { onSwitch: (s: SubView) => void }) {
  const { showToast } = useToast();
  const [stats, setStats] = useState<any>({});

  useEffect(() => {
    api.adminDashboard()
      .then((res) => setStats(res.data || {}))
      .catch(() => setStats({}));
  }, []);

  const c = stats.counts || {};
  const trend = stats.trend_7days || { dates: [], events: [], news: [] };
  const todayAct = stats.today_activity || { pv: 0, uv: 0, dau: 0 };
  const activity = stats.activity_trend || { dates: [], pv: [], uv: [], dau: [] };
  const totalActivity = stats.total_activity || { pv: 0, uv: 0 };

  const handleInvalidate = async () => {
    try {
      await api.adminInvalidateHomepage();
      showToast('首页缓存已刷新', 'success');
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  return (
    <>
      {/* 基础数据网格 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {[
          { label: '用户总数', value: c.users || 0, icon: User, color: '#e3f2fd' },
          { label: '活动总数', value: c.events || 0, icon: Calendar, color: '#d5e3fc' },
          { label: '资讯总数', value: c.news || 0, icon: Newspaper, color: '#FFF9E5' },
          { label: '资源总数', value: c.resources || 0, icon: BookOpen, color: '#f7f9fb' },
        ].map((s, i) => (
          <div key={i} className="sidebar-panel rounded-[2rem] p-6 space-y-4">
            <div className="flex justify-between items-start">
              <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ backgroundColor: s.color }}>
                <s.icon className="w-5 h-5 text-primary" />
              </div>
            </div>
            <div>
              <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest">{s.label}</div>
              <div className="text-3xl font-serif text-primary mt-1">{s.value}</div>
            </div>
          </div>
        ))}
      </div>

      {/* 流量统计网格 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {[
          { label: '今日浏览量 (PV)', value: todayAct.pv || 0, icon: Eye, color: '#e0f2fe' },
          { label: '今日独立访客 (UV)', value: todayAct.uv || 0, icon: Activity, color: '#dcfce7' },
          { label: '今日活跃会员 (DAU)', value: todayAct.dau || 0, icon: Users, color: '#fef9c3' },
        ].map((s, i) => (
          <div key={i} className="sidebar-panel rounded-[2rem] p-6 space-y-4">
            <div className="flex justify-between items-start">
              <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ backgroundColor: s.color }}>
                <s.icon className="w-5 h-5 text-primary" />
              </div>
            </div>
            <div>
              <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest">{s.label}</div>
              <div className="text-3xl font-serif text-primary mt-1">{s.value}</div>
            </div>
          </div>
        ))}
      </div>

      {/* 全站累计访问数据 */}
      <div className="grid grid-cols-1 lg:grid-cols-[0.9fr_1.1fr] gap-6">
        <div className="grid grid-cols-2 gap-4">
          {[
            { label: '总浏览量 (PV)', value: totalActivity.pv || 0, icon: Eye, color: '#e1f3fe' },
            { label: '总独立访客 (UV)', value: totalActivity.uv || 0, icon: Users, color: '#edf3ec' },
          ].map((s) => (
            <div key={s.label} className="sidebar-panel rounded-[2rem] p-5 space-y-4">
              <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ backgroundColor: s.color }}>
                <s.icon className="w-5 h-5 text-accent" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted font-bold uppercase tracking-widest">{s.label}</div>
                <div className="text-3xl font-medium tracking-tight text-charcoal mt-1">{s.value.toLocaleString('zh-CN')}</div>
              </div>
            </div>
          ))}
        </div>
        <div className="sidebar-panel rounded-[2rem] p-6">
          <div className="flex items-center justify-between mb-3">
            <div>
              <h3 className="text-lg font-medium tracking-tight text-charcoal">访问量总览</h3>
              <p className="text-xs text-text-muted mt-1">从数据统计启用后累计计算</p>
            </div>
            <Eye className="w-5 h-5 text-accent" />
          </div>
          <div className="h-36 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={[
                { name: '总浏览量', value: totalActivity.pv || 0, fill: '#1f2a44' },
                { name: '总独立访客', value: totalActivity.uv || 0, fill: '#6f8f75' },
              ]} layout="vertical" margin={{ top: 0, right: 12, left: 8, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="#edf0f4" />
                <XAxis type="number" hide />
                <YAxis type="category" dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 11, fill: '#787774' }} width={70} />
                <Tooltip formatter={(value: number) => [value.toLocaleString('zh-CN'), '数量']} contentStyle={{ borderRadius: '14px', border: '1px solid #eaeaea', fontSize: '12px' }} />
                <Bar dataKey="value" radius={[0, 8, 8, 0]} barSize={22} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 内容发布趋势 */}
        <div className="sidebar-panel rounded-[2.5rem] p-8 space-y-6">
          <div className="flex justify-between items-center">
            <h3 className="font-serif text-lg text-primary">7 天内容发布趋势</h3>
            <button onClick={handleInvalidate} className="btn-secondary !text-xs !py-1.5 !px-3 flex items-center gap-1.5">
              <RefreshCw className="w-3 h-3" /> 刷新首页缓存
            </button>
          </div>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trend.dates?.map((d: string, i: number) => ({ name: d.slice(5), events: trend.events[i], news: trend.news[i] })) || []}>
                <defs>
                  <linearGradient id="colorEv" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#526069" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#526069" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f1f1" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#999' }} />
                <YAxis hide />
                <Tooltip contentStyle={{ borderRadius: '16px', border: 'none', boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)', fontSize: '12px' }} />
                <Area type="monotone" dataKey="events" stroke="#526069" fillOpacity={1} fill="url(#colorEv)" name="活动" />
                <Area type="monotone" dataKey="news" stroke="#ba1a1a" strokeDasharray="5 5" fill="transparent" name="资讯" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* 流量与活跃度趋势 */}
        <div className="sidebar-panel rounded-[2.5rem] p-8 space-y-6">
          <div className="flex justify-between items-center">
            <h3 className="font-serif text-lg text-primary">7 天访客与活跃度趋势</h3>
          </div>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={activity.dates?.map((d: string, i: number) => ({ name: d.slice(5), pv: activity.pv[i], uv: activity.uv[i], dau: activity.dau[i] })) || []}>
                <defs>
                  <linearGradient id="colorPv" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#2563eb" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#2563eb" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="colorUv" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#16a34a" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#16a34a" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="colorDau" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#ca8a04" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#ca8a04" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f1f1" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#999' }} />
                <YAxis hide />
                <Tooltip contentStyle={{ borderRadius: '16px', border: 'none', boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)', fontSize: '12px' }} />
                <Area type="monotone" dataKey="pv" stroke="#2563eb" fillOpacity={1} fill="url(#colorPv)" name="PV (浏览量)" />
                <Area type="monotone" dataKey="uv" stroke="#16a34a" fillOpacity={1} fill="url(#colorUv)" name="UV (独立访客)" />
                <Area type="monotone" dataKey="dau" stroke="#ca8a04" fillOpacity={1} fill="url(#colorDau)" name="DAU (活跃会员)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </>
  );
}

function ContentManager({ type, setType, showToast }: { type: ContentType; setType: (t: ContentType) => void; showToast: (m: string, t?: any) => void }) {
  const [items, setItems] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [editing, setEditing] = useState<any>(null);
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

  return (
    <div className="space-y-6">
      <div className="flex gap-2">
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
                  <div className="font-medium text-zinc-800 truncate">{it.title || it.name || `#${it.id}`}</div>
                  <div className="text-xs text-zinc-500 mt-0.5">
                    ID: {it.id} · {it.status === 1 ? '已发布' : '草稿'} · {it.created_at ? new Date(it.created_at).toLocaleDateString('zh-CN') : ''}
                  </div>
                </div>
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

      {editing && <ContentEditor type={type} item={editing} onClose={() => { setEditing(null); load(); }} showToast={showToast} />}
    </div>
  );
}

function ContentEditor({ type, item, onClose, showToast }: { type: ContentType; item: any; onClose: () => void; showToast: (m: string, t?: any) => void }) {
  const isNew = !item.id;
  const [form, setForm] = useState<any>(() => {
    if (isNew) {
      if (type === 'events') return { title: '', summary: '', content: '', category: '讲座', location: '', start_time: '', end_time: '', cover_url: '', status: 1, is_featured: false };
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

function UserManager({ showToast }: { showToast: (m: string, t?: any) => void }) {
  const [items, setItems] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [keyword, setKeyword] = useState('');
  const [resetTarget, setResetTarget] = useState<any>(null);
  const [newPwd, setNewPwd] = useState('');

  const load = () => {
    setLoading(true);
    const params: Record<string, string> = {};
    if (keyword) params.keyword = keyword;
    api.adminListUsers(params).then((res) => setItems(res.data || [])).catch(() => setItems([])).finally(() => setLoading(false));
  };

  useEffect(() => { load(); }, []);

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
    if (!newPwd || newPwd.length < 6) { showToast('密码至少 6 位', 'error'); return; }
    try {
      await api.adminResetUserPassword(resetTarget.id, newPwd);
      showToast(`已重置 ${resetTarget.username} 的密码`, 'success');
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
              placeholder="搜索用户名/邮箱/昵称..."
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && load()}
              className="app-input w-full rounded-full py-2 pl-9 pr-4"
            />
            <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-400" />
          </div>
          <button onClick={load} className="btn-secondary !py-2 !text-xs">查询</button>
        </div>

        {loading ? (
          <div className="text-center py-12 text-zinc-500">加载中...</div>
        ) : items.length === 0 ? (
          <div className="text-center py-12 text-zinc-500">暂无用户</div>
        ) : (
          <div className="space-y-2">
            {items.map((u) => (
              <div key={u.id} className="flex items-center gap-4 p-3 rounded-xl hover:bg-white/60">
                <div className="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary text-sm font-bold">
                  {u.username?.[0]?.toUpperCase()}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-zinc-800 truncate">{u.username} <span className="text-xs text-zinc-400">#{u.id}</span></div>
                  <div className="text-xs text-zinc-500 mt-0.5">
                    {u.email || '—'} · {u.role || 'member'} · {u.status === 1 ? '✓ 正常' : '✕ 禁用'}
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

      {resetTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setResetTarget(null)}>
          <div className="glass-card rounded-3xl p-6 max-w-md w-full" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-serif text-primary mb-3">重置密码</h3>
            <p className="text-sm text-zinc-600 mb-4">为用户 <b>{resetTarget.username}</b> 设置新密码：</p>
            <input
              type="text"
              placeholder="新密码（至少 6 位）"
              value={newPwd}
              onChange={(e) => setNewPwd(e.target.value)}
              className="app-input w-full rounded-xl py-2.5 px-4 mb-4"
            />
            <div className="flex gap-2">
              <button onClick={() => setResetTarget(null)} className="btn-secondary flex-1">取消</button>
              <button onClick={handleResetPwd} className="btn-primary flex-1">确认重置</button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function CommentManager({ showToast }: { showToast: (m: string, t?: any) => void }) {
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
