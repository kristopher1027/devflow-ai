CREATE TABLE github_repository_import_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    requester_id UUID NOT NULL,

    project_id UUID NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',

    attempts INTEGER NOT NULL DEFAULT 0,

    failure_code TEXT,

    failure_message TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    completed_at TIMESTAMPTZ,

    CONSTRAINT fk_import_jobs_requester
        FOREIGN KEY(requester_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_import_jobs_project
        FOREIGN KEY(project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_import_jobs_status
        CHECK (status IN ('pending', 'running', 'succeeded', 'failed')),

    CONSTRAINT chk_import_jobs_attempts
        CHECK (attempts >= 0)
);

CREATE INDEX idx_import_jobs_project_id
    ON github_repository_import_jobs(project_id);

CREATE INDEX idx_import_jobs_status
    ON github_repository_import_jobs(status);