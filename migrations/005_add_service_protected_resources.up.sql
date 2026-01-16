-- Add protected_resources column to thirdparty_services table for RFC 8693 token exchange
-- This allows services to specify which resource URIs they can provide tokens for

-- Add protected_resources column as TEXT array with default empty array
ALTER TABLE thirdparty_services
ADD COLUMN protected_resources TEXT[] DEFAULT '{}';

-- Create GIN index for efficient array containment lookups (@> operator)
-- This enables fast queries to find services by resource URI
CREATE INDEX idx_thirdparty_services_protected_resources
ON thirdparty_services USING GIN (protected_resources);

-- Add column comment for documentation
COMMENT ON COLUMN thirdparty_services.protected_resources IS
'Array of normalized resource URIs that map to this service for RFC 8693 token exchange. URIs are normalized with trailing slashes removed for consistent matching. Used with @> operator for resource discovery queries.';

-- Add index comment for documentation
COMMENT ON INDEX idx_thirdparty_services_protected_resources IS
'GIN index enabling efficient resource URI lookups for token exchange. Supports @> containment operator for finding services by resource URI.';
