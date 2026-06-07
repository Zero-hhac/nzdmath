import React from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { cn } from '@/src/lib/utils';
import type { OverlayConfig } from '@/src/types/app';

interface AppOverlayProps {
  overlay: OverlayConfig | null;
  onClose: () => void;
}

export const AppOverlay: React.FC<AppOverlayProps> = ({ overlay, onClose }) => {
  return (
    <AnimatePresence>
      {overlay && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-[100] bg-[#526069]/28 backdrop-blur-md p-4 md:p-10"
          onClick={onClose}
        >
          <motion.div
            initial={{ opacity: 0, y: 24, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
            exit={{ opacity: 0, y: 16, scale: 0.98 }}
            transition={{ duration: 0.24, ease: 'easeOut' }}
            className="mx-auto flex max-h-[90vh] w-full max-w-4xl flex-col overflow-hidden rounded-[2rem] border border-white/75 bg-white/92 shadow-[0_30px_120px_rgba(82,96,105,0.22)]"
            onClick={(event) => event.stopPropagation()}
          >
            <div className="flex items-start justify-between gap-6 border-b border-white/80 bg-[#fbfcfd] px-6 py-5 md:px-8">
              <div className="space-y-1">
                <div className="text-[11px] font-bold uppercase tracking-[0.24em] text-zinc-400">
                  Association Workspace
                </div>
                <h3 className="font-serif text-2xl text-primary">{overlay.title}</h3>
                {overlay.subtitle && (
                  <p className="text-sm font-medium text-[#586873]">{overlay.subtitle}</p>
                )}
              </div>
              <button
                onClick={onClose}
                className="flex h-10 w-10 items-center justify-center rounded-full border border-white/70 bg-white/92 text-zinc-500 transition-all hover:text-primary"
                aria-label="关闭弹层"
              >
                <span className="material-symbols-outlined text-[20px]">close</span>
              </button>
            </div>

            <div className="overflow-y-auto bg-white/82 px-6 py-6 md:px-8 md:py-8">{overlay.content}</div>

            {overlay.actions && overlay.actions.length > 0 && (
              <div className="flex flex-wrap justify-end gap-3 border-t border-white/80 bg-[#fbfcfd] px-6 py-5 md:px-8">
                {overlay.actions.map((action) => (
                  <button
                    key={action.label}
                    onClick={action.onClick ?? onClose}
                    className={cn(
                      'rounded-full px-6 py-3 text-sm font-semibold transition-all',
                      action.variant === 'secondary'
                        ? 'btn-secondary'
                        : 'btn-primary',
                    )}
                  >
                    {action.label}
                  </button>
                ))}
              </div>
            )}
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
};
