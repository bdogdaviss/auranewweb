import { motion } from "framer-motion";
import { useDownloadModal } from "../DownloadModalContext";

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

interface MiniStat {
  value: string;
  label: string;
}

interface ProductStripItem {
  category: string;
  name: string;
}

const MINI_STATS: MiniStat[] = [
  { value: "2,000+", label: "Active Users" },
  { value: "24/7", label: "Priority Support" },
  { value: "85%", label: "Lifetime Pro Off" },
];

const PRODUCT_STRIP: ProductStripItem[] = [
  { category: "Free", name: "Starter" },
  { category: "Most Popular", name: "Lifetime Pro" },
  { category: "Teams", name: "Team License" },
];

const TRAFFIC_LIGHTS = ["#ff5f57", "#febc2e", "#28c840"];

export function HeroSection() {
  const { openModal } = useDownloadModal();

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center pt-32 pb-20 px-6 md:px-12 overflow-hidden text-center">
      <div className="absolute left-0 right-0 top-1/3 h-px bg-linear-to-r from-transparent via-white/4 to-transparent pointer-events-none" />
      <div className="absolute left-0 right-0 bottom-1/3 h-px bg-linear-to-r from-transparent via-white/4 to-transparent pointer-events-none" />
      <div className="absolute top-24 left-8 hidden lg:block">
        <div className="w-5 h-5 border-l border-t border-white/10" />
      </div>
      <div className="absolute top-24 right-8 hidden lg:block">
        <div className="w-5 h-5 border-r border-t border-white/10" />
      </div>
      <div className="absolute bottom-16 left-8 hidden lg:block">
        <div className="w-5 h-5 border-l border-b border-white/10" />
      </div>
      <div className="absolute bottom-16 right-8 hidden lg:block">
        <div className="w-5 h-5 border-r border-b border-white/10" />
      </div>

      <div className="relative z-10 w-full max-w-5xl mx-auto flex flex-col items-center">
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0, ease: EASE }}
          className="inline-flex items-center gap-2.5 px-4 py-1.5 rounded-full border border-white/8 bg-white/4 backdrop-blur-sm mb-8"
        >
          <span className="relative flex h-1.5 w-1.5">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-white opacity-25" />
            <span className="relative inline-flex rounded-full h-1.5 w-1.5 bg-white/50" />
          </span>
          <span className="font-mono text-[10px] tracking-[0.15em] text-white/40 uppercase">
            2,000+ Active Users
          </span>
        </motion.div>

        <h1 className="font-display text-[clamp(40px,8vw,128px)] sm:text-[clamp(56px,10vw,128px)] leading-[0.9] sm:leading-[0.88] tracking-[0.02em] mb-6 sm:mb-8 text-center">
          <motion.span
            initial={{ opacity: 0, y: 16, filter: "blur(8px)" }}
            animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
            transition={{ duration: 0.7, delay: 0.08, ease: EASE }}
            className="text-white block"
          >
            ELITE GAMING
          </motion.span>
          <motion.span
            initial={{ opacity: 0, y: 16, filter: "blur(8px)" }}
            animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
            transition={{ duration: 0.7, delay: 0.18, ease: EASE }}
            className="text-white/60 block"
          >
            PERFORMANCE
          </motion.span>
        </h1>

        <motion.p
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.55, delay: 0.3, ease: EASE }}
          className="text-white/40 text-[15px] font-light leading-relaxed max-w-lg mb-10"
        >
          Professional-grade optimization tools engineered to deliver
          measurable competitive advantage. System tweaks, network tuning, GPU
          pipeline — one click, fully reversible.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.55, delay: 0.38, ease: EASE }}
          className="flex flex-wrap gap-3 justify-center mb-14"
        >
          <a
            href="/checkout?product=lifetime"
            className="px-7 py-3.5 bg-white text-black text-[13px] font-semibold tracking-wider rounded-xl transition-all duration-300 hover:shadow-[0_0_50px_rgba(255,255,255,0.18)] hover:-translate-y-0.5"
          >
            Get Lifetime Pro
          </a>
          <button
            type="button"
            onClick={openModal}
            className="inline-flex items-center gap-2 px-7 py-3.5 bg-white/4 border border-white/10 text-white/70 text-[13px] font-medium tracking-wider rounded-xl transition-all duration-300 hover:bg-white/8 hover:border-white/20 hover:text-white hover:-translate-y-0.5 cursor-pointer"
          >
            Download Now
          </button>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ duration: 0.6, delay: 0.46, ease: EASE }}
          className="flex gap-10 items-center justify-center mb-16"
        >
          {MINI_STATS.map((stat) => (
            <div key={stat.label} className="flex flex-col items-center">
              <span className="font-display text-[28px] tracking-wider text-white">
                {stat.value}
              </span>
              <span className="font-mono text-[9px] tracking-[0.18em] text-white/30 uppercase mt-0.5">
                {stat.label}
              </span>
            </div>
          ))}
        </motion.div>

        <motion.div
          initial={{ opacity: 0, y: 24 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.8, delay: 0.55, ease: EASE }}
          className="relative w-full max-w-4xl mx-auto"
        >
          <div className="absolute -inset-6 rounded-3xl bg-white/[0.012] blur-2xl pointer-events-none" />
          <div className="relative rounded-2xl overflow-hidden border border-white/8 bg-black/60 backdrop-blur-sm shadow-[0_40px_120px_rgba(0,0,0,0.8)]">
            <div className="flex items-center justify-between px-4 py-3 border-b border-white/6 bg-white/2">
              <div className="flex gap-1.5">
                {TRAFFIC_LIGHTS.map((color) => (
                  <div
                    key={color}
                    className="w-2.5 h-2.5 rounded-full"
                    style={{ background: color, opacity: 0.6 }}
                  />
                ))}
              </div>
              <div className="font-mono text-[10px] tracking-widest text-white/20">
                AURA OPTIMIZER
              </div>
              <div className="w-12" />
            </div>
            <div className="relative aspect-video bg-black overflow-hidden">
              <video
                src="/static/videos/aura-showcase.mp4"
                autoPlay
                loop
                muted
                playsInline
                className="w-full h-full object-cover"
              />
              <div
                className="absolute inset-0 pointer-events-none"
                style={{
                  backgroundImage:
                    "repeating-linear-gradient(0deg, transparent, transparent 2px, rgba(0,0,0,0.06) 2px, rgba(0,0,0,0.06) 4px)",
                }}
              />
            </div>
            <div className="grid grid-cols-3 divide-x divide-white/6 border-t border-white/6">
              {PRODUCT_STRIP.map((item) => (
                <div
                  key={item.name}
                  className="px-5 py-3.5 text-center bg-white/1"
                >
                  <div className="font-mono text-[8px] tracking-[0.18em] text-white/25 uppercase mb-0.5">
                    {item.category}
                  </div>
                  <div className="font-display text-[18px] tracking-wider text-white">
                    {item.name}
                  </div>
                </div>
              ))}
            </div>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
