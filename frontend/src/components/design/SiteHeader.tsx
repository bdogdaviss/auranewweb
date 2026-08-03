import { Link, useLocation } from "react-router-dom";
import { motion } from "framer-motion";
import type { ComponentType, SVGProps } from "react";
import {
  ContactIcon,
  DiscordIcon,
  GiftNavIcon,
  HomeIcon,
  InfoIcon,
  ProductsIcon,
} from "./icons";

type NavItem = {
  label: string;
  href: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
};

const NAV_ITEMS: NavItem[] = [
  { label: "Home", href: "/", icon: HomeIcon },
  { label: "Products", href: "/products", icon: ProductsIcon },
  { label: "About", href: "/about", icon: InfoIcon },
  { label: "Affiliate", href: "/affiliate", icon: GiftNavIcon },
  { label: "Contact", href: "/contact", icon: ContactIcon },
];

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

const NAV_ITEM_STYLE = {
  padding: "7px 18px",
  transition: "padding 0.4s cubic-bezier(0.22,1,0.36,1), color 0.3s",
} as const;

export function SiteHeader() {
  const { pathname } = useLocation();

  return (
    <header className="fixed top-0 left-0 right-0 z-50 px-4 sm:px-6 pt-4 sm:pt-5 pointer-events-none">
      <div className="flex items-center justify-between pointer-events-auto">
        {/* Logo (left) */}
        <motion.div
          className="shrink-0"
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: EASE }}
        >
          <Link
            to="/"
            className="flex items-center justify-center w-10 h-10 sm:w-11 sm:h-11 rounded-full overflow-hidden border border-white/10 bg-white/4 backdrop-blur-xl hover:border-white/20 hover:bg-white/8 transition-all duration-300"
          >
            <img
              src="/aura-icon.png"
              alt="Aura Optimizer"
              className="w-6 h-6 object-contain"
            />
          </Link>
        </motion.div>

        {/* Center pill nav, absolutely centered */}
        <div className="absolute left-1/2 -translate-x-1/2">
          <motion.div
            className="flex items-center gap-0.5 px-1.5 py-1.5 rounded-full transition-all duration-500 bg-white/3 backdrop-blur-xl border border-white/6"
            initial={{ opacity: 0, y: -8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: EASE, delay: 0.05 }}
          >
            {NAV_ITEMS.map((item) => {
              const isActive =
                item.href === "/"
                  ? pathname === "/"
                  : pathname.startsWith(item.href);
              const Icon = item.icon;

              return (
                <Link
                  key={item.href}
                  to={item.href}
                  className={`relative flex items-center gap-2 rounded-full text-[13px] font-medium tracking-wide transition-colors duration-300 overflow-hidden ${
                    isActive ? "text-white" : "text-white/45 hover:text-white/75"
                  }`}
                  style={NAV_ITEM_STYLE}
                >
                  {isActive && (
                    <div className="absolute inset-0 rounded-full bg-white/9 border border-white/10" />
                  )}
                  <span className="relative z-10 flex items-center shrink-0">
                    <Icon />
                  </span>
                  <span className="relative z-10 overflow-hidden whitespace-nowrap hidden sm:block">
                    {item.label}
                  </span>
                </Link>
              );
            })}
          </motion.div>
        </div>

        {/* Discord CTA (right) */}
        <motion.div
          className="shrink-0"
          initial={{ opacity: 0, y: -8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: EASE, delay: 0.1 }}
        >
          <a
            href="https://discord.com/invite/4TBUw4nBFd"
            target="_blank"
            rel="noreferrer"
            className="flex items-center gap-2.5 rounded-full text-white font-semibold tracking-wide transition-all duration-300 hover:shadow-[0_0_28px_rgba(88,101,242,0.5)] hover:-translate-y-0.5 backdrop-blur-xl border border-white/8"
            style={{ background: "#5865F2", padding: "8px 10px" }}
          >
            <DiscordIcon height={15} width={15} />
            <span className="hidden sm:block text-[13px] pr-1">Discord</span>
          </a>
        </motion.div>
      </div>
    </header>
  );
}
