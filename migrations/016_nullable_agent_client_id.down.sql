DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM agents WHERE client_id IS NULL) THEN
    RAISE EXCEPTION
      'Cannot roll back migration 016: % agent(s) have no client_id — assign a client_id to each before downgrading',
      (SELECT COUNT(*) FROM agents WHERE client_id IS NULL);
  END IF;
END $$;
ALTER TABLE agents ALTER COLUMN client_id SET NOT NULL;
