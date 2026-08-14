import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { cn } from '@/src/lib/utils';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { LoginModal } from './LoginModal';
import { NotificationsBell } from './NotificationsBell';
import { motion, AnimatePresence } from 'motion/react';
import { LogIn, LogOut, User, ChevronDown, ShieldCheck, Menu, X } from 'lucide-react';
import type { TabId } from '@/src/types/app';
import { tabPaths } from '@/src/lib/routes';

interface NavbarProps {
  activeTab: TabId;
  setActiveTab: (tab: TabId) => void;
}

export const Navbar: React.FC<NavbarProps> = ({ activeTab, setActiveTab }) => {
  const { user, isAdmin, logoutUser } = useAuth();
  const { showToast } = useToast();
  const [loginOpen, setLoginOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [mobileOpen, setMobileOpen] = useState(false);

  const navItems = [
    { id: 'home' as const, label: '首页' },
    { id: 'events' as const, label: '活动中心' },
    { id: 'resources' as const, label: '资源存档' },
    { id: 'news' as const, label: '动态资讯' },
    { id: 'about' as const, label: '关于我们' },
    { id: 'chat' as const, label: '聊天室' },
  ];

  const openLogin = () => setLoginOpen(true);
  const closeMobile = () => setMobileOpen(false);

  const handleLogout = async () => {
    await logoutUser();
    setUserMenuOpen(false);
    showToast('已退出登录', 'info');
  };

  return (
    <>
      <div className="fixed top-3 md:top-5 left-1/2 -translate-x-1/2 w-[calc(100%-1.5rem)] max-w-6xl z-50 flex justify-center">
        <nav className="glass-nav rounded-[1.5rem] md:rounded-full px-4 md:px-6 py-3 flex justify-between items-center w-full transition-all duration-300">
          <Link
            to="/"
            className="font-sans text-lg font-bold tracking-tight text-charcoal cursor-pointer hover:opacity-80 transition-opacity"
          >
            数学协会
          </Link>

          <div className="hidden md:flex gap-5 lg:gap-8 items-center">
            {navItems.map((item) => (
              <Link
                key={item.id}
                to={tabPaths[item.id]}
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
              </Link>
            ))}

            <Link
              to={tabPaths.portal}
              className={cn(
                'text-[13px] tracking-wide font-medium transition-all duration-300 hover:text-charcoal hover:-translate-y-0.5',
                activeTab === 'portal' ? 'text-charcoal font-bold' : 'text-text-muted'
              )}
              aria-current={activeTab === 'portal' ? 'page' : undefined}
            >
              会员专区
            </Link>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setMobileOpen((open) => !open)}
              className="md:hidden flex h-9 w-9 items-center justify-center rounded-full border border-border bg-surface text-charcoal cursor-pointer"
              aria-label={mobileOpen ? '关闭导航' : '打开导航'}
              aria-expanded={mobileOpen}
            >
              {mobileOpen ? <X className="h-4 w-4" /> : <Menu className="h-4 w-4" />}
            </button>

            {user ? (
              <>
                <NotificationsBell />
                <div className="relative">
                  <button
                    onClick={() => setUserMenuOpen(!userMenuOpen)}
                    className="flex items-center gap-2 px-3 py-1.5 rounded-full bg-black/[0.02] border border-border transition-all hover:bg-black/[0.04] cursor-pointer"
                  >
                    {user.avatar ? (
                      <img src={user.avatar} alt="" className="w-6 h-6 rounded-full object-cover border border-border" />
                    ) : (
                      <div className="w-6 h-6 rounded-full bg-pastel-blue flex items-center justify-center text-pastel-blue-text text-xs font-bold">
                        {(user.nickname || user.username)?.[0]?.toUpperCase()}
                      </div>
                    )}
                    <span className="text-xs font-semibold text-charcoal hidden sm:inline">
                      {user.nickname || user.username}
                    </span>
                    {isAdmin && (
                      <span className="hidden sm:inline-block px-1.5 py-0.2 rounded text-[10px] font-bold bg-emerald-100 text-emerald-800 border border-emerald-300/60">
                        管理员
                      </span>
                    )}
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
                        {isAdmin && (
                          <button
                            onClick={() => { setActiveTab('admin'); setUserMenuOpen(false); }}
                            className="w-full text-left px-3 py-2 rounded-xl text-sm text-zinc-700 hover:bg-emerald-50 hover:text-emerald-700 flex items-center gap-2 cursor-pointer transition-colors"
                          >
                            <ShieldCheck className="w-4 h-4 text-emerald-600" />
                            进入后台
                          </button>
                        )}
                        <button
                          onClick={() => { setActiveTab('portal'); setUserMenuOpen(false); }}
                          className="w-full text-left px-3 py-2 rounded-xl text-sm text-zinc-700 hover:bg-white/60 flex items-center gap-2 cursor-pointer transition-colors"
                        >
                          <User className="w-4 h-4" />
                          个人中心
                        </button>
                        <button
                          onClick={handleLogout}
                          className="w-full text-left px-3 py-2 rounded-xl text-sm text-rose-500 hover:bg-rose-50 flex items-center gap-2 cursor-pointer transition-colors"
                        >
                          <LogOut className="w-4 h-4" />
                          退出登录
                        </button>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </div>
              </>
            ) : (
              <div className="flex items-center gap-2">
                <button
                  onClick={openLogin}
                  className="btn-accent text-xs !px-5 !py-2 cursor-pointer"
                >
                  <LogIn className="w-3.5 h-3.5" />
                  <span>登录</span>
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
                  <Link
                    key={item.id}
                    to={tabPaths[item.id]}
                    onClick={closeMobile}
                    className={cn(
                      'rounded-2xl px-4 py-3 text-left text-sm font-medium transition-colors',
                      activeTab === item.id ? 'bg-accent text-white' : 'text-charcoal-muted hover:bg-canvas-alt',
                    )}
                    aria-current={activeTab === item.id ? 'page' : undefined}
                  >
                    {item.label}
                  </Link>
                ))}
                {isAdmin && (
                  <Link
                    to="/admin"
                    onClick={closeMobile}
                    className="rounded-2xl px-4 py-3 text-left text-sm font-semibold text-emerald-700 hover:bg-emerald-50 transition-colors flex items-center gap-2"
                  >
                    <ShieldCheck className="w-4 h-4" />
                    进入后台
                  </Link>
                )}
              </div>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      <LoginModal open={loginOpen} onClose={() => setLoginOpen(false)} />
    </>
  );
};
