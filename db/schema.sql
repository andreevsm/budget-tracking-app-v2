-- ============================================================================
-- Budget tracking app — PostgreSQL schema
--
-- Conventions:
--   * Money is NUMERIC(20,4), never float. On transactions the amount is
--     SIGNED: positive = money in, negative = money out.
--   * Every user-owned row carries user_id, so queries filter by owner in one
--     place and ON DELETE CASCADE keeps cleanup simple.
--   * Soft delete (deleted_at) only on transactions — the ledger is append-
--     oriented and you want edits/corrections visible. Lookup tables use
--     is_archived instead.
--   * Budgets are always valued in users.base_currency; conversion of foreign
--     currency transactions happens in queries via exchange_rates.
-- ============================================================================

BEGIN;

CREATE EXTENSION IF NOT EXISTS pgcrypto;  -- gen_random_uuid() on PG < 13

-- ---------------------------------------------------------------------------
-- Enums
-- ---------------------------------------------------------------------------
CREATE TYPE account_type          AS ENUM ('checking', 'savings', 'credit_card', 'cash', 'investment', 'loan', 'other');
CREATE TYPE transaction_type      AS ENUM ('income', 'expense', 'transfer');
CREATE TYPE category_kind         AS ENUM ('income', 'expense');
CREATE TYPE budget_period         AS ENUM ('weekly', 'monthly', 'yearly', 'custom');
CREATE TYPE recurrence_frequency  AS ENUM ('daily', 'weekly', 'monthly', 'yearly');

-- ---------------------------------------------------------------------------
-- Currencies & exchange rates
-- ---------------------------------------------------------------------------
CREATE TABLE currencies (
    code        CHAR(3) PRIMARY KEY,                    -- ISO 4217
    name        TEXT NOT NULL,
    symbol      TEXT NOT NULL,
    minor_units SMALLINT NOT NULL DEFAULT 2
                CHECK (minor_units BETWEEN 0 AND 4)     -- JPY=0, KWD=3
);

CREATE TABLE exchange_rates (
    base     CHAR(3) NOT NULL REFERENCES currencies(code),
    quote    CHAR(3) NOT NULL REFERENCES currencies(code),
    rate     NUMERIC(20, 10) NOT NULL CHECK (rate > 0), -- price of 1 base in quote
    rated_on DATE NOT NULL,
    PRIMARY KEY (base, quote, rated_on)
);

CREATE INDEX idx_exchange_rates_quote ON exchange_rates (quote, rated_on);

-- ---------------------------------------------------------------------------
-- Users
-- ---------------------------------------------------------------------------
CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         TEXT NOT NULL UNIQUE CHECK (email = lower(email)),
    password_hash TEXT NOT NULL,                        -- argon2id / bcrypt, never plaintext
    display_name  TEXT NOT NULL,
    base_currency CHAR(3) NOT NULL REFERENCES currencies(code),
    timezone      TEXT NOT NULL DEFAULT 'UTC',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- ---------------------------------------------------------------------------
-- Accounts (wallets, cards, cash)
-- ---------------------------------------------------------------------------
CREATE TABLE accounts (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id              UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name                 TEXT NOT NULL,
    account_type         account_type NOT NULL DEFAULT 'checking',
    currency             CHAR(3) NOT NULL REFERENCES currencies(code),
    -- Signed: a credit card with debt starts negative.
    opening_balance      NUMERIC(20, 4) NOT NULL DEFAULT 0,
    opening_balance_date DATE NOT NULL DEFAULT CURRENT_DATE,
    credit_limit         NUMERIC(20, 4),
    is_archived          BOOLEAN NOT NULL DEFAULT false,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name),
    CHECK (account_type = 'credit_card' OR credit_limit IS NULL)
);

CREATE INDEX idx_accounts_user ON accounts (user_id) WHERE NOT is_archived;

-- ---------------------------------------------------------------------------
-- Categories (two-level tree: "Food" -> "Groceries")
-- ---------------------------------------------------------------------------
CREATE TABLE categories (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id   UUID REFERENCES categories(id),         -- NULL = top level
    name        TEXT NOT NULL,
    kind        category_kind NOT NULL,
    color       TEXT,                                    -- UI hint, e.g. '#ff6b6b'
    icon        TEXT,
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, name)
    -- Deleting a parent with children is blocked (NO ACTION); reassign or
    -- delete them first. Cycles are prevented in application code.
);

CREATE INDEX idx_categories_user ON categories (user_id, kind) WHERE NOT is_archived;

