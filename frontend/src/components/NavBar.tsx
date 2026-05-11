import { NavLink, Link } from 'react-router-dom';

// SVG icons inlined to match the upstream HTML (and avoid an icon-library dep).
function HomeIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M3 9l9-7 9 7v11a2 2 0 01-2 2H5a2 2 0 01-2-2z" />
      <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
  );
}
function ProductsIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <path d="M21 16V8a2 2 0 00-1-1.73l-7-4a2 2 0 00-2 0l-7 4A2 2 0 002 8v8a2 2 0 001 1.73l7 4a2 2 0 002 0l7-4A2 2 0 0021 16z" />
    </svg>
  );
}
function AboutIcon() {
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <circle cx="12" cy="12" r="10" />
      <path d="M12 16v-4M12 8h.01" />
    </svg>
  );
}
function DiscordIcon() {
  return (
    <svg width="16" height="12" viewBox="0 0 24 18" fill="currentColor" xmlns="http://www.w3.org/2000/svg" aria-hidden>
      <path d="M20.317 1.492a19.7 19.7 0 00-4.885-1.516.074.074 0 00-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 00-5.487 0 12.64 12.64 0 00-.617-1.25.077.077 0 00-.079-.037A19.7 19.7 0 003.677 1.492a.07.07 0 00-.032.027C.533 6.093-.32 10.555.099 14.961a.08.08 0 00.031.055 19.9 19.9 0 005.993 3.03.078.078 0 00.084-.028c.462-.63.874-1.295 1.227-1.994a.076.076 0 00-.041-.106 13.1 13.1 0 01-1.872-.892.077.077 0 01-.008-.128c.126-.094.252-.192.372-.291a.074.074 0 01.078-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 01.078.009c.12.099.246.198.373.292a.077.077 0 01-.006.127 12.3 12.3 0 01-1.873.892.077.077 0 00-.041.107c.36.698.772 1.363 1.225 1.993a.076.076 0 00.084.028 19.84 19.84 0 006.002-3.03.077.077 0 00.032-.054c.5-5.177-.838-9.674-3.549-13.66a.06.06 0 00-.031-.028zM8.02 12.278c-1.183 0-2.157-1.086-2.157-2.419 0-1.334.956-2.419 2.157-2.419 1.21 0 2.176 1.095 2.157 2.42 0 1.332-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.086-2.157-2.419 0-1.334.955-2.419 2.157-2.419 1.21 0 2.176 1.095 2.157 2.42 0 1.332-.946 2.418-2.157 2.418z" />
    </svg>
  );
}

const navClass = ({ isActive }: { isActive: boolean }) =>
  'nav-link' + (isActive ? ' active' : '');

export default function NavBar() {
  return (
    <header className="header" role="banner">
      <div className="header-inner">
        <Link to="/" className="logo">
          Aura Optimizer
        </Link>
        <div className="nav-divider" aria-hidden />
        <nav className="nav-links" aria-label="Primary">
          <NavLink to="/" end className={navClass}>
            <HomeIcon /> Home
          </NavLink>
          <NavLink to="/products" className={navClass}>
            <ProductsIcon /> Products
          </NavLink>
          <NavLink to="/about" className={navClass}>
            <AboutIcon /> About
          </NavLink>
        </nav>
        <div className="nav-divider" aria-hidden />
        <a
          href="https://discord.com/invite/4TBUw4nBFd"
          className="discord-cta"
          target="_blank"
          rel="noopener noreferrer"
        >
          <DiscordIcon />
          Discord
        </a>
      </div>
    </header>
  );
}
