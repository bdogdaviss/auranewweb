import { motion } from 'framer-motion';
import type { Variants } from 'framer-motion';

// FloatText — one-shot per-character entrance animation. Unlike ScrollFloat
// (which uses GSAP ScrollTrigger with scrub:true and reverses on scroll-up),
// this component fires on the first viewport intersection and stays
// animated. No mid-animation flicker when scrolling back up.

const containerVariants: Variants = {
  hidden: { opacity: 1 },
  visible: {
    opacity: 1,
    transition: { staggerChildren: 0.03 },
  },
};

const charVariants: Variants = {
  hidden: { opacity: 0, y: 28, scaleY: 2.3, scaleX: 0.7 },
  visible: {
    opacity: 1,
    y: 0,
    scaleY: 1,
    scaleX: 1,
    transition: { type: 'spring', damping: 14, stiffness: 140 },
  },
};

export default function FloatText({
  children,
  className = '',
  textClassName = '',
}: {
  children: string;
  className?: string;
  textClassName?: string;
}) {
  const chars = children.split('');

  return (
    <motion.h2
      className={`scroll-float ${className}`.trim()}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, amount: 0.4 }}
      variants={containerVariants}
    >
      <span className={`scroll-float-text ${textClassName}`.trim()}>
        {chars.map((char, i) => (
          <motion.span
            key={i}
            className="scroll-float-char"
            variants={charVariants}
            style={{
              display: 'inline-block',
              transformOrigin: '50% 0%',
              whiteSpace: char === ' ' ? 'pre' : undefined,
            }}
          >
            {char === ' ' ? ' ' : char}
          </motion.span>
        ))}
      </span>
    </motion.h2>
  );
}
