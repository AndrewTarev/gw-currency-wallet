CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE wallets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    balance_rub DECIMAL(20, 2) NOT NULL DEFAULT 0,
    balance_usd DECIMAL(20, 2) NOT NULL DEFAULT 0,
    balance_eur DECIMAL(20, 2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);