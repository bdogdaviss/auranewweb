import { useState, type FormEvent } from 'react';
import { motion } from 'framer-motion';
import UpsellBanner from '../components/UpsellBanner';

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

interface CoreValue {
  num: string;
  title: string;
  body: string;
}

const CORE_VALUES: CoreValue[] = [
  {
    num: '01',
    title: 'Community First',
    body: 'Our vibrant Discord community of 2,000+ active members is at the heart of everything we do.',
  },
  {
    num: '02',
    title: 'Real-Time Support',
    body: 'We listen to feedback, provide real-time support, and continuously improve our products based on what competitive gamers actually need to dominate.',
  },
  {
    num: '03',
    title: 'Cutting-Edge Technology',
    body: 'Built on years of research and testing, our optimization algorithms represent the pinnacle of gaming performance technology.',
  },
  {
    num: '04',
    title: 'Tested & Safe',
    body: 'Every product is rigorously tested to ensure maximum effectiveness and complete safety for your system.',
  },
];

function GiftIcon() {
  return (
    <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <polyline points="20 12 20 22 4 22 4 12" />
      <rect x="2" y="7" width="20" height="5" />
      <line x1="12" y1="22" x2="12" y2="7" />
      <path d="M12 7H7.5a2.5 2.5 0 0 1 0-5C11 2 12 7 12 7z" />
      <path d="M12 7h4.5a2.5 2.5 0 0 0 0-5C13 2 12 7 12 7z" />
    </svg>
  );
}

