CREATE TABLE oauth_clients (
    id               SERIAL PRIMARY KEY,
    client_id        TEXT      UNIQUE NOT NULL,
    client_secret    TEXT      NOT NULL,
    name             TEXT      NOT NULL,
    owner_id         INTEGER   REFERENCES users(id) ON DELETE SET NULL,
    redirect_uris    TEXT[]    NOT NULL DEFAULT ARRAY[]::TEXT[],
    grant_types      TEXT[]    NOT NULL DEFAULT ARRAY['password','client_credentials']::TEXT[],
    scopes           TEXT[]    NOT NULL DEFAULT ARRAY['read','write']::TEXT[],
    is_confidential  BOOLEAN   NOT NULL DEFAULT TRUE,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_oauth_clients_client_id ON oauth_clients(client_id);