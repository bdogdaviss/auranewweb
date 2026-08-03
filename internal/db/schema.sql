-- available_keys: the pool of pre-generated, unsold license keys. Rows are
-- DELETED from this table the moment a key is assigned to a buyer (atomic
-- inside the same transaction that inserts into license_keys). Pool size
-- decrements by exactly one per purchase; a sold key can never be re-picked
-- because the row no longer exists here.
CREATE TABLE IF NOT EXISTS available_keys (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    key        TEXT NOT NULL UNIQUE,
    product_id TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_available_keys_product ON available_keys(product_id, id);

-- license_keys: durable record of every issued license (the audit / receipt
-- table). UNIQUE(user_email, product_id) enforces "one license per
-- user-product pair", which makes assignment idempotent against Stripe
-- webhook retries even after a process restart wipes any in-memory state.
CREATE TABLE IF NOT EXISTS license_keys (
    id                       INTEGER PRIMARY KEY AUTOINCREMENT,
    key                      TEXT NOT NULL UNIQUE,
    user_email               TEXT NOT NULL,
    product_id               TEXT NOT NULL,
    stripe_payment_intent_id TEXT,
    created_at               TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    revoked_at               TIMESTAMP NULL,
    UNIQUE (user_email, product_id)
);

CREATE INDEX IF NOT EXISTS idx_license_keys_email ON license_keys(user_email);

-- processed_events: durable dedup for Stripe webhook deliveries. The handler
-- writes here only AFTER the email is sent, so a process crash between
-- license-insert and email-send leaves the event unmarked and Stripe's retry
-- will re-attempt the email (the license row is already there via the
-- UNIQUE constraint).
CREATE TABLE IF NOT EXISTS processed_events (
    event_id     TEXT PRIMARY KEY,
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- users: durable identity anchor for the affiliate program. The legacy
-- in-memory user map is recreated on every login; this table persists the
-- stable referral_code and store_credit balance across restarts. Email is the
-- join key to license_keys.user_email. id mirrors the uuid minted in main.go.
CREATE TABLE IF NOT EXISTS users (
    id                 TEXT PRIMARY KEY,
    email              TEXT NOT NULL UNIQUE,
    name               TEXT,
    referral_code      TEXT NOT NULL UNIQUE,
    referred_by_code   TEXT,
    store_credit_cents INTEGER NOT NULL DEFAULT 0,
    password_hash      TEXT,                       -- bcrypt; NULL for Discord-only accounts
    created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_referral_code ON users(referral_code);

-- referral_events: the idempotent commission ledger. UNIQUE(stripe_event_id)
-- is the hard guarantee that a Stripe webhook retry can never double-credit a
-- referrer — exactly the dedup role processed_events plays for license issue,
-- but carrying the payout data. One row per qualifying purchase.
CREATE TABLE IF NOT EXISTS referral_events (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    stripe_event_id   TEXT NOT NULL UNIQUE,
    payment_intent_id TEXT,
    referrer_user_id  TEXT NOT NULL,
    referrer_code     TEXT NOT NULL,
    buyer_email       TEXT NOT NULL,
    product_id        TEXT NOT NULL,
    credit_cents      INTEGER NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (referrer_user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_referral_events_referrer ON referral_events(referrer_user_id);

-- credit_ledger: append-only audit of every balance change. users.store_credit_cents
-- is the cached running total; this table is the truth/audit trail. reason is
-- 'referral_earn' (+) now, 'redeem' (-) once Phase 2 lands. ref_event_id ties a
-- row back to the Stripe event that caused it.
CREATE TABLE IF NOT EXISTS credit_ledger (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id      TEXT NOT NULL,
    delta_cents  INTEGER NOT NULL,
    reason       TEXT NOT NULL,
    ref_event_id TEXT,
    created_at   TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_credit_ledger_user ON credit_ledger(user_id);

-- redemptions: idempotent ledger for store-credit SPENDING (Phase 2). Same role
-- referral_events plays for earning: UNIQUE(stripe_event_id) guarantees a single
-- debit per Stripe event even across webhook retries. The actual balance change
-- is recorded in credit_ledger with reason='redeem'.
CREATE TABLE IF NOT EXISTS redemptions (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    stripe_event_id   TEXT NOT NULL UNIQUE,
    payment_intent_id TEXT,
    user_id           TEXT NOT NULL,
    product_id        TEXT NOT NULL,
    amount_cents      INTEGER NOT NULL,
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_redemptions_user ON redemptions(user_id);

-- leads: email captures from marketing forms (affiliate-signup, payout-waitlist).
-- UNIQUE(email, source) makes re-submits idempotent.
CREATE TABLE IF NOT EXISTS leads (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    email      TEXT NOT NULL,
    name       TEXT,
    source     TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (email, source)
);

CREATE INDEX IF NOT EXISTS idx_leads_source ON leads(source);

-- Login sessions. Durable so server restarts (and multi-machine deploys)
-- don't log everyone out; the in-memory map in main.go is a read cache.
CREATE TABLE IF NOT EXISTS sessions (
    token      TEXT PRIMARY KEY,
    user_id    TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);
