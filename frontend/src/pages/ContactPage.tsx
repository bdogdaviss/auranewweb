import { motion } from 'framer-motion';
import { AboutIcon, ClockIcon, DiscordIcon } from '../components/design/icons';

const EASE: [number, number, number, number] = [0.22, 1, 0.36, 1];

const FAQ_ITEMS = [
  {
    q: 'How fast do you respond?',
    a: 'Discord support is typically answered within a few hours. Account and billing requests within 24 hours on business days.',
  },
  {
    q: 'I bought a product and need help installing it.',
    a: 'Head to our Discord #support channel for step-by-step installation help with every product.',
  },
];

export default function ContactPage() {
  return (
    <>
      {/* Hero */}
      <section className="pt-36 pb-10 px-6 md:px-12 max-w-4xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, ease: EASE }}
          className="font-mono text-[10px] tracking-[0.2em] text-white/30 uppercase mb-4"
        >
          {'// contact'}
        </motion.div>
        <motion.h1
          initial={{ opacity: 0, y: 16 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.7, ease: EASE, delay: 0.08 }}
          className="font-display text-[clamp(48px,7vw,90px)] leading-[0.88] tracking-[0.02em] text-white mb-6"
        >
          GET IN TOUCH
        </motion.h1>
        <motion.p
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.55, ease: EASE, delay: 0.16 }}
          className="text-white/40 text-[14px] font-light leading-relaxed max-w-md"
        >
          Have a question, issue, or just want to say hi? We&apos;re a small
          team and we actually read every message.
        </motion.p>
      </section>

      {/* Contact cards */}
      <section className="px-6 md:px-12 pb-10 max-w-4xl mx-auto">
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <motion.div
            initial={{ opacity: 0, y: 16 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-60px' }}
            transition={{ duration: 0.55, ease: EASE }}
          >
            <div className="glass-card p-8 h-full flex flex-col">
              <div className="w-10 h-10 rounded-xl bg-white/5 border border-white/8 flex items-center justify-center text-white/50 mb-5">
                <AboutIcon height={18} width={18} />
              </div>
              <h2 className="text-[16px] font-semibold text-white mb-2">
                Account &amp; Billing
              </h2>
              <p className="text-[13px] text-white/38 font-light leading-relaxed mb-6 flex-1">
                Best for billing questions, refund requests, and account
                issues. Manage your license, referral link, and orders from
                your dashboard.
              </p>
              <a
                href="/account"
                className="inline-flex items-center gap-2 px-6 py-3 text-[12px] font-semibold tracking-wider rounded-xl transition-all duration-300 hover:-translate-y-0.5 bg-white text-black hover:shadow-[0_0_40px_rgba(255,255,255,0.15)]"
              >
                Open Account →
              </a>
            </div>
          </motion.div>

          <motion.div
            initial={{ opacity: 0, y: 16 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-60px' }}
            transition={{ duration: 0.55, ease: EASE, delay: 0.06 }}
          >
            <div className="glass-card p-8 h-full flex flex-col">
              <div className="w-10 h-10 rounded-xl bg-white/5 border border-white/8 flex items-center justify-center text-white/50 mb-5">
                <DiscordIcon height={18} width={18} />
              </div>
              <h2 className="text-[16px] font-semibold text-white mb-2">
                Discord Community
              </h2>
              <p className="text-[13px] text-white/38 font-light leading-relaxed mb-6 flex-1">
                Fastest support channel. Get help from the team and 2,000+
                community members in real time.
              </p>
              <a
                href="https://discord.com/invite/4TBUw4nBFd"
                target="_blank"
                rel="noreferrer"
                className="inline-flex items-center gap-2 px-6 py-3 text-[12px] font-semibold tracking-wider rounded-xl transition-all duration-300 hover:-translate-y-0.5 bg-[#5865F2] hover:bg-[#6875f5] text-white hover:shadow-[0_0_40px_rgba(88,101,242,0.4)]"
              >
                Join Discord →
              </a>
            </div>
          </motion.div>
        </div>
      </section>

      {/* Stat chips */}
      <section className="px-6 md:px-12 pb-10 max-w-4xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.55, ease: EASE }}
          className="grid grid-cols-1 sm:grid-cols-3 gap-3"
        >
          <div className="flex items-center gap-4 p-5 rounded-xl border border-white/[0.07] bg-white/2">
            <div className="w-8 h-8 rounded-lg bg-white/5 border border-white/7 flex items-center justify-center text-white/40 shrink-0">
              <ClockIcon height={15} width={15} />
            </div>
            <div>
              <div className="font-display text-[18px] tracking-wider text-white">
                {'< 24h'}
              </div>
              <div className="font-mono text-[9px] tracking-widest text-white/28 uppercase">
                Support response time
              </div>
            </div>
          </div>
          <div className="flex items-center gap-4 p-5 rounded-xl border border-white/[0.07] bg-white/2">
            <div className="w-8 h-8 rounded-lg bg-white/5 border border-white/7 flex items-center justify-center text-white/40 shrink-0">
              <DiscordIcon height={15} width={15} />
            </div>
            <div>
              <div className="font-display text-[18px] tracking-wider text-white">
                Live
              </div>
              <div className="font-mono text-[9px] tracking-widest text-white/28 uppercase">
                Discord community support
              </div>
            </div>
          </div>
        </motion.div>
      </section>

      {/* Mini FAQ */}
      <section className="px-6 md:px-12 pb-24 max-w-4xl mx-auto">
        <motion.div
          initial={{ opacity: 0, y: 8 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: '-60px' }}
          transition={{ duration: 0.5, ease: EASE }}
          className="font-mono text-[9px] tracking-[0.2em] text-white/25 uppercase mb-4"
        >
          Before you write in
        </motion.div>
        <div className="flex flex-col gap-3">
          {FAQ_ITEMS.map((item, i) => (
            <motion.div
              key={item.q}
              initial={{ opacity: 0, y: 16 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: '-60px' }}
              transition={{ duration: 0.55, ease: EASE, delay: i * 0.06 }}
              className="p-5 rounded-xl border border-white/[0.07] bg-white/2"
            >
              <div className="text-[13px] font-semibold text-white/75 mb-1.5">
                {item.q}
              </div>
              <div className="text-[12px] text-white/35 font-light leading-relaxed">
                {item.a}
              </div>
            </motion.div>
          ))}
        </div>
      </section>
    </>
  );
}
