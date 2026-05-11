import { useState } from 'react';
import { motion } from 'framer-motion';

// BackgroundRippleEffect — grid of cells that fires a ripple from the
// clicked cell outward. Each cell's animation delay is proportional to its
// Manhattan-ish distance from the click origin, so the wave propagates.
//
// Adapted from Aceternity UI (Tailwind + framer-motion) → plain CSS classes
// in theme.css under "BackgroundRippleEffect". Pure decoration: the only
// interaction is click-to-ripple.

const ROWS = 12;
const COLS = 24;
const TOTAL = ROWS * COLS;

type Origin = { r: number; c: number; key: number };

export default function BackgroundRippleEffect() {
  const [origin, setOrigin] = useState<Origin | null>(null);

  const fireRipple = (r: number, c: number) => {
    setOrigin((prev) => ({ r, c, key: (prev?.key ?? 0) + 1 }));
  };

  return (
    <div className="bg-ripple" aria-hidden>
      <div
        className="bg-ripple-grid"
        style={
          {
            ['--cols' as string]: COLS,
            ['--rows' as string]: ROWS,
          } as React.CSSProperties
        }
      >
        {Array.from({ length: TOTAL }).map((_, i) => {
          const r = Math.floor(i / COLS);
          const c = i % COLS;
          let delay = 0;
          if (origin) {
            const dx = c - origin.c;
            const dy = r - origin.r;
            const distance = Math.sqrt(dx * dx + dy * dy);
            delay = distance * 0.06;
          }
          return (
            <motion.button
              // No key trick — framer-motion replays the keyframe animation
              // when the transition object's identity changes (origin.key
              // forces a new transition object every click).
              key={i}
              type="button"
              className="bg-ripple-cell"
              onClick={() => fireRipple(r, c)}
              initial={false}
              animate={
                origin
                  ? {
                      backgroundColor: [
                        'rgba(255,255,255,0)',
                        'rgba(255,255,255,0.18)',
                        'rgba(255,255,255,0)',
                      ],
                    }
                  : { backgroundColor: 'rgba(255,255,255,0)' }
              }
              transition={{
                duration: 0.9,
                delay,
                times: [0, 0.4, 1],
                ease: 'easeOut',
                // Tag transition with the click id so framer-motion sees a
                // new transition each click and replays the keyframes.
                ...(origin ? { _replayKey: origin.key } : {}),
              } as object}
              aria-hidden
              tabIndex={-1}
            />
          );
        })}
      </div>
    </div>
  );
}
