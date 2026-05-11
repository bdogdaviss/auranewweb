// Decorative full-page background: image tint, overlay, grid pattern, and
// floating particles. Pure CSS animation, no JS loops.

const PARTICLES = [
  { top: '12%', left: '8%', size: 2, duration: 18, delay: 0 },
  { top: '25%', left: '85%', size: 3, duration: 22, delay: 3 },
  { top: '60%', left: '20%', size: 2, duration: 15, delay: 7 },
  { top: '45%', left: '70%', size: 4, duration: 20, delay: 2 },
  { top: '80%', left: '45%', size: 2, duration: 24, delay: 10 },
  { top: '15%', left: '55%', size: 3, duration: 17, delay: 5 },
  { top: '70%', left: '90%', size: 2, duration: 21, delay: 15 },
  { top: '35%', left: '35%', size: 3, duration: 19, delay: 20 },
];

export default function Background() {
  return (
    <>
      <div className="bg-image" aria-hidden />
      <div className="bg-overlay" aria-hidden />
      <div className="particles" aria-hidden>
        {PARTICLES.map((p, i) => (
          <span
            key={i}
            className="particle"
            style={{
              top: p.top,
              left: p.left,
              width: p.size,
              height: p.size,
              animationDuration: `${p.duration}s`,
              animationDelay: `${p.delay}s`,
            }}
          />
        ))}
      </div>
    </>
  );
}
