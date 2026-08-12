import React, { useEffect, useMemo, useState } from 'react';
import { Search, Download, Folder, FileText, Edit, Code, Save, Eye } from 'lucide-react';
import { DocxViewer } from '@/src/components/DocxViewer';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';
import { useAuth } from '@/src/lib/auth';
import { useToast } from '@/src/lib/toast';

type ResourceItem = {
  id: number;
  title: string;
  summary?: string;
  category?: string;
  file_name?: string;
  file_size?: number;
  file_type?: string;
  file_ext?: string;
  cover_url?: string;
  view_count?: number;
  download_count?: number;
  like_count?: number;
  is_featured?: boolean;
};

const categories = ['全部资源', '课程笔记', '竞赛讲义', '学术模板', '开源代码'];

function formatSize(bytes?: number) {
  if (!bytes) return '—';
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

export const ResourcesView: React.FC<ViewProps> = ({ navigate, openOverlay }) => {
  const { user } = useAuth();
  const { showToast } = useToast();
  const [resources, setResources] = useState<ResourceItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [query, setQuery] = useState('');
  const [activeCategory, setActiveCategory] = useState(categories[0]);

  useEffect(() => {
    setLoading(true);
    api.getResources()
      .then((res) => setResources(res.data || []))
      .catch(() => setResources([]))
      .finally(() => setLoading(false));
  }, []);

  const filtered = useMemo(() => {
    return resources.filter((r) => {
      const matchesCat = activeCategory === '全部资源' ||
        (activeCategory === '课程笔记' && r.category === '课程笔记') ||
        (activeCategory === '竞赛讲义' && r.category === '竞赛讲义') ||
        (activeCategory === '学术模板' && r.category === '学术模板') ||
        (activeCategory === '开源代码' && r.category === '开源代码');
      const matchesQ = [r.title, r.category, r.file_name].filter(Boolean).join(' ').toLowerCase().includes(query.toLowerCase());
      return matchesCat && matchesQ;
    });
  }, [resources, activeCategory, query]);

  const handleDownload = (r: ResourceItem) => {
    if (!user) {
      showToast('请先登录后再下载', 'info');
      return;
    }
    window.open(api.downloadResource(r.id), '_blank');
    showToast('开始下载：' + r.title, 'success');
  };

  return (
    <div className="space-y-12">
      <div className="page-intro flex flex-col md:flex-row justify-between items-center gap-8">
        <div className="space-y-1 text-center md:text-left">
          <div className="section-kicker">Resources</div>
          <h2 className="section-title">资源存档</h2>
          <p className="section-subtitle">知识的对撞与沉淀</p>
        </div>
        <div className="w-full md:w-96 relative">
          <input
            type="text"
            placeholder="搜索文档、模板或笔记..."
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            className="app-input w-full rounded-full py-3 pl-12 pr-4"
          />
          <Search className="w-4 h-4 absolute left-4 top-1/2 -translate-y-1/2 text-zinc-400" />
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        <div className="lg:col-span-1 space-y-6">
          <div className="sidebar-panel space-y-6">
            <h4 className="font-serif text-lg text-primary">分类浏览</h4>
            <div className="space-y-2">
              {categories.map((item) => (
                <button
                  key={item}
                  onClick={() => setActiveCategory(item)}
                  className={`w-full text-left px-4 py-2 rounded-xl text-sm font-medium transition-all ${
                    activeCategory === item ? 'bg-accent-soft text-accent' : 'text-text-muted hover:bg-canvas-alt hover:text-charcoal'
                  }`}
                >
                  {item}
                </button>
              ))}
            </div>
          </div>

          <div className="sidebar-panel bg-gradient-to-br from-white to-accent-soft/70">
            <div className="mb-4 flex items-center gap-3">
              <Folder className="w-5 h-5 text-primary" />
              <h4 className="font-serif text-lg text-primary">作品档案馆</h4>
            </div>
            <p className="mb-4 text-xs font-medium leading-relaxed text-soft-body">
              集中展示竞赛作品、可视化成果与优秀手稿，记录协会成员的探索过程。
            </p>
            <button
              onClick={() => navigate('showcase')}
              className="btn-primary w-full !py-2.5 !text-xs !font-bold"
            >
              进入档案馆
            </button>
          </div>
        </div>

        <div className="lg:col-span-3">
          {loading ? (
            <div className="text-center text-zinc-500 py-20">加载中...</div>
          ) : filtered.length === 0 ? (
            <div className="text-center text-zinc-500 py-20">暂无资源</div>
          ) : (
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              {filtered.map((r) => {
                const ext = r.file_ext?.toLowerCase() || '';
                const Icon = ext === '.pdf' ? FileText : ext === '.zip' ? Folder : ext === '.py' || ext === '.js' ? Code : Edit;
                return (
                  <div key={r.id} className="glass-card rounded-3xl p-6 flex items-start gap-5 group hover:bg-white transition-colors duration-500">
                    <div className="w-12 h-12 sm:w-14 sm:h-14 shrink-0 rounded-2xl bg-accent-soft flex items-center justify-center text-accent border border-accent/10 group-hover:scale-105 transition-transform">
                      <Icon className="w-7 h-7" />
                    </div>
                    <div className="flex-grow space-y-2">
                      <div className="flex justify-between items-start">
                        <span className="text-[10px] font-bold text-zinc-400 uppercase tracking-widest">{r.category || '资源'}</span>
                        <span className="text-[10px] font-bold text-accent bg-accent-soft px-2 py-0.5 rounded-md">{r.file_ext || '资料'}</span>
                      </div>
                      <h3 className="text-base font-medium tracking-tight text-charcoal leading-snug group-hover:text-accent transition-colors">{r.title}</h3>
                      {r.summary && <p className="line-clamp-2 text-xs leading-relaxed text-text-muted">{r.summary}</p>}
                      <div className="flex items-center justify-between pt-2">
                        <div className="flex items-center gap-4 text-[10px] text-text-muted font-medium">
                          <span className="flex items-center gap-1">
                            <Save className="w-3 h-3" /> {formatSize(r.file_size)}
                          </span>
                          <span className="flex items-center gap-1">
                            <Download className="w-3 h-3" /> {r.download_count || 0}
                          </span>
                        </div>
                        <div className="flex gap-1">
                          {(ext === '.docx' || ext === '.doc') && (
                            <button
                              onClick={() => {
                                openOverlay({
                                  title: r.title,
                                  subtitle: '文档预览',
                                  content: <div className="rounded-[1.75rem] border border-border bg-canvas-alt p-6 max-h-[70vh] overflow-y-auto"><DocxViewer fileUrl={api.downloadResource(r.id)} /></div>,
                                  actions: [{ label: '下载源文件', onClick: () => handleDownload(r) }, { label: '关闭', variant: 'secondary' }]
                                });
                              }}
                              className="btn-ghost hover:scale-110 transition-transform flex items-center justify-center p-2"
                              title="预览"
                              aria-label="预览文档"
                            >
                              <Eye className="w-4 h-4" />
                            </button>
                          )}
                          <button
                            onClick={() => handleDownload(r)}
                            className="btn-ghost hover:scale-110 transition-transform flex items-center justify-center p-2"
                            title="下载"
                            aria-label="下载资源"
                          >
                            <Download className="w-4 h-4" />
                          </button>
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
