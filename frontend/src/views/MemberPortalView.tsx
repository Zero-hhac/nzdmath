import React, { useEffect, useRef, useState } from 'react';
import {
  User, BookOpen, Heart, LogIn, LogOut, Download, Save, X, Edit2, Camera,
  Lock, Trash2, Clock, CalendarDays,
} from 'lucide-react';
import type { ViewProps } from '@/src/types/app';
import { api, DEPARTMENTS } from '@/src/lib/api';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { LoginModal } from '@/src/components/LoginModal';

type SubView = 'dashboard' | 'profile' | 'favorites' | 'downloads' | 'password' | 'comments' | 'registrations';

type Favorite = {
  id: number;
  user_id: number;
  target_type: string;
  target_id: number;
  title?: string;
  summary?: string;
  cover_url?: string;
  created_at?: string;
};

type Download = {
  id: number;
  user_id: number;
  resource_id: number;
  ip?: string;
  created_at: string;
};

export const MemberPortalView: React.FC<ViewProps> = ({ navigate }) => {
  const { user, loading: authLoading, refreshUser, logoutUser } = useAuth();
  const { showToast } = useToast();
  const [loginOpen, setLoginOpen] = useState(false);
  const [subView, setSubView] = useState<SubView>('dashboard');

  if (authLoading) {
    return <div className="text-center text-zinc-500 py-20">加载中...</div>;
  }

  if (!user) {
    return (
      <>
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-10">
          <div className="lg:col-span-1 space-y-8">
            <div className="sidebar-panel rounded-[2.5rem] p-10 text-center space-y-6 relative overflow-hidden">
              <div className="absolute top-0 left-0 w-full h-24 bg-gradient-to-r from-accent-soft/70 to-transparent opacity-60"></div>
              <div className="relative pt-6">
                <div className="w-24 h-24 mx-auto rounded-[2rem] border-4 border-white shadow-xl overflow-hidden glass-card p-0 flex items-center justify-center bg-primary/10">
                  <User className="w-10 h-10 text-primary" />
                </div>
                <div className="mt-4">
                  <h3 className="text-2xl font-serif text-primary italic">请先登录</h3>
                  <div className="math-tag mt-2">访客模式</div>
                </div>
              </div>
              <p className="text-sm text-soft-body">登录后可查看和编辑个人资料。</p>
              <button onClick={() => setLoginOpen(true)} className="btn-primary w-full !py-3 !text-xs !font-bold flex items-center justify-center gap-2">
                <LogIn className="w-4 h-4" /> 立即登录
              </button>
            </div>
          </div>
          <div className="lg:col-span-2 space-y-10">
            <div className="page-intro space-y-2">
              <div className="section-kicker">Member Portal</div>
              <h2 className="section-title">会员专属专区</h2>
              <p className="section-subtitle">登录后享受完整会员服务</p>
            </div>
          </div>
        </div>
        <LoginModal open={loginOpen} onClose={() => setLoginOpen(false)} />
      </>
    );
  }

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-10">
      <div className="lg:col-span-1 space-y-8">
        <ProfileCard user={user} onSubView={setSubView} onLogout={logoutUser} />
      </div>
      <div className="lg:col-span-2 space-y-10">
        {subView === 'dashboard' && <Dashboard user={user} onSubView={setSubView} showToast={showToast} />}
        {subView === 'profile' && <ProfileEditor user={user} onSaved={refreshUser} showToast={showToast} />}
        {subView === 'favorites' && <FavoritesView showToast={showToast} />}
        {subView === 'downloads' && <DownloadsView showToast={showToast} />}
        {subView === 'password' && <PasswordView showToast={showToast} />}
        {subView === 'registrations' && <RegistrationsView showToast={showToast} navigate={navigate} />}
      </div>
    </div>
  );
};

