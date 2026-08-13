import { useEffect, useState } from 'react';
import { CalendarDays, CheckCircle2, Undo2, Trash2, Users } from 'lucide-react';
import { api } from '@/src/lib/api';
import { useToast } from '@/src/lib/toast';

/**
 * 活动报名/签到管理页：
 * 左侧为活动汇总（报名/签到人数），点击选择活动后右侧展示报名名单，
 * 支持签到、取消签到、移除报名。新报名按时间倒序展示。
 */
export function AdminRegistrationsPanel() {
  const { showToast } = useToast();
  const [events, setEvents] = useState<any[]>([]);
  const [loadingEvents, setLoadingEvents] = useState(true);
  const [selected, setSelected] = useState<number | null>(null);
  const [items, setItems] = useState<any[]>([]);
  const [loadingItems, setLoadingItems] = useState(false);

  const loadEvents = () => {
    setLoadingEvents(true);
    api.adminEventRegistrationSummary()
      .then((res) => {
        const data = res.data || [];
        setEvents(data);
        setSelected((prev) => prev ?? data[0]?.id ?? null);
      })
      .catch(() => setEvents([]))
      .finally(() => setLoadingEvents(false));
  };

  const loadItems = (eventId: number) => {
    setLoadingItems(true);
    api.adminListEventRegistrations(eventId)
      .then((res) => setItems(res.data || []))
      .catch(() => setItems([]))
      .finally(() => setLoadingItems(false));
  };

  useEffect(() => {
    loadEvents();
  }, []);

  useEffect(() => {
    if (selected) loadItems(selected);
  }, [selected]);

  const refresh = () => {
    loadItems(selected!);
    loadEvents();
  };

  const checkin = async (userId: number) => {
    if (!selected) return;
    try {
      await api.adminCheckinEventRegistration(selected, userId);
      showToast('签到成功', 'success');
      refresh();
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    }
  };

  const uncheckin = async (userId: number) => {
    if (!selected) return;
    try {
      await api.adminUncheckinEventRegistration(selected, userId);
      showToast('已取消签到', 'success');
      refresh();
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    }
  };

  const remove = async (userId: number) => {
    if (!selected) return;
    if (!confirm('确定移除该用户的报名吗？')) return;
    try {
      await api.adminRemoveEventRegistration(selected, userId);
      showToast('已移除报名', 'success');
      refresh();
    } catch (err: any) {
      showToast(err.message || '操作失败', 'error');
    }
  };

  const selectedEvent = events.find((e) => e.id === selected);
  const attendedCount = items.filter((i) => i.status === 2).length;

  return (
    <div className="space-y-6">
      <div className="page-intro space-y-2">
        <h2 className="section-title">活动报名 / 签到</h2>
        <p className="section-subtitle">查看各活动报名进度，为到场会员签到</p>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="sidebar-panel rounded-[2rem] p-5 space-y-2">
          <div className="flex items-center gap-2 px-1 pb-2 text-sm font-bold text-charcoal">
            <CalendarDays className="w-4 h-4 text-primary" /> 活动列表
          </div>
          {loadingEvents ? (
            <div className="text-center py-8 text-zinc-500 text-sm">加载中...</div>
          ) : events.length === 0 ? (
            <div className="text-center py-8 text-zinc-500 text-sm">暂无活动</div>
          ) : (
            events.map((e) => (
              <button
                key={e.id}
                onClick={() => setSelected(e.id)}
                className={`w-full text-left p-3 rounded-xl transition-colors ${
                  selected === e.id ? 'bg-primary/10 border border-primary/30' : 'hover:bg-white/60 border border-transparent'
                }`}
              >
                <div className="text-sm font-medium text-charcoal truncate">{e.title}</div>
                <div className="text-xs text-zinc-500 mt-1 flex flex-wrap gap-x-2">
                  <span>{new Date(e.start_time).toLocaleString('zh-CN')}</span>
                </div>
                <div className="text-xs mt-1.5 font-semibold flex gap-2">
                  <span className="px-2 py-0.5 rounded-full bg-blue-50 text-blue-600">
                    报名 {e.registered}{e.capacity ? `/${e.capacity}` : ''}
                  </span>
                  <span className="px-2 py-0.5 rounded-full bg-emerald-50 text-emerald-600">
                    签到 {e.attended}
                  </span>
                </div>
              </button>
            ))
          )}
        </div>

        <div className="lg:col-span-2 sidebar-panel rounded-[2rem] p-5">
          <div className="flex flex-wrap items-center justify-between gap-3 pb-3 border-b border-border mb-3">
            <div className="flex items-center gap-2 text-sm font-bold text-charcoal">
              <Users className="w-4 h-4 text-primary" />
              {selectedEvent ? selectedEvent.title : '请选择活动'}
            </div>
            {selectedEvent && (
              <div className="text-xs text-zinc-500">
                已报名 {selectedEvent.registered} 人{selectedEvent.capacity ? ` / ${selectedEvent.capacity}` : '（不限）'} · 已签到 {attendedCount} 人
              </div>
            )}
          </div>

          {!selected ? (
            <div className="text-center py-12 text-zinc-500 text-sm">请先在左侧选择活动</div>
          ) : loadingItems ? (
            <div className="text-center py-12 text-zinc-500 text-sm">加载中...</div>
          ) : items.length === 0 ? (
            <div className="text-center py-12 text-zinc-500 text-sm">暂无报名，会员报名成功后会实时出现在这里</div>
          ) : (
            <div className="space-y-2">
              {items.map((it) => (
                <div key={it.id} className="flex flex-wrap items-center gap-3 p-3 rounded-xl hover:bg-white/60 transition-colors">
                  <div className="flex-1 min-w-0">
                    <div className="font-medium text-zinc-800">
                      {it.real_name || it.nickname || it.username}
                      <span className="text-xs text-zinc-400 ml-2">@{it.username}</span>
                    </div>
                    <div className="text-xs text-zinc-500 mt-0.5">
                      {it.class_name || '未填班级'} · {it.department || '未分配'} · 报名于 {new Date(it.registered_at).toLocaleString('zh-CN')}
                    </div>
                  </div>
                  {it.status === 2 ? (
                    <span className="px-3 py-1 rounded-full text-xs font-bold bg-emerald-50 text-emerald-600 border border-emerald-200">
                      已签到 {it.checked_in_at ? new Date(it.checked_in_at).toLocaleString('zh-CN') : ''}
                    </span>
                  ) : (
                    <span className="px-3 py-1 rounded-full text-xs font-bold bg-blue-50 text-blue-600 border border-blue-200">已报名</span>
                  )}
                  <div className="flex items-center gap-1.5">
                    {it.status === 2 ? (
                      <button onClick={() => uncheckin(it.user_id)} className="px-3 py-1.5 rounded-lg text-xs text-zinc-600 hover:bg-white flex items-center gap-1">
                        <Undo2 className="w-3.5 h-3.5" /> 取消签到
                      </button>
                    ) : (
                      <button onClick={() => checkin(it.user_id)} className="px-3 py-1.5 rounded-lg text-xs text-emerald-600 hover:bg-emerald-50 flex items-center gap-1">
                        <CheckCircle2 className="w-3.5 h-3.5" /> 签到
                      </button>
                    )}
                    <button onClick={() => remove(it.user_id)} className="px-3 py-1.5 rounded-lg text-xs text-rose-500 hover:bg-rose-50 flex items-center gap-1">
                      <Trash2 className="w-3.5 h-3.5" /> 移除
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
