import React, { useEffect, useMemo, useState } from 'react';
import { Calendar, MapPin, Plus } from 'lucide-react';
import { AsyncMarkdownViewer } from '@/src/components/AsyncMarkdownViewer';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';

type EventItem = {
  id: number;
  title: string;
  summary?: string;
  content?: string;
  category?: string;
  location?: string;
  start_time?: string;
  end_time?: string;
  cover_url?: string;
  capacity?: number;
  registered_count?: number;
  is_registered?: boolean;
  is_expired?: boolean;
  is_featured?: boolean;
  status?: number;
  created_at?: string;
};

const categoryAll = ['全部', '竞赛', '讲座', '研讨会', '社交'];

function formatDate(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return '';
  return new Intl.DateTimeFormat('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit' }).format(date);
}

export const EventsView: React.FC<ViewProps> = ({ navigate, openOverlay }) => {
  const [events, setEvents] = useState<EventItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const [activeCategory, setActiveCategory] = useState('全部');

  const load = (p: number) => {
    if (p > 1) setLoadingMore(true);
    else setLoading(true);
    api.getEvents(p, 12)
      .then((res) => {
        setEvents((prev) => (p === 1 ? res.data || [] : [...prev, ...(res.data || [])]));
        setTotal(res.total || 0);
        setPage(p);
      })
      .catch(() => {
        if (p === 1) setEvents([]);
      })
      .finally(() => {
        setLoading(false);
        setLoadingMore(false);
      });
  };

  useEffect(() => {
    load(1);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const filteredEvents = useMemo(() => {
    if (activeCategory === '全部') return events;
    return events.filter((e) => e.category === activeCategory);
  }, [events, activeCategory]);

  return (
    <div className="space-y-12">
      <div className="page-intro flex flex-col md:flex-row justify-between items-end gap-6">
        <div className="space-y-2">
          <div className="section-kicker">Events</div>
          <h2 className="section-title">活动中心</h2>
          <p className="section-subtitle italic">Logic in action, community in motion.</p>
        </div>
        <div className="flex max-w-full flex-wrap gap-2">
          {categoryAll.map((cat) => (
            <button
              key={cat}
              onClick={() => setActiveCategory(cat)}
              className={`px-5 py-2 rounded-full text-xs font-semibold transition-all cursor-pointer ${
                activeCategory === cat ? 'bg-accent text-white shadow-sm' : 'surface-subtle text-text-muted hover:border-accent/30 hover:text-accent'
              }`}
            >
              {cat}
            </button>
          ))}
        </div>
      </div>

      {loading ? (
        <div className="text-center text-zinc-500 py-20">加载中...</div>
      ) : filteredEvents.length === 0 ? (
        <div className="text-center text-zinc-500 py-20">暂无活动</div>
      ) : (
        <>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {filteredEvents.map((event) => {
            const date = formatDate(event.start_time);
            return (
                <article key={event.id} className="glass-card rounded-[2rem] overflow-hidden flex flex-col group">
                  <div className="h-44 overflow-hidden relative bg-accent-soft">
                  {event.cover_url ? (
                    <img
                      src={event.cover_url}
                      alt={event.title}
                      className={`w-full h-full object-cover group-hover:scale-110 transition-transform duration-700 ${event.is_expired ? 'saturate-[.72]' : ''}`}
                    />
                  ) : (
                    <div className="w-full h-full bg-gradient-to-br from-accent-soft to-white flex items-center justify-center text-accent text-3xl font-medium">
                      {event.title?.[0] || '数'}
                    </div>
                  )}
                  {/* 过期仅轻微压暗一点点，保留原色观感 */}
                  {event.is_expired && <div className="absolute inset-0 bg-black/10" />}
                  {event.category && (
                    <div className="absolute top-4 left-4">
                      <span className="math-tag !bg-white/90 backdrop-blur-md shadow-sm">{event.category}</span>
                    </div>
                  )}
                  {event.is_expired && (
                    <div className="absolute top-4 right-4">
                      <span className="px-3 py-1 rounded-full text-[10px] font-bold bg-zinc-800/85 text-white backdrop-blur-md shadow-sm">
                        已过期
                      </span>
                    </div>
                  )}
                </div>

                <div className="p-8 flex-grow flex flex-col justify-between space-y-6">
                  <div className="space-y-3">
                    {date && (
                      <div className="flex items-center text-[10px] uppercase tracking-widest text-accent font-bold">
                        <Calendar className="w-3.5 h-3.5 mr-1" />
                        {date}
                      </div>
                    )}
                    <h3 className="text-xl font-medium tracking-tight text-charcoal leading-snug group-hover:text-accent transition-colors">
                      {event.title}
                    </h3>
                    <p className="text-sm text-soft-body line-clamp-3 font-medium leading-relaxed">
                      {event.summary || '点击查看详情'}
                    </p>
                  </div>

                  <div className="pt-6 border-t border-border flex items-center justify-between gap-4">
                    {event.location && (
                      <div className="flex min-w-0 items-center text-xs text-text-muted font-medium">
                        <MapPin className="w-3.5 h-3.5 mr-1" />
                        {event.location}
                      </div>
                    )}
                    <button
                      onClick={() =>
                        openOverlay({
                          title: event.title,
                          subtitle: `${date} · ${event.location || ''}`,
                          content: (
                            <div className="space-y-6">
                              {event.cover_url && (
                                <div className="aspect-[16/8] overflow-hidden rounded-[1.75rem]">
                                  <img src={event.cover_url} alt={event.title} className="h-full w-full object-cover" />
                                </div>
                              )}
                              <div className="rounded-[1.75rem] border border-border bg-canvas-alt p-6 max-h-[60vh] overflow-y-auto">
                                <AsyncMarkdownViewer type="events" id={event.id} initialContent={event.content} />
                              </div>
                            </div>
                          ),
                          // 报名栏固定在卡片底部，不随详情内容滚动
                          footer: <EventRegisterActions event={event} onNeedLogin={() => navigate('portal')} />,
                        })
                      }
                      className="btn-secondary !h-10 !w-10 !px-0 !py-0 flex items-center justify-center"
                      aria-label="查看活动详情"
                    >
                      <Plus className="w-5 h-5" />
                    </button>
                  </div>
                </div>
              </article>
            );
          })}
          </div>
          {events.length < total && (
            <div className="flex justify-center pt-4">
              <button
                onClick={() => load(page + 1)}
                disabled={loadingMore}
                className="btn-secondary !px-6 !py-2.5"
              >
                {loadingMore ? '加载中...' : `加载更多活动（${events.length}/${total}）`}
              </button>
            </div>
          )}
        </>
      )}
    </div>
  );
};

/**
 * 活动报名面板：展示报名人数/名额，登录后可报名/取消报名。
 * 详情接口按登录态返回 is_registered，报名后本地刷新。
 */
function EventRegisterActions({ event, onNeedLogin }: { event: EventItem; onNeedLogin: () => void }) {
  const { user } = useAuth();
  const { showToast } = useToast();
  const [detail, setDetail] = useState<{ registered_count: number; is_registered: boolean; is_expired: boolean } | null>(null);
  const [busy, setBusy] = useState(false);

  const load = () => {
    api.getEvent(event.id)
      .then((res) => {
        setDetail({
          registered_count: res.data?.registered_count ?? 0,
          is_registered: res.data?.is_registered ?? false,
          is_expired: res.data?.is_expired ?? false,
        });
      })
      .catch(() => {});
  };

  useEffect(() => {
    load();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [event.id]);

  const registeredCount = detail?.registered_count ?? event.registered_count ?? 0;
  const isRegistered = detail?.is_registered ?? false;
  const isExpired = detail?.is_expired ?? event.is_expired ?? false;
  const started = event.start_time ? new Date(event.start_time).getTime() < Date.now() : false;

  const doRegister = async () => {
    if (!user) {
      onNeedLogin();
      return;
    }
    setBusy(true);
    try {
      await api.registerEvent(event.id);
      showToast('报名成功，活动当天凭姓名签到', 'success');
      load();
    } catch (err: any) {
      showToast(err.message || '报名失败', 'error');
    } finally {
      setBusy(false);
    }
  };

  const doCancel = async () => {
    if (!confirm('确定取消报名吗？')) return;
    setBusy(true);
    try {
      await api.cancelEventRegistration(event.id);
      showToast('已取消报名', 'success');
      load();
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    } finally {
      setBusy(false);
    }
  };

  return (
    <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="text-sm text-zinc-600">
        <div className="font-semibold text-charcoal">
          报名人数：{registeredCount}
          {event.capacity ? ` / ${event.capacity} 人` : '（不限名额）'}
        </div>
        {isExpired ? (
          <div className="text-xs text-rose-500 mt-1">活动已过期，停止报名</div>
        ) : isRegistered ? (
          <div className="text-xs text-emerald-600 mt-1">已报名，活动当天凭姓名签到</div>
        ) : started ? (
          <div className="text-xs text-zinc-400 mt-1">活动已开始，停止报名</div>
        ) : null}
      </div>
      {isExpired || started ? null : isRegistered ? (
        <button onClick={doCancel} disabled={busy} className="btn-secondary !py-2.5 !text-xs shrink-0">
          {busy ? '处理中...' : '取消报名'}
        </button>
      ) : (
        <button onClick={doRegister} disabled={busy} className="btn-primary !py-2.5 !text-xs shrink-0">
          {busy ? '处理中...' : user ? '立即报名' : '登录后报名'}
        </button>
      )}
    </div>
  );
}
