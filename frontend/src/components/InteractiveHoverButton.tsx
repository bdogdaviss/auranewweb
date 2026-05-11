import type { ReactNode } from 'react';
import { Link } from 'react-router-dom';

// Polymorphic Magic-UI-style "interactive hover" button:
//   - on hover, a small dot at the left scales up and fills the button with
//     white; the visible label fades + slides right; a duplicate label + arrow
//     slides in from the left in dark text on the new white background.
//
// Render-element rules (mutually exclusive, evaluated in order):
//   - `to`   → React Router <Link> (client-side nav, internal route)
//   - `href` → plain <a> (external link)
//   - else   → <button> (use onClick)
//
// The Magic UI version is Tailwind + lucide-react; this is a plain-CSS,
// inline-SVG re-implementation so the project doesn't take on those deps.

type CommonProps = {
  children: ReactNode;
  className?: string;
};

type LinkVariant = CommonProps & {
  to: string;
  href?: never;
  onClick?: never;
  external?: never;
};
type AnchorVariant = CommonProps & {
  to?: never;
  href: string;
  external?: boolean;
  onClick?: never;
};
type ButtonVariant = CommonProps & {
  to?: never;
  href?: never;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
  external?: never;
};

type Props = LinkVariant | AnchorVariant | ButtonVariant;

function ArrowRightIcon() {
  return (
    <svg
      className="ihb-arrow"
      width="16"
      height="16"
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2.2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden
    >
      <line x1="5" y1="12" x2="19" y2="12" />
      <polyline points="12 5 19 12 12 19" />
    </svg>
  );
}

function Inner({ children }: { children: ReactNode }) {
  return (
    <>
      <span className="ihb-row">
        <span className="ihb-dot" aria-hidden />
        <span className="ihb-label">{children}</span>
      </span>
      <span className="ihb-overlay" aria-hidden>
        <span>{children}</span>
        <ArrowRightIcon />
      </span>
    </>
  );
}

export function InteractiveHoverButton(props: Props) {
  const cls = `ihb ${props.className ?? ''}`.trim();

  if ('to' in props && props.to !== undefined) {
    return (
      <Link to={props.to} className={cls}>
        <Inner>{props.children}</Inner>
      </Link>
    );
  }

  if ('href' in props && props.href !== undefined) {
    const external = props.external ?? /^https?:/i.test(props.href);
    return (
      <a
        href={props.href}
        className={cls}
        {...(external ? { target: '_blank', rel: 'noopener noreferrer' } : {})}
      >
        <Inner>{props.children}</Inner>
      </a>
    );
  }

  const { onClick, type = 'button' } = props as ButtonVariant;
  return (
    <button type={type} className={cls} onClick={onClick}>
      <Inner>{props.children}</Inner>
    </button>
  );
}

export default InteractiveHoverButton;
