export default function AboutPage() {
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
