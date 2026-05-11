import { Link } from 'react-router-dom';

// Hero with a video placeholder. When the user uploads img/hero-bg.mp4 we'll
// swap the <div className="hero-poster"> for an autoplay-muted-loop <video>.
const HERO_VIDEO_SRC = '/static/images/hero-bg.mp4';

export default function Hero() {
  return (
    <section className="hero">
      <div className="hero-left">
        <div className="hero-badge animate-hero" style={{ animationDelay: '0.2s' }}>
          <span className="badge-dot" aria-hidden />
          <span>2,000+ Active Users</span>
        </div>
        <h1 className="hero-title animate-hero" style={{ animationDelay: '0.4s' }}>
          Elite Gaming
          <br />
          <span className="gradient-text">Performance</span>
        </h1>
        <p className="hero-subtitle animate-hero" style={{ animationDelay: '0.6s' }}>
          Aura Optimizer offers professional-grade optimization tools engineered to deliver
          measurable competitive advantage in gaming performance.
        </p>
        <div className="hero-cta-group animate-hero" style={{ animationDelay: '0.8s' }}>
          <Link to="/products" className="btn-primary">
            View Products
          </Link>
          <a
            href="https://discord.com/invite/4TBUw4nBFd"
            className="btn-secondary"
            target="_blank"
            rel="noopener noreferrer"
          >
            Join Discord
          </a>
        </div>
      </div>

      <div className="hero-right animate-hero" style={{ animationDelay: '0.8s' }}>
        <div className="hero-image-wrapper">
          <video
            className="hero-video"
            autoPlay
            muted
            loop
            playsInline
            poster=""
            onError={(e) => {
              // Hide the video and fall back to the gradient poster underneath.
              (e.currentTarget as HTMLVideoElement).style.display = 'none';
            }}
          >
            <source src={HERO_VIDEO_SRC} type="video/mp4" />
          </video>
          <div className="hero-poster" aria-hidden>
            Aura&nbsp;·&nbsp;Optimizer
          </div>
        </div>
      </div>
    </section>
  );
}
