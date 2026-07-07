CREATE TABLE order_items (
  id    BIGINT PRIMARY KEY,
  total NUMERIC(10,2) NOT NULL
);
-- 'order_item' → 'order_items': Levenshtein similarity > 0.7, same columns → TABLE_RENAMED