-- ---------------------------------------------------------------------------
-- Recurring transactions (rent, salary, subscriptions)
-- ---------------------------------------------------------------------------
CREATE TABLE recurring_transactions (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id      UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id   UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    category_id  UUID REFERENCES categories(id) ON DELETE SET NULL,
    type         transaction_type NOT NULL CHECK (type <> 'transfer'),
    amount       NUMERIC(20, 4) NOT NULL CHECK (amount > 0),  -- always positive here
    frequency    recurrence_frequency NOT NULL,
    interval     INT NOT NULL DEFAULT 1 CHECK (interval > 0), -- every N periods
    day_of_month SMALLINT CHECK (day_of_month BETWEEN 1 AND 31),
    next_due_on  DATE NOT NULL,
    end_on       DATE,
    payee        TEXT,
    description  TEXT,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_recurring_due ON recurring_transactions (next_due_on) WHERE is_active;

-- ---------------------------------------------------------------------------
-- Transactions (the ledger; a transfer is TWO linked rows)
-- ---------------------------------------------------------------------------
CREATE TABLE transactions (
    id                        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id                   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id                UUID NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    category_id               UUID REFERENCES categories(id) ON DELETE SET NULL,
    type                      transaction_type NOT NULL,
    -- Signed amount, always in the account's currency (join to get it):
    --   income leg of transfer  > 0, expense / out leg of transfer < 0
    amount                    NUMERIC(20, 4) NOT NULL CHECK (amount <> 0),
    occurred_on               DATE NOT NULL,             -- user-facing date
    payee                     TEXT,                      -- merchant / counterparty
    description               TEXT,
    notes                     TEXT,
    -- Groups the two legs of a transfer: same value on the out-row (-X) and
    -- the in-row (+X). Pairing is enforced in app code.
    transfer_group_id         UUID,
    recurring_transaction_id  UUID REFERENCES recurring_transactions(id) ON DELETE SET NULL,
    external_id               TEXT,                      -- dedup key for CSV / bank imports
    created_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at                TIMESTAMPTZ,
    CHECK (type <> 'transfer' OR category_id IS NULL),   -- transfers carry no category
    CHECK (type <> 'income'   OR amount > 0),
    CHECK (type <> 'expense'  OR amount < 0)
);

CREATE INDEX idx_transactions_user_date     ON transactions (user_id, occurred_on DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_account_date  ON transactions (account_id, occurred_on DESC);
CREATE INDEX idx_transactions_category      ON transactions (category_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_transactions_transfer      ON transactions (transfer_group_id) WHERE transfer_group_id IS NOT NULL;
-- Import dedup: one row per external id per account.
CREATE UNIQUE INDEX uq_transactions_external ON transactions (account_id, external_id)
    WHERE external_id IS NOT NULL AND deleted_at IS NULL;

-- ---------------------------------------------------------------------------
-- Tags (many-to-many with transactions)
-- ---------------------------------------------------------------------------
CREATE TABLE tags (
    id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name    TEXT NOT NULL,
    UNIQUE (user_id, name)
);

CREATE TABLE transaction_tags (
    transaction_id UUID NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id         UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);

-- ---------------------------------------------------------------------------
-- Budgets: a budget covers a period and limits spending per category
-- ---------------------------------------------------------------------------
CREATE TABLE budgets (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name        TEXT NOT NULL,
    period      budget_period NOT NULL DEFAULT 'monthly',
    start_date  DATE NOT NULL,                           -- align to period start in app
    end_date    DATE,                                    -- NULL = recurring, never ends
    rollover    BOOLEAN NOT NULL DEFAULT false,          -- unspent amount carries over
    is_archived BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CHECK (end_date IS NULL OR end_date > start_date)
);

CREATE TABLE budget_items (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    budget_id    UUID NOT NULL REFERENCES budgets(id) ON DELETE CASCADE,
    category_id  UUID NOT NULL REFERENCES categories(id),
    amount_limit NUMERIC(20, 4) NOT NULL CHECK (amount_limit > 0),  -- in users.base_currency
    UNIQUE (budget_id, category_id)
);

CREATE INDEX idx_budgets_user ON budgets (user_id) WHERE NOT is_archived;

-- ---------------------------------------------------------------------------
-- updated_at maintenance (drop these if you set it in Go instead)
-- ---------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION touch_updated_at() RETURNS trigger AS $$
BEGIN
    NEW.updated_at := now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_users_touch    BEFORE UPDATE ON users                   FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
CREATE TRIGGER trg_accounts_touch BEFORE UPDATE ON accounts                FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
CREATE TRIGGER trg_cats_touch     BEFORE UPDATE ON categories              FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
CREATE TRIGGER trg_txns_touch     BEFORE UPDATE ON transactions            FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
CREATE TRIGGER trg_budgets_touch  BEFORE UPDATE ON budgets                 FOR EACH ROW EXECUTE FUNCTION touch_updated_at();
CREATE TRIGGER trg_recurs_touch   BEFORE UPDATE ON recurring_transactions   FOR EACH ROW EXECUTE FUNCTION touch_updated_at();

-- ---------------------------------------------------------------------------
-- Views
-- ---------------------------------------------------------------------------

-- Current balance per account = opening + sum of live transactions.
CREATE VIEW account_balances AS
SELECT a.id AS account_id,
       a.user_id,
       a.name,
       a.account_type,
       a.currency,
       a.opening_balance + COALESCE(SUM(t.amount), 0) AS balance
FROM accounts a
LEFT JOIN transactions t
       ON t.account_id = a.id AND t.deleted_at IS NULL
GROUP BY a.id;

-- Monthly spend per category (in each account's native currency).
CREATE VIEW monthly_category_expenses AS
SELECT t.user_id,
       date_trunc('month', t.occurred_on)::date AS month,
       t.category_id,
       a.currency,
       SUM(-t.amount) AS spent
FROM transactions t
JOIN accounts a ON a.id = t.account_id
WHERE t.type = 'expense'
  AND t.deleted_at IS NULL
GROUP BY 1, 2, 3, 4;

-- ---------------------------------------------------------------------------
-- Seed currencies
-- ---------------------------------------------------------------------------
INSERT INTO currencies (code, name, symbol, minor_units) VALUES
    ('RUB', 'Russian ruble',       '₽', 2),
    ('USD', 'US dollar',           '$', 2),
    ('EUR', 'Euro',                '€', 2),
    ('GBP', 'Pound sterling',      '£', 2),
    ('KZT', 'Kazakhstani tenge',   '₸', 2),
    ('AED', 'UAE dirham',          'د.إ', 2),
    ('CNY', 'Chinese yuan',        '¥', 2),
    ('JPY', 'Japanese yen',        '¥', 0);

COMMIT;