function ProfileCard({ user, onSubView, onLogout }: { user: any; onSubView: (s: SubView) => void; onLogout: () => Promise<void> }) {
  return (
    <div className="sidebar-panel rounded-[2.5rem] p-10 text-center space-y-6 relative overflow-hidden">
              <div className="absolute top-0 left-0 w-full h-24 bg-gradient-to-r from-accent-soft/70 to-transparent opacity-60"></div>
      <div className="relative pt-6">
        {user.avatar ? (
          <img src={user.avatar} alt="" className="w-24 h-24 mx-auto rounded-[2rem] border-4 border-white shadow-xl object-cover" />
        ) : (
          <div className="w-24 h-24 mx-auto rounded-[2rem] border-4 border-white shadow-xl glass-card p-0 flex items-center justify-center bg-primary/10">
            <User className="w-10 h-10 text-primary" />
          </div>
        )}
        <div className="mt-4">
          <h3 className="text-2xl font-serif text-primary italic">{user.nickname || user.username}</h3>
          <div className="math-tag mt-2">{user.role === 'admin' ? '管理员' : '正式会员'}</div>
        </div>
      </div>

      <div className="space-y-3">
        {[
          { label: '个人资料', icon: User, view: 'profile' as SubView },
          { label: '我的收藏', icon: Heart, view: 'favorites' as SubView },
          { label: '下载历史', icon: Download, view: 'downloads' as SubView },
          { label: '我的活动', icon: CalendarDays, view: 'registrations' as SubView },
          { label: '修改密码', icon: Lock, view: 'password' as SubView },
        ].map((item) => (
          <button
            key={item.view}
            onClick={() => onSubView(item.view)}
            className="btn-secondary w-full !py-2.5 !text-xs !font-bold flex items-center justify-center gap-2"
          >
            <item.icon className="w-4 h-4" />
            {item.label}
          </button>
        ))}
        <button
          onClick={onLogout}
          className="w-full !py-2.5 !text-xs !font-bold flex items-center justify-center gap-2 text-rose-500 hover:bg-rose-50 rounded-full border border-rose-200 transition-colors"
        >
          <LogOut className="w-4 h-4" /> 退出登录
        </button>
      </div>
    </div>
  );
}

function Dashboard({ user, onSubView, showToast }: { user: any; onSubView: (s: SubView) => void; showToast: (m: string, t?: any) => void }) {
  return (
    <>
      <div className="page-intro space-y-2">
        <div className="section-kicker">Member Portal</div>
        <h2 className="section-title">会员专属专区</h2>
        <p className="section-subtitle">欢迎回来，{user.nickname || user.username}！</p>
      </div>

      {(!user.real_name || !user.class_name || !user.department) && (
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 rounded-2xl border border-amber-200 bg-amber-50 px-5 py-4">
          <div className="flex-1 text-sm text-amber-800">
            <span className="font-bold">资料待完善：</span>请补充姓名、班级与部门，方便协会组织活动和内部联系。
          </div>
          <button onClick={() => onSubView('profile')} className="btn-primary !py-2 !text-xs shrink-0">
            <Edit2 className="w-3.5 h-3.5" /> 去完善
          </button>
        </div>
      )}

      <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
        {[
          { label: '个人资料', icon: User, view: 'profile' as SubView, color: 'from-accent-soft/70 to-surface' },
          { label: '我的收藏', icon: Heart, view: 'favorites' as SubView, color: 'from-canvas-alt to-surface' },
          { label: '下载历史', icon: Download, view: 'downloads' as SubView, color: 'from-canvas-alt to-surface' },
          { label: '我的活动', icon: CalendarDays, view: 'registrations' as SubView, color: 'from-canvas-alt to-surface' },
        ].map((item) => (
          <button
            key={item.view}
            onClick={() => onSubView(item.view)}
            className={`glass-card rounded-2xl p-5 bg-gradient-to-br ${item.color} hover:scale-[1.02] transition-transform`}
          >
            <item.icon className="w-6 h-6 text-primary mb-3" />
            <div className="text-sm font-bold text-zinc-700">{item.label}</div>
          </button>
        ))}
      </div>

      <div className="sidebar-panel rounded-[2rem] p-8 space-y-4">
        <h3 className="font-serif text-xl text-primary">快速操作</h3>
        <div className="grid grid-cols-2 gap-3">
          <button onClick={() => onSubView('profile')} className="surface-subtle p-4 text-left">
            <Edit2 className="w-4 h-4 text-primary mb-2" />
            <div className="text-sm font-medium">编辑资料</div>
          </button>
          <button onClick={() => onSubView('password')} className="surface-subtle p-4 text-left">
            <Lock className="w-4 h-4 text-primary mb-2" />
            <div className="text-sm font-medium">修改密码</div>
          </button>
        </div>
      </div>
    </>
  );
}

