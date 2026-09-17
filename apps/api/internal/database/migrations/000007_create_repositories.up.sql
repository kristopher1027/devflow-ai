CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    project_id UUID NOT NULL,

    provider TEXT NOT NULL,

    external_id TEXT NOT NULL,

    owner TEXT NOT NULL,

    name TEXT NOT NULL,

    full_name TEXT NOT NULL,

    default_branch TEXT NOT NULL,

    html_url TEXT NOT NULL,

    clone_url TEXT NOT NULL,

    is_private BOOLEAN NOT NULL DEFAULT false,

    sync_status TEXT NOT NULL DEFAULT 'pending',

    last_synced_at TIMESTAMP WITH TIME ZONE,

    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),

    CONSTRAINT fk_repositories_project
        FOREIGN KEY(project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_repositories_provider_external_id
        UNIQUE(provider, external_id),

    CONSTRAINT chk_repositories_provider
        CHECK (provider IN ('github')),

    CONSTRAINT chk_repositories_sync_status
        CHECK (
            sync_status IN (
                'pending',
                'syncing',
                'synced',
                'failed'
            )
        )
);

CREATE INDEX idx_repositories_project_id
    ON repositories(project_id);