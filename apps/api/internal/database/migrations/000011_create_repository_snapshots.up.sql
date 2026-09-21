CREATE TABLE repository_snapshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    repository_id UUID NOT NULL,

    commit_sha TEXT NOT NULL,

    branch TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_repository_snapshots_repository
        FOREIGN KEY (repository_id)
        REFERENCES repositories(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_repository_snapshots_commit_sha
        CHECK (length(trim(commit_sha)) > 0),

    CONSTRAINT chk_repository_snapshots_branch
        CHECK (length(trim(branch)) > 0),

    CONSTRAINT uq_repository_snapshots_repository_commit
        UNIQUE (repository_id, commit_sha)
);

CREATE INDEX idx_repository_snapshots_repository_id
    ON repository_snapshots(repository_id);

CREATE INDEX idx_repository_snapshots_created_at
    ON repository_snapshots(created_at DESC);