import type { TabId } from '@/src/types/app';

export const tabPaths: Record<TabId, string> = {
  home: '/',
  events: '/events',
  resources: '/resources',
  showcase: '/showcase',
  news: '/news',
  about: '/about',
  chat: '/chat',
  admin: '/admin',
  portal: '/portal',
};

export function tabFromPath(pathname: string): TabId {
  for (const [tab, path] of Object.entries(tabPaths)) {
    if (path === pathname) return tab as TabId;
  }
  return 'home';
}
