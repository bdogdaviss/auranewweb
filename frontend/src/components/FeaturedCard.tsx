import { motion } from 'framer-motion';
import type { Product } from '../data/products';
import Carousel from './Carousel';
import { useReveal } from '../hooks/useReveal';
import { useDownloadModal } from './DownloadModalContext';

function CartIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <circle cx="9" cy="21" r="1" />
      <circle cx="20" cy="21" r="1" />
      <path d="M1 1h4l2.68 13.39a2 2 0 002 1.61h9.72a2 2 0 002-1.61L23 6H6" />
    </svg>
  );
}

// FlyonUI-style intersect animation for the highlighted card:
// scale+opacity+blur fade-in from bottom-right with smooth spring, and the
// CTA button scales in with a 400ms delayed bouncier spring.
const lifetimeCardVariants = {
  hidden: { scale: 0, opacity: 0, filter: 'blur(5px)' },
  visible: {
    scale: 1,
    opacity: 1,
    filter: 'blur(0px)',
    transition: {
      type: 'spring' as const,
      damping: 18,
      stiffness: 200,
      when: 'beforeChildren' as const,
    },
  },
};

const lifetimeButtonVariants = {
  hidden: { scale: 0 },
  visible: {
    scale: 1,
    transition: { type: 'spring' as const, damping: 10, stiffness: 300, delay: 0.4 },
  },
};

type Props = {
  product: Product;
};

export default function FeaturedCard({ product }: Props) {
  const usesCarousel = product.images.length > 1;
  const ref = useReveal<HTMLElement>();
  const isLifetime = product.id === 'lifetime';
  const isFreeDownload = product.id === 'free';
  const { openModal } = useDownloadModal();

  // Intercept the click for the free Starter product so it opens the
  // shared confirmation modal instead of navigating to /download.
  const handleCtaClick = (e: React.MouseEvent) => {
    if (isFreeDownload) {
      e.preventDefault();
      openModal();
    }
  };

  const inner = (
    <>
      {product.badge && <span className="card-badge">{product.badge}</span>}

      {usesCarousel ? (
        <Carousel slides={product.images} alt={product.name} />
      ) : (
        <div className="card-image-wrap placeholder-active">
          {product.images[0] && (
            <img
              src={product.images[0]}
              alt={product.name}
              className="card-image"
              onError={(e) => {
                (e.currentTarget as HTMLImageElement).style.display = 'none';
              }}
            />
          )}
        </div>
      )}

      <h3 className="card-name">{product.name}</h3>
      <p className="card-desc">{product.description}</p>

      {product.originalPriceLabel ? (
        <div className="card-price sale-price">
          <span className="price-original">{product.originalPriceLabel}</span>
          <span className="price-current">{product.priceLabel}</span>
          {product.saleTag && <span className="price-sale-tag">{product.saleTag}</span>}
        </div>
      ) : (
        <div className="card-price">{product.priceLabel}</div>
      )}

      <ul className="card-features has-dividers">
        {product.features.map((f) => (
          <li key={f}>{f}</li>
        ))}
      </ul>

      {isLifetime ? (
        <motion.a
          href={product.cta.href}
          className="card-purchase-btn"
          variants={lifetimeButtonVariants}
          {...(product.cta.external
            ? { target: '_blank', rel: 'noopener noreferrer' }
            : {})}
        >
          <CartIcon /> {product.cta.label}
        </motion.a>
      ) : (
        <a
          href={product.cta.href}
          className="card-purchase-btn"
          onClick={handleCtaClick}
          {...(product.cta.external
            ? { target: '_blank', rel: 'noopener noreferrer' }
            : {})}
        >
          <CartIcon /> {product.cta.label}
        </a>
      )}
    </>
  );

  if (isLifetime) {
    return (
      <motion.article
        className={`featured-card featured-card--${product.id}`}
        style={{ transformOrigin: 'bottom right' }}
        variants={lifetimeCardVariants}
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true, amount: 0.5 }}
      >
        {inner}
      </motion.article>
    );
  }

  return (
    <article
      ref={ref}
      className={`featured-card featured-card--${product.id} reveal-on-scroll`}
    >
      {inner}
    </article>
  );
}
