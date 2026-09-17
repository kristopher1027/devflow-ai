CREATE TABLE github_connections (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    workspace_id UUID NOT NULL,

    installation_id TEXT NOT NULL,

    account_login TEXT NOT NULL,

    status TEXT NOT NULL DEFAULT 'pending',

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_github_connections_workspace
        FOREIGN KEY(workspace_id)
        REFERENCES workspaces(id)
        ON DELETE CASCADE,

    CONSTRAINT uq_github_connections_workspace
        UNIQUE(workspace_id),

    CONSTRAINT uq_github_connections_installation
        UNIQUE(installation_id),

    CONSTRAINT chk_github_connections_status
        CHECK (
            status IN (
                'pending',
                'active',
                'disconnected',
                'error'
            )
        )
);

CREATE INDEX idx_github_connections_workspace_id
    ON github_connections(workspace_id);