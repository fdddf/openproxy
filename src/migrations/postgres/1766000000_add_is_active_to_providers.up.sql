-- Add is_active field to providers table
ALTER TABLE providers 
ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
