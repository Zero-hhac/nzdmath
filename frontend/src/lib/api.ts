import { request, tokenStore, exportDownload } from './http';

export const DEPARTMENTS = ['组织部', '办公室', '宣传部', '外联部'];

export const api = {
  getHome: () => request('/home'),
  trackPageView: () => request('/active/track', { method: 'POST' }),
  getEvents: (page = 1, pageSize = 12) => request(`/events?page=${page}&page_size=${pageSize}`),
  getEvent: (id: number) => request(`/events/${id}`),
  getNews: (page = 1, pageSize = 10) => request(`/news?page=${page}&page_size=${pageSize}`),
  getNewsDetail: (id: number) => request(`/news/${id}`),
  getResources: () => request('/resources'),
  getResource: (id: number) => request(`/resources/${id}`),
  downloadResource: (id: number) => `/api/v1/resources/download/${id}`,
  getShowcases: (params?: Record<string, string>) =>
    request('/showcases' + (params ? '?' + new URLSearchParams(params) : '')),
  getShowcase: (id: number) => request(`/showcases/${id}`),
  getComments: (targetType: string, targetId: number, page = 1) =>
    request(`/comments?target_type=${targetType}&target_id=${targetId}&page=${page}`),

  userLogin: (username: string, password: string) =>
    request<{ token: string; user: any }>('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
    }),
  userRegister: (data: { username: string; password: string; nickname?: string; email?: string; real_name: string; class_name: string; department: string }) =>
    request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  userLogout: () => request('/auth/logout', { method: 'POST' }),
  forgotPassword: (username: string, email: string) =>
    request<{ message?: string; dev_code?: string }>('/auth/forgot-password', {
      method: 'POST',
      body: JSON.stringify({ username, email }),
    }),
  resetPassword: (username: string, email: string, code: string, newPassword: string) =>
    request('/auth/reset-password', {
      method: 'POST',
      body: JSON.stringify({ username, email, code, new_password: newPassword }),
    }),
  changePassword: (oldPassword: string, newPassword: string) =>
    request('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
    }),

  getProfile: () => request('/profile'),
  updateProfile: (data: { nickname?: string; avatar?: string; bio?: string; email?: string; real_name?: string; class_name?: string; department?: string }) =>
    request('/profile', { method: 'PUT', body: JSON.stringify(data) }),
  uploadAvatar: (formData: FormData) =>
    request<{ avatar: string }>('/user/avatar', { method: 'POST', body: formData }),

  getFavorites: () => request('/member/favorites'),
  addFavorite: (targetType: string, targetId: number) =>
    request('/member/favorites', {
      method: 'POST',
      body: JSON.stringify({ target_type: targetType, target_id: targetId }),
    }),
  removeFavorite: (id: number) =>
    request(`/member/favorites/${id}`, { method: 'DELETE' }),

  registerEvent: (eventId: number) =>
    request(`/member/events/${eventId}/register`, { method: 'POST' }),
  cancelEventRegistration: (eventId: number) =>
    request(`/member/events/${eventId}/register`, { method: 'DELETE' }),
  getMyRegistrations: () => request('/member/events/registrations'),

  getNotifications: (page = 1, pageSize = 20, unreadOnly = false) =>
    request(`/member/notifications?page=${page}&page_size=${pageSize}&unread_only=${unreadOnly ? 1 : 0}`),
  getUnreadNotificationCount: () => request<{ count: number }>('/member/notifications/unread-count'),
  markNotificationRead: (id: number) =>
    request(`/member/notifications/${id}/read`, { method: 'POST' }),
  markAllNotificationsRead: () =>
    request('/member/notifications/read-all', { method: 'POST' }),

  getMyDownloads: (page = 1) => request(`/member/downloads?page=${page}`),

  addComment: (data: { target_type: string; target_id: number; content: string; rating?: number; parent_id?: number }) =>
    request('/comments', { method: 'POST', body: JSON.stringify(data) }),
  deleteComment: (id: number) => request(`/comments/${id}`, { method: 'DELETE' }),
  toggleCommentLike: (id: number) => request(`/comments/${id}/like`, { method: 'POST' }),

  chatJoin: () => request<{ online_count: number }>('/chat/join', { method: 'POST' }),
  chatLeave: () => request('/chat/leave', { method: 'POST' }),
  getChatMessages: (options: { afterId?: number; beforeId?: number; limit?: number; afterDeleteMs?: number } = {}, signal?: AbortSignal) => {
    const params = new URLSearchParams();
    if (options.afterId) params.set('after_id', String(options.afterId));
    if (options.beforeId) params.set('before_id', String(options.beforeId));
    params.set('limit', String(options.limit || 50));
    if (options.afterDeleteMs) params.set('after_delete_ms', String(options.afterDeleteMs));
    const query = params.toString();
    return request<{
      messages: any[];
      online_count: number;
      deleted_ids?: number[];
      deleted_at_ms?: number;
      has_more?: boolean;
      next_before_id?: number;
    }>(`/chat/messages?${query}`, { signal });
  },
  sendChatText: (content: string) =>
    request('/chat/messages', {
      method: 'POST',
      body: JSON.stringify({ content }),
    }),
  sendChatFile: (formData: FormData) =>
    request('/chat/messages/file', { method: 'POST', body: formData }),
  deleteChatMessage: (id: number) => request(`/chat/messages/${id}`, { method: 'DELETE' }),

  adminLogin: (username: string, password: string) =>
    request<{ token: string; admin: any }>('/admin/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
      tokenType: 'admin',
    }),
  adminLogout: () => request('/admin/auth/logout', { method: 'POST', tokenType: 'admin' }),
  adminChangePassword: (oldPassword: string, newPassword: string) =>
    request('/admin/auth/password', {
      method: 'PUT',
      body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
      tokenType: 'admin',
    }),

  adminDashboard: () => request<{
    counts: any;
    today_activity: { pv: number; uv: number; dau: number };
    activity_trend: { dates: string[]; pv: number[]; uv: number[]; dau: number[] };
    total_activity: { pv: number; uv: number };
  }>('/admin/dashboard', { tokenType: 'admin' }),
  adminInvalidateHomepage: () => request('/admin/homepage/invalidate', { method: 'POST', tokenType: 'admin' }),

  adminListEvents: (params?: Record<string, string>) =>
    request('/admin/events' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminGetEvent: (id: number) => request(`/admin/events/${id}`, { tokenType: 'admin' }),
  adminCreateEvent: (data: any) =>
    request('/admin/events', { method: 'POST', body: JSON.stringify(data), tokenType: 'admin' }),
  adminUpdateEvent: (id: number, data: any) =>
    request(`/admin/events/${id}`, { method: 'PUT', body: JSON.stringify(data), tokenType: 'admin' }),
  adminDeleteEvent: (id: number) =>
    request(`/admin/events/${id}`, { method: 'DELETE', tokenType: 'admin' }),
  adminToggleEventFeature: (id: number) =>
    request(`/admin/events/${id}/feature`, { method: 'PATCH', tokenType: 'admin' }),
  adminSetEventExpired: (id: number, isExpired: boolean) =>
    request(`/admin/events/${id}/expired`, {
      method: 'PATCH',
      body: JSON.stringify({ is_expired: isExpired }),
      tokenType: 'admin',
    }),

  adminListNews: (params?: Record<string, string>) =>
    request('/admin/news' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminGetNews: (id: number) => request(`/admin/news/${id}`, { tokenType: 'admin' }),
  adminCreateNews: (data: any) =>
    request('/admin/news', { method: 'POST', body: JSON.stringify(data), tokenType: 'admin' }),
  adminUpdateNews: (id: number, data: any) =>
    request(`/admin/news/${id}`, { method: 'PUT', body: JSON.stringify(data), tokenType: 'admin' }),
  adminDeleteNews: (id: number) =>
    request(`/admin/news/${id}`, { method: 'DELETE', tokenType: 'admin' }),

  adminListResources: (params?: Record<string, string>) =>
    request('/admin/resources' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminGetResource: (id: number) => request(`/admin/resources/${id}`, { tokenType: 'admin' }),
  adminCreateResource: (data: FormData) =>
    request('/admin/resources', { method: 'POST', body: data, tokenType: 'admin' }),
  adminUpdateResource: (id: number, data: any) =>
    request(`/admin/resources/${id}`, { method: 'PUT', body: JSON.stringify(data), tokenType: 'admin' }),
  adminDeleteResource: (id: number) =>
    request(`/admin/resources/${id}`, { method: 'DELETE', tokenType: 'admin' }),

  adminListShowcases: (params?: Record<string, string>) =>
    request('/admin/showcases' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminGetShowcase: (id: number) => request(`/admin/showcases/${id}`, { tokenType: 'admin' }),
  adminCreateShowcase: (data: any) =>
    request('/admin/showcases', { method: 'POST', body: JSON.stringify(data), tokenType: 'admin' }),
  adminUpdateShowcase: (id: number, data: any) =>
    request(`/admin/showcases/${id}`, { method: 'PUT', body: JSON.stringify(data), tokenType: 'admin' }),
  adminDeleteShowcase: (id: number) =>
    request(`/admin/showcases/${id}`, { method: 'DELETE', tokenType: 'admin' }),

  adminListUsers: (params?: Record<string, string>) =>
    request('/admin/users' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminExportUsers: (department?: string) =>
    exportDownload('/admin/users/export' + (department ? `?department=${encodeURIComponent(department)}` : '')),
  adminSetUserStatus: (id: number, status: 0 | 1) =>
    request(`/admin/users/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }), tokenType: 'admin' }),
  adminResetUserPassword: (id: number, newPassword: string) =>
    request(`/admin/users/${id}/reset-password`, { method: 'POST', body: JSON.stringify({ new_password: newPassword }), tokenType: 'admin' }),
  adminDeleteUser: (id: number) =>
    request(`/admin/users/${id}`, { method: 'DELETE', tokenType: 'admin' }),
  adminBatchSetUserStatus: (ids: number[], status: 0 | 1) =>
    request('/admin/users/batch-status', { method: 'POST', body: JSON.stringify({ ids, status }), tokenType: 'admin' }),
  adminBatchResetUserPassword: (ids: number[], newPassword: string) =>
    request('/admin/users/batch-reset-password', { method: 'POST', body: JSON.stringify({ ids, new_password: newPassword }), tokenType: 'admin' }),
  adminBatchDeleteUsers: (ids: number[]) =>
    request('/admin/users/batch-delete', { method: 'POST', body: JSON.stringify({ ids }), tokenType: 'admin' }),

  adminListComments: (params?: Record<string, string>) =>
    request('/admin/comments' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminDeleteComment: (id: number) =>
    request(`/admin/comments/${id}`, { method: 'DELETE', tokenType: 'admin' }),
  adminSetCommentStatus: (id: number, status: 0 | 1) =>
    request(`/admin/comments/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }), tokenType: 'admin' }),

  adminListChatMessages: (params?: Record<string, string>) =>
    request('/admin/chat/messages' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminDeleteChatMessage: (id: number) =>
    request(`/admin/chat/messages/${id}`, { method: 'DELETE', tokenType: 'admin' }),

  adminListEventRegistrations: (eventId: number) =>
    request(`/admin/events/${eventId}/registrations`, { tokenType: 'admin' }),
  adminEventRegistrationSummary: () =>
    request('/admin/events/registration-summary', { tokenType: 'admin' }),
  adminCheckinEventRegistration: (eventId: number, userId: number) =>
    request(`/admin/events/${eventId}/registrations/${userId}/checkin`, { method: 'POST', tokenType: 'admin' }),
  adminUncheckinEventRegistration: (eventId: number, userId: number) =>
    request(`/admin/events/${eventId}/registrations/${userId}/uncheckin`, { method: 'POST', tokenType: 'admin' }),
  adminRemoveEventRegistration: (eventId: number, userId: number) =>
    request(`/admin/events/${eventId}/registrations/${userId}`, { method: 'DELETE', tokenType: 'admin' }),

  adminSendNotification: (data: {
    title: string;
    content: string;
    type?: string;
    target: { mode: 'all' | 'department' | 'users'; department?: string; usernames?: string[] };
  }) =>
    request('/admin/notifications', {
      method: 'POST',
      body: JSON.stringify(data),
      tokenType: 'admin',
    }),
  adminListNotificationBatches: (page = 1, pageSize = 20) =>
    request(`/admin/notifications?page=${page}&page_size=${pageSize}`, { tokenType: 'admin' }),
};

export { tokenStore };
