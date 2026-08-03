import { motion } from "framer-motion";
import type { ComponentType, SVGProps } from "react";

import {
  ChipIcon,
  TerminalWindowIcon,
  WifiIcon,
  StackIcon,
  PulseIcon,
  ShieldCheckIcon,
} from "./icons";

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

const VIEWPORT = { once: true, margin: "-80px" } as const;

type Feature = {
  index: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  title: string;
  description: string;
};

const FEATURES: Feature[] = [
  {
    index: "01",
    icon: ChipIcon,
    title: "System Tweaks",
    description:
      "OS-level tuning that strips background noise from your boot, services, and scheduler so the game gets the CPU.",
  },
  {
    index: "02",
    icon: WifiIcon,
    title: "Network Stack",
    description:
      "Latency-first network tuning: route, MTU, TCP, and DNS configured for competitive ping consistency.",
  },
  {
    index: "03",
    icon: StackIcon,
    title: "GPU Pipeline",
    description:
      "Render-side tuning: driver flags, frame pacing, shader-cache placement, and power-state pinning to lock your FPS floor.",
  },
  {
    index: "04",
    icon: TerminalWindowIcon,
    title: "Macro Engine",
    description:
      "Precision keyboard + mouse automation with sub-millisecond timing, plus build profiles you can hot-swap per game.",
  },
  {
    index: "05",
    icon: PulseIcon,
    title: "Real-Time Monitor",
    description:
      "Live dashboard of FPS, latency, frametime variance, and CPU/GPU load so you actually see what your changes did.",
  },
  {
    index: "06",
    icon: ShieldCheckIcon,
    title: "Safe & Reversible",
    description:
      "Every tweak is snapshotted before it runs. One click to restore your prior state if a game breaks or you want stock back.",
  },
];

export function FeaturesSection() {
  return (
    <section className="relative z-10 py-28 px-6 md:px-12 max-w-7xl mx-auto">
      <div className="text-center mb-16">
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={VIEWPORT}
          transition={{ duration: 0.5, ease: EASE }}
          className="font-mono text-[10px] tracking-[0.2em] text-white/30 uppercase mb-4"
        >
          {"// what you get"}
        </motion.div>
        <motion.h2
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={VIEWPORT}
          transition={{ duration: 0.6, delay: 0.05, ease: EASE }}
          className="font-display text-[clamp(48px,7vw,88px)] leading-[0.9] tracking-[0.02em] text-white mb-5"
        >
          BUILT FOR THE
          <br />
          <span className="text-white/60">COMPETITIVE EDGE</span>
        </motion.h2>
        <motion.p
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={VIEWPORT}
          transition={{ duration: 0.55, delay: 0.1, ease: EASE }}
          className="text-white/35 text-[14px] font-light leading-relaxed max-w-md mx-auto"
        >
          Every millisecond between your brain and your screen is contested.
          Aura tunes the whole path — OS, network, GPU, input — so you reclaim
          the frames you paid for.
        </motion.p>
      </div>
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {FEATURES.map(({ index, icon: Icon, title, description }, i) => (
          <motion.div
            key={index}
            initial={{ opacity: 0, y: 24 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={VIEWPORT}
            transition={{ duration: 0.6, delay: i * 0.07, ease: EASE }}
            className="relative p-8 rounded-xl border border-white/8 bg-white/2 backdrop-blur-sm group transition-colors duration-300 hover:bg-white/4"
          >
            <div className="flex items-start justify-between mb-6">
              <div className="w-9 h-9 rounded-lg bg-white/5 border border-white/7 flex items-center justify-center">
                <Icon className="text-white/50 group-hover:text-white/70 transition-colors duration-300 text-[17px]" />
              </div>
              <span className="font-mono text-[11px] text-white/10 tracking-widest">
                {index}
              </span>
            </div>
            <h3 className="text-[15px] font-semibold text-white mb-3 tracking-tight">
              {title}
            </h3>
            <p className="text-[13px] text-white/35 leading-relaxed font-light">
              {description}
            </p>
            <div className="absolute bottom-0 left-0 right-0 h-px bg-linear-to-r from-transparent via-white/15 to-transparent scale-x-0 group-hover:scale-x-100 transition-transform duration-500 origin-center" />
          </motion.div>
        ))}
      </div>
    </section>
  );
}
