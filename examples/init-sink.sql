-- ============================================
-- SINK DATABASE INITIALIZATION
-- Mirror tables from source (no constraints on FKs for flexibility)
-- ============================================

CREATE TABLE IF NOT EXISTS users (
    id INT PRIMARY KEY,
    username VARCHAR(100),
    email VARCHAR(255),
    full_name VARCHAR(255),
    status VARCHAR(20),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    id INT PRIMARY KEY,
    name VARCHAR(255),
    sku VARCHAR(50),
    price NUMERIC(12, 2),
    stock INT,
    category VARCHAR(100),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS orders (
    id INT PRIMARY KEY,
    user_id INT,
    product_id INT,
    quantity INT,
    total_amount NUMERIC(12, 2),
    status VARCHAR(20),
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- Indexes for query performance on sink side
CREATE INDEX idx_sink_users_email ON users(email);
CREATE INDEX idx_sink_orders_user_id ON orders(user_id);
CREATE INDEX idx_sink_products_category ON products(category);
