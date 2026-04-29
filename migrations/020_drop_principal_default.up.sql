-- Remove the default '' from principal now that all writers populate it.
-- Apply only after all application instances have been updated to pass principal.
ALTER TABLE authorization_sessions ALTER COLUMN principal DROP DEFAULT;
