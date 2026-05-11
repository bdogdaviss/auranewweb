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
import { FEATURED } from '../data/products';

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
            {FEATURED.map((p, i) =>
              p.id === 'lifetime' ? (
                // Lifetime card owns its own intersect animation (scale +
                // blur + spring); skip the parent Reveal to avoid stacking
                // two opacity/transform passes on the same element.
                <FeaturedCard key={p.id} product={p} />
              ) : (
                <Reveal key={p.id} delay={i * 0.08} amount={0.15}>
                  <FeaturedCard product={p} />
                </Reveal>
              )
            )}
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
