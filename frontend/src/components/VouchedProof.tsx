import { useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

// VouchedProof — DM screenshot with a hover-card popup. Small thumbnail
// stays in the row; hovering it pops a 600px-wide spring-scaled overlay
// above so the DM is actually readable. Click → full-size in a new tab.
//
// Inspired by animate-ui's HoverCard pattern (radix-ui + motion) but
// implemented locally with framer-motion to avoid the radix-ui install.

const IMG = '/static/images/joshy.png';

export default function VouchedProof() {
  const [hovered, setHovered] = useState(false);

  return (
    <div
      className="vouched-proof-wrap"
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
    >
      <a
        href={IMG}
        target="_blank"
        rel="noopener noreferrer"
        className="vouched-proof"
        aria-label="Open full-size DM screenshot from Joshreyli"
        aria-describedby={hovered ? 'vouched-proof-popup' : undefined}
      >
        <img src={IMG} alt="DM from Joshreyli vouching for Aura Optimizer" />
      </a>

      <AnimatePresence>
        {hovered && (
          <motion.div
            id="vouched-proof-popup"
            role="tooltip"
            className="vouched-proof-popup"
            initial={{ opacity: 0, scale: 0.85, y: 8 }}
            animate={{ opacity: 1, scale: 1, y: 0 }}
            exit={{ opacity: 0, scale: 0.9, y: 6 }}
            transition={{ type: 'spring', stiffness: 300, damping: 25 }}
          >
            <img src={IMG} alt="" />
          </motion.div>
        )}
      </AnimatePresence>
    </div>
  );
}
