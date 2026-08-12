import React, { useEffect, useMemo, useState } from 'react';
import { motion } from 'motion/react';
import {
  ArrowRight,
  ArrowUpRight,
  Award,
  BookMarked,
  CalendarDays,
  GalleryVerticalEnd,
  GraduationCap,
  Info,
  Sparkles,
  Trophy,
  Users,
} from 'lucide-react';
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

type EventItem = {
  id: number | string;
  title: string;
  summary?: string;
  category?: string;
  location?: string;
  start_time?: string;
  end_time?: string;
  cover_url?: string;
  status?: number;
  is_featured?: boolean;
};

const WEEKDAYS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

function formatEventDate(iso?: string) {
  if (!iso) return { date: '—', time: '—', weekday: '', monthDay: '' };
  const d = new Date(iso);
  if (isNaN(d.getTime())) return { date: '—', time: '—', weekday: '', monthDay: '' };
  const pad = (n: number) => String(n).padStart(2, '0');
  return {
    monthDay: `${d.getMonth() + 1}月${d.getDate()}日`,
    weekday: WEEKDAYS[d.getDay()],
    time: `${pad(d.getHours())}:${pad(d.getMinutes())}`,
  };
}

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

  const nextEvent: EventItem | null = useMemo(() => {
    if (!data) return null;
    const now = Date.now();
    const pool: EventItem[] = [
      ...(data.featured_events || []),
      ...(data.recent_events || []),
    ];
    const upcoming = pool
      .filter((e) => e?.start_time && new Date(e.start_time).getTime() >= now)
      .sort((a, b) => new Date(a.start_time!).getTime() - new Date(b.start_time!).getTime());
    if (upcoming.length > 0) return upcoming[0];
    return pool[0] || null;
  }, [data]);

  const quickLinks = [
    { title: '近期活动', desc: '查看本周讲座、工作坊与竞赛安排。', action: '前往活动中心', tab: 'events' as const, icon: 'event', badge: '新人起点' as (string | null), accent: true, tone: 'blue' },
    { title: '重点资源', desc: '快速进入讲义、模板、课程笔记与档案。', action: '浏览资源库', tab: 'resources' as const, icon: 'library_books', badge: null, accent: false, tone: 'green' },
    { title: '会员服务', desc: '解锁内部存档、导师预约与协作入口。', action: '进入会员专区', tab: 'portal' as const, icon: 'workspace_premium', badge: null, accent: false, tone: 'red' },
    { title: '作品档案', desc: '查看竞赛作品、可视化成果与优秀手稿。', action: '打开档案馆', tab: 'showcase' as const, icon: 'gallery_thumbnail', badge: null, accent: false, tone: 'gold' },
  ];

  const quickTones: Record<string, { card: string; icon: string }> = {
    blue: { card: '!border-pastel-blue !bg-pastel-blue/25 hover:!bg-pastel-blue/40', icon: 'bg-pastel-blue text-pastel-blue-text' },
    green: { card: '!border-pastel-green !bg-pastel-green/25 hover:!bg-pastel-green/40', icon: 'bg-pastel-green text-pastel-green-text' },
    red: { card: '!border-pastel-red !bg-pastel-red/20 hover:!bg-pastel-red/35', icon: 'bg-pastel-red text-pastel-red-text' },
    gold: { card: '!border-[#f2e1b3] !bg-[#fcf8ed] hover:!bg-[#faf0d7]', icon: 'bg-[#fcf8ed] text-[#8c6d14]' },
  };

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
      steps: [
        { index: '01', label: '提交意向', text: '填写基础信息与感兴趣方向。' },
        { index: '02', label: '参加迎新沙龙', text: '了解协会节奏与导师体系。' },
        { index: '03', label: '解锁会员权限', text: '获得内部资源与协作入口。' },
      ],
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
      desc: '把近期最值得参加的活动、最热门的讨论和最新公告合并到一处。',
      action: '打开本周推荐',
      picks: [
        ...((data?.recent_events || []).slice(0, 2).map((ev: any) => ({
          kind: '活动' as const,
          title: ev.title,
        }))),
        ...((data?.recent_news || []).slice(0, 2).map((nw: any) => ({
          kind: '资讯' as const,
          title: nw.title,
        }))),
      ],
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
      <section className="grid lg:grid-cols-12 gap-10 lg:gap-8 items-center min-h-[70vh]">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
          className="lg:col-span-7 flex flex-col justify-center"
        >
          <div className="eyebrow-accent mb-6">
            <span className="w-1.5 h-1.5 rounded-full bg-accent" />
            广西农业职业技术大学 · 数爱会
          </div>

          <h1 className="display-title mb-7">
            一起学数学，<br className="hidden md:block" />
            比一个人<span className="text-accent">走得更远</span>。
          </h1>

          <p className="text-[1.05rem] leading-[1.85] text-charcoal-muted max-w-xl font-medium mb-10">
            我们是校园里一群喜欢数学的人——讲拓扑、聊数论、组队建模、刷考研真题。每周都有活动、每年都有竞赛、随时都能找到同路人。
          </p>

          <div className="flex flex-wrap gap-3">
            <button onClick={() => navigate('events')} className="btn-accent group">
              <CalendarDays className="h-[18px] w-[18px]" />
              查看近期活动
              <div className="btn-icon-wrapper !bg-white/15">
                <ArrowRight className="h-4 w-4 text-white" />
              </div>
            </button>
            <button onClick={() => navigate('about')} className="btn-secondary">
              <Info className="h-4 w-4" />
              了解我们
            </button>
          </div>
        </motion.div>

        <motion.div
          initial={{ opacity: 0, x: 24 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.8, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
          className="lg:col-span-5 relative w-full"
        >
          <div className="bento-card !p-7 relative overflow-hidden">
            <div className="flex items-center justify-between mb-5">
              <span className="eyebrow-accent !bg-accent !text-white !text-[10px]">
                <span className="w-1.5 h-1.5 rounded-full bg-white animate-pulse" />
                下一场活动
              </span>
              {nextEvent?.category && (
                <span className="text-[10px] uppercase tracking-[0.18em] text-text-muted font-bold">
                  {nextEvent.category}
                </span>
              )}
            </div>

            {nextEvent ? (
              <>
                <h3 className="text-2xl font-medium tracking-tight text-charcoal leading-snug mb-4 line-clamp-2">
                  {nextEvent.title}
                </h3>
                {nextEvent.summary && (
                  <p className="text-sm text-charcoal-muted leading-relaxed mb-6 line-clamp-2">
                    {nextEvent.summary}
                  </p>
                )}

                <div className="flex items-end gap-6 pt-5 border-t border-border">
                  {(() => {
                    const f = formatEventDate(nextEvent.start_time);
                    return (
                      <div>
                        <div className="text-[10px] uppercase tracking-[0.2em] text-text-muted font-bold mb-1">时间</div>
                        <div className="text-xl font-semibold text-accent tracking-tight">
                          {f.monthDay}
                          <span className="ml-2 text-sm font-medium text-text-muted">{f.weekday}</span>
                        </div>
                        <div className="text-sm text-charcoal-muted mt-0.5">{f.time} 始</div>
                      </div>
                    );
                  })()}
                  {nextEvent.location && (
                    <div className="flex-1 min-w-0">
                      <div className="text-[10px] uppercase tracking-[0.2em] text-text-muted font-bold mb-1">地点</div>
                      <div className="text-sm font-medium text-charcoal truncate">{nextEvent.location}</div>
                    </div>
                  )}
                </div>

                <button
                  onClick={() => navigate('events')}
                  className="mt-6 w-full btn-ghost justify-center"
                >
                  查看活动详情
                  <ArrowRight className="h-4 w-4" />
                </button>
              </>
            ) : (
              <div className="py-10 text-center">
                <div className="text-text-muted text-sm">暂无安排中的活动，去活动中心看看？</div>
                <button onClick={() => navigate('events')} className="btn-ghost mt-4">
                  浏览活动中心
                </button>
              </div>
            )}
          </div>

          <div className="grid grid-cols-2 gap-3 mt-4">
            <div className="bento-card !p-4 text-center !border-pastel-blue !bg-pastel-blue/20">
              <div className="text-2xl font-semibold text-charcoal tracking-tight">2022</div>
              <div className="text-[10px] uppercase tracking-[0.18em] text-text-muted font-bold mt-1">协会成立</div>
            </div>
            <div className="bento-card !p-4 text-center !border-pastel-green !bg-pastel-green/20">
              <div className="text-2xl font-semibold text-charcoal tracking-tight">4</div>
              <div className="text-[10px] uppercase tracking-[0.18em] text-text-muted font-bold mt-1">核心职能部门</div>
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
              className={`bento-card text-left group relative ${quickTones[item.tone]?.card || ''}`}
            >
              {item.badge && (
                  <span className="absolute top-5 right-5 inline-flex items-center gap-1 rounded-full bg-accent text-white px-2.5 py-0.5 text-[10px] font-semibold tracking-wide">
                  <Sparkles className="h-3 w-3" />
                  {item.badge}
                </span>
              )}
              <div className={`mb-6 w-10 h-10 rounded-lg flex items-center justify-center ${quickTones[item.tone]?.icon || 'bg-black/[0.04] text-charcoal'}`}>
                {item.icon === 'event' && <CalendarDays className="h-5 w-5" />}
                {item.icon === 'library_books' && <BookMarked className="h-5 w-5" />}
                {item.icon === 'workspace_premium' && <Award className="h-5 w-5" />}
                {item.icon === 'gallery_thumbnail' && <GalleryVerticalEnd className="h-5 w-5" />}
              </div>
              <h3 className="text-xl font-medium tracking-tight text-charcoal mb-3">{item.title}</h3>
              <p className="text-[13px] leading-relaxed text-text-muted">{item.desc}</p>
              <div className="mt-8 inline-flex items-center gap-2 text-[13px] font-bold text-charcoal group-hover:text-accent transition-colors">
                {item.action}
                <ArrowUpRight className="h-4 w-4 transition-transform group-hover:translate-x-0.5 group-hover:-translate-y-0.5" />
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
                    <ArrowRight className="h-4 w-4" />
                  </div>
                </div>
              </button>
            ))}
          </div>
        </div>

        <div className="lg:col-span-5 space-y-6">
          {curatedEntries.map((item: any) => (
            <div key={item.title} className="bento-card !p-7">
              <h3 className="text-lg font-medium tracking-tight text-charcoal mb-2">{item.title}</h3>
              <p className="text-[13px] leading-relaxed text-text-muted mb-5">{item.desc}</p>

              {item.steps && (
                <div className="space-y-2.5 mb-5">
                  {item.steps.map((s: any) => (
                    <div key={s.index} className="flex items-start gap-3 text-sm">
                      <span className="shrink-0 w-7 h-7 rounded-full bg-accent-soft text-accent text-xs font-bold flex items-center justify-center tracking-wider">
                        {s.index}
                      </span>
                      <div className="pt-0.5">
                        <span className="font-semibold text-charcoal">{s.label}</span>
                        <span className="text-text-muted ml-1.5">{s.text}</span>
                      </div>
                    </div>
                  ))}
                </div>
              )}

              {item.picks && item.picks.length > 0 && (
                <ul className="space-y-2 mb-5">
                  {item.picks.map((p: any, i: number) => (
                    <li key={i} className="flex items-start gap-2.5 text-sm">
                      <span className={`shrink-0 px-1.5 py-0.5 rounded text-[10px] font-bold tracking-wide ${p.kind === '活动' ? 'bg-pastel-blue text-pastel-blue-text' : 'bg-pastel-green text-pastel-green-text'}`}>
                        {p.kind}
                      </span>
                      <span className="text-charcoal-muted line-clamp-1 flex-1">{p.title}</span>
                    </li>
                  ))}
                </ul>
              )}

              <button onClick={item.onClick} className="btn-ghost !text-xs !py-2 !px-4 mt-2">
                {item.action}
                <ArrowRight className="h-3.5 w-3.5" />
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

        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="bento-card flex flex-col justify-between !p-8 md:!p-10">
            <div>
              <div className="w-11 h-11 rounded-xl bg-accent-soft text-accent flex items-center justify-center mb-6">
                <Sparkles className="h-[22px] w-[22px]" />
              </div>
              <h3 className="text-2xl font-medium tracking-tight text-charcoal mb-3">学术卓越</h3>
              <p className="text-charcoal-muted leading-relaxed text-sm">
                提供深度理论探讨空间，涵盖拓扑学、数论与现代代数。每两周一次的研讨会与读书会，是纯粹数学爱好者的主场。
              </p>
            </div>
            <div className="flex flex-wrap gap-1.5 pt-6 mt-4 border-t border-border">
              <span className="tag-kinpaku">拓扑学</span>
              <span className="tag-kinpaku">泛函分析</span>
              <span className="tag-kinpaku">代数几何</span>
            </div>
          </div>

          <div className="bento-card flex flex-col justify-between !p-8 md:!p-10">
            <div>
              <div className="w-11 h-11 rounded-xl bg-pastel-green text-pastel-green-text flex items-center justify-center mb-6">
                <Users className="h-[22px] w-[22px]" />
              </div>
              <h3 className="text-2xl font-medium tracking-tight text-charcoal mb-3">社区连接</h3>
              <p className="text-charcoal-muted leading-relaxed text-sm">
                打破孤岛，建立跨学科的对话机制。定期的学术沙龙、在线工作坊与晚间讨论室，让想法在这里碰撞。
              </p>
            </div>
            <div className="pt-6 mt-4 border-t border-border text-xs text-text-muted">
              <span className="font-semibold text-charcoal">每周</span> · 沙龙 · 读书会 · 跨组讨论
            </div>
          </div>

          <div className="bento-card flex flex-col justify-between !p-8 md:!p-10">
            <div>
              <div className="w-11 h-11 rounded-xl bg-pastel-red text-pastel-red-text flex items-center justify-center mb-6">
                <Trophy className="h-[22px] w-[22px]" />
              </div>
              <h3 className="text-2xl font-medium tracking-tight text-charcoal mb-3">竞赛战绩</h3>
              <p className="text-charcoal-muted leading-relaxed text-sm">
                数学建模、国赛、考研数学——历年都有学长学姐带队备赛。我们整理了历年真题、参考解答与内部培训讲义。
              </p>
            </div>
            <div className="pt-6 mt-4 border-t border-border text-xs text-text-muted">
              <span className="font-semibold text-charcoal">每年</span> · 校赛 · 区域赛 · 国赛
            </div>
          </div>

          <div className="bento-card flex flex-col justify-between !p-8 md:!p-10">
            <div>
              <div className="w-11 h-11 rounded-xl bg-pastel-blue text-pastel-blue-text flex items-center justify-center mb-6">
                <GraduationCap className="h-[22px] w-[22px]" />
              </div>
              <h3 className="text-2xl font-medium tracking-tight text-charcoal mb-3">会员成长</h3>
              <p className="text-charcoal-muted leading-relaxed text-sm">
                注册会员可解锁内部存档、导师预约与小组协作入口。我们用一条清晰的学习路径，陪每位新同学走完第一年。
              </p>
            </div>
            <button
              onClick={() => navigate('portal')}
              className="self-start pt-6 mt-4 border-t border-border w-full text-left text-xs text-text-muted hover:text-accent transition-colors flex items-center justify-between"
            >
              <span><span className="font-semibold text-charcoal">会员专享</span> · 内部资源 · 导师 1v1</span>
              <ArrowRight className="h-4 w-4" />
            </button>
          </div>
        </div>
      </motion.section>
    </div>
  );
};
