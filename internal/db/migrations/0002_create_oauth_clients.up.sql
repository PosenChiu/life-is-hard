CREATE TABLE oauth_clients (
    id            SERIAL        PRIMARY KEY,
    client_id     TEXT          UNIQUE NOT NULL,
    client_secret TEXT          NOT NULL,
    owner_id      INTEGER       REFERENCES users(id) ON DELETE SET NULL,
    grant_types   TEXT[]        NOT NULL DEFAULT ARRAY['password','client_credentials']::TEXT[],
    created_at    TIMESTAMPTZ   NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ   NOT NULL DEFAULT now()
);

CREATE INDEX idx_oauth_clients_client_id ON oauth_clients(client_id);