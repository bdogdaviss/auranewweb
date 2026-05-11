import type { ReactNode } from 'react';
import { motion } from 'framer-motion';

// Reveal — wraps a section and fades it up from opacity-0 + y:50 the first
// time it crosses ~20% into the viewport. Uses framer-motion's whileInView
// + viewport.once so the animation runs once per section. Matches the
// "sections appear from dark/nothing" pattern.

type RevealProps = {
  children: ReactNode;
  delay?: number;
  className?: string;
  // How much of the element needs to be visible before triggering (0–1).
  amount?: number;
};

export default function Reveal({ children, delay = 0, className, amount = 0.2 }: RevealProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 50 }}
      whileInView={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.6, delay, ease: 'easeOut' }}
      viewport={{ once: true, amount }}
      className={className}
    >
      {children}
    </motion.div>
  );
}
