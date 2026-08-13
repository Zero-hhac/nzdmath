const BASE = '/api/v1';

const USER_TOKEN_KEY = 'user_token';
const ADMIN_TOKEN_KEY = 'admin_token';

export const tokenStore = {
  getUser: () => localStorage.getItem(USER_TOKEN_KEY) || '',
  setUser: (t: string) => localStorage.setItem(USER_TOKEN_KEY, t),
  clearUser: () => localStorage.removeItem(USER_TOKEN_KEY),

  getAdmin: () => localStorage.getItem(ADMIN_TOKEN_KEY) || '',
  setAdmin: (t: string) => localStorage.setItem(ADMIN_TOKEN_KEY, t),
  clearAdmin: () => localStorage.removeItem(ADMIN_TOKEN_KEY),
};

export type ApiResponse<T = any> = {
  code: number;
  msg: string;
  data: T;
  total?: number;
  page?: number;
  pageSize?: number;
};

export async function request<T = any>(
  url: string,
  options: RequestInit & { tokenType?: 'user' | 'admin' } = {},
): Promise<ApiResponse<T>> {
  const { tokenType = 'user', ...fetchOptions } = options;
  const headers: Record<string, string> = {
    ...(fetchOptions.headers as Record<string, string> | undefined),
  };
  const t = tokenType === 'admin' ? tokenStore.getAdmin() : tokenStore.getUser();
  if (t) headers['Authorization'] = `Bearer ${t}`;
  if (fetchOptions.body && !(fetchOptions.body instanceof FormData)) {
    headers['Content-Type'] = headers['Content-Type'] || 'application/json';
  }

  const res = await fetch(`${BASE}${url}`, { ...fetchOptions, headers });
  let json: ApiResponse<T>;
  try {
    json = await res.json();
  } catch {
    throw new Error(`请求失败 (${res.status})`);
  }
  if (json.code !== 200 && json.code !== 0) {
    throw new Error(json.msg || '请求失败');
  }
  return json;
}

export async function exportDownload(url: string): Promise<void> {
  const t = tokenStore.getAdmin();
  const headers: Record<string, string> = {};
  if (t) headers['Authorization'] = `Bearer ${t}`;
  const res = await fetch(`${BASE}${url}`, { headers });
  const contentType = res.headers.get('Content-Type') || '';
  if (!res.ok || contentType.includes('application/json')) {
    let msg = `导出失败 (${res.status})`;
    try {
      const data = await res.json();
      if (data?.msg) msg = data.msg;
    } catch {
      // ignore
    }
    throw new Error(msg);
  }
  const blob = await res.blob();
  const disposition = res.headers.get('Content-Disposition') || '';
  let filename = 'members.xlsx';
  const utf8 = disposition.match(/filename\*=UTF-8''([^;]+)/i);
  if (utf8 && utf8[1]) {
    filename = decodeURIComponent(utf8[1]);
  } else {
    const plain = disposition.match(/filename="?([^";]+)"?/i);
    if (plain && plain[1]) filename = plain[1];
  }
  const objectUrl = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = objectUrl;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  URL.revokeObjectURL(objectUrl);
}

/**
 * 带鉴权拉取静态文件（聊天室文件已挂 JWT 静态路由，<img>/<a> 无法带
 * Authorization 头，必须走 fetch + blob URL）。
 */
export async function authFetchBlob(
  url: string,
  tokenType: 'user' | 'admin' = 'user',
): Promise<Blob> {
  const t = tokenType === 'admin' ? tokenStore.getAdmin() : tokenStore.getUser();
  const res = await fetch(url, { headers: t ? { Authorization: `Bearer ${t}` } : {} });
  if (!res.ok) throw new Error(`加载失败 (${res.status})`);
  return res.blob();
}
