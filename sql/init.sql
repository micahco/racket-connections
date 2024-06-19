/*
 * TABLES
 */
 
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);
-- The scs package will automatically delete expired sessions
CREATE INDEX sessions_expiry_idx ON sessions (expiry);

CREATE TABLE IF NOT EXISTS verifications (
    token VARCHAR(100) NOT NULL PRIMARY KEY,
    email CITEXT NOT NULL,
    expiry TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    email CITEXT UNIQUE NOT NULL,
    password_hash CHAR(60) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sports (
    id BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    sport_id INT NOT NULL,
    skill_level INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE CASCADE
);

/*
 * VIEWS
 */

-- Returns the last three posts for each sport
CREATE OR REPLACE VIEW latest_posts AS
WITH LatestPosts AS (
    SELECT
        p.id,
        p.user_id,
        p.sport_id,
        p.skill_level,
        p.created_at,
        ROW_NUMBER() OVER (PARTITION BY p.sport_id ORDER BY p.id DESC) AS row_num
    FROM
        posts p
)
SELECT
    p.id AS post_id,
    p.skill_level AS skill_level,
    p.created_at AS created_at,
    u.id AS user_id,
    u.name AS user_name,
    s.name AS sport_name
FROM
    LatestPosts p
    JOIN users u ON p.user_id = u.id
    JOIN sports s ON p.sport_id = s.id
WHERE
    p.row_num <= 3;
