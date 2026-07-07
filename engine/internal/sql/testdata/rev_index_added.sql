CREATE TABLE users (id BIGINT PRIMARY KEY, email TEXT);
CREATE INDEX idx_email ON users(email);
