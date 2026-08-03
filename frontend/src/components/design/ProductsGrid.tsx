import { motion } from "framer-motion";
import type { MouseEvent, ReactNode } from "react";
import { useDownloadModal } from "../DownloadModalContext";

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

type Product = {
  name: string;
  badge: string;
  description: string;
  features: string[];
  price: string;
  priceSuffix?: string;
  origPrice?: string;
  discount?: string;
  href?: string;
  ctaLabel: string;
};

const PRODUCTS: Product[] = [
  {
    name: "Starter",
    badge: "Free",
    description:
      "Perfect for casual gamers. Get started with basic optimization, RAM cleanup, and startup management at no cost.",
    features: [
      "Basic optimization",
      "RAM cleanup",
      "Startup manager",
      "Community support",
    ],
    price: "Free",
    ctaLabel: "Download Free →",
  },
  {
    name: "Lifetime Pro",
    badge: "Most Popular",
    description:
      "One-time payment, forever access. The complete optimization suite with advanced AI tuning and priority support.",
    features: [
      "Everything in Starter",
      "Advanced AI tuning",
      "FPS boost guarantee",
      "Priority 24/7 support",
      "Lifetime updates",
      "Multi-PC license",
    ],
    price: "$15",
    priceSuffix: "/one-time",
    origPrice: "$99.99",
    discount: "85% OFF",
    href: "/checkout?product=lifetime",
    ctaLabel: "Purchase Now →",
  },
  {
    name: "Team License",
    badge: "Teams",
    description:
      "For esports teams and gaming cafes. Manage multiple PCs with a team dashboard, white-label options, and API access.",
    features: [
      "5 PC licenses",
      "Team dashboard",
      "White-label option",
      "API access",
      "Dedicated manager",
    ],
    price: "$149.99",
    priceSuffix: "/one-time",
    origPrice: "$299.99",
    discount: "50% OFF",
    href: "/checkout?product=team",
    ctaLabel: "Purchase Now →",
  },
];

function handleGlow(e: MouseEvent<HTMLDivElement>) {
  const rect = e.currentTarget.getBoundingClientRect();
  e.currentTarget.style.setProperty(
    "--mx",
    ((e.clientX - rect.left) / rect.width) * 100 + "%"
  );
  e.currentTarget.style.setProperty(
    "--my",
    ((e.clientY - rect.top) / rect.height) * 100 + "%"
  );
}

function CardShell({
  product,
  children,
}: {
  product: Product;
  children: ReactNode;
}) {
  const { openModal } = useDownloadModal();

  if (product.href) {
    return (
      <a href={product.href} className="block group h-full cursor-pointer">
        {children}
      </a>
    );
  }
  return (
    <button
      type="button"
      onClick={openModal}
      className="block group h-full w-full text-left cursor-pointer"
    >
      {children}
    </button>
  );
}

export function ProductsGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-5">
      {PRODUCTS.map((product, i) => (
        <motion.div
          key={product.name}
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-60px" }}
          transition={{ duration: 0.55, delay: i * 0.05, ease: EASE }}
        >
          <CardShell product={product}>
            <div
              className="glass-card overflow-hidden h-full flex flex-col"
              onMouseMove={handleGlow}
            >
              <div className="relative aspect-video">
                <div className="relative w-full h-full bg-black/40 overflow-hidden">
                  <div
                    className="absolute inset-0 z-10 pointer-events-none"
                    style={{
                      backgroundImage:
                        "repeating-linear-gradient(0deg,transparent,transparent 2px,rgba(0,0,0,0.05) 2px,rgba(0,0,0,0.05) 4px)",
                    }}
                  />
                  <img
                    src="/static/images/newdashboard.png"
                    alt={product.name}
                    className="absolute inset-0 w-full h-full object-cover opacity-60"
                  />
                  <div
                    className="absolute inset-0 z-10 pointer-events-none"
                    style={{
                      background:
                        "linear-gradient(to top,rgba(0,0,0,0.65) 0%,transparent 60%)",
                    }}
                  />
                  <div className="absolute bottom-3 left-4 z-20 font-mono text-[9px] tracking-widest text-white/25 uppercase">
                    {product.name}
                  </div>
                </div>
              </div>
              <div className="p-6 flex flex-col flex-1">
                <div className="font-mono text-[9px] tracking-[0.18em] text-white/30 uppercase mb-2">
                  {product.badge}
                </div>
                <div className="font-display text-[26px] tracking-wider text-white mb-3">
                  {product.name}
                </div>
                <div className="flex flex-wrap gap-1.5 mb-3">
                  {product.features.map((feature) => (
                    <span
                      key={feature}
                      className="font-mono text-[9px] tracking-widest text-white/28 uppercase px-2 py-1 rounded-md bg-white/4 border border-white/6"
                    >
                      {feature}
                    </span>
                  ))}
                </div>
                <p className="text-[12px] text-white/35 leading-relaxed font-light mb-5 flex-1">
                  {product.description}
                </p>
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-display text-[24px] tracking-wider text-white">
                      {product.price}
                      {product.priceSuffix && (
                        <span className="font-mono text-[9px] text-white/30 tracking-widest ml-1.5">
                          {product.priceSuffix}
                        </span>
                      )}
                    </div>
                    {product.origPrice && (
                      <div className="font-mono text-[9px] tracking-widest text-white/25 mt-0.5">
                        <span className="line-through">{product.origPrice}</span>
                        <span className="ml-1.5 text-white/40">
                          {product.discount}
                        </span>
                      </div>
                    )}
                  </div>
                  <div className="px-4 py-2 bg-white/5 border border-white/10 text-white/75 text-[11px] font-semibold tracking-wider rounded-lg group-hover:bg-white group-hover:text-z-black group-hover:border-transparent transition-all duration-300 whitespace-nowrap">
                    {product.ctaLabel}
                  </div>
                </div>
              </div>
            </div>
          </CardShell>
        </motion.div>
      ))}
    </div>
  );
}
