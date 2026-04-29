-- Remove sessions created without a principal (any rows that pre-date migration 018).
-- Authorization sessions have a 10-minute TTL so this only affects sessions
-- created during a rolling deploy window.
-- NOTE: DROP DEFAULT is deferred to migration 020 so older application instances
-- can still write rows using the default '' during a rolling deploy of 019.
DELETE FROM authorization_sessions WHERE principal = '';
