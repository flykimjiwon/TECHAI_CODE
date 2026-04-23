-- SuperSOL Demo Database Schema
-- 신한은행 슈퍼솔 시연용 데이터베이스

CREATE TABLE IF NOT EXISTS customers (
    customer_id     SERIAL PRIMARY KEY,
    name            VARCHAR(50)  NOT NULL,
    phone           VARCHAR(20),
    membership      VARCHAR(20)  DEFAULT 'STANDARD',  -- STANDARD, PREMIER, VIP
    created_at      TIMESTAMP    DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS accounts (
    account_id      SERIAL PRIMARY KEY,
    customer_id     INTEGER      REFERENCES customers(customer_id),
    account_number  VARCHAR(20)  NOT NULL UNIQUE,
    account_name    VARCHAR(50)  NOT NULL,
    bank_name       VARCHAR(20)  DEFAULT '신한은행',
    balance         BIGINT       NOT NULL DEFAULT 0,
    account_type    VARCHAR(20)  DEFAULT 'CHECKING',  -- CHECKING, SAVINGS, ISA
    is_primary      BOOLEAN      DEFAULT FALSE,
    created_at      TIMESTAMP    DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    tx_id           SERIAL PRIMARY KEY,
    account_id      INTEGER      REFERENCES accounts(account_id),
    tx_type         VARCHAR(10)  NOT NULL,  -- DEPOSIT, WITHDRAW, TRANSFER
    amount          BIGINT       NOT NULL,
    balance_after   BIGINT       NOT NULL,
    description     VARCHAR(100),
    counterparty    VARCHAR(50),
    category        VARCHAR(30),            -- SALARY, FOOD, TRANSPORT, SHOPPING, TRANSFER
    created_at      TIMESTAMP    DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS cards (
    card_id         SERIAL PRIMARY KEY,
    customer_id     INTEGER      REFERENCES customers(customer_id),
    card_name       VARCHAR(50)  NOT NULL,
    card_number     VARCHAR(20),
    monthly_usage   BIGINT       DEFAULT 0,
    card_type       VARCHAR(20)  DEFAULT 'CREDIT'
);

CREATE TABLE IF NOT EXISTS stocks (
    stock_id        SERIAL PRIMARY KEY,
    customer_id     INTEGER      REFERENCES customers(customer_id),
    symbol          VARCHAR(20)  NOT NULL,
    name            VARCHAR(50)  NOT NULL,
    quantity        INTEGER      DEFAULT 0,
    avg_price       DECIMAL(12,2),
    current_price   DECIMAL(12,2),
    change_pct      DECIMAL(6,2),
    market          VARCHAR(10)  DEFAULT 'KR',  -- KR, US
    is_watchlist    BOOLEAN      DEFAULT FALSE
);

-- 인덱스: 거래내역 조회 성능용
CREATE INDEX IF NOT EXISTS idx_transactions_account_date
    ON transactions(account_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_transactions_created_at
    ON transactions(created_at);

CREATE INDEX IF NOT EXISTS idx_accounts_customer
    ON accounts(customer_id);
