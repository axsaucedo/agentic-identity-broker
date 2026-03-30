ALTER TABLE agents ADD COLUMN redirect_uris TEXT[] NOT NULL DEFAULT '{}';
ALTER TABLE agents ADD COLUMN allowed_scopes TEXT[] NOT NULL DEFAULT '{}';
