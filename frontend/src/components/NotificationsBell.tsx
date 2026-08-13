import React, { useEffect, useRef, useState } from 'react';
import { Bell, CheckCheck, X } from 'lucide-react';
import { motion } from 'motion/react';
import { createPortal } from 'react-dom';
import { api } from '@/src/lib/api';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';
import { connectNotifyWS } from '@/src/lib/ws';
import { useLocation } from 'react-router-dom';

const notifTypeLabels: Record<string, string> = {
  system: '系统通知',
  activity: '活动通知',
  reward: '获奖通知',
};

/**
 * 顶栏通知铃铛：红点数字显示未读条数，点击展开通知面板。
 * 路由变化时刷新未读数；面板内支持单条已读/全部已读/加载更多。
 */
export function NotificationsBell() {
  const { user } = useAuth();
  const { showToast } = useToast();
  const location = useLocation();
  const [open, setOpen] = useState(false);
  const [unread, setUnread] = useState(0);
  const [items, setItems] = useState<any[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [expanded, setExpanded] = useState<number | null>(null);
  const triggerRef = useRef<HTMLButtonElement>(null);
  const panelRef = useRef<HTMLDivElement>(null);

  const refreshUnread = () => {
    api.getUnreadNotificationCount()
      .then((res) => setUnread(res.data?.count ?? 0))
      .catch(() => {});
  };

  const load = (p: number) => {
    setLoading(true);
    api.getNotifications(p, 20)
      .then((res) => {
        setItems((prev) => (p === 1 ? res.data || [] : [...prev, ...(res.data || [])]));
        setTotal(res.total || 0);
        setPage(p);
      })
      .catch(() => {})
      .finally(() => setLoading(false));
  };

  // 打开面板时拉最新一页；路由变化时刷新未读角标
  useEffect(() => {
    if (open) load(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  useEffect(() => {
    refreshUnread();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.pathname]);

  // 实时推送：管理员发通知 / 报名成功自动通知 → 红点立即刷新（复用 WebSocket 通道）
  useEffect(() => {
    if (!user) return;
    const disconnect = connectNotifyWS(refreshUnread);
    return disconnect;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user?.id]);

  // 点击面板与铃铛之外关闭（面板经 portal 渲染到 body，需分别判定两个 ref）。
  // 用捕获阶段的 pointerdown：即使页面其他脚本 stopPropagation 也能收到。
  useEffect(() => {
    const handler = (e: Event) => {
      const t = e.target as Node;
      if (panelRef.current?.contains(t)) return;
      if (triggerRef.current?.contains(t)) return;
      setOpen(false);
    };
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false);
    };
    document.addEventListener('pointerdown', handler, true);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('pointerdown', handler, true);
      document.removeEventListener('keydown', onKey);
    };
  }, []);

  const markRead = async (id: number) => {
    try {
      await api.markNotificationRead(id);
      setItems((prev) => prev.map((i) => (i.id === id ? { ...i, is_read: true } : i)));
      refreshUnread();
    } catch {
      // ignore
    }
  };

  const markAll = async () => {
    try {
      await api.markAllNotificationsRead();
      setItems((prev) => prev.map((i) => ({ ...i, is_read: true })));
      refreshUnread();
      showToast('已全部标记为已读', 'success');
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    }
  };

  return (
    <div className="relative">
      <button
        ref={triggerRef}
        onClick={() => setOpen((o) => !o)}
        className="relative flex h-9 w-9 items-center justify-center rounded-full border border-border bg-surface text-charcoal transition-colors hover:bg-white"
        aria-label={`通知，${unread} 条未读`}
      >
        <Bell className="h-4 w-4" />
        {unread > 0 && (
          <span className="absolute -top-1.5 -right-1.5 min-w-[18px] h-[18px] px-1 rounded-full bg-rose-500 text-white text-[10px] font-bold flex items-center justify-center shadow-md">
            {unread > 99 ? '99+' : unread}
          </span>
        )}
      </button>

      {createPortal(
        open ? (
          <motion.div
            key="bell-panel-root"
            ref={panelRef}
            className="fixed inset-0 z-40 pointer-events-none flex items-start justify-center"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
          >
            {/* 手机端半透明遮罩，点击关闭 */}
            <div className="absolute inset-0 bg-black/20 md:hidden pointer-events-auto" onClick={() => setOpen(false)} />
            <motion.div
              initial={{ opacity: 0, y: -8, scale: 0.98 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              className="pointer-events-auto relative mt-24 md:mt-20 w-[calc(100vw-1.5rem)] max-w-[26rem] max-h-[70vh] flex flex-col bg-surface border border-border rounded-2xl shadow-[0_20px_60px_rgba(31,42,68,0.16)] overflow-hidden"
            >
            <div className="flex items-center justify-between px-4 py-3 border-b border-border bg-white/40">
              <div className="flex items-center gap-2">
                <Bell className="w-4 h-4 text-primary" />
                <span className="text-sm font-bold text-charcoal">消息通知</span>
                {unread > 0 && (
                  <span className="px-2 py-0.5 rounded-full bg-rose-50 text-rose-500 text-[10px] font-bold border border-rose-200">
                    {unread} 未读
                  </span>
                )}
              </div>
              {unread > 0 && (
                <button
                  onClick={markAll}
                  className="flex items-center gap-1 px-2.5 py-1.5 rounded-full text-[11px] font-semibold text-primary hover:bg-primary/10"
                >
                  <CheckCheck className="w-3.5 h-3.5" /> 全部已读
                </button>
              )}
              <button
                onClick={() => setOpen(false)}
                className="ml-1 flex h-7 w-7 items-center justify-center rounded-full text-zinc-400 hover:bg-black/5 hover:text-zinc-600"
                aria-label="关闭通知面板"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto">
              {loading && items.length === 0 ? (
                <div className="text-center text-zinc-500 text-sm py-10">加载中...</div>
              ) : items.length === 0 ? (
                <div className="text-center text-zinc-400 text-sm py-10">
                  暂无通知
                  <div className="text-xs mt-1">报名成功、管理员发送的通知会出现在这里</div>
                </div>
              ) : (
                <div className="divide-y divide-border/60">
                  {items.map((it) => (
                    <div
                      key={it.id}
                      className={`px-4 py-3 cursor-pointer transition-colors hover:bg-white/60 ${it.is_read ? '' : 'bg-primary/5'}`}
                      onClick={() => {
                        setExpanded((prev) => (prev === it.id ? null : it.id));
                        if (!it.is_read) markRead(it.id);
                      }}
                    >
                      <div className="flex items-start gap-2.5">
                        <div className={`w-2 h-2 mt-1.5 rounded-full shrink-0 ${it.is_read ? 'bg-zinc-200' : 'bg-primary'}`} />
                        <div className="flex-1 min-w-0">
                          <div className="flex flex-wrap items-center gap-1.5">
                            <span className="px-1.5 py-0.5 rounded text-[10px] font-bold bg-zinc-100 text-zinc-500">
                              {notifTypeLabels[it.type] || '通知'}
                            </span>
                            <span className={`text-sm ${it.is_read ? 'text-zinc-500' : 'font-semibold text-charcoal'}`}>
                              {it.title}
                            </span>
                          </div>
                          <div className="text-[11px] text-zinc-400 mt-0.5">
                            {new Date(it.created_at).toLocaleString('zh-CN')}
                          </div>
                          {expanded === it.id && (
                            <div className="mt-2 text-[13px] text-zinc-600 whitespace-pre-wrap leading-relaxed">
                              {it.content}
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {items.length < total && (
              <div className="p-2 border-t border-border">
                <button
                  onClick={() => load(page + 1)}
                  disabled={loading}
                  className="w-full py-2 rounded-xl text-xs font-semibold text-zinc-500 hover:bg-white/60"
                >
                  {loading ? '加载中...' : `加载更多（${items.length}/${total}）`}
                </button>
              </div>
            )}
            </motion.div>
          </motion.div>
        ) : null,
        document.body,
      )}
    </div>
  );
}
