// Source of truth for product display data on the marketing pages.
// Note: the prices shown here are display-only — the backend's main.go Products
// array is the authoritative source for charge amounts. Keep in sync.

export type Product = {
  id: string;
  name: string;
  description: string;
  badge?: string;
  priceLabel: string;
  originalPriceLabel?: string;
  saleTag?: string;
  features: string[];
  cta: {
    label: string;
    href: string;
    external?: boolean;
  };
  images: string[];
  comingSoon?: boolean;
};

export const FEATURED: Product[] = [
  {
    id: 'free',
    name: 'Starter',
    description:
      'Perfect for casual gamers. Get started with basic optimization, RAM cleanup, and startup management at no cost.',
    badge: 'Free',
    priceLabel: 'Free',
    features: ['Basic optimization', 'RAM cleanup', 'Startup manager', 'Community support'],
    cta: { label: 'Download Free', href: '/download' },
    images: ['/static/images/newdashboard.png'],
  },
  {
    id: 'lifetime',
    name: 'Lifetime Pro',
    description:
      'One-time payment, forever access. The complete optimization suite with advanced AI tuning and priority support.',
    badge: 'Most Popular',
    priceLabel: '$15',
    originalPriceLabel: '$99.99',
    saleTag: '85% OFF',
    features: [
      'Everything in Starter',
      'Advanced AI tuning',
      'FPS boost guarantee',
      'Priority 24/7 support',
      'Lifetime updates',
      'Multi-PC license',
    ],
    cta: { label: 'Purchase Now', href: '/checkout?product=lifetime' },
    images: [
      '/static/images/newdashboard.png',
      '/static/images/newnetwork.png',
      '/static/images/newoptimizer.png',
    ],
  },
  {
    id: 'team',
    name: 'Team License',
    description:
      'For esports teams and gaming cafes. Manage multiple PCs with a team dashboard, white-label options, and API access.',
    badge: 'Teams',
    priceLabel: '$149.99',
    originalPriceLabel: '$299.99',
    saleTag: '50% OFF',
    features: ['5 PC licenses', 'Team dashboard', 'White-label option', 'API access', 'Dedicated manager'],
    cta: { label: 'Purchase Now', href: '/checkout?product=team' },
    images: ['/static/images/newnetwork.png'],
  },
];
