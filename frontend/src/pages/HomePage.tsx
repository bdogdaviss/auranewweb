import { Link } from 'react-router-dom';
import Hero from '../components/Hero';
import FeaturedCard from '../components/FeaturedCard';
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
      <Hero />

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
    </section>
  );
}
