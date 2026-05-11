-- license_keys: durable record of every issued license.
-- UNIQUE(user_email, product_id) enforces "one license per user-product pair"
-- at the DB layer, which is what makes the webhook handler safe against retries
-- after a process restart wipes the in-memory event dedup cache.
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