function ProfileEditor({ user, onSaved, showToast }: { user: any; onSaved: () => Promise<void>; showToast: (m: string, t?: any) => void }) {
  const fileRef = useRef<HTMLInputElement>(null);
  const [form, setForm] = useState({
    nickname: user.nickname || '',
    bio: user.bio || '',
    email: user.email || '',
    real_name: user.real_name || '',
    class_name: user.class_name || '',
    department: user.department || '',
  });
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form.real_name.trim() || !form.class_name.trim() || !form.department) {
      showToast('请填写姓名、班级并选择部门', 'error');
      return;
    }
    setSaving(true);
    try {
      await api.updateProfile(form);
      await onSaved();
      showToast('资料更新成功', 'success');
    } catch (err: any) {
      showToast(err.message || '更新失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  const handleAvatarChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const fd = new FormData();
      fd.append('file', file);
      const res = await api.uploadAvatar(fd);
      await api.updateProfile({ avatar: res.data.avatar });
      await onSaved();
      showToast('头像已更新', 'success');
    } catch (err: any) {
      showToast(err.message || '上传失败', 'error');
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="page-intro space-y-2">
        <h2 className="section-title">个人资料</h2>
        <p className="section-subtitle">管理你的会员信息</p>
      </div>

      <div className="sidebar-panel rounded-[2rem] p-8 space-y-6">
        <div className="flex items-center gap-4">
          {user.avatar ? (
            <img src={user.avatar} alt="" className="w-20 h-20 rounded-2xl object-cover border-4 border-white shadow-md" />
          ) : (
            <div className="w-20 h-20 rounded-2xl bg-primary/10 flex items-center justify-center">
              <User className="w-8 h-8 text-primary" />
            </div>
          )}
          <div className="flex-1">
            <input ref={fileRef} type="file" accept="image/*" onChange={handleAvatarChange} className="hidden" />
            <button
              type="button"
              onClick={() => fileRef.current?.click()}
              disabled={uploading}
              className="btn-secondary !text-xs !py-2 !px-4 flex items-center gap-2"
            >
              <Camera className="w-3.5 h-3.5" /> {uploading ? '上传中...' : '更换头像'}
            </button>
            <p className="text-xs text-zinc-400 mt-2">支持 JPG/PNG/GIF/WebP，&lt; 5MB</p>
          </div>
        </div>

        <form onSubmit={handleSave} className="space-y-4">
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">用户名</label>
            <input type="text" value={user.username} disabled className="app-input w-full rounded-xl py-2.5 px-4 bg-zinc-100" />
          </div>
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">昵称</label>
            <input type="text" value={form.nickname} onChange={(e) => setForm({ ...form, nickname: e.target.value })} className="app-input w-full rounded-xl py-2.5 px-4" autoComplete="nickname" />
          </div>
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">姓名 <span className="text-rose-500">*</span></label>
            <input type="text" value={form.real_name} onChange={(e) => setForm({ ...form, real_name: e.target.value })} className="app-input w-full rounded-xl py-2.5 px-4" autoComplete="name" />
          </div>
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">班级 <span className="text-rose-500">*</span></label>
            <input type="text" value={form.class_name} onChange={(e) => setForm({ ...form, class_name: e.target.value })} className="app-input w-full rounded-xl py-2.5 px-4" placeholder="请输入所在班级" />
          </div>
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">部门 <span className="text-rose-500">*</span></label>
            <select value={form.department} onChange={(e) => setForm({ ...form, department: e.target.value })} className="app-input w-full rounded-xl py-2.5 px-4">
              <option value="">请选择部门</option>
              {DEPARTMENTS.map((d) => (
                <option key={d} value={d}>{d}</option>
              ))}
            </select>
          </div>
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">邮箱</label>
            <input type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} className="app-input w-full rounded-xl py-2.5 px-4" autoComplete="email" />
          </div>
          <div>
            <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">个人简介</label>
            <textarea value={form.bio} onChange={(e) => setForm({ ...form, bio: e.target.value })} rows={4} className="app-input w-full rounded-xl py-2.5 px-4" />
          </div>
          <button type="submit" disabled={saving} className="btn-primary flex items-center gap-2">
            <Save className="w-4 h-4" /> {saving ? '保存中...' : '保存修改'}
          </button>
        </form>
      </div>
    </div>
  );
}

