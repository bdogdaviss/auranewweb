import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Background from './components/Background';
import BackgroundRippleEffect from './components/BackgroundRippleEffect';
import AnimatedNav from './components/AnimatedNav';
import Footer from './components/Footer';
import HomePage from './pages/HomePage';
import ProductsPage from './pages/ProductsPage';
import AboutPage from './pages/AboutPage';
import AccountPage from './pages/AccountPage';
import UiShowcase from './pages/UiShowcase';
import { DownloadModalProvider } from './components/DownloadModalContext';
import DownloadModal from './components/DownloadModal';
import { ToastProvider } from './components/ToastContext';
import ToastContainer from './components/Toast';
import { useReferralCapture } from './hooks/useReferralCapture';

export default function App() {
  useReferralCapture();
  return (
    <BrowserRouter>
      <ToastProvider>
        <DownloadModalProvider>
          <Background />
          <BackgroundRippleEffect />
          <AnimatedNav />
          <main className="main-content" role="main">
            <Routes>
              <Route path="/" element={<HomePage />} />
              <Route path="/products" element={<ProductsPage />} />
              <Route path="/pricing" element={<ProductsPage />} />
              <Route path="/about" element={<AboutPage />} />
              <Route path="/account" element={<AccountPage />} />
              <Route path="/ui" element={<UiShowcase />} />
              <Route path="*" element={<HomePage />} />
            </Routes>
          </main>
          <Footer />
          <DownloadModal />
        </DownloadModalProvider>
        <ToastContainer />
      </ToastProvider>
    </BrowserRouter>
  );
}
