import { Component, lazy, Suspense, useEffect, useState, type ReactNode } from 'react';
import { BrowserRouter, Navigate, Route, Routes, useLocation, useNavigate } from 'react-router-dom';
import { Navbar } from './components/Navbar';
import { Footer } from './components/Footer';
import { MathBackground } from './components/MathBackground';
import { AppOverlay } from './components/AppOverlay';
import { motion } from 'motion/react';
import { AuthProvider } from './lib/auth';
import { ToastProvider } from './lib/toast';
import type { OverlayConfig, TabId } from './types/app';
import { tabPaths, tabFromPath } from './lib/routes';
import { api } from './lib/api';

const HomeView = lazy(() => import('./views/HomeView').then((m) => ({ default: m.HomeView })));
const EventsView = lazy(() => import('./views/EventsView').then((m) => ({ default: m.EventsView })));
const ResourcesView = lazy(() => import('./views/ResourcesView').then((m) => ({ default: m.ResourcesView })));
const ShowcaseView = lazy(() => import('./views/ShowcaseView').then((m) => ({ default: m.ShowcaseView })));
const NewsView = lazy(() => import('./views/NewsView').then((m) => ({ default: m.NewsView })));
const AboutView = lazy(() => import('./views/AboutView').then((m) => ({ default: m.AboutView })));
const AdminView = lazy(() => import('./views/AdminView').then((m) => ({ default: m.AdminView })));
const ChatView = lazy(() => import('./views/ChatView').then((m) => ({ default: m.ChatView })));
const PortalView = lazy(() => import('./views/MemberPortalView').then((m) => ({ default: m.MemberPortalView })));

function PageLoading() {
  return (
    <div className="flex items-center justify-center py-24 text-sm font-medium text-text-muted">
      加载中...
    </div>
  );
}

class ErrorBoundary extends Component<{ children: ReactNode }, { hasError: boolean }> {
  state = { hasError: false };

  static getDerivedStateFromError() {
    return { hasError: true };
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex min-h-screen items-center justify-center bg-canvas px-6 text-center">
          <div>
            <p className="text-lg font-semibold text-charcoal">页面暂时无法显示</p>
            <p className="mt-2 text-sm text-text-muted">请刷新页面重试。</p>
          </div>
        </div>
      );
    }
    return this.props.children;
  }
}


export default function App() {
  return (
    <BrowserRouter>
      <ToastProvider>
        <AuthProvider>
          <ErrorBoundary>
            <AppShell />
          </ErrorBoundary>
        </AuthProvider>
      </ToastProvider>
    </BrowserRouter>
  );
}

function AppShell() {
  const navigate = useNavigate();
  const location = useLocation();
  const [overlay, setOverlay] = useState<OverlayConfig | null>(null);

  const activeTab = tabFromPath(location.pathname);

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
    api.trackPageView().catch(() => {});
  }, [location.pathname]);

  const go = (tab: TabId) => {
    setOverlay(null);
    navigate(tabPaths[tab] ?? '/');
  };

  const openOverlay = (config: OverlayConfig) => setOverlay(config);

  const isChat = location.pathname === '/chat';
  const isAdmin = location.pathname === '/admin';

  return (
    <div className={`min-h-screen flex flex-col selection:bg-charcoal/10 selection:text-charcoal ${isChat ? 'h-[100dvh] overflow-hidden' : ''}`}>
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[130] focus:rounded-full focus:bg-charcoal focus:px-4 focus:py-2 focus:text-white"
      >
        跳到主要内容
      </a>
      <MathBackground />
      <Navbar activeTab={activeTab} setActiveTab={go} />

      <main
        id="main-content"
        tabIndex={-1}
        className={`flex-grow w-full max-w-6xl mx-auto ${
          isChat
            ? 'pt-20 md:pt-24 pb-3 px-3 sm:px-6 h-[100dvh] flex flex-col min-h-0 overflow-hidden'
            : isAdmin
            ? 'pt-20 md:pt-24 pb-8 px-4 sm:px-6 md:px-8'
            : 'pt-28 md:pt-32 pb-24 md:pb-32 px-4 sm:px-6 md:px-10'
        }`}
      >
        <motion.div
          key={location.pathname}
          initial={{ opacity: 0, x: 10 }}
          animate={{ opacity: 1, x: 0 }}
          transition={{ duration: 0.3, ease: 'easeInOut' }}
          className={isChat ? 'flex-1 flex flex-col min-h-0 h-full' : ''}
        >
          <Suspense fallback={<PageLoading />}>
            <Routes>
              <Route path="/" element={<HomeView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/events" element={<EventsView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/resources" element={<ResourcesView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/showcase" element={<ShowcaseView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/news" element={<NewsView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/about" element={<AboutView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/chat" element={<ChatView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/admin" element={<AdminView navigate={go} openOverlay={openOverlay} />} />
              <Route path="/portal" element={<PortalView navigate={go} openOverlay={openOverlay} />} />
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </Suspense>
        </motion.div>
      </main>

      {!isChat && <Footer openOverlay={openOverlay} />}
      <AppOverlay overlay={overlay} onClose={() => setOverlay(null)} />
    </div>
  );
}
