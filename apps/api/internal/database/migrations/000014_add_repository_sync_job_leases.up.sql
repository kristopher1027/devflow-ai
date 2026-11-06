ALTER TABLE repository_sync_jobs
    ADD COLUMN lease_owner TEXT,
    ADD COLUMN lease_expires_at TIMESTAMPTZ;

ALTER TABLE repository_sync_jobs
    ADD CONSTRAINT chk_repository_sync_jobs_lease_pair
    CHECK ((lease_owner IS NULL) = (lease_expires_at IS NULL));

CREATE INDEX idx_repository_sync_jobs_claimable
    ON repository_sync_jobs (status, lease_expires_at, created_at, id);
