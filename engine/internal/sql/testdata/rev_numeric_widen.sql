CREATE TABLE invoices (
  id     BIGINT PRIMARY KEY,
  amount NUMERIC(18,2) NOT NULL
);
-- NUMERIC(10,2) → NUMERIC(18,2): precision widens, same scale → COLUMN_TYPE_WIDENED
