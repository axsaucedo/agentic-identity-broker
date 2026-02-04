-- Rollback: Remove protected_resources column and index from thirdparty_oauth2_services table

-- Remove index first (to avoid dependency issues)
DROP INDEX IF EXISTS idx_thirdparty_oauth2_services_protected_resources;

-- Remove the protected_resources column
ALTER TABLE thirdparty_oauth2_services
DROP COLUMN IF EXISTS protected_resources;
