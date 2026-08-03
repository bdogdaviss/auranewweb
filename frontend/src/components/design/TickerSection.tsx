const phrases = [
  "System Tweaks",
  "Network Stack",
  "GPU Pipeline",
  "Macro Engine",
  "Real-Time Monitor",
  "AI Tuning",
  "FPS Boost Guarantee",
  "Safe & Reversible",
];

const maskGradient =
  "linear-gradient(to right, transparent, black 10%, black 90%, transparent)";

export function TickerSection() {
  return (
    <div
      className="relative overflow-hidden border-y border-white/5 py-3"
      style={{ maskImage: maskGradient, WebkitMaskImage: maskGradient }}
    >
      <div className="flex w-max animate-ticker">
        {[...phrases, ...phrases].map((phrase, index) => (
          <span
            key={index}
            className="inline-flex items-center gap-4 px-8 font-mono text-[9px] tracking-[0.2em] text-white/20 uppercase whitespace-nowrap"
          >
            {phrase}
            <span className="text-white/8">◆</span>
          </span>
        ))}
      </div>
    </div>
  );
}
