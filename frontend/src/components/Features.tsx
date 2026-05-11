import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { useReveal } from '../hooks/useReveal';

type Variant = 'image' | 'orbital' | 'icon';

type Feature = {
  title: string;
  description: string;
  alt: string;
  image?: string;
  hasTrending?: boolean;
  trendingValue?: string;
  imageSize?: 'narrow' | 'wide';
  variant?: Variant;
  icon?: 'cog' | 'cpu' | 'shield' | 'keyboard' | 'gauge' | 'wifi';
};

// 5 satellites + 1 center. Files come from auranewweb/static/images/, served
// at /static/images/ by the Go server's static-file handler.
const ORBITAL_CENTER = '/static/images/showcase.png';
const ORBITAL_SATELLITES = [
  '/static/images/newpfp.png',
  '/static/images/kise.jpg',
  '/static/images/narpzfn.jpeg',
  '/static/images/sjay.jpeg',
  '/static/images/rezy.jpeg',
];

function OrbitalCluster() {
  return (
    <div className="orbital" aria-hidden>
      <svg className="orbital-rings" viewBox="0 0 200 200" preserveAspectRatio="xMidYMid meet">
        <circle cx="100" cy="100" r="42" fill="none" stroke="rgba(255,255,255,0.07)" />
        <circle cx="100" cy="100" r="72" fill="none" stroke="rgba(255,255,255,0.05)" />
        <circle cx="100" cy="100" r="95" fill="none" stroke="rgba(255,255,255,0.035)" />
      </svg>
      <div className="orbital-icon orbital-center">
        <img src={ORBITAL_CENTER} alt="" loading="lazy" />
      </div>
      {ORBITAL_SATELLITES.map((src, i) => (
        <div key={src} className={`orbital-icon orbital-sat orbital-sat-${i}`}>
          <img src={src} alt="" loading="lazy" />
        </div>
      ))}
    </div>
  );
}

function TrendingUpIcon() {
  return (
    <svg
      width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#00a63e"
      strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden
    >
      <path d="M16 7h6v6" />
      <path d="m22 7-8.5 8.5-5-5L2 17" />
    </svg>
  );
}

// One inline SVG per icon variant. 64×64 viewBox; rendered at ~88px on the card.
const ICON_PROPS = {
  width: 64, height: 64, viewBox: '0 0 24 24', fill: 'none',
  stroke: 'currentColor', strokeWidth: 1.4, strokeLinecap: 'round' as const,
  strokeLinejoin: 'round' as const, 'aria-hidden': true,
};
const ICONS = {
  cog: (
    <svg {...ICON_PROPS}>
      <circle cx="12" cy="12" r="3" />
      <path d="M19.4 15a1.7 1.7 0 0 0 .3 1.8l.1.1a2 2 0 1 1-2.8 2.8l-.1-.1a1.7 1.7 0 0 0-1.8-.3 1.7 1.7 0 0 0-1 1.5V21a2 2 0 1 1-4 0v-.1A1.7 1.7 0 0 0 9 19.4a1.7 1.7 0 0 0-1.8.3l-.1.1a2 2 0 1 1-2.8-2.8l.1-.1a1.7 1.7 0 0 0 .3-1.8 1.7 1.7 0 0 0-1.5-1H3a2 2 0 1 1 0-4h.1A1.7 1.7 0 0 0 4.6 9a1.7 1.7 0 0 0-.3-1.8l-.1-.1a2 2 0 1 1 2.8-2.8l.1.1a1.7 1.7 0 0 0 1.8.3H9a1.7 1.7 0 0 0 1-1.5V3a2 2 0 1 1 4 0v.1a1.7 1.7 0 0 0 1 1.5 1.7 1.7 0 0 0 1.8-.3l.1-.1a2 2 0 1 1 2.8 2.8l-.1.1a1.7 1.7 0 0 0-.3 1.8V9a1.7 1.7 0 0 0 1.5 1H21a2 2 0 1 1 0 4h-.1a1.7 1.7 0 0 0-1.5 1z" />
    </svg>
  ),
  cpu: (
    <svg {...ICON_PROPS}>
      <rect x="4" y="4" width="16" height="16" rx="2" />
      <rect x="9" y="9" width="6" height="6" />
      <path d="M9 1v3M15 1v3M9 20v3M15 20v3M20 9h3M20 14h3M1 9h3M1 14h3" />
    </svg>
  ),
  shield: (
    <svg {...ICON_PROPS}>
      <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
      <path d="m9 12 2 2 4-4" />
    </svg>
  ),
  keyboard: (
    <svg {...ICON_PROPS}>
      <rect x="2" y="6" width="20" height="12" rx="2" />
      <path d="M6 10h.01M10 10h.01M14 10h.01M18 10h.01M7 14h10" />
    </svg>
  ),
  gauge: (
    <svg {...ICON_PROPS}>
      <path d="M12 14a2 2 0 1 0-2-2" />
      <path d="m13.4 10.6 3.6-3.6" />
      <path d="M21 12a9 9 0 1 1-18 0c0-2.5 1-4.8 2.7-6.5" />
    </svg>
  ),
  wifi: (
    <svg {...ICON_PROPS}>
      <path d="M5 12.55a11 11 0 0 1 14.08 0" />
      <path d="M1.42 9a16 16 0 0 1 21.16 0" />
      <path d="M8.53 16.11a6 6 0 0 1 6.95 0" />
      <line x1="12" y1="20" x2="12.01" y2="20" />
    </svg>
  ),
};

