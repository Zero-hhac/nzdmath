import React, { useState } from 'react';
import { cn } from '@/src/lib/utils';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { LoginModal } from './LoginModal';
import { motion, AnimatePresence } from 'motion/react';
import { LogIn, LogOut, User, ChevronDown, ShieldCheck, Menu, X } from 'lucide-react';
import type { TabId } from '@/src/types/app';

interface NavbarProps {
  activeTab: TabId;
  setActiveTab: (tab: TabId) => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activeTab, setActiveTab }) => {
  const { user, admin, logoutUser, logoutAdmin } = useAuth();
  const { showToast } = useToast();
  const [loginOpen, setLoginOpen] = useState(false);
  const [loginMode, setLoginMode] = useState<'user' | 'admin'>('user');
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [adminMenuOpen, setAdminMenuOpen] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const navItems = [
    { id: 'home' as const, label: '首页' },
    { id: 'events' as const, label: '活动中心' },
    { id: 'resources' as const, label: '资源存档' },
    { id: 'news' as const, label: '动态资讯' },
    { id: 'about' as const, label: '关于我们' },
    { id: 'chat' as const, label: '聊天室' },
  ];

  const openUserLogin = () => { setLoginMode('user'); setLoginOpen(true); };
  const openAdminLogin = () => { setLoginMode('admin'); setLoginOpen(true); };
  const selectTab = (tab: TabId) => {
    setActiveTab(tab);
    setMobileOpen(false);
  };

  const handleLogout = async () => {
    await logoutUser();
    setUserMenuOpen(false);
    showToast('已退出登录', 'info');
  };

  const handleAdminLogout = async () => {
    await logoutAdmin();
    setAdminMenuOpen(false);
    showToast('已退出后台', 'info');
  };

  return (
    <>
      <div className="fixed top-3 md:top-5 left-1/2 -translate-x-1/2 w-[calc(100%-1.5rem)] max-w-6xl z-50 flex justify-center">
        <nav className="glass-nav rounded-[1.5rem] md:rounded-full px-4 md:px-6 py-3 flex justify-between items-center w-full transition-all duration-300">
          <div
            className="font-sans text-lg font-bold tracking-tight text-charcoal cursor-pointer hover:opacity-80 transition-opacity"
            onClick={() => selectTab('home')}
          >
            数学协会
          </div>

          <div className="hidden md:flex gap-5 lg:gap-8 items-center">
            {navItems.map((item) => (
              <button
                key={item.id}
                onClick={() => selectTab(item.id)}
                className={cn(
                  'relative py-1 text-[13px] tracking-wide font-medium transition-all duration-300 hover:text-charcoal hover:-translate-y-0.5',
                  activeTab === item.id
                    ? 'text-charcoal font-bold'
                    : 'text-text-muted'
                )}
                aria-current={activeTab === item.id ? 'page' : undefined}
              >
                {item.label}
                {activeTab === item.id && (
                  <motion.div
                    layoutId="nav-underline"
                    className="absolute -bottom-1 left-1/2 -translate-x-1/2 h-1 w-1 bg-charcoal rounded-full"
                  />
                )}
              </button>
            ))}

            <button
              onClick={() => selectTab('portal')}
              className={cn(
                'text-[13px] tracking-wide font-medium transition-all duration-300 hover:text-charcoal hover:-translate-y-0.5',
                activeTab === 'portal' ? 'text-charcoal font-bold' : 'text-text-muted'
              )}
              aria-current={activeTab === 'portal' ? 'page' : undefined}
            >
              会员专区
            </button>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setMobileOpen((open) => !open)}
              className="md:hidden flex h-9 w-9 items-center justify-center rounded-full border border-border bg-surface text-charcoal"
              aria-label={mobileOpen ? '关闭导航' : '打开导航'}
              aria-expanded={mobileOpen}
            >
              {mobileOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
            </button>
            {admin ? (
              <div className="relative">
                <button
                  onClick={() => setAdminMenuOpen(!adminMenuOpen)}
                  className="flex items-center gap-2 px-4 py-2 rounded-full bg-emerald-50 text-emerald-700 text-xs font-semibold border border-emerald-200"
                >
                  <ShieldCheck className="w-3.5 h-3.5" />
                  {admin.username}
                  <ChevronDown className="w-3 h-3" />
                </button>
                <AnimatePresence>
                  {adminMenuOpen && (
                    <motion.div
                      initial={{ opacity: 0, y: -8 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -8 }}
                      className="absolute right-0 top-full mt-2 w-44 bg-surface border border-border rounded-2xl p-2 shadow-[0_12px_40px_rgba(0,0,0,0.06)]"
                      onMouseLeave={() => setAdminMenuOpen(false)}
                    >
                      <button
                        onClick={() => { setActiveTab('admin'); setAdminMenuOpen(false); }}
                        className="w-full text-left px-3 py-2 rounded-xl text-sm text-zinc-700 hover:bg-white/60 flex items-center gap-2"
                      >
                        <User className="w-4 h-4" />
                        进入后台
                      </button>
                      <button
                        onClick={handleAdminLogout}
                        className="w-full text-left px-3 py-2 rounded-xl text-sm text-rose-500 hover:bg-rose-50 flex items-center gap-2"
                      >
                        <LogOut className="w-4 h-4" />
                        退出后台
                      </button>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            ) : null}

            {user ? (
              <div className="relative">
                <button
                  onClick={() => setUserMenuOpen(!userMenuOpen)}
                  className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-black/[0.02] border border-border transition-all hover:bg-black/[0.04]"
                >
                  {user.avatar ? (
                    <img src={user.avatar} alt="" className="w-6 h-6 rounded-full object-cover" />
                  ) : (
                    <div className="w-6 h-6 rounded-full bg-pastel-blue flex items-center justify-center text-pastel-blue-text text-xs font-bold">
                      {(user.nickname || user.username)?.[0]?.toUpperCase()}
                    </div>
                  )}
                  <span className="text-xs font-semibold text-charcoal hidden sm:inline">
                    {user.nickname || user.username}
                  </span>
                  <ChevronDown className="w-3 h-3 text-text-muted" />
                </button>
                <AnimatePresence>
                  {userMenuOpen && (
                    <motion.div
                      initial={{ opacity: 0, y: -8 }}
                      animate={{ opacity: 1, y: 0 }}
                      exit={{ opacity: 0, y: -8 }}
                      className="absolute right-0 top-full mt-2 w-44 bg-surface border border-border rounded-2xl p-2 shadow-[0_12px_40px_rgba(0,0,0,0.06)]"
                      onMouseLeave={() => setUserMenuOpen(false)}
                    >
                      <button
                        onClick={() => { setActiveTab('portal'); setUserMenuOpen(false); }}
                        className="w-full text-left px-3 py-2 rounded-xl text-sm text-zinc-700 hover:bg-white/60 flex items-center gap-2"
                      >
                        <User className="w-4 h-4" />
                        个人中心
                      </button>
                      <button
                        onClick={handleLogout}
                        className="w-full text-left px-3 py-2 rounded-xl text-sm text-rose-500 hover:bg-rose-50 flex items-center gap-2"
                      >
                        <LogOut className="w-4 h-4" />
                        退出登录
                      </button>
                    </motion.div>
                  )}
                </AnimatePresence>
              </div>
            ) : (
              <div className="flex items-center gap-2">
                <button
                  onClick={openUserLogin}
                  className="btn-accent text-xs !px-4 !py-2"
                >
                  <LogIn className="w-3.5 h-3.5" />
                  <span className="hidden sm:inline">登录</span>
                </button>
                <button
                  onClick={openAdminLogin}
                  className="hidden lg:block text-text-muted text-[10px] font-medium hover:text-charcoal transition-colors px-2"
                  title="管理员登录"
                >
                  管理入口
                </button>
              </div>
            )}
          </div>
        </nav>
        <AnimatePresence>
          {mobileOpen && (
            <motion.div
              initial={{ opacity: 0, y: -8 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -8 }}
              className="absolute top-full mt-2 w-full rounded-3xl border border-border bg-surface/95 p-3 shadow-[0_20px_60px_rgba(31,42,68,0.12)] backdrop-blur-xl md:hidden"
            >
              <div className="grid grid-cols-1 gap-1">
                {[...navItems, { id: 'portal' as const, label: '会员专区' }].map((item) => (
                  <button
                    key={item.id}
                    onClick={() => selectTab(item.id)}
                    className={cn(
                      'rounded-2xl px-4 py-3 text-left text-sm font-medium transition-colors',
                      activeTab === item.id ? 'bg-accent text-white' : 'text-charcoal-muted hover:bg-canvas-alt',
                    )}
                    aria-current={activeTab === item.id ? 'page' : undefined}
                  >
                    {item.label}
                  </button>
                ))}
              </div>
              {!admin && (
                <button onClick={openAdminLogin} className="mt-2 w-full px-4 py-2 text-left text-xs text-text-muted">
                  管理员入口
                </button>
              )}
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <LoginModal open={loginOpen} onClose={() => setLoginOpen(false)} defaultMode={loginMode} />
    </>
  );
};
