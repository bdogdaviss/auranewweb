import { motion } from 'framer-motion';
import { ProductsGrid } from '../components/design/ProductsGrid';
import UpsellBanner from '../components/UpsellBanner';
import { useToast } from '../components/ToastContext';

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

export default function ProductsPage() {
  const { toast } = useToast();

  return (
    <>
      <section className="pt-36 pb-8 px-6 md:px-12 max-w-7xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: EASE }}
          className="font-mono text-[11px] tracking-[0.2em] text-white/40 uppercase mb-2"
        >
          premium products
        </motion.div>
      </section>

      <div className="max-w-7xl mx-auto px-6 md:px-12">
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
      </div>

      <section className="pb-28 px-6 md:px-12 max-w-7xl mx-auto pt-8">
        <ProductsGrid />
      </section>
    </>
  );
}
