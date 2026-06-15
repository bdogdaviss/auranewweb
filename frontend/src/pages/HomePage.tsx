import { Link } from 'react-router-dom';
import Hero3DStage from '../components/Hero3DStage';
import Features, { FeaturesHeader } from '../components/Features';
import HexGrid from '../components/HexGrid';
import Lamp from '../components/Lamp';
import FloatText from '../components/FloatText';
import FeaturedCard from '../components/FeaturedCard';
import CTASection from '../components/CTASection';
import Reveal from '../components/Reveal';
import Separator from '../components/Separator';
import VouchedProof from '../components/VouchedProof';
import UpsellBanner from '../components/UpsellBanner';
import { FEATURED } from '../data/products';

function CrownIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M2 7l5 5 5-7 5 7 5-5v11a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V7z" />
    </svg>
  );
}

function ArrowIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <line x1="5" y1="12" x2="19" y2="12" />
      <polyline points="12 5 19 12 12 19" />
    </svg>
  );
}

export default function HomePage() {
  return (
    <section className="tab-content active">
      <Hero3DStage />

      <HexGrid />

      <UpsellBanner
        dismissId="home-lifetime"
        title="Unlock Lifetime Pro — 85% off today"
        message={
          <>
            You're running the free build. Go Pro for AI tuning, the FPS-boost guarantee &amp;
            priority support — <a href="/pricing">save $84.99</a> with a one-time payment.
          </>
        }
        imageData={{ kind: 'icon', component: <CrownIcon /> }}
        primaryAction={{
          label: 'Get Lifetime Pro',
          accessibilityLabel: 'Upgrade to Aura Lifetime Pro',
          href: '/checkout?product=lifetime',
        }}
        secondaryAction={{
          label: 'See features',
          accessibilityLabel: 'Compare Aura Optimizer plans',
          href: '/pricing',
        }}
        dismissButton={{ accessibilityLabel: 'Dismiss the Lifetime Pro offer' }}
      />

      {/* Subtle break between hero/dashboard and the Features intro lamp —
          the most important divider on the page. */}
      <Separator color="green" />

      <Lamp>
        <FeaturesHeader />
      </Lamp>

      <Reveal>
        <Features hideHeader />
      </Reveal>

      <Reveal>
        <div className="vouched-row">
          <FloatText>Vouched by Joshreyli</FloatText>
          <VouchedProof />
        </div>
      </Reveal>

      {/* Transition from features → monetization. Quiet, centered, short. */}
      <Separator color="green" />

      <Reveal>
        <div className="featured-section">
          <div className="section-header scroll-animate visible">
            <span className="section-label">
              <span className="label-dot" aria-hidden /> Featured Products
            </span>
            <h2 className="section-title gradient-text">Top Picks</h2>
          </div>

          <div className="featured-grid">
            {FEATURED.map((p) => (
              <FeaturedCard key={p.id} product={p} />
            ))}
          </div>

          <div className="featured-view-all">
            <Link to="/products" className="btn-view-all">
              View All Products <ArrowIcon />
            </Link>
          </div>
        </div>
      </Reveal>

      <Reveal>
        <CTASection />
      </Reveal>
    </section>
  );
}
