import { useState } from 'react';
import {
  Activity, User, ShieldCheck,
  FileText, LogIn, LogOut, MessageSquare, KeyRound, Bell, CalendarCheck,
} from 'lucide-react';
import type { ViewProps } from '@/src/types/app';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { AdminLoginModal } from './admin/AdminLoginModal';
import { DashboardPanel } from './admin/DashboardPanel';
import { ContentPanel } from './admin/ContentPanel';
import { UsersPanel } from './admin/UsersPanel';
import { CommentsPanel } from './admin/CommentsPanel';
import { AdminRegistrationsPanel } from './admin/AdminRegistrationsPanel';
import { NotificationsPanel } from './admin/NotificationsPanel';
import { AccountPanel } from './admin/AccountPanel';
import type { SubView, ContentType } from './admin/types';

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
      <header className="page-intro flex flex-col gap-4 lg:flex-row lg:items-center lg:justify-between rounded-3xl p-5 lg:p-8">
        <div className="flex items-center gap-4">
          <div className="w-12 h-12 lg:w-16 lg:h-16 rounded-full overflow-hidden border-4 border-white shadow-lg bg-primary/10 flex items-center justify-center shrink-0">
            <ShieldCheck className="w-6 h-6 lg:w-7 lg:h-7 text-primary" />
          </div>
          <div>
            <h2 className="text-xl lg:text-2xl font-serif text-primary">管理员控制台</h2>
            <p className="section-subtitle !text-sm">已登录为 {admin.username}</p>
          </div>
        </div>
        <div className="flex flex-wrap gap-2">
          <button onClick={() => setSubView('dashboard')} className={subBtnCls(subView === 'dashboard')}>
            <Activity className="w-4 h-4" /> 概览
          </button>
          <button onClick={() => setSubView('content')} className={subBtnCls(subView === 'content')}>
            <FileText className="w-4 h-4" /> 内容
          </button>
          <button onClick={() => setSubView('users')} className={subBtnCls(subView === 'users')}>
            <User className="w-4 h-4" /> 用户
          </button>
          <button onClick={() => setSubView('registrations')} className={subBtnCls(subView === 'registrations')}>
            <CalendarCheck className="w-4 h-4" /> 报名
          </button>
          <button onClick={() => setSubView('comments')} className={subBtnCls(subView === 'comments')}>
            <MessageSquare className="w-4 h-4" /> 评论
          </button>
          <button onClick={() => setSubView('notifications')} className={subBtnCls(subView === 'notifications')}>
            <Bell className="w-4 h-4" /> 通知
          </button>
          <button onClick={() => setSubView('account')} className={subBtnCls(subView === 'account')}>
            <KeyRound className="w-4 h-4" /> 账号
          </button>
          <button
            onClick={async () => { await logoutAdmin(); showToast('已退出后台', 'info'); }}
            className="px-4 py-2 rounded-full text-xs font-semibold bg-rose-50 text-rose-500 border border-rose-200 flex items-center gap-1.5"
          >
            <LogOut className="w-3.5 h-3.5" /> 退出
          </button>
        </div>
      </header>

      {subView === 'dashboard' && <DashboardPanel onSwitch={setSubView} />}
      {subView === 'content' && <ContentPanel type={contentType} setType={setContentType} />}
      {subView === 'users' && <UsersPanel />}
      {subView === 'registrations' && <AdminRegistrationsPanel />}
      {subView === 'comments' && <CommentsPanel />}
      {subView === 'notifications' && <NotificationsPanel />}
      {subView === 'account' && <AccountPanel />}
    </div>
  );
};

function subBtnCls(active: boolean) {
  return `px-4 py-2 rounded-full text-xs font-semibold flex items-center gap-1.5 transition-colors ${
    active ? 'bg-primary text-white shadow-md' : 'bg-white/60 text-zinc-600 hover:bg-white'
  }`;
}