function FavoritesView({ showToast }: { showToast: (m: string, t?: any) => void }) {
  const [items, setItems] = useState<Favorite[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getFavorites()
      .then((res) => setItems(res.data || []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, []);

  const remove = async (id: number) => {
    try {
      await api.removeFavorite(id);
      setItems(items.filter((it) => it.id !== id));
      showToast('已取消收藏', 'info');
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="page-intro space-y-2">
        <h2 className="section-title">我的收藏</h2>
        <p className="section-subtitle">{items.length} 项</p>
      </div>

      {loading ? (
        <div className="text-center text-zinc-500 py-12">加载中...</div>
      ) : items.length === 0 ? (
        <div className="sidebar-panel rounded-[2rem] p-12 text-center">
          <Heart className="w-10 h-10 text-zinc-300 mx-auto mb-3" />
          <p className="text-zinc-500">还没有收藏任何内容</p>
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((fav) => (
            <div key={fav.id} className="glass-card rounded-2xl p-5 flex items-center gap-4">
              <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
                <BookOpen className="w-5 h-5 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="font-medium text-zinc-800 truncate">{fav.title || fav.target_type + ' #' + fav.target_id}</div>
                <div className="text-xs text-zinc-500 mt-1">
                  {fav.target_type} · {fav.created_at ? new Date(fav.created_at).toLocaleDateString('zh-CN') : ''}
                </div>
              </div>
              <button
                onClick={() => remove(fav.id)}
                className="text-rose-500 hover:bg-rose-50 px-3 py-1.5 rounded-xl text-sm flex items-center gap-1.5"
              >
                <Trash2 className="w-4 h-4" /> 取消
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function DownloadsView({ showToast }: { showToast: (m: string, t?: any) => void }) {
  const [items, setItems] = useState<Download[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getMyDownloads(1)
      .then((res) => setItems(res.data || []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="space-y-6">
      <div className="page-intro space-y-2">
        <h2 className="section-title">下载历史</h2>
        <p className="section-subtitle">最近 20 条</p>
      </div>

      {loading ? (
        <div className="text-center text-zinc-500 py-12">加载中...</div>
      ) : items.length === 0 ? (
        <div className="sidebar-panel rounded-[2rem] p-12 text-center">
          <Download className="w-10 h-10 text-zinc-300 mx-auto mb-3" />
          <p className="text-zinc-500">还没有下载记录</p>
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((d) => (
            <div key={d.id} className="glass-card rounded-2xl p-5 flex items-center gap-4">
              <div className="w-12 h-12 rounded-xl bg-primary/10 flex items-center justify-center">
                <Download className="w-5 h-5 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="font-medium text-zinc-800">资源 #{d.resource_id}</div>
                <div className="text-xs text-zinc-500 mt-1 flex items-center gap-3">
                  <span className="flex items-center gap-1">
                    <Clock className="w-3 h-3" />
                    {new Date(d.created_at).toLocaleString('zh-CN')}
                  </span>
                  {d.ip && <span>IP: {d.ip}</span>}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function PasswordView({ showToast }: { showToast: (m: string, t?: any) => void }) {
  const [oldPwd, setOldPwd] = useState('');
  const [newPwd, setNewPwd] = useState('');
  const [confirmPwd, setConfirmPwd] = useState('');
  const [saving, setSaving] = useState(false);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (newPwd.length < 6 || !/(?=.*[A-Za-z])(?=.*\d)/.test(newPwd)) {
      showToast('新密码至少 6 位，且必须同时包含字母和数字', 'error');
      return;
    }
    if (newPwd !== confirmPwd) { showToast('两次输入不一致', 'error'); return; }
    setSaving(true);
    try {
      await api.changePassword(oldPwd, newPwd);
      showToast('密码已修改，请重新登录', 'success');
      setOldPwd(''); setNewPwd(''); setConfirmPwd('');
    } catch (err: any) {
      showToast(err.message || '修改失败', 'error');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="page-intro space-y-2">
        <h2 className="section-title">修改密码</h2>
        <p className="section-subtitle">为了账户安全，建议定期更换</p>
      </div>

      <form onSubmit={submit} className="sidebar-panel rounded-[2rem] p-8 space-y-4">
        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">当前密码</label>
          <input type="password" value={oldPwd} onChange={(e) => setOldPwd(e.target.value)} className="app-input w-full rounded-xl py-2.5 px-4" autoComplete="current-password" />
        </div>
        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">新密码（至少 6 位，需包含字母和数字）</label>
          <input type="password" value={newPwd} onChange={(e) => setNewPwd(e.target.value)} className="app-input w-full rounded-xl py-2.5 px-4" autoComplete="new-password" />
        </div>
        <div>
          <label className="text-xs font-bold text-zinc-500 uppercase tracking-wider mb-2 block">确认新密码</label>
          <input type="password" value={confirmPwd} onChange={(e) => setConfirmPwd(e.target.value)} className="app-input w-full rounded-xl py-2.5 px-4" autoComplete="new-password" />
        </div>
        <button type="submit" disabled={saving} className="btn-primary flex items-center gap-2">
          <Lock className="w-4 h-4" /> {saving ? '提交中...' : '提交修改'}
        </button>
      </form>
    </div>
  );
}

/** 我的活动：报名列表 + 取消报名 */
function RegistrationsView({ showToast, navigate }: { showToast: (m: string, t?: any) => void; navigate: (tab: any) => void }) {
  const [items, setItems] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const load = () => {
    setLoading(true);
    api.getMyRegistrations()
      .then((res) => setItems(res.data || []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    load();
  }, []);

  const cancel = async (eventId: number) => {
    if (!confirm('确定取消报名吗？')) return;
    try {
      await api.cancelEventRegistration(eventId);
      showToast('已取消报名', 'success');
      load();
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    }
  };

  return (
    <div className="space-y-6">
      <div className="page-intro space-y-2">
        <h2 className="section-title">我的活动</h2>
        <p className="section-subtitle">已报名的活动与签到状态</p>
      </div>
      {loading ? (
        <div className="text-center text-zinc-500 py-12">加载中...</div>
      ) : items.length === 0 ? (
        <div className="sidebar-panel rounded-[2rem] p-10 text-center text-zinc-500">
          还没有报名任何活动，去 <button className="text-primary" onClick={() => navigate('events')}>活动中心</button> 看看吧
        </div>
      ) : (
        <div className="space-y-3">
          {items.map((it) => {
            const started = it.start_time ? new Date(it.start_time).getTime() < Date.now() : false;
            return (
              <div key={it.event_id} className="sidebar-panel rounded-2xl p-5 flex flex-wrap items-center gap-4">
                <div className="flex-1 min-w-0">
                  <div className="font-medium text-charcoal truncate">{it.event_title}</div>
                  <div className="text-xs text-zinc-500 mt-1 flex flex-wrap gap-x-3 gap-y-1">
                    <span>{new Date(it.start_time).toLocaleString('zh-CN')}</span>
                    {it.event_location && <span>📍 {it.event_location}</span>}
                    <span>报名于 {new Date(it.registered_at).toLocaleString('zh-CN')}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  {it.status === 2 ? (
                    <span className="px-3 py-1 rounded-full text-xs font-bold bg-emerald-50 text-emerald-600 border border-emerald-200">
                      已签到 {it.checked_in_at ? new Date(it.checked_in_at).toLocaleString('zh-CN') : ''}
                    </span>
                  ) : (
                    <span className="px-3 py-1 rounded-full text-xs font-bold bg-blue-50 text-blue-600 border border-blue-200">
                      已报名
                    </span>
                  )}
                  {it.status === 1 && !started && (
                    <button onClick={() => cancel(it.event_id)} className="btn-ghost !px-3 !py-1.5 !text-xs text-rose-500">
                      取消报名
                    </button>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
}
