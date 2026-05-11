import { BrowserRouter, Routes, Route } from 'react-router-dom';
import Background from './components/Background';
import BackgroundRippleEffect from './components/BackgroundRippleEffect';
import AnimatedNav from './components/AnimatedNav';
import StickyBanner from './components/StickyBanner';
import Footer from './components/Footer';
import HomePage from './pages/HomePage';
import ProductsPage from './pages/ProductsPage';
import AboutPage from './pages/AboutPage';
import { DownloadModalProvider } from './components/DownloadModalContext';
import DownloadModal from './components/DownloadModal';

export default function App() {
  return (
    <BrowserRouter>
      <DownloadModalProvider>
        <Background />
        <BackgroundRippleEffect />
        <StickyBanner />
        <AnimatedNav />
        <main className="main-content" role="main">
          <Routes>
            <Route path="/" element={<HomePage />} />
            <Route path="/products" element={<ProductsPage />} />
            <Route path="/pricing" element={<ProductsPage />} />
            <Route path="/about" element={<AboutPage />} />
            <Route path="*" element={<HomePage />} />
          </Routes>
        </main>
        <Footer />
        <DownloadModal />
      </DownloadModalProvider>
    </BrowserRouter>
  );
}
