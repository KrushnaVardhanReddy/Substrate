CREATE TYPE order_status AS ENUM ('pending', 'shipped', 'delivered');
CREATE TABLE orders (
  id     BIGINT PRIMARY KEY,
  status TEXT
);
