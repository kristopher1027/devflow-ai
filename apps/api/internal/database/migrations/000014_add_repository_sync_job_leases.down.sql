DROP INDEX IF EXISTS idx_repository_sync_jobs_claimable;

ALTER TABLE repository_sync_jobs
    DROP CONSTRAINT IF EXISTS chk_repository_sync_jobs_lease_pair,
    DROP COLUMN IF EXISTS lease_expires_at,
    DROP COLUMN IF EXISTS lease_owner;
