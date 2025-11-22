CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE IF NOT EXISTS teams
(
    team_name TEXT PRIMARY KEY,
    CHECK (length(trim(team_name)) > 0 )
);

CREATE TABLE IF NOT EXISTS users
(
    user_id   TEXT PRIMARY KEY,
    username  TEXT    NOT NULL,
    team_name TEXT    NOT NULL REFERENCES teams (team_name) ON DELETE RESTRICT,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    CHECK (length(trim(username)) > 0 and length(trim(user_id)) > 0)
);

CREATE TABLE IF NOT EXISTS pull_requests
(
    pull_request_id   TEXT PRIMARY KEY,
    pull_request_name TEXT        NOT NULL,
    author_id         TEXT        NOT NULL REFERENCES users (user_id) ON DELETE RESTRICT,
    status            pr_status   NOT NULL DEFAULT 'OPEN',
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    merged_at         TIMESTAMPTZ,
    UNIQUE (pull_request_name, author_id),
    CHECK (length(trim(pull_request_id)) > 0 and length(trim(pull_request_name)) > 0),
    CHECK ((status = 'MERGED' AND merged_at IS NOT NULL) OR (status = 'OPEN' AND merged_at IS NULL))
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers
(
    pull_request_id TEXT        NOT NULL REFERENCES pull_requests (pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT        NOT NULL REFERENCES users (user_id) ON DELETE RESTRICT,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE OR REPLACE FUNCTION prevent_merged_reopen()
    RETURNS TRIGGER
    LANGUAGE plpgsql
AS
$$
BEGIN
    IF OLD.status = 'MERGED' AND NEW.status = 'OPEN' THEN
        RAISE EXCEPTION 'Cannot change status from MERGED back to OPEN'
            USING ERRCODE = '45000';
    END IF;
    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_prevent_merged_reopen
    BEFORE UPDATE OF status
    ON pull_requests
    FOR EACH ROW
    WHEN (OLD.status IS DISTINCT FROM NEW.status)
EXECUTE FUNCTION prevent_merged_reopen();

CREATE INDEX IF NOT EXISTS idx_users_team_name_active
    ON users (team_name, is_active);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id_status
    ON pull_requests (author_id, status);

CREATE INDEX IF NOT EXISTS idx_pull_request_reviewers_reviewer_id
    ON pull_request_reviewers (reviewer_id);
