CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     username VARCHAR(50) UNIQUE NOT NULL,
     email VARCHAR(100) UNIQUE NOT NULL,
     password_hash VARCHAR(255) NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS accounts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    account_number VARCHAR(50) UNIQUE NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'RUB',
    balance NUMERIC(15, 2) DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS cards (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     user_id UUID REFERENCES users(id) ON DELETE CASCADE,
     account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
     encrypted_number TEXT NOT NULL,
     encrypted_expiration TEXT NOT NULL,
     cvv_hash TEXT NOT NULL,
     hmac_signature TEXT NOT NULL,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender_account_id UUID REFERENCES accounts(id),
    receiver_account_id UUID REFERENCES accounts(id),
    amount NUMERIC(15, 2) NOT NULL,
    transaction_type VARCHAR(20) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS credits (
   id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
   user_id UUID REFERENCES users(id),
   account_id UUID REFERENCES accounts(id),
   principal NUMERIC(15, 2) NOT NULL,
   interest_rate NUMERIC(5, 2) NOT NULL,
   term_months INT NOT NULL,
   status VARCHAR(20) NOT NULL DEFAULT 'active',
   created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payment_schedules (
     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
     credit_id UUID REFERENCES credits(id) ON DELETE CASCADE,
     payment_date TIMESTAMP WITH TIME ZONE NOT NULL,
     total_payment NUMERIC(15, 2) NOT NULL,
     principal_part NUMERIC(15, 2) NOT NULL,
     interest_part NUMERIC(15, 2) NOT NULL,
     status VARCHAR(20) NOT NULL DEFAULT 'pending',
     penalty NUMERIC(15, 2) DEFAULT 0.00,
     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
     updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