export default function AboutPage() {
  const [name, setName] = useState('');
  const [email, setEmail] = useState('');
  const [joined, setJoined] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Persists the lead via /api/leads, so the success message is actually true.
  const handleJoin = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!name.trim() || !email.trim()) return;
    setSubmitting(true);
    setError('');
    try {
      const res = await fetch('/api/leads', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name.trim(), email: email.trim(), source: 'affiliate-signup' }),
      });
      if (!res.ok) throw new Error('Could not sign you up. Try again.');
      setJoined(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Something went wrong.');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <>
      {/* Hero */}
      <section className="pt-36 pb-8 px-6 md:px-12 max-w-6xl mx-auto">
        <div className="text-center">
          <motion.div
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, ease: EASE }}
            className="font-mono text-[10px] tracking-[0.2em] text-white/30 uppercase mb-4"
          >
            {'// who we are'}
          </motion.div>
          <motion.h1
            initial={{ opacity: 0, y: 16 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.7, ease: EASE, delay: 0.08 }}
            className="font-display text-[clamp(52px,9vw,110px)] leading-[0.88] tracking-[0.02em] text-white"
          >
            ABOUT
            <br />
            <span className="text-white/60">AURA</span>
            <br />
            OPTIMIZER
          </motion.h1>
        </div>
      </section>

      {/* Story card */}
      <section className="px-6 md:px-12 pb-8 max-w-6xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.6, ease: EASE }}
          className="relative rounded-2xl border border-white/[0.07] bg-white/2 p-10 md:p-14 overflow-hidden"
        >
          <div className="absolute top-0 left-[8%] right-[8%] h-px bg-linear-to-r from-transparent via-white/20 to-transparent" />
          <img
            src="/aura-icon.png"
            alt="Aura Optimizer"
            className="w-11 h-11 absolute top-5 right-8 select-none opacity-60 object-contain"
          />
          <p className="font-display text-[clamp(26px,4vw,48px)] leading-[1.05] tracking-[0.02em] text-white mb-5">
            Leading provider of
            <br />
            <span className="text-white/20">elite gaming</span>
            <br />
            optimization tools.
          </p>
          <p className="text-[14px] text-white/38 font-light leading-relaxed max-w-lg">
            Delivering professional-grade performance solutions to competitive
            gamers worldwide. Every player deserves access to the tools that
            deliver real, measurable results in competitive gaming.
          </p>
        </motion.div>
      </section>

      {/* Affiliate signup — wired to /api/leads */}
      <section className="px-6 md:px-12 pb-8 max-w-6xl mx-auto">
        <UpsellBanner
          dismissId="about-join"
          title="Join the Aura affiliate program"
          message={
            joined ? (
              <>
                You're on the list! <a href="/account">Sign in</a> anytime to grab your referral link
                and start earning store credit.
              </>
            ) : (
              'Drop your name and email — we\'ll set you up to earn store credit on every sale.'
            )
          }
          imageData={{ kind: 'icon', component: <GiftIcon /> }}
        >
          {!joined && (
            <UpsellBanner.Form
              onSubmit={handleJoin}
              submitButtonText={submitting ? 'Joining…' : 'Join'}
              submitButtonAccessibilityLabel="Join the Aura affiliate program"
              submitButtonDisabled={submitting}
            >
              <input
                className="upsell-input"
                type="text"
                placeholder="Your name"
                aria-label="Your name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                required
              />
              <input
                className="upsell-input"
                type="email"
                placeholder="you@example.com"
                aria-label="Your email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </UpsellBanner.Form>
          )}
          {error && <p style={{ color: '#f87171', fontSize: '0.8rem', marginTop: 8 }}>{error}</p>}
        </UpsellBanner>
      </section>

      {/* Main grid */}
      <section className="px-6 md:px-12 pb-16 max-w-6xl mx-auto">
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          {/* Our Mission */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-60px' }}
            transition={{ duration: 0.6, ease: EASE }}
            className="glass-card p-8 h-full"
          >
            <h2 className="font-display text-[22px] tracking-wider text-white mb-5">
              OUR MISSION
            </h2>
            <p className="text-[14px] text-white/40 leading-relaxed font-light mb-4">
              We are dedicated to providing gamers with the{' '}
              <strong className="text-white/70 font-medium">
                most advanced optimization tools
              </strong>{' '}
              available.
            </p>
            <p className="text-[14px] text-white/40 leading-relaxed font-light">
              Our mission is to level the playing field and give every player
              access to professional-grade performance enhancements that
              deliver{' '}
              <strong className="text-white/70 font-medium">
                real, measurable results
              </strong>{' '}
              in competitive gaming.
            </p>
          </motion.div>

          {/* Core Values */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-60px' }}
            transition={{ duration: 0.6, ease: EASE, delay: 0.05 }}
            className="glass-card p-8 h-full"
          >
            <h2 className="font-display text-[22px] tracking-wider text-white mb-5">
              CORE VALUES
            </h2>
            <div className="flex flex-col gap-3">
              {CORE_VALUES.map((value) => (
                <div
                  key={value.num}
                  className="flex items-start gap-4 p-3.5 rounded-xl bg-white/2 border border-white/5"
                >
                  <span className="font-display text-[20px] text-white/12 tracking-wider mt-0.5 shrink-0">
                    {value.num}
                  </span>
                  <div>
                    <div className="text-[13px] font-semibold text-white/80 mb-0.5">
                      {value.title}
                    </div>
                    <div className="text-[12px] text-white/32 font-light leading-snug">
                      {value.body}
                    </div>
                  </div>
                </div>
              ))}
            </div>
          </motion.div>

          {/* Discord banner */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-60px' }}
            transition={{ duration: 0.6, ease: EASE, delay: 0.1 }}
            className="md:col-span-2"
          >
            <div className="relative rounded-2xl border border-dashed border-white/8 p-12 text-center overflow-hidden">
              <div
                className="absolute inset-0 rounded-2xl"
                style={{
                  background:
                    'radial-gradient(circle at 50% 0%, rgba(255,255,255,0.025) 0%, transparent 60%)',
                }}
              />
              <div className="relative z-10">
                <div className="font-display text-[clamp(30px,5vw,54px)] tracking-wider text-white mb-3">
                  JOIN 2,000+ PLAYERS
                </div>
                <p className="text-[14px] text-white/38 font-light mb-8 max-w-sm mx-auto leading-relaxed">
                  2,000+ players sharing tweaks, configs, and benchmark
                  results. Get optimization help and early access to new
                  releases.
                </p>
                <a
                  href="https://discord.com/invite/4TBUw4nBFd"
                  target="_blank"
                  rel="noreferrer"
                  className="inline-flex items-center gap-3 px-8 py-4 bg-[#5865F2] hover:bg-[#6875f5] text-white text-[13px] font-semibold tracking-wider rounded-xl transition-all duration-300 hover:shadow-[0_0_40px_rgba(88,101,242,0.4)] hover:-translate-y-0.5"
                >
                  Join Now
                </a>
              </div>
            </div>
          </motion.div>
        </div>
      </section>
    </>
  );
}
