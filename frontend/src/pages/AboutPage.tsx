import { useState, type FormEvent } from 'react';
import UpsellBanner from '../components/UpsellBanner';

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
    <section className="tab-content active">
      <div className="section-header scroll-animate visible">
        <span className="section-label">
          <span className="label-dot" aria-hidden /> Who We Are
        </span>
        <h2 className="section-title gradient-text">About Aura Optimizer</h2>
        <p className="section-subtitle">
          Leading provider of elite gaming optimization tools, delivering professional-grade
          performance solutions to competitive gamers worldwide.
        </p>
      </div>

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

      <div className="featured-grid">
        <article className="featured-card">
          <h3 className="card-name">Our Mission</h3>
          <p className="card-desc">
            We are dedicated to providing gamers with the most advanced optimization tools available.
            Our mission is to level the playing field and give every player access to
            professional-grade performance enhancements that deliver real, measurable results in
            competitive gaming.
          </p>
        </article>
        <article className="featured-card">
          <h3 className="card-name">Community First</h3>
          <p className="card-desc">
            Our vibrant Discord community of 2,000+ active members is at the heart of everything we
            do. We listen to feedback, provide real-time support, and continuously improve our
            products based on what competitive gamers actually need to dominate.
          </p>
        </article>
        <article className="featured-card">
          <h3 className="card-name">Cutting-Edge Technology</h3>
          <p className="card-desc">
            Built on years of research and testing, our optimization algorithms represent the
            pinnacle of gaming performance technology. Every product is rigorously tested to ensure
            maximum effectiveness and complete safety for your system.
          </p>
        </article>
      </div>
    </section>
  );
}
