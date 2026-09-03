-- The APIKey model has carried quota/used/reset_time fields with no matching
-- columns, so any INSERT built from the struct failed. Add them.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS quota BIGINT,
    ADD COLUMN IF NOT EXISTS used BIGINT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reset_time TIMESTAMPTZ;
