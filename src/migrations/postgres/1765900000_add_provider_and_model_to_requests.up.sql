ALTER TABLE requests
    ADD COLUMN IF NOT EXISTS provider_id BIGINT REFERENCES providers(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS model_id BIGINT REFERENCES models(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_requests_provider_id ON requests(provider_id);
CREATE INDEX IF NOT EXISTS idx_requests_model_id ON requests(model_id);
