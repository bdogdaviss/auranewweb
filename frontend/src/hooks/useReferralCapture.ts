import { useEffect } from 'react';

// useReferralCapture persists an inbound ?ref=CODE affiliate code into a
// first-party cookie. The Go checkout / payment-intent handler reads this
// cookie server-side (it can't see localStorage), then stamps the code onto the
// Stripe PaymentIntent so the webhook can credit the referrer. Last touch wins;
// the cookie lasts 30 days.
const REF_PARAM = 'ref';
const REF_COOKIE = 'ref';
const MAX_AGE_SECONDS = 60 * 60 * 24 * 30; // 30 days

export function useReferralCapture(): void {
  useEffect(() => {
    const raw = new URLSearchParams(window.location.search).get(REF_PARAM)?.trim();
    if (!raw) return;
    // Keep it conservative: only short alphanumeric codes, matching the
    // server's generated format, to avoid storing junk from crafted URLs.
    if (!/^[A-Za-z0-9]{4,32}$/.test(raw)) return;
    // Stored referral codes are always lowercase — normalize so a link shared
    // with capitals still matches the referrer.
    const code = raw.toLowerCase();
    const secure = window.location.protocol === 'https:' ? '; secure' : '';
    document.cookie = `${REF_COOKIE}=${encodeURIComponent(code)}; path=/; max-age=${MAX_AGE_SECONDS}; samesite=lax${secure}`;
  }, []);
}
