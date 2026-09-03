-- Supports the retention sweeper and the dashboard's time-ordered listing.
CREATE INDEX IF NOT EXISTS idx_requests_created_at ON requests(created_at);
