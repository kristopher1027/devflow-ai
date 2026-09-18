CREATE TABLE repository_sync_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    repository_id UUID NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',

    attempts INTEGER NOT NULL DEFAULT 0,

    failure_code TEXT,

    failure_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    completed_at TIMESTAMPTZ,

    CONSTRAINT fk_repository_sync_jobs_repository
        FOREIGN KEY (repository_id)
        REFERENCES repositories(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_repository_sync_jobs_status
        CHECK (
            status IN (
                'pending',
                'running',
                'succeeded',
                'failed'
            )
        ),

    CONSTRAINT chk_repository_sync_jobs_attempts
        CHECK (attempts >= 0)
);

CREATE INDEX idx_repository_sync_jobs_repository_id
    ON repository_sync_jobs(repository_id);

CREATE INDEX idx_repository_sync_jobs_status
    ON repository_sync_jobs(status);