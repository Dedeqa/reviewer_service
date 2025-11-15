CREATE TABLE IF NOT EXISTS teams
(
    name
    TEXT
    PRIMARY
    KEY
);
CREATE TABLE IF NOT EXISTS users
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
    );
CREATE TABLE IF NOT EXISTS team_users
(
    team_name
    TEXT
    REFERENCES
    teams
(
    name
) ON DELETE CASCADE,
    user_id UUID REFERENCES users
(
    id
)
  ON DELETE CASCADE,
    PRIMARY KEY
(
    team_name,
    user_id
)
    );
CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');
CREATE TABLE IF NOT EXISTS prs
(
    id
    UUID
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
),
    title TEXT NOT NULL,
    author_id UUID REFERENCES users
(
    id
) ON DELETE SET NULL,
    team_name TEXT REFERENCES teams
(
    name
)
  ON DELETE SET NULL,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP
  WITH TIME ZONE DEFAULT now()
    );
CREATE TABLE IF NOT EXISTS pr_reviewers
(
    pr_id
    UUID
    REFERENCES
    prs
(
    id
) ON DELETE CASCADE,
    reviewer_id UUID REFERENCES users
(
    id
)
  ON DELETE CASCADE,
    assigned_at TIMESTAMP
  WITH TIME ZONE DEFAULT now(),
    PRIMARY KEY
(
    pr_id,
    reviewer_id
)
    );
-- Extensions for convenience (pg >= 13)
CREATE
EXTENSION IF NOT EXISTS pgcrypto;
-- Indexes to speed up queries (moderate dataset)
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_prs_status ON prs(status);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON
    pr_reviewers(reviewer_id)
