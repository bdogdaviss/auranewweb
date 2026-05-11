import type { ReactNode } from 'react';
import { motion } from 'framer-motion';

// Lamp — section divider with a cyan conic-gradient beam that shines down
// from a thin horizontal "filament" line. Each piece (the two cones, the
// soft blob, the filament) animates from a narrow→wide width when scrolled
// into view, so the lamp "ignites" on entry.
//
// Adapted from Aceternity UI's LampContainer (Tailwind + motion/react) →
// plain CSS in theme.css under "Lamp". The masks/blurs that hide the
// bottom of each cone are also plain divs with mask-image gradients.

const beamTransition = { delay: 0.3, duration: 0.8, ease: 'easeInOut' as const };

export default function Lamp({ children }: { children?: ReactNode }) {
  return (
    <section className="lamp-section">
      <div className="lamp-stage">
        {/* Left cone */}
        <motion.div
          initial={{ opacity: 0.5, width: '15rem' }}
          whileInView={{ opacity: 1, width: '30rem' }}
          viewport={{ once: true }}
          transition={beamTransition}
          className="lamp-cone lamp-cone-left"
        >
          <div className="lamp-cone-mask-bottom" />
          <div className="lamp-cone-mask-side lamp-cone-mask-side-left" />
        </motion.div>

        {/* Right cone (mirrored) */}
        <motion.div
          initial={{ opacity: 0.5, width: '15rem' }}
          whileInView={{ opacity: 1, width: '30rem' }}
          viewport={{ once: true }}
          transition={beamTransition}
          className="lamp-cone lamp-cone-right"
        >
          <div className="lamp-cone-mask-side lamp-cone-mask-side-right" />
          <div className="lamp-cone-mask-bottom lamp-cone-mask-bottom-right" />
        </motion.div>

        {/* Bottom darkener — hides cone bleed below the filament line */}
        <div className="lamp-bottom-blur" />
        {/* Glass-like blur band sitting right under the filament */}
        <div className="lamp-glass" />
        {/* Soft cyan halo */}
        <div className="lamp-glow" />

        {/* Cyan blob behind filament, grows on entry */}
        <motion.div
          initial={{ width: '8rem' }}
          whileInView={{ width: '16rem' }}
          viewport={{ once: true }}
          transition={beamTransition}
          className="lamp-blob"
        />

        {/* The filament line itself */}
        <motion.div
          initial={{ width: '15rem' }}
          whileInView={{ width: '30rem' }}
          viewport={{ once: true }}
          transition={beamTransition}
          className="lamp-line"
        />

        {/* Top cover — hides the half of the cones above the filament */}
        <div className="lamp-top-cover" />
      </div>

      {children && <div className="lamp-children">{children}</div>}
    </section>
  );
}
