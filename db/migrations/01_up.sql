CREATE TABLE IF NOT EXISTS "user"
(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    coins INTEGER NOT NULL DEFAULT 1000 CHECK (COINS >= 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "product"
(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    price INTEGER NOT NULL CHECK (price > 0),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "transaction"
(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    amount INTEGER NOT NULL CHECK (amount > 0),
    from_user_id INTEGER NOT NULL,
    to_user_id INTEGER NOT NULL,
    FOREIGN KEY (from_user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    FOREIGN KEY (TO_user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "purchase"
(
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id INTEGER NOT NULL,
    product_id INTEGER NOT NULL,
    FOREIGN KEY (user_id) REFERENCES "user"(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES "product"(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COPY "product"(name, price)
    FROM '/docker-entrypoint-initdb.d/products.csv'
    WITH (FORMAT csv, HEADER true,  DELIMITER ';');

CREATE UNIQUE INDEX idx_user_username ON "user" (username);
CREATE INDEX idx_user_id ON "user" (id);
CREATE INDEX idx_purchase_user_id ON "purchase" (user_id);
CREATE INDEX idx_transaction_from_user ON "transaction" (from_user_id);
CREATE INDEX idx_transaction_to_user ON "transaction" (to_user_id);
