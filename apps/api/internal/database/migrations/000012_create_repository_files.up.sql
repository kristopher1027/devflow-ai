CREATE TABLE repository_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    snapshot_id UUID NOT NULL,

    path TEXT NOT NULL,

    size_bytes BIGINT NOT NULL,

    language TEXT,

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_repository_files_snapshot
        FOREIGN KEY (snapshot_id)
        REFERENCES repository_snapshots(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_repository_files_path
        CHECK (length(trim(path)) > 0),

    CONSTRAINT chk_repository_files_size
        CHECK (size_bytes >= 0),

    CONSTRAINT uq_repository_files_snapshot_path
        UNIQUE (snapshot_id, path)
);

CREATE INDEX idx_repository_files_snapshot_id
    ON repository_files(snapshot_id);

CREATE INDEX idx_repository_files_path
    ON repository_files(path);