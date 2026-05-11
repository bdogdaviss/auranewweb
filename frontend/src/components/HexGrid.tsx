import { useEffect, useMemo, useRef, useState } from 'react';
import { motion, AnimatePresence } from 'framer-motion';

// HexGrid — hexagonally-packed circular avatars with a profile tooltip on
// hover. Rows alternate between N and N-1 capacity and nest vertically via
// a negative top margin so the circles tile like a honeycomb.
//
// Adapted from a Tailwind/framer-motion source → plain CSS in theme.css
// under "HexGrid". Real Aura community avatars from /static/images/ stand
// in for the source's pravatar mock.

const AVATARS = [
  '/static/images/ajerrs.jpg',
  '/static/images/bokify.jpeg',
  '/static/images/kise.jpg',
  '/static/images/narpzfn.jpeg',
  '/static/images/perksfn.jpeg',
  '/static/images/rezy.jpeg',
  '/static/images/sjay.jpeg',
  '/static/images/newpfp.png',
];

const COVERS = [
  '/static/images/newdashboard.png',
  '/static/images/newnetwork.png',
  '/static/images/newoptimizer.png',
  '/static/images/showcase.png',
];

const HANDLES = ['ajerrs', 'bokify', 'kise', 'narpz', 'perks', 'rezy', 'sjay', 'aura'];
const ROLES = ['Valorant', 'Apex Legends', 'CS2', 'Fortnite', 'Overwatch 2', 'Rainbow Six'];

const TESTIMONIALS = [
  'Frame pacing is rock solid now. Got my 240 fps lock back.',
  'Network tuning cut my ping by 18ms. Real numbers.',
  'Set it up once. Undo if I need stock. Couldnt be easier.',
  'Macro engine is sub-millisecond. Feels native to my hardware.',
  'Found startup apps Windows hid from me. Boot is half the time.',
  '47% more FPS in Apex after one pass. Crazy gains.',
  'Best $15 Ive spent on this rig all year.',
  'Support actually replies on Discord. Real engineers, not bots.',
];

type Member = {
  id: string;
  name: string;
  role: string;
  message: string;
  avatar: string;
  cover: string;
};

const GRID_LENGTH = 48;
const ITEM_SIZE = 64;
const GAP = 36;

const TEAM: Member[] = Array.from({ length: GRID_LENGTH }, (_, i) => ({
  id: `member-${i}`,
  name: HANDLES[i % HANDLES.length] + (i >= HANDLES.length ? `_${Math.floor(i / HANDLES.length)}` : ''),
  role: ROLES[i % ROLES.length],
  message: TESTIMONIALS[i % TESTIMONIALS.length],
  avatar: AVATARS[i % AVATARS.length],
  cover: COVERS[i % COVERS.length],
}));

const containerVariants = {
  hidden: { opacity: 0 },
  visible: { opacity: 1, transition: { staggerChildren: 0.02, delayChildren: 0.1 } },
};

const itemVariants = {
  hidden: { opacity: 0, scale: 0, y: 20 },
  visible: {
    opacity: 1,
    scale: 1,
    y: 0,
    transition: { type: 'spring' as const, stiffness: 200, damping: 20 },
  },
};

function QuoteIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
      <path d="M3 21c3 0 7-1 7-8V5H3v8h4c0 4-3 5-4 5v3zm11 0c3 0 7-1 7-8V5h-7v8h4c0 4-3 5-4 5v3z" />
    </svg>
  );
}

function ProfileCard({ member }: { member: Member }) {
  return (
    <motion.div
      role="tooltip"
      className="hexgrid-tooltip"
      initial={{ opacity: 0, scale: 0.8, y: 15 }}
      animate={{ opacity: 1, scale: 1, y: -15 }}
      exit={{ opacity: 0, scale: 0.9, y: 15 }}
      transition={{ type: 'spring', stiffness: 400, damping: 25 }}
    >
      <div className="hexgrid-tooltip-cover">
        <img src={member.cover} alt="" loading="lazy" />
        <div className="hexgrid-tooltip-cover-shade" />
      </div>
      <div className="hexgrid-tooltip-body">
        <div className="hexgrid-tooltip-avatar-row">
          <img src={member.avatar} alt={member.name} className="hexgrid-tooltip-avatar" />
        </div>
        <h3 className="hexgrid-tooltip-name">@{member.name}</h3>
        <p className="hexgrid-tooltip-role">{member.role}</p>
        <div className="hexgrid-tooltip-quote">
          <span className="hexgrid-tooltip-quote-icon" aria-hidden>
            <QuoteIcon />
          </span>
          <p>{member.message}</p>
        </div>
      </div>
      <div className="hexgrid-tooltip-arrow" aria-hidden />
    </motion.div>
  );
}

function AvatarItem({ user }: { user: Member }) {
  const [hovered, setHovered] = useState(false);
  return (
    <motion.div
      className="hexgrid-item"
      variants={itemVariants}
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{ zIndex: hovered ? 50 : 1 }}
    >
      <motion.button
        type="button"
        aria-label={`View profile of ${user.name}`}
        aria-expanded={hovered}
        whileHover={{ scale: 1.15 }}
        whileTap={{ scale: 0.95 }}
        className="hexgrid-avatar"
      >
        <img src={user.avatar} alt={user.name} loading="lazy" />
      </motion.button>
      <AnimatePresence>{hovered && <ProfileCard member={user} />}</AnimatePresence>
    </motion.div>
  );
}

export default function HexGrid() {
  const [columns, setColumns] = useState(8);
  const containerRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const updateLayout = () => {
      const width = Math.min(window.innerWidth - 32, 1280);
      const itemWidth = ITEM_SIZE + GAP;
      const cols = Math.min(Math.max(Math.floor(width / itemWidth), 3), 14);
      setColumns(cols);
    };
    updateLayout();
    window.addEventListener('resize', updateLayout);
    return () => window.removeEventListener('resize', updateLayout);
  }, []);

  // Hex packing geometry — nest rows by sin(60°) of the row height.
  const sin60 = 0.866;
  const effectiveHeight = (ITEM_SIZE + GAP) * sin60;
  const negativeMargin = -(ITEM_SIZE - effectiveHeight + GAP * (1 - sin60));

  const rows = useMemo(() => {
    const out: { items: Member[]; isEven: boolean }[] = [];
    let i = 0;
    let isEven = true;
    while (i < TEAM.length) {
      const capacity = isEven ? columns : columns - 1;
      const chunk = TEAM.slice(i, i + capacity);
      if (chunk.length > 0) out.push({ items: chunk, isEven });
      i += capacity;
      isEven = !isEven;
    }
    return out;
  }, [columns]);

  return (
    <section className="hexgrid-section">
      <div className="hexgrid-intro">
        <h2 className="hexgrid-title">
          Built for scale. <br /> Trusted by players.
        </h2>
        <p className="hexgrid-subtitle">
          Hover any face to read what they got out of Aura.
        </p>
      </div>
      <motion.div
        ref={containerRef}
        className="hexgrid"
        variants={containerVariants}
        initial="hidden"
        whileInView="visible"
        viewport={{ once: true, amount: 0.1 }}
        aria-label="Aura community grid"
      >
        {rows.map((row, ri) => (
          <div
            key={ri}
            className="hexgrid-row"
            style={{
              marginTop: ri > 0 ? `${negativeMargin}px` : 0,
              gap: `${GAP}px`,
            }}
          >
            {row.items.map((user) => (
              <AvatarItem key={user.id} user={user} />
            ))}
          </div>
        ))}
      </motion.div>
    </section>
  );
}
