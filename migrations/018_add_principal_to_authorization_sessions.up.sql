ALTER TABLE authorization_sessions ADD COLUMN principal TEXT NOT NULL DEFAULT '';
-- Remove sessions created without a principal (pre-018 rows). Authorization sessions
-- have a 10-minute TTL, so this only affects sessions created during rolling deploys.
DELETE FROM authorization_sessions WHERE principal = '';
ALTER TABLE authorization_sessions ALTER COLUMN principal DROP DEFAULT;