const FEATURES: Feature[] = [
  {
    title: 'System Tweaks',
    description:
      'OS-level tuning that strips background noise from your boot, services, and scheduler so the game gets the CPU.',
    alt: 'system tweaks',
    variant: 'icon',
    icon: 'cog',
  },
  {
    title: 'Network Stack',
    description:
      'Latency-first network tuning: route, MTU, TCP, and DNS configured for competitive ping consistency.',
    alt: 'network screenshot',
    image: '/static/images/newnetwork.png',
    variant: 'image',
    hasTrending: true,
    trendingValue: '−18ms',
  },
  {
    title: 'GPU Pipeline',
    description:
      'Render-side tuning: driver flags, frame pacing, shader-cache placement, and power-state pinning to lock your FPS floor.',
    alt: 'optimizer screenshot',
    image: '/static/images/newoptimizer.png',
    variant: 'image',
  },
  {
    title: 'Macro Engine',
    description:
      'Precision keyboard + mouse automation with sub-millisecond timing, plus build profiles you can hot-swap per game.',
    alt: 'macro engine',
    variant: 'icon',
    icon: 'keyboard',
  },
  {
    title: 'Real-Time Monitor',
    description:
      'Live dashboard of FPS, latency, frametime variance, and CPU/GPU load so you actually see what your changes did.',
    alt: 'dashboard screenshot',
    image: '/static/images/newdashboard.png',
    variant: 'image',
  },
  {
    title: 'Safe & Reversible',
    description:
      'Every tweak is snapshotted before it runs. One click to restore your prior state if a game breaks or you want stock back.',
    alt: 'safe & reversible',
    variant: 'icon',
    icon: 'shield',
  },
];

export function FeaturesHeader() {
  return (
    <>
      <button type="button" className="features-eyebrow" tabIndex={-1}>
        What you get
      </button>
      <h2 className="features-title">Built for the competitive edge.</h2>
      <p className="features-subtitle">
        Every millisecond between your brain and your screen is contested. Aura tunes the whole
        path — OS, network, GPU, input — so you reclaim the frames you paid for.
      </p>
    </>
  );
}

export default function Features({ hideHeader = false }: { hideHeader?: boolean }) {
  // hoveredIndex drives the shared-layoutId spotlight that slides between
  // cards. AnimatePresence handles enter/exit so the spotlight fades in on
  // the first hover and fades out when the cursor leaves the grid entirely.
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);

  return (
    <section className="features-section">
      {!hideHeader && <FeaturesHeader />}
      <div className="features-grid features-grid--six">
        {FEATURES.map((f, i) => (
          <div
            key={i}
            className="feature-card-wrap"
            onMouseEnter={() => setHoveredIndex(i)}
            onMouseLeave={() => setHoveredIndex(null)}
          >
            <AnimatePresence>
              {hoveredIndex === i && (
                <motion.span
                  layoutId="feature-hover-bg"
                  className="feature-hover-bg"
                  initial={{ opacity: 0 }}
                  animate={{ opacity: 1, transition: { duration: 0.15 } }}
                  exit={{ opacity: 0, transition: { duration: 0.15, delay: 0.2 } }}
                />
              )}
            </AnimatePresence>
            <FeatureCard feature={f} />
          </div>
        ))}
      </div>
    </section>
  );
}

function FeatureCard({ feature: f }: { feature: Feature }) {
  const ref = useReveal<HTMLElement>();
  return (
    <article ref={ref} className="feature-card reveal-on-scroll">
      {f.hasTrending && (
        <div className="feature-trending">
          <TrendingUpIcon />
          <span>{f.trendingValue ?? ''}</span>
        </div>
      )}
      <div className="feature-image-wrap">
        {f.variant === 'orbital' ? (
          <OrbitalCluster />
        ) : f.variant === 'icon' && f.icon ? (
          <div className="feature-icon">{ICONS[f.icon]}</div>
        ) : (
          <img
            className={`feature-image${
              f.imageSize === 'narrow'
                ? ' feature-image--narrow'
                : f.imageSize === 'wide'
                  ? ' feature-image--wide'
                  : ''
            }`}
            src={f.image}
            alt={f.alt}
            loading="lazy"
          />
        )}
      </div>
      <h3 className="feature-card-title">{f.title}</h3>
      <p className="feature-card-desc">{f.description}</p>
    </article>
  );
}
