CREATE TABLE oauth_clients (
    id            SERIAL        PRIMARY KEY,
    client_id     TEXT          UNIQUE NOT NULL,
    client_secret TEXT          NOT NULL,
    owner_id      INTEGER       NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    grant_types   TEXT[]        NOT NULL DEFAULT ARRAY['password','client_credentials']::TEXT[],
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_oauth_clients_client_id ON oauth_clients(client_id);

INSERT INTO oauth_clients (client_id, client_secret, owner_id, grant_types)
SELECT
  'admin',
  'admin',
  id,
  ARRAY['password','client_credentials','refresh_token']::TEXT[]
FROM users
WHERE name = 'admin';
