CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS pull_request_statuses
(
    id   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL UNIQUE
);

INSERT INTO pull_request_statuses (name)
VALUES ('OPEN'),
       ('MERGED')
ON CONFLICT (name) DO NOTHING;



CREATE TABLE IF NOT EXISTS users
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    username   VARCHAR(64) NOT NULL UNIQUE,
    is_active  BOOLEAN     NOT NULL     DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);



CREATE TABLE IF NOT EXISTS teams
(
    id         UUID PRIMARY KEY                  DEFAULT gen_random_uuid(),
    name       VARCHAR(64)              NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);



CREATE TABLE IF NOT EXISTS team_user
(
    team_id UUID NOT NULL,
    user_id UUID NOT NULL,
    PRIMARY KEY (team_id, user_id),
    FOREIGN KEY (team_id) REFERENCES teams (id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_team_team_id ON team_user (team_id);
CREATE INDEX IF NOT EXISTS idx_user_team_user_id ON team_user (user_id);



CREATE TABLE IF NOT EXISTS pull_requests
(
    id         UUID PRIMARY KEY         DEFAULT gen_random_uuid(),
    title      VARCHAR(64) NOT NULL,
    author_id  UUID        NOT NULL,
    status_id  UUID        NOT NULL,
    merged_at  TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (author_id) REFERENCES users (id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (status_id) REFERENCES pull_request_statuses (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pull_requests_author_id ON pull_requests (author_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status_id ON pull_requests (status_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_created_at ON pull_requests (created_at DESC);



CREATE TABLE IF NOT EXISTS pull_request_reviewers
(
    pull_request_id UUID NOT NULL            DEFAULT gen_random_uuid(),
    reviewer_id     UUID NOT NULL,
    assigned_at     TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, reviewer_id),
    FOREIGN KEY (reviewer_id) REFERENCES users (id)
        ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (pull_request_id) REFERENCES pull_requests (id)
        ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_pull_request_reviewers_user_id ON pull_request_reviewers (reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pull_request_reviewers_pr_id ON pull_request_reviewers (pull_request_id);



CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE TRIGGER update_users_updated_at
    BEFORE UPDATE
    ON users
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_teams_updated_at
    BEFORE UPDATE
    ON teams
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

CREATE OR REPLACE TRIGGER update_pull_requests_updated_at
    BEFORE UPDATE
    ON pull_requests
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();