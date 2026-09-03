-- Remove is_active field from providers table
ALTER TABLE providers 
DROP COLUMN IF EXISTS is_active;
