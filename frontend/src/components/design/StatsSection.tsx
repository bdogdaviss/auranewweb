import { motion } from "framer-motion";

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

interface Stat {
  value: string;
  label: string;
  sublabel: string;
}

const STATS: Stat[] = [
  { value: "2K+", label: "Active Users", sublabel: "and growing daily" },
  { value: "−18ms", label: "Latency Cut", sublabel: "from system tweaks" },
  { value: "3", label: "Products", sublabel: "starter, pro, team" },
  { value: "85%", label: "Off Lifetime Pro", sublabel: "one-time payment" },
];

export function StatsSection() {
  return (
    <section className="relative z-10 py-16 px-6 md:px-12">
      <motion.div
        initial={{ opacity: 0, scaleX: 0.9 }}
        whileInView={{ opacity: 1, scaleX: 1 }}
        viewport={{ once: true, margin: "-80px" }}
        transition={{ duration: 0.6, ease: EASE }}
        className="max-w-7xl mx-auto rounded-2xl border border-white/[0.07] bg-white/2 backdrop-blur-sm overflow-hidden"
      >
        <div className="h-px bg-linear-to-r from-transparent via-white/20 to-transparent" />
        <div className="grid grid-cols-2 lg:grid-cols-4 divide-x divide-y lg:divide-y-0 divide-white/6">
          {STATS.map((stat, i) => (
            <motion.div
              key={stat.label}
              initial={{ opacity: 0, y: 12 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-80px" }}
              transition={{ duration: 0.5, delay: i * 0.08, ease: EASE }}
              className="px-8 py-8 text-center group hover:bg-white/2 transition-colors duration-300"
            >
              <div className="font-display text-[42px] tracking-wider text-white mb-1">
                {stat.value}
              </div>
              <div className="font-mono text-[9px] tracking-[0.18em] text-white/35 uppercase">
                {stat.label}
              </div>
              <div className="font-mono text-[9px] tracking-widest text-white/18 mt-0.5">
                {stat.sublabel}
              </div>
            </motion.div>
          ))}
        </div>
      </motion.div>
    </section>
  );
}
