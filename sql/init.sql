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
    sport_name TEXT UNIQUE NOT NULL
);
INSERT INTO sports (sport_name) 
VALUES 
    ('Tennis'),
    ('Badminton'),
    ('Table Tennis'),
    ('Pickleball'),
    ('Racquetball'),
    ('Squash')
ON CONFLICT (sport_name) DO NOTHING;

CREATE TABLE IF NOT EXISTS posts (
    id BIGSERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    sport_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (sport_id) REFERENCES sports(id) ON DELETE CASCADE
);
