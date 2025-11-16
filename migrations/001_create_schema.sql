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
    TEXT
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
)::text,
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
    user_id TEXT REFERENCES users
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
DO
$$
BEGIN
    IF
NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pr_status') THEN
CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');
END IF;
END
$$;
CREATE TABLE IF NOT EXISTS prs
(
    id
    TEXT
    PRIMARY
    KEY
    DEFAULT
    gen_random_uuid
(
)::text,
    title TEXT NOT NULL,
    author_id TEXT REFERENCES users
(
    id
) ON DELETE SET NULL,
    team_name TEXT REFERENCES teams
(
    name
)
  ON DELETE SET NULL,
    status pr_status NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMPTZ DEFAULT now
(
),
    merged_at TIMESTAMPTZ
    );
CREATE TABLE IF NOT EXISTS pr_reviewers
(
    pr_id
    TEXT
    REFERENCES
    prs
(
    id
) ON DELETE CASCADE,
    reviewer_id TEXT REFERENCES users
(
    id
)
  ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ DEFAULT now
(
),
    PRIMARY KEY
(
    pr_id,
    reviewer_id
)
    );
CREATE
EXTENSION IF NOT EXISTS pgcrypto;
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_prs_status ON prs(status);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON
    pr_reviewers(reviewer_id);