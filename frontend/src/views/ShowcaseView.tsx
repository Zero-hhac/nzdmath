import React, { useEffect, useMemo, useState } from 'react';
import { Search, Heart, ArrowRight } from 'lucide-react';
import { AsyncMarkdownViewer } from '@/src/components/AsyncMarkdownViewer';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';

type ShowcaseItem = {
  id: number;
  title: string;
  author?: string;
  field?: string;
  competition?: string;
  summary?: string;
  cover_url?: string;
  h5_url?: string;
  view_count?: number;
  is_featured?: boolean;
};

const categories = ['全部领域', '几何学', '代数学', '算法与计算', '拓扑学', '数论'];
const sorts = ['浏览最多', '最新发布'];

export const ShowcaseView: React.FC<ViewProps> = ({ navigate, openOverlay }) => {
  const [items, setItems] = useState<ShowcaseItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [field, setField] = useState('全部领域');
  const [query, setQuery] = useState('');
  const [sort, setSort] = useState('浏览最多');
  const [activeH5Url, setActiveH5Url] = useState<string | null>(null);

  useEffect(() => {
    setLoading(true);
    const params: Record<string, string> = {};
    if (field !== '全部领域') params.field = field;
    if (query) params.keyword = query;
    api.getShowcases(params)
      .then((res) => setItems(res.data || []))
      .catch(() => setItems([]))
      .finally(() => setLoading(false));
  }, [field, query]);

  const sorted = useMemo(() => {
    const arr = [...items];
    if (sort === '最新发布') {
      arr.sort((a, b) => b.id - a.id);
    } else {
      arr.sort((a, b) => (b.view_count || 0) - (a.view_count || 0));
    }
    return arr;
  }, [items, sort]);

  return (
    <div className="space-y-10 py-6">
      <div className="page-intro flex flex-col gap-4 md:flex-row md:items-end md:justify-between">
        <div className="space-y-2">
          <div className="section-kicker">Showcase Archive</div>
          <h1 className="font-serif text-4xl text-primary md:text-5xl">数韵之美 - 作品档案馆</h1>
          <p className="section-subtitle">汇集竞赛作品、可视化成果与优秀手稿。</p>
        </div>
        <div className="flex flex-wrap gap-3">
          <button onClick={() => navigate('resources')} className="btn-secondary">返回资源库</button>
          <button onClick={() => navigate('events')} className="btn-primary">查看相关赛事</button>
        </div>
      </div>

      <div className="flex flex-col gap-6 lg:flex-row">
        <aside className="w-full space-y-6 lg:w-72">
          <div className="sidebar-panel">
            <h3 className="mb-4 border-b border-white/50 pb-3 font-serif text-xl text-primary">领域分类</h3>
            <div className="space-y-2">
              {categories.map((item) => (
                <button
                  key={item}
                  onClick={() => setField(item)}
                  className={`w-full rounded-xl px-4 py-2 text-left text-sm font-medium transition-all ${
                    field === item ? 'bg-primary-container text-primary' : 'text-[#5d6d77] hover:bg-white/55'
                  }`}
                >
                  {item}
                </button>
              ))}
            </div>
          </div>

          <div className="sidebar-panel">
            <h3 className="mb-4 border-b border-white/50 pb-3 font-serif text-xl text-primary">热门标签</h3>
            <div className="flex flex-wrap gap-2">
              {['# 黎曼猜想', '# 庞加莱猜想', '# 分形几何', '# 组合数学'].map((tag) => (
                <span key={tag} className="math-tag">{tag}</span>
              ))}
            </div>
          </div>
        </aside>

        <section className="min-w-0 flex-1 space-y-8">
          <div className="page-intro flex flex-col gap-4 p-5 xl:flex-row xl:items-center xl:justify-between">
            <div className="relative w-full xl:max-w-xs">
              <input
                type="text"
                value={query}
                onChange={(e) => setQuery(e.target.value)}
                placeholder="搜索作品、作者或关键词..."
                className="app-input w-full rounded-full py-3 pl-11 pr-4"
              />
              <Search className="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-zinc-400" />
            </div>

            <select
              value={sort}
              onChange={(e) => setSort(e.target.value)}
              className="app-input rounded-xl px-4 py-3 text-primary"
            >
              {sorts.map((s) => (
                <option key={s}>{s}</option>
              ))}
            </select>
          </div>

          {loading ? (
            <div className="text-center text-zinc-500 py-20">加载中...</div>
          ) : sorted.length === 0 ? (
            <div className="text-center text-zinc-500 py-20">暂无作品</div>
          ) : (
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2 xl:grid-cols-3">
              {sorted.map((item) => (
                <article key={item.id} className="glass-card group overflow-hidden rounded-[2rem]">
                  <div className="relative h-52 overflow-hidden">
                    {item.cover_url ? (
                      <img src={item.cover_url} alt={item.title} className="h-full w-full object-cover transition-transform duration-500 group-hover:scale-105" />
                    ) : (
                      <div className="h-full w-full bg-gradient-to-br from-[#e3f2fd] to-[#d5e3fc] flex items-center justify-center text-primary text-2xl font-serif">
                        {item.title?.[0] || '作'}
                      </div>
                    )}
                    {item.field && (
                      <div className="absolute left-4 top-4">
                        <span className="math-tag !bg-white/85">{item.field}</span>
                      </div>
                    )}
                    {item.h5_url && (
                      <div className="absolute right-4 top-4 z-10">
                        <span className="math-tag !bg-amber-500 !text-white flex items-center gap-1 font-semibold shadow-sm">
                          <span className="material-symbols-outlined text-[14px]">play_circle</span>
                          H5 演示
                        </span>
                      </div>
                    )}
                    {item.h5_url && (
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          setActiveH5Url(item.h5_url || null);
                        }}
                        className="absolute inset-0 bg-black/40 opacity-0 group-hover:opacity-100 flex items-center justify-center transition-opacity duration-300 z-10"
                      >
                        <div className="flex items-center gap-2 rounded-full bg-white/90 px-5 py-2.5 text-sm font-semibold text-primary shadow-lg hover:scale-105 transition-all">
                          <span className="material-symbols-outlined text-[18px]">play_circle</span>
                          立即播放演示
                        </div>
                      </button>
                    )}
                  </div>

                  <div className="flex flex-col gap-4 p-6">
                    <div>
                      <h2 className="mb-2 line-clamp-2 font-serif text-xl leading-snug text-primary">{item.title}</h2>
                      <p className="mb-3 text-sm font-semibold text-[#606f78]">作者：{item.author || '匿名'}</p>
                      <p className="line-clamp-3 text-sm font-medium leading-7 text-soft-body">{item.summary}</p>
                    </div>

                    <div className="flex items-center justify-between border-t border-white/50 pt-4">
                      <button
                        onClick={() => openOverlay({
                          title: item.title,
                          subtitle: `${item.author || ''} · ${item.field || ''} · ${item.competition || ''}`,
                          content: (
                            <div className="space-y-5">
                              {item.cover_url && (
                                <div className="aspect-[16/8] overflow-hidden rounded-[1.75rem]">
                                  <img src={item.cover_url} alt={item.title} className="h-full w-full object-cover" />
                                </div>
                              )}
                              {item.h5_url && (
                                <div className="flex items-center justify-between p-4 rounded-2xl bg-amber-500/5 border border-amber-500/20">
                                  <div className="flex items-center gap-3">
                                    <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-amber-500/10 text-amber-600">
                                      <span className="material-symbols-outlined text-[20px]">presentation_play</span>
                                    </div>
                                    <div>
                                      <h4 className="text-sm font-semibold text-zinc-800">在线 H5 幻灯片演示</h4>
                                      <p className="text-xs text-zinc-500">本作品包含交互式幻灯片网页。</p>
                                    </div>
                                  </div>
                                  <button
                                    onClick={() => {
                                      setActiveH5Url(item.h5_url || null);
                                    }}
                                    className="flex items-center gap-1 text-sm font-bold bg-amber-500 hover:bg-amber-600 text-white rounded-full px-4 py-2 transition-all shadow-sm"
                                  >
                                    <span className="material-symbols-outlined text-[16px]">play_circle</span>
                                    全屏播放
                                  </button>
                                </div>
                              )}
                              <div className="rounded-[1.75rem] border border-white/70 bg-[#f7f9fb] p-6 max-h-[60vh] overflow-y-auto">
                                <AsyncMarkdownViewer type="showcases" id={item.id} initialContent={item.summary} />
                              </div>
                            </div>
                          ),
                          actions: [
                            { label: '收藏作品', onClick: () => navigate('portal') },
                            { label: '查看赛事', variant: 'secondary', onClick: () => navigate('events') },
                          ],
                        })}
                        className="flex items-center gap-2 text-sm font-medium text-[#677782]"
                      >
                        <Heart className="w-4 h-4" />
                        {item.view_count || 0}
                      </button>

                      {item.h5_url ? (
                        <button
                          onClick={() => setActiveH5Url(item.h5_url || null)}
                          className="flex items-center gap-1 text-xs font-semibold px-3 py-1.5 rounded-full bg-amber-500/10 text-amber-600 hover:bg-amber-500 hover:text-white transition-all shadow-sm"
                        >
                          <span className="material-symbols-outlined text-[14px]">play_circle</span>
                          演示 PPT
                        </button>
                      ) : (
                        <button
                          onClick={() => openOverlay({
                            title: '作品已归档',
                            subtitle: item.title,
                            content: (
                              <div className="space-y-4">
                                <div className="glass-card rounded-[1.5rem] p-5">
                                  <p className="text-sm font-medium leading-7 text-zinc-600">
                                    该作品已纳入"作品档案馆"归档结构。
                                  </p>
                                </div>
                              </div>
                            ),
                            actions: [{ label: '知道了' }],
                          })}
                          className="text-primary"
                        >
                          <ArrowRight className="w-4 h-4" />
                        </button>
                      )}
                    </div>
                  </div>
                </article>
              ))}
            </div>
          )}
        </section>
      </div>

      {activeH5Url && (
        <div className="fixed inset-0 z-[110] bg-[#1a1c1e]/95 backdrop-blur-md animate-in fade-in duration-300">
          <button
            onClick={() => setActiveH5Url(null)}
            className="absolute top-6 right-6 z-[120] flex h-10 w-10 items-center justify-center rounded-full border border-white/20 bg-black/50 text-white/80 hover:text-white hover:bg-black/75 hover:scale-105 transition-all cursor-pointer shadow-lg"
            aria-label="关闭播放"
          >
            <span className="material-symbols-outlined text-[20px]">close</span>
          </button>
          <iframe
            src={activeH5Url}
            className="w-full h-full border-none"
            title="H5 Showcase Presentation"
            allowFullScreen
          />
        </div>
      )}
    </div>
  );
};