import { useEffect } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useDownloadModal } from './DownloadModalContext';

// DownloadModal — confirmation popup that gates the actual download.
// Anchored to the App root via context so both the hero "Download Now"
// button and the Starter card's "Download Free" button can open it.

const DOWNLOAD_URL =
  'https://github.com/bdogdaviss/auranewweb/releases/download/v2.0.11/aura-setup.exe';
const VERSION = 'v2.0.11';
const FILE_NAME = 'aura-setup.exe';
const FILE_SIZE = '166.7 MB';

function CloseIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <line x1="18" y1="6" x2="6" y2="18" />
      <line x1="6" y1="6" x2="18" y2="18" />
    </svg>
  );
}

function DownloadIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
      <polyline points="7 10 12 15 17 10" />
      <line x1="12" y1="15" x2="12" y2="3" />
    </svg>
  );
}

export default function DownloadModal() {
  const { isOpen, closeModal } = useDownloadModal();

  // Esc-to-close + lock body scroll while modal is open.
  useEffect(() => {
    if (!isOpen) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') closeModal();
    };
    document.addEventListener('keydown', onKey);
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = prevOverflow;
    };
  }, [isOpen, closeModal]);

  return (
    <AnimatePresence>
      {isOpen && (
        <div className="download-modal-root">
          <motion.button
            type="button"
            className="download-modal-backdrop"
            onClick={closeModal}
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.18 }}
            aria-label="Close download dialog"
          />
          <motion.div
            className="download-modal"
            role="dialog"
            aria-modal="true"
            aria-labelledby="download-modal-title"
            initial={{ opacity: 0, scale: 0.92, y: 20 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.95, y: 12 }}
            transition={{ type: 'spring', stiffness: 320, damping: 26 }}
          >
            <button
              type="button"
              className="download-modal-close"
              onClick={closeModal}
              aria-label="Close"
            >
              <CloseIcon />
            </button>

            <div className="download-modal-icon" aria-hidden>
              <DownloadIcon />
            </div>

            <h2 id="download-modal-title" className="download-modal-title">
              Download Aura Optimizer
            </h2>

            <div className="download-modal-meta">
              <span className="download-modal-tag">{VERSION}</span>
              <span className="download-modal-sep" aria-hidden>·</span>
              <span>Windows 10 / 11</span>
              <span className="download-modal-sep" aria-hidden>·</span>
              <span>{FILE_SIZE}</span>
              <span className="download-modal-sep" aria-hidden>·</span>
              <span>{FILE_NAME}</span>
            </div>

            <p className="download-modal-desc">
              Direct download from our GitHub release. Run as administrator
              after the install completes — Aura needs elevated permissions
              to apply system-level tweaks.
            </p>

            <a
              href={DOWNLOAD_URL}
              className="download-modal-cta"
              onClick={closeModal}
              rel="noopener noreferrer"
            >
              <DownloadIcon />
              Download File
            </a>

            <button
              type="button"
              className="download-modal-cancel"
              onClick={closeModal}
            >
              Cancel
            </button>
          </motion.div>
        </div>
      )}
    </AnimatePresence>
  );
}
