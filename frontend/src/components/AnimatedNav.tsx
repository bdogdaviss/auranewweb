import { useRef, useState } from 'react';
import { motion, useScroll, useMotionValueEvent } from 'framer-motion';
import { Link, useLocation } from 'react-router-dom';

// AnimatedNav — pill-shaped top nav that auto-collapses to a circle on
// scroll-down and re-expands on scroll-up (or click). Spring physics from
// framer-motion; plain CSS classes in theme.css under "AnimatedNav".
//
// Adapted from a Tailwind/shadcn original; the icon imports (Navigation,
// Menu from lucide-react) are inlined as small SVGs so we don't pull in
// react-icons or lucide just for two glyphs.

type NavItem = {
  name: string;
  href: string;
  external?: boolean;
};

const NAV_ITEMS: NavItem[] = [
  { name: 'Home', href: '/' },
  { name: 'Products', href: '/products' },
  { name: 'About', href: '/about' },
  { name: 'Discord', href: 'https://discord.com/invite/4TBUw4nBFd', external: true },
];

const EXPAND_SCROLL_THRESHOLD = 80;

const containerVariants = {
  expanded: {
    y: 0,
    opacity: 1,
    width: 'auto',
    transition: {
      type: 'spring' as const,
      damping: 20,
      stiffness: 300,
      staggerChildren: 0.07,
      delayChildren: 0.2,
    },
  },
  collapsed: {
    y: 0,
    opacity: 1,
    width: '3rem',
    transition: {
      type: 'spring' as const,
      damping: 20,
      stiffness: 300,
      when: 'afterChildren' as const,
      staggerChildren: 0.05,
      staggerDirection: -1,
    },
  },
};

const logoVariants = {
  expanded: { opacity: 1, x: 0, rotate: 0, transition: { type: 'spring' as const, damping: 15 } },
  collapsed: { opacity: 0, x: -25, rotate: -180, transition: { duration: 0.3 } },
};

const itemVariants = {
  expanded: { opacity: 1, x: 0, scale: 1, transition: { type: 'spring' as const, damping: 15 } },
  collapsed: { opacity: 0, x: -20, scale: 0.95, transition: { duration: 0.2 } },
};

const collapsedIconVariants = {
  expanded: { opacity: 0, scale: 0.8, transition: { duration: 0.2 } },
  collapsed: {
    opacity: 1,
    scale: 1,
    transition: { type: 'spring' as const, damping: 15, stiffness: 300, delay: 0.15 },
  },
};

function NavigationGlyph() {
  return (
    <svg
      className="anav-icon"
      viewBox="0 0 24 24"
      width="22"
      height="22"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <polygon points="3 11 22 2 13 21 11 13 3 11" />
    </svg>
  );
}

function MenuGlyph() {
  return (
    <svg
      className="anav-icon"
      viewBox="0 0 24 24"
      width="22"
      height="22"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <line x1="4" y1="6" x2="20" y2="6" />
      <line x1="4" y1="12" x2="20" y2="12" />
      <line x1="4" y1="18" x2="20" y2="18" />
    </svg>
  );
}

export default function AnimatedNav() {
  const [isExpanded, setExpanded] = useState(true);
  const { scrollY } = useScroll();
  const lastScrollY = useRef(0);
  const scrollPositionOnCollapse = useRef(0);
  const location = useLocation();

  useMotionValueEvent(scrollY, 'change', (latest) => {
    const previous = lastScrollY.current;

    if (isExpanded && latest > previous && latest > 150) {
      setExpanded(false);
      scrollPositionOnCollapse.current = latest;
    } else if (
      !isExpanded &&
      latest < previous &&
      scrollPositionOnCollapse.current - latest > EXPAND_SCROLL_THRESHOLD
    ) {
      setExpanded(true);
    }

    lastScrollY.current = latest;
  });

  const handleNavClick = (e: React.MouseEvent) => {
    if (!isExpanded) {
      e.preventDefault();
      setExpanded(true);
    }
  };

  return (
    <div className="anav-fixed">
      <motion.nav
        initial={{ y: -80, opacity: 0 }}
        animate={isExpanded ? 'expanded' : 'collapsed'}
        variants={containerVariants}
        whileHover={!isExpanded ? { scale: 1.1 } : {}}
        whileTap={!isExpanded ? { scale: 0.95 } : {}}
        onClick={handleNavClick}
        className={`anav ${!isExpanded ? 'anav-collapsed' : ''}`.trim()}
      >
        <motion.div variants={logoVariants} className="anav-logo">
          <NavigationGlyph />
        </motion.div>

        <motion.div className={`anav-items ${!isExpanded ? 'anav-items-locked' : ''}`.trim()}>
          {NAV_ITEMS.map((item) => {
            const isActive = !item.external && location.pathname === item.href;
            const className = `anav-link ${isActive ? 'anav-link-active' : ''}`.trim();
            if (item.external) {
              return (
                <motion.a
                  key={item.name}
                  href={item.href}
                  variants={itemVariants}
                  onClick={(e) => e.stopPropagation()}
                  className={className}
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {item.name}
                </motion.a>
              );
            }
            return (
              <motion.span key={item.name} variants={itemVariants} className="anav-link-wrap">
                <Link
                  to={item.href}
                  className={className}
                  onClick={(e) => e.stopPropagation()}
                >
                  {item.name}
                </Link>
              </motion.span>
            );
          })}
        </motion.div>

        <div className="anav-collapsed-icon">
          <motion.div
            variants={collapsedIconVariants}
            animate={isExpanded ? 'expanded' : 'collapsed'}
          >
            <MenuGlyph />
          </motion.div>
        </div>
      </motion.nav>
    </div>
  );
}
