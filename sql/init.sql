/*
 * SESSIONS
 */
CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    data BYTEA NOT NULL,
    expiry TIMESTAMPTZ NOT NULL
);
-- The scs package will automatically delete expired sessions
CREATE INDEX sessions_expiry_idx ON sessions (expiry);

/*
 * TABLES
 */
CREATE TABLE IF NOT EXISTS skill_level_ (
    id_ SERIAL PRIMARY KEY,
    name_ TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS sport_ (
    id_ SERIAL PRIMARY KEY,
    name_ TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS day_of_week_ (
    id_ SERIAL PRIMARY KEY,
    name_ TEXT UNIQUE,
    abbrev_ VARCHAR(3) UNIQUE
);

CREATE TABLE IF NOT EXISTS time_of_day_ (
    id_ SERIAL PRIMARY KEY,
    name_ TEXT UNIQUE,
    abbrev_ VARCHAR(3) UNIQUE
);

CREATE TABLE IF NOT EXISTS contact_method_ (
    id_ SERIAL PRIMARY KEY,
    name_ TEXT UNIQUE
);

CREATE TABLE IF NOT EXISTS user_ (
    id_ BIGSERIAL PRIMARY KEY,
    name_ TEXT NOT NULL,
    email_ CITEXT UNIQUE NOT NULL,
    password_hash_ CHAR(60) NOT NULL,
    created_at_ TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS timeslot_ (
    user_id_ INT NOT NULL,
    day_id_ INT NOT NULL,
    time_id_ INT NOT NULL,
    PRIMARY KEY (user_id_, day_id_, time_id_),
    FOREIGN KEY (user_id_) REFERENCES user_(id_),
    FOREIGN KEY (day_id_) REFERENCES day_of_week_(id_),
    FOREIGN KEY (time_id_) REFERENCES time_of_day_(id_)
);

CREATE TABLE IF NOT EXISTS contact_ (
    id_ BIGSERIAL PRIMARY KEY,
    value_ TEXT NOT NULL,
    user_id_ INT NOT NULL,
    contact_method_id_ INT NOT NULL,
    FOREIGN KEY (user_id_) REFERENCES user_(id_) ON DELETE CASCADE,
    FOREIGN KEY (contact_method_id_) REFERENCES contact_method_(id_) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS post_ (
    id_ BIGSERIAL PRIMARY KEY,
    comment_ TEXT NOT NULL DEFAULT '',
    created_at_ TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id_ INT NOT NULL,
    sport_id_ INT NOT NULL,
    skill_level_id_ INT NOT NULL,
    FOREIGN KEY (user_id_) REFERENCES user_(id_) ON DELETE CASCADE,
    FOREIGN KEY (sport_id_) REFERENCES sport_(id_) ON DELETE CASCADE,
    FOREIGN KEY (skill_level_id_) REFERENCES skill_level_(id_) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS verification_ (
    token_ VARCHAR(100) NOT NULL PRIMARY KEY,
    email_ CITEXT NOT NULL,
    expiry_ TIMESTAMPTZ NOT NULL,
    created_at_ TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/*
 * DATA
 */
INSERT INTO day_of_week_ (name_, abbrev_) 
VALUES 
    ('monday', 'mon'),
    ('tuesday', 'tue'),
    ('wednesday', 'wed'),
    ('thursday', 'thu'),
    ('friday', 'fri'),
    ('saturday', 'sat'),
    ('sunday', 'sun')
ON CONFLICT (name_) DO NOTHING;

INSERT INTO time_of_day_ (name_, abbrev_) 
VALUES 
    ('morning', 'mor'),
    ('afternoon', 'aft'),
    ('evening', 'eve')
ON CONFLICT (name_) DO NOTHING;

INSERT INTO contact_method_ (name_) 
VALUES 
    ('email'),
    ('phone'),
    ('other')
ON CONFLICT (name_) DO NOTHING;

INSERT INTO sport_ (name_) 
VALUES 
    ('tennis'),
    ('badminton'),
    ('table tennis'),
    ('pickleball'),
    ('racquetball'),
    ('squash')
ON CONFLICT (name_) DO NOTHING;

INSERT INTO skill_level_ (name_) 
VALUES 
    ('beginner'),
    ('novice'),
    ('intermediate'),
    ('advanced'),
    ('expert')
ON CONFLICT (name_) DO NOTHING;
