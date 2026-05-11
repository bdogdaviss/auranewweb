import { createContext, useContext, useState, type ReactNode } from 'react';

// Shared state for the download confirmation modal. Mounted at App root so
// both the hero StatefulDownloadButton and the Starter FeaturedCard can
// open the same modal without prop-drilling.

type Ctx = {
  isOpen: boolean;
  openModal: () => void;
  closeModal: () => void;
};

const DownloadModalContext = createContext<Ctx | null>(null);

export function DownloadModalProvider({ children }: { children: ReactNode }) {
  const [isOpen, setIsOpen] = useState(false);
  return (
    <DownloadModalContext.Provider
      value={{
        isOpen,
        openModal: () => setIsOpen(true),
        closeModal: () => setIsOpen(false),
      }}
    >
      {children}
    </DownloadModalContext.Provider>
  );
}

export function useDownloadModal(): Ctx {
  const ctx = useContext(DownloadModalContext);
  if (!ctx) {
    throw new Error('useDownloadModal must be used inside <DownloadModalProvider>');
  }
  return ctx;
}
