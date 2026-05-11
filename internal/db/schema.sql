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
