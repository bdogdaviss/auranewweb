import FeaturedCard from '../components/FeaturedCard';
import { FEATURED } from '../data/products';

// Placeholder grid: reuses the same FeaturedCard component as the Home tab.
// The dedicated ProductsPage layout from the original HTML (with the 6-card
// grid including Coming Soon tiles) will land after the user uploads the
// remaining assets. Until then this gives a working, navigable page.

export default function ProductsPage() {
  return (
    <section className="tab-content active">
      <div className="section-header scroll-animate visible">
        <span className="section-label">
          <span className="label-dot" aria-hidden /> Premium Products
        </span>
        <h2 className="section-title gradient-text">Gaming Optimization Suite</h2>
        <p className="section-subtitle">
          Professional tools engineered for competitive performance and maximum gaming advantage.
        </p>
      </div>

      <div className="featured-grid">
        {FEATURED.map((p) => (
          <FeaturedCard key={p.id} product={p} />
        ))}
      </div>
    </section>
  );
}
