import React, { useEffect, useState } from 'react';
import { Navbar } from './components/Navbar';
import { Footer } from './components/Footer';
import { MathBackground } from './components/MathBackground';
import { HomeView } from './views/HomeView';
import { EventsView } from './views/EventsView';
import { ResourcesView } from './views/ResourcesView';
import { ShowcaseView } from './views/ShowcaseView';
import { NewsView } from './views/NewsView';
import { AboutView } from './views/AboutView';
import { AdminView } from './views/AdminView';
import { MemberPortalView as PortalView } from './views/MemberPortalView';
import { AppOverlay } from './components/AppOverlay';
import { motion, AnimatePresence } from 'motion/react';
import { AuthProvider } from './lib/auth';
import { ToastProvider } from './lib/toast';
import type { OverlayConfig, TabId } from './types/app';
import { api } from './lib/api';


export default function App() {
  return (
    <ToastProvider>
      <AuthProvider>
        <AppShell />
      </AuthProvider>
    </ToastProvider>
  );
}

function AppShell() {
  const [activeTab, setActiveTab] = useState<TabId>('home');
  const [overlay, setOverlay] = useState<OverlayConfig | null>(null);

  useEffect(() => {
    window.scrollTo({ top: 0, behavior: 'smooth' });
    api.trackPageView().catch(() => {});
  }, [activeTab]);

  const navigate = (tab: TabId) => {
    setOverlay(null);
    setActiveTab(tab);
  };

  const openOverlay = (config: OverlayConfig) => setOverlay(config);

  const renderView = () => {
    switch (activeTab) {
      case 'home': return <HomeView navigate={navigate} openOverlay={openOverlay} />;
      case 'events': return <EventsView navigate={navigate} openOverlay={openOverlay} />;
      case 'resources': return <ResourcesView navigate={navigate} openOverlay={openOverlay} />;
      case 'showcase': return <ShowcaseView navigate={navigate} openOverlay={openOverlay} />;
      case 'news': return <NewsView navigate={navigate} openOverlay={openOverlay} />;
      case 'about': return <AboutView navigate={navigate} openOverlay={openOverlay} />;
      case 'admin': return <AdminView navigate={navigate} openOverlay={openOverlay} />;
      case 'portal': return <PortalView navigate={navigate} openOverlay={openOverlay} />;
      default: return <HomeView navigate={navigate} openOverlay={openOverlay} />;
    }
  };

  return (
    <div className="min-h-screen flex flex-col selection:bg-charcoal/10 selection:text-charcoal">
      <MathBackground />
      <Navbar activeTab={activeTab} setActiveTab={navigate} />

      <main className="flex-grow pt-32 pb-32 px-6 md:px-10 max-w-5xl mx-auto w-full overflow-hidden">
        <AnimatePresence mode="wait">
          <motion.div
            key={activeTab}
            initial={{ opacity: 0, x: 10 }}
            animate={{ opacity: 1, x: 0 }}
            exit={{ opacity: 0, x: -10 }}
            transition={{ duration: 0.3, ease: 'easeInOut' }}
          >
            {renderView()}
          </motion.div>
        </AnimatePresence>
      </main>

      <Footer openOverlay={openOverlay} />
      <AppOverlay overlay={overlay} onClose={() => setOverlay(null)} />
    </div>
  );
}