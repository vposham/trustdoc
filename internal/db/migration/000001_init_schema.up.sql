-- Create tables

-- documents table maintains the documents uploaded by the users and its metadata.
CREATE TABLE documents
(
    id              BIGSERIAL PRIMARY KEY,
    doc_id          VARCHAR(255) NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    file_name       VARCHAR(255) NOT NULL,
    blockchain_hash VARCHAR(64)  NOT NULL,
    uploaded_by     BIGINT       NOT NULL,
    modified_at     timestamptz  NOT NULL DEFAULT NOW(),
    uploaded_at     timestamptz  NOT NULL DEFAULT NOW(),
    last_updated_at timestamptz  NOT NULL DEFAULT NOW()
);

-- doc_id is unique
ALTER TABLE documents
    ADD CONSTRAINT doc_id_un_key UNIQUE (doc_id);

-- users table maintains the user who uploaded the document.
CREATE TABLE users
(
    id              BIGSERIAL PRIMARY KEY,
    user_id         VARCHAR(50)  NOT NULL,
    first_name      VARCHAR(100) NOT NULL,
    last_name       VARCHAR(100) NOT NULL,
    is_active       BOOLEAN      NOT NULL DEFAULT TRUE,
    created_at      timestamptz  NOT NULL DEFAULT NOW(),
    last_updated_at timestamptz  NOT NULL DEFAULT NOW()
);

-- user_id is unique
ALTER TABLE users
    ADD CONSTRAINT user_id_un_key UNIQUE (user_id);

-- foreign key constraint between documents and users table to know the document ownership
ALTER TABLE documents
    ADD CONSTRAINT document_owner_user_fkey FOREIGN KEY (uploaded_by) REFERENCES users (id);

-- Create the trigger function to update the last_updated_at column
CREATE OR REPLACE
    FUNCTION update_change_timestamp_column()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS
$function$
BEGIN
    NEW.last_updated_at = NOW();
    RETURN new;
END;
$function$;


-- Create triggers
CREATE TRIGGER update_documents_change_timestamp
    BEFORE
        UPDATE
    ON
        documents
    FOR EACH ROW
EXECUTE FUNCTION update_change_timestamp_column();

CREATE TRIGGER update_users_change_timestamp
    BEFORE
        UPDATE
    ON
        users
    FOR EACH ROW
EXECUTE FUNCTION update_change_timestamp_column();
