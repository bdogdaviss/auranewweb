import FeaturedCard from '../components/FeaturedCard';
import UpsellBanner from '../components/UpsellBanner';
import { useToast } from '../components/ToastContext';
import { FEATURED } from '../data/products';

// Placeholder grid: reuses the same FeaturedCard component as the Home tab.
// The dedicated ProductsPage layout from the original HTML (with the 6-card
// grid including Coming Soon tiles) will land after the user uploads the
// remaining assets. Until then this gives a working, navigable page.

export default function ProductsPage() {
  const { toast } = useToast();

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

      <UpsellBanner
        dismissId="products-referral"
        title="Earn $5.99 for every friend who upgrades"
        message="Share your referral link and pocket store credit each time someone buys Lifetime Pro through it. No cap, paid automatically."
        imageData={{
          kind: 'image',
          width: 96,
          component: <img src="/static/images/showcase.png" alt="" />,
        }}
        primaryAction={{
          label: 'Open affiliate dashboard',
          accessibilityLabel: 'Open your Aura affiliate dashboard',
          href: '/account',
          onClick: () => toast.success('Opening affiliate dashboard...'),
        }}
        dismissButton={{ accessibilityLabel: 'Dismiss the referral offer' }}
      />

      <div className="featured-grid">
        {FEATURED.map((p) => (
          <FeaturedCard key={p.id} product={p} />
        ))}
      </div>
    </section>
  );
}
