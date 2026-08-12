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
  actions?: OverlayAction[];
}

export interface ViewProps {
  navigate: (tab: TabId) => void;
  openOverlay: (config: OverlayConfig) => void;
}
