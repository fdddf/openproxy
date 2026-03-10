-- Remove OAuth2 fields from providers table
ALTER TABLE providers 
DROP COLUMN IF EXISTS client_id,
DROP COLUMN IF EXISTS client_secret,
DROP COLUMN IF EXISTS access_token,
DROP COLUMN IF EXISTS refresh_token,
DROP COLUMN IF EXISTS token_expiry,
DROP COLUMN IF EXISTS auth_url,
DROP COLUMN IF EXISTS token_url,
DROP COLUMN IF EXISTS redirect_url,
DROP COLUMN IF EXISTS scopes;
