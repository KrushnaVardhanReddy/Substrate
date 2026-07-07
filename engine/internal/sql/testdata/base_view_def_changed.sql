CREATE TABLE products (id BIGINT PRIMARY KEY, name TEXT, active BOOLEAN);
CREATE VIEW active_products AS SELECT id, name FROM products WHERE active = true;
