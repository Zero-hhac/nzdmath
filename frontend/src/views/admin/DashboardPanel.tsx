import { useEffect, useState } from 'react';
import {
  AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  BarChart, Bar,
} from 'recharts';
import {
  Eye, User, Users, BookOpen, Activity, Calendar, Newspaper, RefreshCw,
} from 'lucide-react';
import { api } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';
import type { SubView } from './types';

export function DashboardPanel({ onSwitch }: { onSwitch: (s: SubView) => void }) {
  const { showToast } = useToast();
  const [stats, setStats] = useState<any>({});

  useEffect(() => {
    api.adminDashboard()
      .then((res) => setStats(res.data || {}))
      .catch(() => setStats({}));
  }, []);

  const c = stats.counts || {};
  const trend = stats.trend_7days || { dates: [], events: [], news: [] };
  const todayAct = stats.today_activity || { pv: 0, uv: 0, dau: 0 };
  const activity = stats.activity_trend || { dates: [], pv: [], uv: [], dau: [] };
  const totalActivity = stats.total_activity || { pv: 0, uv: 0 };

  const handleInvalidate = async () => {
    try {
      await api.adminInvalidateHomepage();
      showToast('首页缓存已刷新', 'success');
    } catch (err: any) {
      showToast(err.message, 'error');
    }
  };

  return (
    <>
      {/* 基础数据网格 */}
      <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
        {[
          { label: '用户总数', value: c.users || 0, icon: User, color: '#e3f2fd' },
          { label: '活动总数', value: c.events || 0, icon: Calendar, color: '#d5e3fc' },
          { label: '资讯总数', value: c.news || 0, icon: Newspaper, color: '#FFF9E5' },
          { label: '资源总数', value: c.resources || 0, icon: BookOpen, color: '#f7f9fb' },
        ].map((s, i) => (
          <div key={i} className="sidebar-panel rounded-[2rem] p-6 space-y-4">
            <div className="flex justify-between items-start">
              <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ backgroundColor: s.color }}>
                <s.icon className="w-5 h-5 text-primary" />
              </div>
            </div>
            <div>
              <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest">{s.label}</div>
              <div className="text-3xl font-serif text-primary mt-1">{s.value}</div>
            </div>
          </div>
        ))}
      </div>

      {/* 流量统计网格 */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {[
          { label: '今日浏览量 (PV)', value: todayAct.pv || 0, icon: Eye, color: '#e0f2fe' },
          { label: '今日独立访客 (UV)', value: todayAct.uv || 0, icon: Activity, color: '#dcfce7' },
          { label: '今日活跃会员 (DAU)', value: todayAct.dau || 0, icon: Users, color: '#fef9c3' },
        ].map((s, i) => (
          <div key={i} className="sidebar-panel rounded-[2rem] p-6 space-y-4">
            <div className="flex justify-between items-start">
              <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ backgroundColor: s.color }}>
                <s.icon className="w-5 h-5 text-primary" />
              </div>
            </div>
            <div>
              <div className="text-[10px] text-zinc-400 font-bold uppercase tracking-widest">{s.label}</div>
              <div className="text-3xl font-serif text-primary mt-1">{s.value}</div>
            </div>
          </div>
        ))}
      </div>

      {/* 全站累计访问数据 */}
      <div className="grid grid-cols-1 lg:grid-cols-[0.9fr_1.1fr] gap-6">
        <div className="grid grid-cols-2 gap-4">
          {[
            { label: '总浏览量 (PV)', value: totalActivity.pv || 0, icon: Eye, color: '#e1f3fe' },
            { label: '总独立访客 (UV)', value: totalActivity.uv || 0, icon: Users, color: '#edf3ec' },
          ].map((s) => (
            <div key={s.label} className="sidebar-panel rounded-[2rem] p-5 space-y-4">
              <div className="w-10 h-10 rounded-xl flex items-center justify-center" style={{ backgroundColor: s.color }}>
                <s.icon className="w-5 h-5 text-accent" />
              </div>
              <div>
                <div className="text-[10px] text-text-muted font-bold uppercase tracking-widest">{s.label}</div>
                <div className="text-3xl font-medium tracking-tight text-charcoal mt-1">{s.value.toLocaleString('zh-CN')}</div>
              </div>
            </div>
          ))}
        </div>
        <div className="sidebar-panel rounded-[2rem] p-6">
          <div className="flex items-center justify-between mb-3">
            <div>
              <h3 className="text-lg font-medium tracking-tight text-charcoal">访问量总览</h3>
              <p className="text-xs text-text-muted mt-1">从数据统计启用后累计计算</p>
            </div>
            <Eye className="w-5 h-5 text-accent" />
          </div>
          <div className="h-36 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={[
                { name: '总浏览量', value: totalActivity.pv || 0, fill: '#1f2a44' },
                { name: '总独立访客', value: totalActivity.uv || 0, fill: '#6f8f75' },
              ]} layout="vertical" margin={{ top: 0, right: 12, left: 8, bottom: 0 }}>
                <CartesianGrid strokeDasharray="3 3" horizontal={false} stroke="#edf0f4" />
                <XAxis type="number" hide />
                <YAxis type="category" dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 11, fill: '#787774' }} width={70} />
                <Tooltip formatter={(value: number) => [value.toLocaleString('zh-CN'), '数量']} contentStyle={{ borderRadius: '14px', border: '1px solid #eaeaea', fontSize: '12px' }} />
                <Bar dataKey="value" radius={[0, 8, 8, 0]} barSize={22} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* 内容发布趋势 */}
        <div className="sidebar-panel rounded-[2.5rem] p-8 space-y-6">
          <div className="flex justify-between items-center">
            <h3 className="font-serif text-lg text-primary">7 天内容发布趋势</h3>
            <button onClick={handleInvalidate} className="btn-secondary !text-xs !py-1.5 !px-3 flex items-center gap-1.5">
              <RefreshCw className="w-3 h-3" /> 刷新首页缓存
            </button>
          </div>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={trend.dates?.map((d: string, i: number) => ({ name: d.slice(5), events: trend.events[i], news: trend.news[i] })) || []}>
                <defs>
                  <linearGradient id="colorEv" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#526069" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#526069" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f1f1" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#999' }} />
                <YAxis hide />
                <Tooltip contentStyle={{ borderRadius: '16px', border: 'none', boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)', fontSize: '12px' }} />
                <Area type="monotone" dataKey="events" stroke="#526069" fillOpacity={1} fill="url(#colorEv)" name="活动" />
                <Area type="monotone" dataKey="news" stroke="#ba1a1a" strokeDasharray="5 5" fill="transparent" name="资讯" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* 流量与活跃度趋势 */}
        <div className="sidebar-panel rounded-[2.5rem] p-8 space-y-6">
          <div className="flex justify-between items-center">
            <h3 className="font-serif text-lg text-primary">7 天访客与活跃度趋势</h3>
          </div>
          <div className="h-64 w-full">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={activity.dates?.map((d: string, i: number) => ({ name: d.slice(5), pv: activity.pv[i], uv: activity.uv[i], dau: activity.dau[i] })) || []}>
                <defs>
                  <linearGradient id="colorPv" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#2563eb" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#2563eb" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="colorUv" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#16a34a" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#16a34a" stopOpacity={0} />
                  </linearGradient>
                  <linearGradient id="colorDau" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#ca8a04" stopOpacity={0.1} />
                    <stop offset="95%" stopColor="#ca8a04" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" vertical={false} stroke="#f1f1f1" />
                <XAxis dataKey="name" axisLine={false} tickLine={false} tick={{ fontSize: 10, fill: '#999' }} />
                <YAxis hide />
                <Tooltip contentStyle={{ borderRadius: '16px', border: 'none', boxShadow: '0 10px 15px -3px rgba(0,0,0,0.1)', fontSize: '12px' }} />
                <Area type="monotone" dataKey="pv" stroke="#2563eb" fillOpacity={1} fill="url(#colorPv)" name="PV (浏览量)" />
                <Area type="monotone" dataKey="uv" stroke="#16a34a" fillOpacity={1} fill="url(#colorUv)" name="UV (独立访客)" />
                <Area type="monotone" dataKey="dau" stroke="#ca8a04" fillOpacity={1} fill="url(#colorDau)" name="DAU (活跃会员)" />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </>
  );
}
