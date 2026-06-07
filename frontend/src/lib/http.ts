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
