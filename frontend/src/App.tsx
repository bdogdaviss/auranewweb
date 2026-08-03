import { useEffect } from 'react';
import { BrowserRouter, Routes, Route, useLocation } from 'react-router-dom';
import { CanvasBackground } from './components/design/CanvasBackground';
import { SiteHeader } from './components/design/SiteHeader';
import { SiteFooter } from './components/design/SiteFooter';
import HomePage from './pages/HomePage';
import ProductsPage from './pages/ProductsPage';
import AboutPage from './pages/AboutPage';
import AccountPage from './pages/AccountPage';
import ContactPage from './pages/ContactPage';
import UiShowcase from './pages/UiShowcase';
import { DownloadModalProvider } from './components/DownloadModalContext';
import DownloadModal from './components/DownloadModal';
import { ToastProvider } from './components/ToastContext';
import ToastContainer from './components/Toast';
import { useReferralCapture } from './hooks/useReferralCapture';

// Client-side navigations keep the previous scroll position; jump to the top
// on every route change. behavior:'instant' bypasses the global
// `scroll-behavior: smooth` so the new page doesn't visibly scroll up.
function ScrollToTop() {
  const { pathname } = useLocation();
  useEffect(() => {
    window.scrollTo({ top: 0, left: 0, behavior: 'instant' });
  }, [pathname]);
  return null;
}

export default function App() {
  useReferralCapture();
  return (
    <BrowserRouter>
      <ScrollToTop />
      <ToastProvider>
        <DownloadModalProvider>
          <div className="relative min-h-screen overflow-x-hidden">
            <CanvasBackground />
            <SiteHeader />
            <main className="main-content relative z-10" role="main">
              <Routes>
                <Route path="/" element={<HomePage />} />
                <Route path="/products" element={<ProductsPage />} />
                <Route path="/pricing" element={<ProductsPage />} />
                <Route path="/about" element={<AboutPage />} />
                <Route path="/account" element={<AccountPage />} />
                <Route path="/affiliate" element={<AccountPage />} />
                <Route path="/contact" element={<ContactPage />} />
                <Route path="/ui" element={<UiShowcase />} />
                <Route path="*" element={<HomePage />} />
              </Routes>
            </main>
            <SiteFooter />
            <DownloadModal />
          </div>
        </DownloadModalProvider>
        <ToastContainer />
      </ToastProvider>
    </BrowserRouter>
  );
}
