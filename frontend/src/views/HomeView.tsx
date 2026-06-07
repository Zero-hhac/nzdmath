import React, { useEffect, useState } from 'react';
import { motion } from 'motion/react';
import { Formula } from '@/src/components/Formula';
import type { ViewProps } from '@/src/types/app';
import { api } from '@/src/lib/api';

type HomeData = {
  recent_events?: any[];
  featured_events?: any[];
  recent_news?: any[];
  featured_news?: any[];
  featured_resources?: any[];
  recent_showcases?: any[];
};

export const HomeView: React.FC<ViewProps> = ({ navigate, openOverlay }) => {
  const [data, setData] = useState<HomeData | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    setLoading(true);
    api.getHome()
      .then((res) => setData(res.data || null))
      .catch(() => setData(null))
      .finally(() => setLoading(false));
  }, []);

  const featuredEvents = (data?.featured_events && data.featured_events.length > 0)
    ? data.featured_events.slice(0, 3)
    : (data?.recent_events || []).slice(0, 3);
  const recentNews = (data?.recent_news || []).slice(0, 3);
  const featuredResources = (data?.featured_resources || []).slice(0, 3);

  const quickLinks = [
    { title: '近期活动', desc: '查看本周讲座、工作坊与竞赛安排。', action: '前往活动中心', tab: 'events' as const, icon: 'event' },
    { title: '重点资源', desc: '快速进入讲义、模板、课程笔记与档案。', action: '浏览资源库', tab: 'resources' as const, icon: 'library_books' },
    { title: '会员服务', desc: '解锁内部存档、导师预约与协作入口。', action: '进入会员专区', tab: 'portal' as const, icon: 'workspace_premium' },
    { title: '作品档案', desc: '查看竞赛作品、可视化成果与优秀手稿。', action: '打开档案馆', tab: 'showcase' as const, icon: 'gallery_thumbnail' },
  ];

  const featuredEntries = [];
  if (featuredEvents.length > 0) {
    featuredEntries.push({ label: '重点活动', title: featuredEvents[0].title, desc: featuredEvents[0].summary || '了解更多活动详情...', tab: 'events' as const, kicker: '活动' });
  }
  if (recentNews.length > 0) {
    featuredEntries.push({ label: '最新资讯', title: recentNews[0].title, desc: recentNews[0].summary || '了解更多资讯...', tab: 'news' as const, kicker: '动态' });
  }
  if (featuredResources.length > 0) {
    featuredEntries.push({ label: '资源推荐', title: featuredResources[0].title, desc: featuredResources[0].summary || '探索更多文档与代码...', tab: 'resources' as const, kicker: '文档' });
  }
  if (featuredEntries.length === 0) {
    featuredEntries.push({ label: '暂无推荐', title: '暂无置顶内容', desc: '管理员尚未发布任何推荐信息。', tab: 'events' as const, kicker: '提示' });
  }

  const curatedEntries = [
    {
      title: '新成员路径',
      desc: '从零开始认识协会、了解入会流程、规划你的第一条学习路径。',
      action: '查看加入流程',
      onClick: () =>
        openOverlay({
          title: '加入数爱会',
          subtitle: '新成员引导',
          content: (
            <div className="grid gap-5 md:grid-cols-3">
              {[
                ['01', '提交意向', '填写基础信息与感兴趣方向，我们会匹配最适合你的活动组。'],
                ['02', '参加迎新沙龙', '通过一次开放沙龙了解协会节奏、资源和导师体系。'],
                ['03', '解锁会员权限', '完成审核后进入会员专区，获得内部资源、活动预约和协作入口。'],
              ].map(([index, label, text]) => (
                <div key={index} className="bento-card !p-6">
                  <div className="mb-3 text-xs font-bold uppercase tracking-[0.24em] text-text-muted">{index}</div>
                  <h4 className="mb-2 text-xl text-charcoal font-medium tracking-tight">{label}</h4>
                  <p className="text-sm font-medium leading-7 text-text-muted">{text}</p>
                </div>
              ))}
            </div>
          ),
          actions: [
            { label: '前往会员专区', onClick: () => navigate('portal') },
            { label: '稍后再看', variant: 'secondary' },
          ],
        }),
    },
    {
      title: '本周推荐',
      desc: '把近期最值得参加的活动、最热门的讨论和最新公告合并到一个工作台里。',
      action: '打开本周推荐',
      onClick: () =>
        openOverlay({
          title: '本周推荐',
          subtitle: '精选动态',
          content: (
            <div className="grid gap-4 md:grid-cols-2">
              {(data?.recent_events || []).slice(0, 2).map((ev: any) => (
                <div key={'ev' + ev.id} className="bento-card !p-5">
                  <div className="mb-2 text-xs font-bold uppercase tracking-[0.2em] text-text-muted">最新活动</div>
                  <p className="text-sm font-medium leading-7 text-charcoal-muted line-clamp-2">{ev.title}</p>
                </div>
              ))}
              {(data?.recent_news || []).slice(0, 2).map((nw: any) => (
                <div key={'nw' + nw.id} className="bento-card !p-5">
                  <div className="mb-2 text-xs font-bold uppercase tracking-[0.2em] text-text-muted">最新资讯</div>
                  <p className="text-sm font-medium leading-7 text-charcoal-muted line-clamp-2">{nw.title}</p>
                </div>
              ))}
            </div>
          ),
          actions: [{ label: '查看动态资讯', onClick: () => navigate('news') }],
        }),
    },
  ];

  return (
    <div className="space-y-32 py-12">
      {/* Hero Section */}
      <section className="grid lg:grid-cols-12 gap-12 lg:gap-8 items-center min-h-[70vh]">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          className="lg:col-span-7 flex flex-col justify-center"
        >
          <div className="inline-flex items-center px-3 py-1 bg-pastel-blue text-pastel-blue-text text-[11px] uppercase tracking-[0.15em] font-medium rounded-full w-fit mb-8">
            官方门户
          </div>

          <h1 className="display-title mb-8">
            探索无限可能
          </h1>

          <p className="text-[1.1rem] leading-[1.8] text-charcoal-muted max-w-xl font-medium mb-10">
            广西农业职业技术大学数学爱好者协会（简称数爱会）是交流数学史、数学文化、数学建模、数学竞赛和考研备战经验的学习型社团。我们以“交流思想、提高能力、团队协作、开拓创新”为宗旨，汇聚同路人共同成长。
          </p>

          <div className="flex flex-wrap gap-4">
            <button onClick={() => navigate('events')} className="btn-primary group">
              查看近期活动
              <div className="btn-icon-wrapper">
                <span className="material-symbols-outlined text-[16px] text-charcoal">arrow_forward</span>
              </div>
            </button>
            <button onClick={() => navigate('resources')} className="btn-secondary">
              浏览重点资源
            </button>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, x: 24 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.8, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
          className="lg:col-span-5 relative w-full"
        >
           <div className="doppel-shell w-full min-h-[360px] flex items-center justify-center cursor-default">
             <div className="doppel-core w-full h-full flex flex-col items-center justify-center p-8 gap-8 relative overflow-hidden">
                <div className="absolute top-6 left-6 text-text-muted opacity-30">
                  <Formula expression={String.raw`\int e^{-x^2}\,dx`} className="text-xl" />
                </div>
                <div className="absolute bottom-6 right-6 text-text-muted opacity-30">
                  <Formula expression={String.raw`e^{i\pi}+1=0`} className="text-xl" />
                </div>
                
                <div className="text-center z-10">
                  <div className="text-5xl font-light text-charcoal mb-2 tracking-tight">500+</div>
                  <div className="text-[10px] uppercase tracking-[0.2em] text-text-muted font-bold">活跃成员</div>
                </div>
                
                <div className="w-16 h-px bg-border"></div>
                
                <div className="text-center z-10">
                  <div className="text-4xl font-light text-charcoal mb-2 tracking-tight">200+</div>
                  <div className="text-[10px] uppercase tracking-[0.2em] text-text-muted font-bold">学术资源</div>
                </div>
             </div>
           </div>
        </motion.div>
      </section>

      {/* Quick Access Bento */}
      <motion.section 
        initial={{ opacity: 0, y: 24 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
        className="space-y-12"
      >
        <div className="flex flex-col gap-3 md:flex-row md:items-end md:justify-between">
          <div>
            <h2 className="section-title">官网入口导览</h2>
            <p className="mt-4 text-charcoal-muted leading-relaxed font-medium">一眼看懂站内能做什么、下一步该去哪。</p>
          </div>
        </div>

        <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-4">
          {quickLinks.map((item, i) => (
            <motion.button
              key={item.title}
              onClick={() => navigate(item.tab)}
              initial={{ opacity: 0, y: 12 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ duration: 0.5, delay: i * 0.1, ease: [0.16, 1, 0.3, 1] }}
              className="bento-card text-left group"
            >
              <div className="mb-6 w-10 h-10 rounded-lg bg-black/[0.04] flex items-center justify-center text-charcoal">
                <span className="material-symbols-outlined text-[20px]">{item.icon}</span>
              </div>
              <h3 className="text-xl font-medium tracking-tight text-charcoal mb-3">{item.title}</h3>
              <p className="text-[13px] leading-relaxed text-text-muted">{item.desc}</p>
              <div className="mt-8 inline-flex items-center gap-2 text-[13px] font-bold text-charcoal group-hover:text-pastel-blue-text transition-colors">
                {item.action}
                <span className="material-symbols-outlined text-[16px] transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5">north_east</span>
              </div>
            </motion.button>
          ))}
        </div>
      </motion.section>

      {/* Featured Content Asymmetrical Split */}
      <motion.section 
        initial={{ opacity: 0, y: 24 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
        className="grid gap-6 lg:grid-cols-12 items-start"
      >
        <div className="lg:col-span-7 bento-card">
          <h2 className="text-3xl font-medium tracking-tight text-charcoal mb-8">近期重点</h2>
          <div className="space-y-6">
            {featuredEntries.map((item) => (
              <button
                key={item.title}
                onClick={() => navigate(item.tab)}
                className="w-full text-left group border-b border-border last:border-0 pb-6 last:pb-0"
              >
                <div className="flex items-start justify-between">
                  <div>
                    <div className="mb-2 text-[10px] font-bold uppercase tracking-[0.2em] text-text-muted">{item.kicker}</div>
                    <h3 className="text-xl font-medium text-charcoal mb-2 group-hover:text-pastel-blue-text transition-colors">{item.title}</h3>
                    <p className="text-sm text-text-muted leading-relaxed">{item.desc}</p>
                  </div>
                  <div className="w-8 h-8 rounded-full bg-black/[0.02] flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity">
                    <span className="material-symbols-outlined text-[16px]">arrow_forward</span>
                  </div>
                </div>
              </button>
            ))}
          </div>
        </div>

        <div className="lg:col-span-5 space-y-6">
          {curatedEntries.map((item) => (
            <div key={item.title} className="bento-card !p-8">
              <h3 className="text-xl font-medium tracking-tight text-charcoal mb-3">{item.title}</h3>
              <p className="text-[13px] leading-relaxed text-text-muted mb-6">{item.desc}</p>
              <button onClick={item.onClick} className="btn-secondary !text-xs !py-2 !px-4">
                {item.action}
              </button>
            </div>
          ))}
        </div>
      </motion.section>

      {/* Core Dimensions */}
      <motion.section 
        initial={{ opacity: 0, y: 24 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, amount: 0.2 }}
        transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
        className="space-y-12"
      >
        <div className="text-center max-w-2xl mx-auto">
          <h2 className="section-title mb-4">核心维度</h2>
          <p className="text-charcoal-muted leading-relaxed font-medium">融合学术严谨与现代连接，让内容、活动与协作彼此支撑。</p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="bento-card md:col-span-2 flex flex-col justify-between">
            <div>
              <div className="w-10 h-10 rounded-lg bg-pastel-blue text-pastel-blue-text flex items-center justify-center mb-6">
                <span className="material-symbols-outlined text-[20px]" style={{ fontVariationSettings: "'FILL' 1" }}>auto_awesome</span>
              </div>
              <h3 className="text-2xl font-medium tracking-tight text-charcoal mb-4">学术卓越</h3>
              <p className="text-charcoal-muted leading-relaxed max-w-md text-sm">
                提供深度理论探讨空间，涵盖拓扑学、数论与现代代数。我们致力于推动纯粹数学的边界，为学者提供一流的研讨平台。
              </p>
            </div>
            <div className="flex gap-2 pt-8 mt-auto">
              <span className="tag-kinpaku">拓扑学</span>
              <span className="tag-kinpaku">泛函分析</span>
              <span className="tag-kinpaku">代数拓补</span>
            </div>
          </div>

          <div className="bento-card flex flex-col justify-between">
            <div>
              <div className="w-10 h-10 rounded-lg bg-pastel-green text-pastel-green-text flex items-center justify-center mb-6">
                <span className="material-symbols-outlined text-[20px]" style={{ fontVariationSettings: "'FILL' 1" }}>groups</span>
              </div>
              <h3 className="text-2xl font-medium tracking-tight text-charcoal mb-4">社区连接</h3>
              <p className="text-charcoal-muted leading-relaxed text-sm">
                打破孤岛，建立跨学科的对话机制。定期的学术沙龙与在线工作坊，让思想在这里碰撞出新的火花。
              </p>
            </div>
          </div>
        </div>
      </motion.section>
    </div>
  );
};
