CREATE TABLE repository_chunks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    file_id UUID NOT NULL,

    chunk_index INTEGER NOT NULL,

    start_line INTEGER NOT NULL,

    end_line INTEGER NOT NULL,

    character_count INTEGER NOT NULL,

    content TEXT NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT fk_repository_chunks_file
        FOREIGN KEY (file_id)
        REFERENCES repository_files(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_repository_chunks_index
        CHECK (chunk_index >= 0),

    CONSTRAINT chk_repository_chunks_start_line
        CHECK (start_line >= 1),

    CONSTRAINT chk_repository_chunks_end_line
        CHECK (end_line >= start_line),

    CONSTRAINT chk_repository_chunks_character_count
        CHECK (character_count > 0),

    CONSTRAINT chk_repository_chunks_content
        CHECK (length(content) > 0),

    CONSTRAINT uq_repository_chunks_file_index
        UNIQUE (file_id, chunk_index)
);

CREATE INDEX idx_repository_chunks_file_id
    ON repository_chunks(file_id);
