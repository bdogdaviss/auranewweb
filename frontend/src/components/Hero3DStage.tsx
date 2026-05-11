import { useEffect, useState } from 'react';
import { InteractiveHoverButton } from './InteractiveHoverButton';
import StatefulDownloadButton from './StatefulDownloadButton';

// Hero3DStage — centered hero text with a perspective-rotated dashboard mockup
// floating underneath. Scroll tilts the mockup forward for a subtle parallax.
// Plain CSS classes live in theme.css under "Hero3DStage".
const MOCKUP_SRC = '/static/images/newdashboard.png';

export default function Hero3DStage() {
  const [scrollY, setScrollY] = useState(0);

  useEffect(() => {
    const onScroll = () => setScrollY(window.scrollY);
    window.addEventListener('scroll', onScroll, { passive: true });
    return () => window.removeEventListener('scroll', onScroll);
  }, []);

  // Tilt forward as the user scrolls down. Clamp so we never invert.
  // 47deg at top → ~12deg by the time the stage scrolls off-screen.
  const tiltX = Math.max(12, 47 - scrollY * 0.06);
  const lift = Math.min(60, scrollY * 0.12);

  return (
    <section className="hero3d">
      <div className="hero3d-text">
        <div className="hero-badge animate-hero" style={{ animationDelay: '0.2s' }}>
          <span className="badge-dot" aria-hidden />
          <span>2,000+ Active Users</span>
        </div>
        <h1 className="hero3d-title animate-hero" style={{ animationDelay: '0.4s' }}>
          Elite Gaming
          <br />
          <span className="gradient-text">Performance</span>
        </h1>
        <p className="hero3d-subtitle animate-hero" style={{ animationDelay: '0.6s' }}>
          Professional-grade optimization tools engineered to deliver measurable
          competitive advantage. System tweaks, network tuning, GPU pipeline —
          one click, fully reversible.
        </p>
        <div className="hero3d-cta animate-hero" style={{ animationDelay: '0.8s' }}>
          <InteractiveHoverButton href="/checkout?product=lifetime">
            Get Lifetime Pro
          </InteractiveHoverButton>
          <StatefulDownloadButton>Download Now</StatefulDownloadButton>
        </div>
      </div>

      <div className="hero3d-stage" aria-hidden>
        <div
          className="hero3d-mockup"
          style={{
            transform: `translateY(${-lift}px) scale(1.05) rotateX(${tiltX}deg)`,
          }}
        >
          <img
            src={MOCKUP_SRC}
            alt=""
            className="hero3d-mockup-img"
            onError={(e) => {
              (e.currentTarget as HTMLImageElement).style.display = 'none';
            }}
          />
          <div className="hero3d-glow" />
        </div>
      </div>
    </section>
  );
}
