import React, { useEffect, useMemo, useState } from 'react';
import { Calendar, MapPin, Plus } from 'lucide-react';
import { AsyncMarkdownViewer } from '@/src/components/AsyncMarkdownViewer';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';

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
  const [activeCategory, setActiveCategory] = useState('全部');

  useEffect(() => {
    setLoading(true);
    api.getEvents()
      .then((res) => setEvents(res.data || []))
      .catch(() => setEvents([]))
      .finally(() => setLoading(false));
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
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
          {filteredEvents.map((event) => {
            const date = formatDate(event.start_time);
            return (
                <article key={event.id} className="glass-card rounded-[2rem] overflow-hidden flex flex-col group">
                  <div className="h-44 overflow-hidden relative bg-accent-soft">
                  {event.cover_url ? (
                    <img src={event.cover_url} alt={event.title} className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-700" />
                  ) : (
                    <div className="w-full h-full bg-gradient-to-br from-accent-soft to-white flex items-center justify-center text-accent text-3xl font-medium">
                      {event.title?.[0] || '数'}
                    </div>
                  )}
                  {event.category && (
                    <div className="absolute top-4 left-4">
                      <span className="math-tag !bg-white/90 backdrop-blur-md shadow-sm">{event.category}</span>
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
                          actions: [
                            { label: '前往会员专区', onClick: () => navigate('portal') },
                            {
                              label: event.category === '竞赛' ? '查看作品归档' : '查看会员服务',
                              variant: 'secondary',
                              onClick: () => navigate(event.category === '竞赛' ? 'showcase' : 'portal'),
                            },
                          ],
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
      )}
    </div>
  );
};
