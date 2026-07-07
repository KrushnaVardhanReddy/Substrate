CREATE TABLE products (id BIGINT PRIMARY KEY, name TEXT, active BOOLEAN);
CREATE VIEW active_products AS SELECT id, name FROM products WHERE active = true AND id > 0;
-- same columns (id, name), different WHERE clause → VIEW_DEFINITION_CHANGED
