import type { ReactNode } from 'react';

export type TabId =
  | 'home'
  | 'events'
  | 'resources'
  | 'showcase'
  | 'news'
  | 'about'
  | 'chat'
  | 'admin'
  | 'portal';

export interface OverlayAction {
  label: string;
  onClick?: () => void;
  variant?: 'primary' | 'secondary';
}

export interface OverlayConfig {
  title: string;
  subtitle?: string;
  content: ReactNode;
  /** 固定在卡片底部、不随内容滚动的自定义区域（如活动报名栏） */
  footer?: ReactNode;
  actions?: OverlayAction[];
}

export interface ViewProps {
  navigate: (tab: TabId) => void;
  openOverlay: (config: OverlayConfig) => void;
}
