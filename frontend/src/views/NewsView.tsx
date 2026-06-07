import React, { useEffect, useState } from 'react';
import { Formula } from '@/src/components/Formula';
import { AsyncMarkdownViewer } from '@/src/components/AsyncMarkdownViewer';
import { ArrowRight } from 'lucide-react';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';

type NewsItem = {
  id: number;
  title: string;
  summary?: string;
  content?: string;
  category?: string;
  tag?: string;
  cover_url?: string;
  published_at?: string;
  created_at?: string;
};

export const NewsView: React.FC<ViewProps> = ({ navigate, openOverlay }) => {
  const [news, setNews] = useState<NewsItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getNews()
      .then((res) => setNews(res.data || []))
      .catch(() => setNews([]))
      .finally(() => setLoading(false));
  }, []);

  return (
    <div className="grid grid-cols-1 lg:grid-cols-3 gap-12">
      <div className="lg:col-span-2 space-y-12">
        <div className="page-intro space-y-2">
          <div className="section-kicker">News</div>
          <h2 className="section-title">动态资讯</h2>
          <p className="section-subtitle">从黑板到世界，追踪每一个逻辑步长</p>
        </div>

        {loading ? (
          <div className="text-center text-zinc-500 py-20">加载中...</div>
        ) : news.length === 0 ? (
          <div className="text-center text-zinc-500 py-20">暂无资讯</div>
        ) : (
          <div className="space-y-6">
            {news.map((item) => {
              const date = item.published_at || item.created_at;
              const dateStr = date ? new Date(date).toISOString().slice(0, 10).replace(/-/g, '.') : '';
              return (
                <div key={item.id} className="glass-card rounded-[2rem] p-8 space-y-4 hover:bg-white transition-all duration-300">
                  <div className="flex justify-between items-center">
                    <div className="flex gap-4 items-center">
                      <span className="px-3 py-1 bg-zinc-100 text-zinc-500 text-[10px] font-bold rounded-md uppercase tracking-wider">
                        {item.category || '资讯'}
                      </span>
                      <span className="text-xs text-zinc-400 font-medium font-mono">{dateStr}</span>
                    </div>
                    {item.tag && (
                      <span className="math-tag !bg-white border-zinc-200">{item.tag}</span>
                    )}
                  </div>
                  <h3 className="text-2xl font-serif text-primary leading-tight">{item.title}</h3>
                  <p className="text-soft-body font-medium leading-relaxed">
                    {item.summary}
                  </p>
                  <button
                    onClick={() =>
                      openOverlay({
                        title: item.title,
                        subtitle: `${item.category || ''} · ${dateStr}`,
                        content: (
                          <div className="space-y-5">
                            {item.cover_url && (
                              <div className="aspect-[16/8] overflow-hidden rounded-[1.75rem]">
                                <img src={item.cover_url} alt={item.title} className="h-full w-full object-cover" />
                              </div>
                            )}
                            <div className="rounded-[1.75rem] border border-white/70 bg-[#f7f9fb] p-6 max-h-[60vh] overflow-y-auto">
                              <AsyncMarkdownViewer type="news" id={item.id} initialContent={item.content} />
                            </div>
                            {/* Content placeholder removed for a cleaner markdown-only read */}
                          </div>
                        ),
                        actions: [
                          { label: '查看相关资源', onClick: () => navigate('resources') },
                          { label: '关闭', variant: 'secondary' },
                        ],
                      })
                    }
                    className="btn-ghost flex items-center gap-1 pt-2 group"
                  >
                    阅读全文 <ArrowRight className="w-4 h-4 group-hover:translate-x-1 transition-transform" />
                  </button>
                </div>
              );
            })}
          </div>
        )}
      </div>

      <div className="space-y-10">
        <div className="sidebar-panel space-y-6 bg-gradient-to-br from-white/88 to-[#e3f2fd]/42">
          <div className="flex items-center gap-3">
            <div className="w-10 h-10 rounded-xl bg-primary text-white flex items-center justify-center shadow-lg">
              <span className="material-symbols-outlined">question_mark</span>
            </div>
            <h4 className="font-serif text-xl text-primary">今日一题</h4>
          </div>
          <div className="panel-inset text-center">
            <p className="text-soft-strong font-medium mb-4 italic">
              证明：对于任何正整数 <Formula expression="n" className="inline-block px-0.5" />，存在一个含有 <Formula expression="n" className="inline-block px-0.5" /> 个连续合数的序列。
            </p>
            <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest mb-6">难度: ★★★☆☆</div>
            <button
              onClick={() => navigate('portal')}
              className="btn-primary w-full !py-3 !text-xs !font-bold"
            >
              提交答案
            </button>
          </div>
        </div>

        <div className="sidebar-panel space-y-6">
          <h4 className="font-serif text-xl text-primary border-b border-white pb-4">热门讨论</h4>
          <div className="space-y-4">
            {[
              { t: '关于黎曼猜想的非正式讨论', cCount: 42 },
              { t: '如何有效地用 LaTeX 绘制复杂几何图形？', cCount: 56 },
              { t: '求助：调和级数发散速度的直观理解', cCount: 12 },
            ].map((topic, index) => (
              <button
                key={index}
                onClick={() => navigate('portal')}
                className="flex w-full justify-between items-center group cursor-pointer pb-4 border-b border-white/30 last:border-0 last:pb-0 text-left"
              >
                <span className="text-sm font-medium text-soft-body group-hover:text-primary transition-colors line-clamp-1 flex-grow pr-4 italic">
                  {topic.t}
                </span>
                <span className="text-[10px] font-bold text-zinc-400">{topic.cCount}</span>
              </button>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};