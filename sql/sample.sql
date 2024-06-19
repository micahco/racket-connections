INSERT INTO users (name, email, password_hash)
VALUES
    ('John Doe', 'johnd@oregonstate.edu', 'hsfdasdhfjkshdfjlkhsa'),
    ('Sarah Jane', 'sarahd@oregonstate.edu', 'hsfdasdhfjkshdfjlkhsa'),
    ('Chad Thunder', 'chadt@oregonstate.edu', 'hsfdasdhfjkshdfjlkhsa'),
    ('Zendaya', 'zend@oregonstate.edu', 'hsfdasdhfjkshdfjlkhsa')
ON CONFLICT (email) DO NOTHING;

INSERT INTO sports (name) 
VALUES 
    ('Tennis'),
    ('Badminton'),
    ('Table Tennis'),
    ('Pickleball'),
    ('Racquetball'),
    ('Squash')
ON CONFLICT (name) DO NOTHING;

INSERT INTO posts (user_id, sport_id, skill_level)
VALUES
    (1, 1, 2),
    (2, 1, 3),
    (3, 1, 4),
    (3, 2, 2),
    (2, 2, 5),
    (1, 2, 2);
