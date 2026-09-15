WITH ranked_current_keys AS (
    SELECT id,
           ROW_NUMBER() OVER (
               ORDER BY activates_at DESC, created_at DESC, id DESC
           ) AS row_num
    FROM signing_keys
    WHERE removed_at IS NULL
      AND is_current = true
)
UPDATE signing_keys
SET is_current = false
WHERE id IN (
    SELECT id
    FROM ranked_current_keys
    WHERE row_num > 1
);

-- PostgreSQL does not support DEFERRABLE partial unique indexes, so current-key
-- transitions must demote existing current rows before promoting or inserting the
-- next current row. Keep SetCurrent/CreateAndSetCurrent as ordered multi-statement
-- updates to avoid transient uniqueness conflicts.
CREATE UNIQUE INDEX idx_signing_keys_single_current_active
    ON signing_keys (is_current)
    WHERE removed_at IS NULL AND is_current = true;
