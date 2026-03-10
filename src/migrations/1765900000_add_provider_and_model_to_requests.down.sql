DROP INDEX IF EXISTS idx_requests_provider_id;
DROP INDEX IF EXISTS idx_requests_model_id;

ALTER TABLE requests
    DROP COLUMN IF EXISTS provider_id,
    DROP COLUMN IF EXISTS model_id;
