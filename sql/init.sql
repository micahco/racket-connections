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

CREATE TABLE IF NOT EXISTS skill_levels (
    value INT PRIMARY KEY,
    description TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS sports (
    id BIGSERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    comment TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    skill_level INT NOT NULL,
    user_id INT NOT NULL,
    sport_id INT NOT NULL,
    FOREIGN KEY (skill_level) REFERENCES skill_levels(value) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE CASCADE
);

INSERT INTO sports (name) 
VALUES 
    ('tennis'),
    ('badminton'),
    ('table tennis'),
    ('pickleball'),
    ('racquetball'),
    ('squash')
ON CONFLICT (name) DO NOTHING;

INSERT INTO skill_levels (value, description) 
VALUES 
    (1, 'beginner'),
    (2, 'novice'),
    (3, 'intermediate'),
    (4, 'advanced'),
    (5, 'expert')
ON CONFLICT (value) DO NOTHING;

/*
 * VIEWS
 */
CREATE OR REPLACE VIEW numbered_posts_by_sport AS
SELECT
        *,
        ROW_NUMBER() OVER (
            PARTITION BY sport_id
            ORDER BY id
        )
    FROM
        posts;

CREATE OR REPLACE VIEW post_details AS
SELECT
    p.id,
    comment,
    p.created_at,
    l.value AS skill_level,
    l.description AS skill_level_description,
    user_id,
    u.name AS user_name,
    sport_id,
    s.name AS sport_name
FROM
    posts p
    INNER JOIN skill_levels l ON p.skill_level = l.value
    INNER JOIN users u ON p.user_id = u.id
    INNER JOIN sports s ON p.sport_id = s.id;

CREATE OR REPLACE VIEW latest_posts AS
SELECT
    p.id,
    comment,
    p.created_at,
    l.value AS skill_level,
    l.description AS skill_level_description,
    user_id,
    u.name AS user_name,
    sport_id,
    s.name AS sport_name
FROM
    numbered_posts_by_sport p
    INNER JOIN skill_levels l ON p.skill_level = l.value
    INNER JOIN users u ON p.user_id = u.id
    INNER JOIN sports s ON p.sport_id = s.id
WHERE
    ROW_NUMBER <= 3;
