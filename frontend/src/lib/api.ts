import { request, tokenStore } from './http';

export const api = {
  getHome: () => request('/home'),
  trackPageView: () => request('/active/track', { method: 'POST' }),
  getEvents: () => request('/events'),
  getEvent: (id: number) => request(`/events/${id}`),
  getNews: () => request('/news'),
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
  userRegister: (data: { username: string; password: string; nickname?: string; email?: string }) =>
    request('/auth/register', {
      method: 'POST',
      body: JSON.stringify(data),
    }),
  userLogout: () => request('/auth/logout', { method: 'POST' }),
  changePassword: (oldPassword: string, newPassword: string) =>
    request('/auth/change-password', {
      method: 'POST',
      body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
    }),

  getProfile: () => request('/profile'),
  updateProfile: (data: { nickname?: string; avatar?: string; bio?: string; email?: string }) =>
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

  getMyDownloads: (page = 1) => request(`/member/downloads?page=${page}`),

  addComment: (data: { target_type: string; target_id: number; content: string; rating?: number; parent_id?: number }) =>
    request('/comments', { method: 'POST', body: JSON.stringify(data) }),
  deleteComment: (id: number) => request(`/comments/${id}`, { method: 'DELETE' }),
  toggleCommentLike: (id: number) => request(`/comments/${id}/like`, { method: 'POST' }),

  adminLogin: (username: string, password: string) =>
    request<{ token: string; admin: any }>('/admin/auth/login', {
      method: 'POST',
      body: JSON.stringify({ username, password }),
      tokenType: 'admin',
    }),
  adminLogout: () => request('/admin/auth/logout', { method: 'POST', tokenType: 'admin' }),

  adminDashboard: () => request('/admin/dashboard', { tokenType: 'admin' }),
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

  uploadResource: (data: FormData) =>
    request('/resources/upload', { method: 'POST', body: data }),

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
  adminSetUserStatus: (id: number, status: 0 | 1) =>
    request(`/admin/users/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }), tokenType: 'admin' }),
  adminResetUserPassword: (id: number, newPassword: string) =>
    request(`/admin/users/${id}/reset-password`, { method: 'POST', body: JSON.stringify({ new_password: newPassword }), tokenType: 'admin' }),
  adminDeleteUser: (id: number) =>
    request(`/admin/users/${id}`, { method: 'DELETE', tokenType: 'admin' }),

  adminListComments: (params?: Record<string, string>) =>
    request('/admin/comments' + (params ? '?' + new URLSearchParams(params) : ''), { tokenType: 'admin' }),
  adminDeleteComment: (id: number) =>
    request(`/admin/comments/${id}`, { method: 'DELETE', tokenType: 'admin' }),
  adminSetCommentStatus: (id: number, status: 0 | 1) =>
    request(`/admin/comments/${id}/status`, { method: 'PATCH', body: JSON.stringify({ status }), tokenType: 'admin' }),
};

export { tokenStore };
