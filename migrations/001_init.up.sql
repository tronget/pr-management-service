CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE teams
(
    team_name TEXT PRIMARY KEY,
    CHECK (length(trim(team_name)) > 0 )
);

CREATE TABLE users
(
    user_id    TEXT PRIMARY KEY,
    username   TEXT        NOT NULL,
    team_name  TEXT        NOT NULL REFERENCES teams (team_name) ON DELETE RESTRICT,
    is_active  BOOLEAN     NOT NULL DEFAULT TRUE,
    CHECK (length(trim(username)) > 0 and length(trim(user_id)) > 0)
);


CREATE TABLE pull_requests
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

CREATE TABLE pull_request_reviewers
(
    pull_request_id TEXT        NOT NULL REFERENCES pull_requests (pull_request_id) ON DELETE CASCADE,
    reviewer_id     TEXT        NOT NULL REFERENCES users (user_id) ON DELETE RESTRICT,
    assigned_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, reviewer_id)
);

