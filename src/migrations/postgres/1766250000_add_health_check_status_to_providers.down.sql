ALTER TABLE providers
DROP COLUMN IF EXISTS health_check_status,
DROP COLUMN IF EXISTS health_check_checked;
