import { useState, type ReactNode, type FormEvent } from 'react';
import './UpsellBanner.css';

// UpsellBanner — an Aura-themed take on Gestalt's BannerUpsell. Same prop shape
// (title, message, imageData, primaryAction, secondaryAction, dismissButton,
// and an UpsellBanner.Form subcomponent) but styled to the dark emerald theme
// and dependency-free. Used to promote upgrades/offers across the site.

type UpsellAction = {
  label: string;
  accessibilityLabel: string;
  href?: string;
  onClick?: () => void;
  target?: '_blank' | '_self';
};

type ImageData = {
  component: ReactNode;
  // 'icon' renders in a rounded emerald chip; 'image' renders at `width` (px).
  kind?: 'icon' | 'image';
  width?: number;
};

type UpsellBannerProps = {
  title: string;
  message: ReactNode;
  imageData?: ImageData;
  primaryAction?: UpsellAction;
  secondaryAction?: UpsellAction;
  dismissButton?: { accessibilityLabel?: string; onDismiss?: () => void };
  // Persisted dismissal key: once dismissed, this banner stays hidden for the
  // user (honors the "don't keep showing a dismissed upsell" guideline).
  dismissId?: string;
  children?: ReactNode; // UpsellBanner.Form
};

function ActionEl({ action, variant }: { action: UpsellAction; variant: 'primary' | 'secondary' }) {
  const cls = `upsell-btn upsell-btn-${variant}`;
  if (action.href) {
    return (
      <a
        className={cls}
        href={action.href}
        target={action.target}
        rel={action.target === '_blank' ? 'noopener noreferrer' : undefined}
        aria-label={action.accessibilityLabel}
        onClick={action.onClick}
      >
        {action.label}
      </a>
    );
  }
  return (
    <button type="button" className={cls} aria-label={action.accessibilityLabel} onClick={action.onClick}>
      {action.label}
    </button>
  );
}

function UpsellBannerBase({
  title,
  message,
  imageData,
  primaryAction,
  secondaryAction,
  dismissButton,
  dismissId,
  children,
}: UpsellBannerProps) {
  const storageKey = dismissId ? `aura_upsell_dismissed_${dismissId}` : null;
  const [dismissed, setDismissed] = useState(() => {
    if (!storageKey) return false;
    try {
      return localStorage.getItem(storageKey) === '1';
    } catch {
      return false;
    }
  });

  if (dismissed) return null;

  const handleDismiss = () => {
    if (storageKey) {
      try {
        localStorage.setItem(storageKey, '1');
      } catch {
        /* ignore storage failures */
      }
    }
    setDismissed(true);
    dismissButton?.onDismiss?.();
  };

  return (
    <div className="upsell" role="region" aria-label={title}>
      {imageData && (
        <div
          className={`upsell-media ${imageData.kind === 'image' ? 'upsell-media-img' : 'upsell-icon-wrap'}`}
          style={imageData.kind === 'image' && imageData.width ? { width: imageData.width } : undefined}
        >
          {imageData.component}
        </div>
      )}
      <div className="upsell-body">
        <h3 className="upsell-title">{title}</h3>
        <div className="upsell-message">{message}</div>
        {children}
        {(primaryAction || secondaryAction) && (
          <div className="upsell-actions">
            {primaryAction && <ActionEl action={primaryAction} variant="primary" />}
            {secondaryAction && <ActionEl action={secondaryAction} variant="secondary" />}
          </div>
        )}
      </div>
      {dismissButton && (
        <button
          type="button"
          className="upsell-dismiss"
          aria-label={dismissButton.accessibilityLabel ?? 'Dismiss this banner'}
          onClick={handleDismiss}
        >
          ×
        </button>
      )}
    </div>
  );
}

type FormProps = {
  children: ReactNode;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  submitButtonText: string;
  submitButtonAccessibilityLabel: string;
  submitButtonDisabled?: boolean;
};

function UpsellForm({
  children,
  onSubmit,
  submitButtonText,
  submitButtonAccessibilityLabel,
  submitButtonDisabled,
}: FormProps) {
  return (
    <form className="upsell-form" onSubmit={onSubmit}>
      <div className="upsell-form-fields">{children}</div>
      <button
        type="submit"
        className="upsell-btn upsell-btn-primary"
        aria-label={submitButtonAccessibilityLabel}
        disabled={submitButtonDisabled}
      >
        {submitButtonText}
      </button>
    </form>
  );
}

const UpsellBanner = Object.assign(UpsellBannerBase, { Form: UpsellForm });
export default UpsellBanner;
