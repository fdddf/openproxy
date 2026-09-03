-- Create oauth_sessions table for storing OAuth2 session data including PKCE parameters
CREATE TABLE IF NOT EXISTS oauth_sessions (
    id SERIAL PRIMARY KEY,
    session_id VARCHAR(255) NOT NULL UNIQUE,
    provider_id INTEGER NOT NULL,
    state VARCHAR(255) NOT NULL,
    code_verifier TEXT NOT NULL,
    code_challenge VARCHAR(255) NOT NULL,
    code_challenge_method VARCHAR(20) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ
);

-- Create index on provider_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_provider_id ON oauth_sessions(provider_id);

-- Create index on state for faster lookups during callback
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_state ON oauth_sessions(state);

-- Create index on expires_at to help with cleanup of expired sessions
CREATE INDEX IF NOT EXISTS idx_oauth_sessions_expires_at ON oauth_sessions(expires_at);
