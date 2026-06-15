import type { Product } from '../data/products';
import Carousel from './Carousel';
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

type Props = {
  product: Product;
};

export default function FeaturedCard({ product }: Props) {
  const usesCarousel = product.images.length > 1;
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
    </>
  );

  return (
    <article className={`featured-card featured-card--${product.id}`}>
      {inner}
    </article>
  );
}
