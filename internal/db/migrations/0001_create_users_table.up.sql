CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);
INSERT INTO users (name, email, password_hash, is_admin)
VALUES
  ('admin', 'admin@example.com', '$2a$10$2Two2V3hfv.TpJnLjQ5awOGlFpzIr4bXGQyUTBStDw8PRO/oble/K', TRUE);
