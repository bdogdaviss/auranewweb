export default function Footer() {
  return (
    <footer className="footer">
      <div className="footer-inner">
        <div className="footer-grid">
          <div className="footer-brand">
            <div className="footer-logo">Aura Optimizer</div>
            <p className="footer-desc">
              Professional gaming optimization tools engineered for competitive players.
              Boost your performance and gain the edge you need to dominate.
            </p>
          </div>
          <div>
            <h4 className="footer-heading">Products</h4>
            <ul className="footer-links">
              <li><a href="/products">Starter</a></li>
              <li><a href="/products">Lifetime Pro</a></li>
              <li><a href="/products">Team License</a></li>
            </ul>
          </div>
          <div>
            <h4 className="footer-heading">Support</h4>
            <ul className="footer-links">
              <li>
                <a href="https://discord.com/invite/4TBUw4nBFd" target="_blank" rel="noopener noreferrer">
                  Discord
                </a>
              </li>
              <li><a href="#">Documentation</a></li>
              <li><a href="#">FAQ</a></li>
              <li><a href="#">Contact</a></li>
            </ul>
          </div>
          <div>
            <h4 className="footer-heading">Legal</h4>
            <ul className="footer-links">
              <li><a href="#">Terms of Service</a></li>
              <li><a href="#">Privacy Policy</a></li>
              <li><a href="#">Refund Policy</a></li>
            </ul>
          </div>
        </div>
        <div className="footer-bottom">
          © {new Date().getFullYear()} Aura Optimizer. All rights reserved.
        </div>
      </div>
    </footer>
  );
}
