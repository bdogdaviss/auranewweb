import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useDownloadModal } from './DownloadModalContext';

// StatefulDownloadButton — Aceternity-style three-state button.
//   idle    → "Download Now"
//   loading → spinner + "Preparing"
//   success → check + "Ready"  → opens the DownloadModal
//
// Brief artificial delay so the user gets visual feedback that the click
// was registered before the confirmation modal appears.

type Status = 'idle' | 'loading' | 'success';

function CheckIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <polyline points="20 6 9 17 4 12" />
    </svg>
  );
}

const variants = {
  initial: { opacity: 0, y: 8 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -8 },
};

export default function StatefulDownloadButton({
  children = 'Download Now',
}: {
  children?: React.ReactNode;
}) {
  const [status, setStatus] = useState<Status>('idle');
  const { openModal } = useDownloadModal();

  const handleClick = async () => {
    if (status !== 'idle') return;
    setStatus('loading');
    await new Promise((r) => setTimeout(r, 900));
    setStatus('success');
    await new Promise((r) => setTimeout(r, 400));
    openModal();
    // Reset back to idle so the button is ready for next click after the
    // user dismisses the modal.
    setTimeout(() => setStatus('idle'), 400);
  };

  return (
    <motion.button
      type="button"
      className="stateful-btn"
      onClick={handleClick}
      whileTap={status === 'idle' ? { scale: 0.97 } : undefined}
      disabled={status !== 'idle'}
      aria-busy={status === 'loading'}
      aria-live="polite"
    >
      <AnimatePresence mode="wait" initial={false}>
        {status === 'idle' && (
          <motion.span
            key="idle"
            className="stateful-btn-content"
            variants={variants}
            initial="initial"
            animate="animate"
            exit="exit"
            transition={{ duration: 0.2 }}
          >
            {children}
          </motion.span>
        )}
        {status === 'loading' && (
          <motion.span
            key="loading"
            className="stateful-btn-content"
            variants={variants}
            initial="initial"
            animate="animate"
            exit="exit"
            transition={{ duration: 0.2 }}
          >
            <span className="stateful-btn-spinner" aria-hidden />
            Preparing
          </motion.span>
        )}
        {status === 'success' && (
          <motion.span
            key="success"
            className="stateful-btn-content"
            variants={variants}
            initial="initial"
            animate="animate"
            exit="exit"
            transition={{ duration: 0.2 }}
          >
            <CheckIcon />
            Ready
          </motion.span>
        )}
      </AnimatePresence>
    </motion.button>
  );
}
