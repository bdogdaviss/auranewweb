import { HeroSection } from '../components/design/HeroSection';
import { TickerSection } from '../components/design/TickerSection';
import { FeaturesSection } from '../components/design/FeaturesSection';
import { StatsSection } from '../components/design/StatsSection';
import UpsellBanner from '../components/UpsellBanner';

function CrownIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <path d="M2 7l5 5 5-7 5 7 5-5v11a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1V7z" />
    </svg>
  );
}

export default function HomePage() {
  return (
    <>
      <HeroSection />

      <div className="max-w-7xl mx-auto px-6 md:px-12">
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
      </div>

      <TickerSection />
      <FeaturesSection />
      <StatsSection />
    </>
  );
}
