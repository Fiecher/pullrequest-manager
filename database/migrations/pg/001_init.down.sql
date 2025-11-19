DROP TRIGGER IF EXISTS update_pull_requests_updated_at ON pull_requests;
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

DROP FUNCTION IF EXISTS update_updated_at_column;

DROP INDEX IF EXISTS idx_pull_request_reviewers_pr_id;
DROP INDEX IF EXISTS idx_pull_request_reviewers_user_id;

DROP TABLE IF EXISTS pull_request_reviewers;

DROP INDEX IF EXISTS idx_pull_requests_created_at;
DROP INDEX IF EXISTS idx_pull_requests_status_id;
DROP INDEX IF EXISTS idx_pull_requests_author_id;

DROP TABLE IF EXISTS pull_requests;

DROP INDEX IF EXISTS idx_user_team_user_id;
DROP INDEX IF EXISTS idx_user_team_team_id;

DROP TABLE IF EXISTS team_user;
DROP TABLE IF EXISTS teams;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS pull_request_statuses;

DROP EXTENSION IF EXISTS pgcrypto;
