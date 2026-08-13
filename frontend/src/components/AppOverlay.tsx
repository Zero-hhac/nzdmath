import React, { useEffect } from 'react';
import { AnimatePresence, motion } from 'motion/react';
import { X } from 'lucide-react';
import { cn } from '@/src/lib/utils';
import type { OverlayConfig } from '@/src/types/app';

interface AppOverlayProps {
  overlay: OverlayConfig | null;
  onClose: () => void;
}

export const AppOverlay: React.FC<AppOverlayProps> = ({ overlay, onClose }) => {
  useEffect(() => {
    if (!overlay) return;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKeyDown);
    return () => window.removeEventListener('keydown', onKeyDown);
  }, [overlay, onClose]);

  return (
    <AnimatePresence>
      {overlay && (
        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          className="fixed inset-0 z-[100] bg-accent/20 backdrop-blur-md p-3 md:p-10"
          onClick={onClose}
        >
          <motion.div
            initial={{ opacity: 0, y: 24, scale: 0.98 }}
            animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 16, scale: 0.98 }}
          transition={{ duration: 0.24, ease: 'easeOut' }}
          className="mx-auto flex max-h-[94vh] md:max-h-[90vh] w-full max-w-4xl flex-col overflow-hidden rounded-[1.5rem] md:rounded-[2rem] border border-border bg-surface shadow-[0_30px_120px_rgba(31,42,68,0.2)]"
          role="dialog"
          aria-modal="true"
          aria-label={overlay.title}
          onClick={(event) => event.stopPropagation()}
          >
            <div className="flex items-start justify-between gap-6 border-b border-border bg-canvas-alt px-5 py-5 md:px-8">
              <div className="space-y-1">
                <div className="text-[11px] font-bold uppercase tracking-[0.24em] text-zinc-400">
                  Association Workspace
                </div>
                <h3 className="text-2xl font-medium tracking-tight text-charcoal">{overlay.title}</h3>
                {overlay.subtitle && (
                  <p className="text-sm font-medium text-text-muted">{overlay.subtitle}</p>
                )}
              </div>
              <button
                onClick={onClose}
                className="flex h-10 w-10 items-center justify-center rounded-full border border-border bg-surface text-text-muted transition-all hover:text-accent"
                aria-label="关闭弹层"
              >
                <X className="h-5 w-5" />
              </button>
            </div>

            <div className="overflow-y-auto bg-surface px-5 py-6 md:px-8 md:py-8">{overlay.content}</div>

            {overlay.footer && (
              <div className="flex-shrink-0 border-t border-border bg-canvas-alt px-5 py-5 md:px-8">
                {overlay.footer}
              </div>
            )}

            {overlay.actions && overlay.actions.length > 0 && (
              <div className="flex flex-wrap justify-end gap-3 border-t border-border bg-canvas-alt px-5 py-5 md:px-8">
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
