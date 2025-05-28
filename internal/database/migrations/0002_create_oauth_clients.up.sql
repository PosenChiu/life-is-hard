CREATE TABLE oauth_clients (
    client_id     TEXT          PRIMARY KEY,
    client_secret TEXT          NOT NULL,
    user_id       INTEGER       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    grant_types   TEXT[]        NOT NULL DEFAULT ARRAY['password','client_credentials','refresh_token']::TEXT[],
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);


INSERT INTO oauth_clients (client_id, client_secret, user_id)
SELECT 'admin', 'admin', id FROM users WHERE name = 'admin';
