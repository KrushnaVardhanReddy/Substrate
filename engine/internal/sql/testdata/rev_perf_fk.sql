CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT
);

ALTER TABLE orders ADD CONSTRAINT fk_orders_user FOREIGN KEY (user_id) REFERENCES users (id);
