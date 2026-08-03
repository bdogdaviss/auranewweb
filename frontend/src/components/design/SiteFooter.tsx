import { DiscordIcon } from "./icons";

type FooterLinkGroup = {
  heading: string;
  links: { label: string; href: string; external?: boolean }[];
};

const linkGroups: FooterLinkGroup[] = [
  {
    heading: "Products",
    links: [
      { label: "Starter", href: "/download" },
      { label: "Lifetime Pro", href: "/checkout?product=lifetime" },
      { label: "Team License", href: "/checkout?product=team" },
    ],
  },
  {
    heading: "Support",
    links: [
      {
        label: "Discord",
        href: "https://discord.com/invite/4TBUw4nBFd",
        external: true,
      },
      { label: "Documentation", href: "#" },
      { label: "FAQ", href: "#" },
    ],
  },
  {
    heading: "Legal",
    links: [
      { label: "Terms of Service", href: "#" },
      { label: "Privacy Policy", href: "#" },
      { label: "Refund Policy", href: "#" },
    ],
  },
];

export function SiteFooter() {
  return (
    <footer className="relative z-10 mt-16">
      <div className="h-px bg-linear-to-r from-transparent via-white/8 to-transparent" />
      <div className="max-w-7xl mx-auto px-6 md:px-12 py-14">
        <div className="grid grid-cols-2 md:grid-cols-5 gap-10 mb-12">
          <div className="col-span-2 md:col-span-2">
            <div className="flex items-center gap-2.5 mb-4">
              <img
                src="/aura-icon.png"
                alt="Aura Optimizer"
                className="w-6 h-6 object-contain"
              />
              <span className="font-display text-[20px] tracking-widest text-white/70">
                AURA OPTIMIZER
              </span>
            </div>
            <p className="font-mono text-[10px] tracking-widest text-white/20 leading-relaxed uppercase max-w-60">
              Boost your performance.
              <br />
              Gain the edge.
            </p>
            <a
              href="https://discord.com/invite/4TBUw4nBFd"
              target="_blank"
              rel="noreferrer"
              className="inline-flex items-center gap-2 mt-5 px-4 py-2 rounded-full border text-[11px] font-semibold tracking-wide transition-all duration-300"
              style={{
                background: "rgba(88,101,242,0.2)",
                borderColor: "rgba(88,101,242,0.3)",
                color: "#8b9cf5",
              }}
            >
              <DiscordIcon height={12} width={12} />
              Join Discord
            </a>
          </div>
          {linkGroups.map((group) => (
            <div key={group.heading}>
              <div className="font-mono text-[9px] tracking-[0.18em] text-white/25 uppercase mb-4">
                {group.heading}
              </div>
              <ul className="flex flex-col gap-3">
                {group.links.map((link) => (
                  <li key={link.label}>
                    <a
                      href={link.href}
                      {...(link.external
                        ? { target: "_blank", rel: "noreferrer" }
                        : {})}
                      className="font-mono text-[10px] tracking-widest text-white/35 hover:text-white/65 uppercase transition-colors duration-200"
                    >
                      {link.label}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
        <div className="h-px bg-linear-to-r from-transparent via-white/5 to-transparent mb-6" />
        <div className="flex flex-col sm:flex-row items-center justify-between gap-3">
          <span className="font-mono text-[9px] tracking-[0.12em] text-white/18 uppercase">
            © 2026 Aura Optimizer. All rights reserved.
          </span>
          <span className="font-mono text-[9px] tracking-[0.12em] text-white/18 uppercase">
            <a
              href="https://discord.com/invite/4TBUw4nBFd"
              target="_blank"
              rel="noreferrer"
              className="hover:text-white/40 transition-colors duration-200"
            >
              Powered by Discord
            </a>
          </span>
        </div>
      </div>
    </footer>
  );
}
