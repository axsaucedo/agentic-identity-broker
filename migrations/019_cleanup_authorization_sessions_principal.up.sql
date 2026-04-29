-- Remove sessions created without a principal (any rows that pre-date migration 018).
-- Authorization sessions have a 10-minute TTL so this only affects sessions
-- created during a rolling deploy window.
DELETE FROM authorization_sessions WHERE principal = '';
ALTER TABLE authorization_sessions ALTER COLUMN principal DROP DEFAULT;
