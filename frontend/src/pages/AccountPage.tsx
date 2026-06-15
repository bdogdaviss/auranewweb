import { useEffect, useState } from 'react';
import { useToast } from '../components/ToastContext';

interface Referral {
  date: string;
  product: string;
  commission: string;
  status: string;
}

interface SummaryData {
  referral_code: string;
  referral_link: string;
  store_credit_cents: number;
  store_credit_display: string;
  total_earned_cents: number;
  total_earned_display: string;
  referral_count: number;
  referrals: Referral[];
  commission_display: string;
  referred_by_code: string;
  referred_by_name: string;
  user: {
    name: string;
    email: string;
  };
}

export default function AccountPage() {
  const { toast } = useToast();

  const [data, setData] = useState<SummaryData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [redirecting, setRedirecting] = useState(false);

  // Editable code
  const [editCode, setEditCode] = useState('');
  const [savingCode, setSavingCode] = useState(false);
  const [codeMsg, setCodeMsg] = useState<{ text: string; type: 'success' | 'error' | 'info' } | null>(null);

  // Payout waitlist
  const [waitlistEmail, setWaitlistEmail] = useState('');
  const [waitlistSubmitting, setWaitlistSubmitting] = useState(false);
  const [waitlistSuccess, setWaitlistSuccess] = useState(false);
  const [waitlistError, setWaitlistError] = useState('');

  useEffect(() => {
    async function load() {
      setLoading(true);
      setError(null);
      try {
        const res = await fetch('/api/account/summary', {
          credentials: 'include',
        });
        if (res.status === 401) {
          setRedirecting(true);
          // Full page load to the server-rendered login template (client-side
          // navigate won't work for non-React auth routes like /login).
          window.location.href = '/login';
          return;
        }
        if (!res.ok) {
          throw new Error('Failed to load account data');
        }
        const json: SummaryData = await res.json();
        setData(json);
        setEditCode(json.referral_code);
      } catch (e: any) {
        setError(e.message || 'Could not load your affiliate dashboard.');
      } finally {
        setLoading(false);
      }
    }
    load();
  }, []);

  const copyToClipboard = async (text: string, label: string) => {
    try {
      await navigator.clipboard.writeText(text);
      toast.success(`${label} copied!`);
    } catch {
      toast.error('Failed to copy');
    }
  };

  const saveReferralCode = async () => {
    if (!data) return;
    const code = editCode.trim();
    if (!code) {
      setCodeMsg({ text: 'Code cannot be empty.', type: 'error' });
      return;
    }
    setSavingCode(true);
    setCodeMsg({ text: 'Saving…', type: 'info' });
    try {
      const res = await fetch('/api/account/referral-code', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        credentials: 'include',
        body: JSON.stringify({ code }),
      });
      const json = await res.json().catch(() => ({}));
      if (!res.ok) {
        throw new Error(json.error || 'Could not save code.');
      }
      // Update local data
      const newCode = json.code || code;
      const newLink = json.link || data.referral_link.replace(/\/\?ref=.*/, `/?ref=${newCode}`);
      setData({
        ...data,
        referral_code: newCode,
        referral_link: newLink,
      });
      setEditCode(newCode);
      setCodeMsg({ text: 'Saved successfully.', type: 'success' });
      toast.success('Referral code updated!');
    } catch (e: any) {
      setCodeMsg({ text: e.message, type: 'error' });
    } finally {
      setSavingCode(false);
      // Clear message after delay
      setTimeout(() => setCodeMsg(null), 2500);
    }
  };

  const submitWaitlist = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!waitlistEmail.trim()) return;
    setWaitlistSubmitting(true);
    setWaitlistError('');
    try {
      const res = await fetch('/api/leads', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          email: waitlistEmail.trim(),
          source: 'payout-waitlist',
        }),
      });
      if (!res.ok) throw new Error('Failed');
      setWaitlistSuccess(true);
      setWaitlistEmail('');
      toast.success("You're on the list!");
    } catch {
      setWaitlistError("Couldn't save — try again.");
    } finally {
      setWaitlistSubmitting(false);
    }
  };

  if (redirecting) {
    return (
      <div className="max-w-5xl mx-auto px-6 py-16 text-center text-gray-400">
        Redirecting to login…
      </div>
    );
  }

  if (loading) {
    return (
      <div className="max-w-5xl mx-auto px-6 py-16 text-center text-gray-400">
        Loading your affiliate dashboard…
      </div>
    );
  }

  if (error || !data) {
    return (
      <div className="max-w-5xl mx-auto px-6 py-16">
        <div className="bg-red-500/10 border border-red-500/30 rounded-xl p-8 text-center">
          <p className="text-red-400 mb-4">{error || 'Unable to load account data. You may need to log in.'}</p>
          <a href="/login" className="text-emerald-400 hover:underline">
            Log in to continue
          </a>
        </div>
      </div>
    );
  }

  return (
    <section className="tab-content active">
      <div className="max-w-6xl mx-auto px-6 py-12">
      {/* Cash payouts upsell */}
      {!waitlistSuccess && (
        <div className="mb-8 relative flex flex-col sm:flex-row sm:items-center gap-4 rounded-2xl border border-emerald-500/25 bg-gradient-to-br from-emerald-500/10 to-white/[0.02] p-5">
          <div className="shrink-0 w-12 h-12 rounded-xl bg-emerald-500/10 flex items-center justify-center text-emerald-400">
            <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M18 8a6 6 0 0 0-12 0c0 7-3 9-3 9h18s-3-2-3-9" />
              <path d="M13.73 21a2 2 0 0 1-3.46 0" />
            </svg>
          </div>
          <div className="flex-1">
            <h3 className="font-semibold text-white">Cash payouts are coming</h3>
            <p className="text-sm text-gray-400">Today you earn store credit. Want to withdraw to your bank instead? Get notified when payouts launch.</p>
          </div>
          <form onSubmit={submitWaitlist} className="flex gap-2 w-full sm:w-auto">
            <input
              type="email"
              required
              placeholder="you@example.com"
              value={waitlistEmail}
              onChange={(e) => setWaitlistEmail(e.target.value)}
              className="flex-1 sm:w-56 bg-black/30 border border-white/10 rounded-lg px-3 py-2 text-sm text-white placeholder-gray-500 focus:outline-none focus:border-aura-500/50"
            />
            <button
              type="submit"
              disabled={waitlistSubmitting}
              className="bg-aura-500 hover:bg-aura-600 text-white text-sm font-semibold px-4 py-2 rounded-lg whitespace-nowrap disabled:opacity-60"
            >
              {waitlistSubmitting ? 'Saving…' : 'Notify me'}
            </button>
          </form>
          {waitlistError && <p className="text-red-400 text-sm absolute -bottom-5 right-0">{waitlistError}</p>}
        </div>
      )}
      {waitlistSuccess && (
        <div className="mb-8 rounded-2xl border border-emerald-500/30 bg-emerald-500/10 p-5 text-emerald-400">
          Thanks — you're on the list. We'll email you when payouts launch.
        </div>
      )}

      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold tracking-tight">Affiliate Dashboard</h1>
        <a href="/logout" className="text-sm text-gray-400 hover:text-white">Log out</a>
      </div>

      <div className="grid lg:grid-cols-3 gap-6">
        {/* LEFT COLUMN */}
        <div className="lg:col-span-2 space-y-6">
          <h2 className="text-sm text-gray-400 font-medium tracking-wider">OVERVIEW</h2>

          {/* Earnings */}
          <div className="panel p-6">
            <div className="flex items-center justify-between mb-5">
              <h3 className="font-semibold">Earnings</h3>
              <span className="text-xs px-3 py-1.5 bg-[#1a1a1d] border border-white/10 rounded text-gray-400">All time</span>
            </div>
            <div className="grid sm:grid-cols-3 gap-3">
              <div className="inset p-4">
                <p className="text-xs text-gray-500 mb-1">Available credit</p>
                <p className="text-2xl font-bold text-emerald-400">{data.store_credit_display}</p>
              </div>
              <div className="inset p-4">
                <p className="text-xs text-gray-500 mb-1">Total earned</p>
                <p className="text-2xl font-bold">{data.total_earned_display}</p>
              </div>
              <div className="inset p-4">
                <p className="text-xs text-gray-500 mb-1">Referrals</p>
                <p className="text-2xl font-bold">{data.referral_count}</p>
              </div>
            </div>
          </div>

          {/* Your code */}
          <div className="panel p-6">
            <h3 className="font-semibold mb-4">Your code</h3>
            <div className="inset flex items-center gap-2 px-4 py-3">
              <input
                type="text"
                readOnly
                value={data.referral_code}
                className="flex-1 bg-transparent text-sm text-white tracking-widest uppercase focus:outline-none"
              />
              <button
                onClick={() => copyToClipboard(data.referral_code, 'Code')}
                className="text-gray-400 hover:text-white p-1"
                aria-label="Copy code"
              >
                <CopyIcon />
              </button>
            </div>
            <div className="flex gap-6 mt-3 text-xs text-gray-500">
              <span>Referrals: <span className="text-gray-300">{data.referral_count}</span></span>
              <span>Earned: <span className="text-gray-300">{data.total_earned_display}</span></span>
            </div>
          </div>

          {/* Referral link */}
          <div className="panel p-6">
            <h3 className="font-semibold mb-4">Referral link</h3>
            <div className="inset flex items-center gap-2 px-4 py-3">
              <input
                type="text"
                readOnly
                value={data.referral_link}
                className="flex-1 bg-transparent text-sm text-emerald-400 focus:outline-none"
              />
              <button
                onClick={() => copyToClipboard(data.referral_link, 'Link')}
                className="text-gray-400 hover:text-white p-1"
                aria-label="Copy link"
              >
                <CopyIcon />
              </button>
            </div>
            <div className="flex gap-6 mt-3 text-xs text-gray-500">
              <span>Commission: <span className="text-gray-300">{data.commission_display}</span> per sale</span>
              <span>Destination: <span className="text-gray-300">/</span></span>
            </div>

            {/* Prominent share buttons for instant gratification */}
            <div className="mt-4">
              <p className="text-xs text-gray-500 mb-2 font-medium">Share your link</p>
              <div className="flex flex-wrap gap-2">
                <button
                  onClick={() => copyToClipboard(data.referral_link, 'Referral link')}
                  className="btn"
                >
                  📋 Copy Link
                </button>
                <a
                  href={`https://x.com/intent/tweet?text=${encodeURIComponent(`Check out Aura Optimizer! Use my referral link: ${data.referral_link}`)}`}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="btn bg-black hover:bg-[#111] text-white"
                  onClick={() => toast.success('Opening X...')}
                >
                  𝕏 Share on X
                </a>
                <button
                  onClick={() => {
                    const msg = `Hey! I've been using Aura Optimizer and it's awesome. Use my referral link to get store credit: ${data.referral_link}`;
                    copyToClipboard(msg, 'Discord message');
                  }}
                  className="btn bg-[#5865F2] hover:bg-[#4752C4]"
                >
                  💬 Copy for Discord
                </button>
                <a
                  href={`mailto:?subject=${encodeURIComponent('Join me on Aura Optimizer')}&body=${encodeURIComponent(`Hey!\n\nI've been using Aura Optimizer and think you'll love it too.\n\nUse my referral link: ${data.referral_link}\n\n`)}`}
                  className="btn bg-gray-700 hover:bg-gray-600"
                >
                  ✉️ Email Link
                </a>
              </div>
              <p className="text-[10px] text-gray-500 mt-2">Every share can earn you $5.99 in store credit per Lifetime Pro sale.</p>
            </div>
          </div>

          {/* Referrals table */}
          <div className="panel p-6">
            <h3 className="font-semibold mb-4">Referrals</h3>
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="text-left text-xs text-gray-500 border-b border-white/10">
                    <th className="py-2 font-medium">Date</th>
                    <th className="py-2 font-medium">Product</th>
                    <th className="py-2 font-medium">Status</th>
                    <th className="py-2 font-medium text-right">Commission</th>
                  </tr>
                </thead>
                <tbody>
                  {data.referrals.length > 0 ? (
                    data.referrals.map((r, idx) => (
                      <tr key={idx} className="border-b border-white/5">
                        <td className="py-3 text-gray-300">{r.date}</td>
                        <td className="py-3 text-gray-300">{r.product}</td>
                        <td className="py-3">
                          <span className="text-emerald-400 text-xs bg-emerald-500/10 px-2 py-0.5 rounded">{r.status}</span>
                        </td>
                        <td className="py-3 text-right text-gray-300">{r.commission}</td>
                      </tr>
                    ))
                  ) : (
                    <tr>
                      <td colSpan={4} className="py-8 text-center text-gray-600 text-sm">
                        No referrals yet — share your link to start earning.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        </div>

        {/* RIGHT COLUMN - Details */}
        <div className="space-y-6">
          <h2 className="text-sm text-gray-400 font-medium tracking-wider">DETAILS</h2>

          <div className="panel p-6">
            <div className="flex items-center gap-3 mb-5">
              <div className="w-10 h-10 rounded-full bg-emerald-900/40 flex items-center justify-center text-emerald-400 text-lg font-medium">
                {data.user.name?.[0]?.toUpperCase() || data.user.email[0]?.toUpperCase()}
              </div>
              <div>
                <p className="font-semibold">{data.user.name || 'User'}</p>
                <p className="text-xs text-gray-500">{data.user.email}</p>
              </div>
            </div>

            {data.referred_by_code && (
              <div className="mb-4">
                <p className="text-xs text-gray-500 mb-1">Referred by</p>
                <div className="inset px-3 py-2 text-sm text-emerald-400">
                  {data.referred_by_name ? `${data.referred_by_name} (${data.referred_by_code})` : data.referred_by_code}
                </div>
              </div>
            )}

            <p className="text-xs text-gray-500 mb-1.5">Social Name (your referral code)</p>
            <div className="flex items-center gap-2">
              <input
                type="text"
                value={editCode}
                onChange={(e) => setEditCode(e.target.value)}
                className="inset flex-1 px-3 py-2 text-sm text-white focus:outline-none focus:border-emerald-500/50"
              />
              <button
                onClick={saveReferralCode}
                disabled={savingCode}
                className="bg-emerald-600 hover:bg-emerald-500 text-white text-sm font-semibold px-4 py-2 rounded-lg disabled:opacity-60"
              >
                {savingCode ? 'Saving…' : 'Save'}
              </button>
            </div>
            {codeMsg && (
              <p className={`text-xs mt-2 ${codeMsg.type === 'success' ? 'text-emerald-400' : codeMsg.type === 'error' ? 'text-red-400' : 'text-gray-400'}`}>
                {codeMsg.text}
              </p>
            )}

            <div className="border-t border-white/10 mt-6 pt-5">
              <p className="text-xs text-gray-500 mb-1">Commission</p>
              <p className="text-sm text-gray-300">{data.commission_display} per referred sale</p>
            </div>
          </div>

          <div className="panel p-6">
            <p className="text-xs text-gray-500 mb-1">Payout</p>
            <p className="text-sm text-gray-300">
              Paid as <span className="text-emerald-400">store credit</span> — automatically applied at checkout on your next purchase.
            </p>
          </div>
        </div>
      </div>
    </div>
    </section>
  );
}

function CopyIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
      <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
      <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
    </svg>
  );
}
