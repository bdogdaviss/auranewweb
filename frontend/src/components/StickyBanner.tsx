import { useEffect, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { Link } from 'react-router-dom';

// StickyBanner — fixed top-of-page banner. Sets a CSS custom property on
// :root (--banner-h) while visible; AnimatedNav and .main-content read
// that var to add the right top offset so nothing overlaps. Dismissal is
// persisted to sessionStorage so it stays gone for the current tab.
//
// Adapted from Aceternity UI's StickyBanner (Tailwind + motion/react) to
// plain CSS in theme.css under "StickyBanner".

const STORAGE_KEY = 'aura-banner-dismissed-v1';
const BANNER_HEIGHT_PX = 44;

function CloseIcon() {
  return (
    <svg
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}

export default function StickyBanner() {
  // Render as hidden on first paint so SSR/CSR don't flash a banner the user
  // dismissed earlier. The effect below flips to visible once we've checked
  // sessionStorage.
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    if (sessionStorage.getItem(STORAGE_KEY) === '1') return;
    setVisible(true);
  }, []);

  // Drive the layout offset for the nav + main via a CSS variable on :root.
  useEffect(() => {
    const root = document.documentElement;
    if (visible) {
      root.style.setProperty('--banner-h', `${BANNER_HEIGHT_PX}px`);
    } else {
      root.style.removeProperty('--banner-h');
    }
    return () => {
      root.style.removeProperty('--banner-h');
    };
  }, [visible]);

  const dismiss = () => {
    setVisible(false);
    sessionStorage.setItem(STORAGE_KEY, '1');
  };

  return (
    <AnimatePresence>
      {visible && (
        <motion.div
          className="sticky-banner"
          initial={{ y: -50, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          exit={{ y: -50, opacity: 0 }}
          transition={{ type: 'spring', damping: 22, stiffness: 280 }}
          role="region"
          aria-label="Site announcement"
        >
          <p className="sticky-banner-text">
            Lifetime Pro is <strong>85% off</strong> — $15 one-time, forever updates.{' '}
            <Link to="/products" className="sticky-banner-link">
              Get it now →
            </Link>
          </p>
          <button
            type="button"
            className="sticky-banner-close"
            aria-label="Dismiss announcement"
            onClick={dismiss}
          >
            <CloseIcon />
          </button>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
