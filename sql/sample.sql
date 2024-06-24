INSERT INTO users (name, email, password_hash)
VALUES
    ('John Doe', 'johnd@oregonstate.edu', 'pswd'),
    ('Sarah Jane', 'sarahd@oregonstate.edu', 'pswd'),
    ('Chad Thunder', 'chadt@oregonstate.edu', 'pswd'),
    ('Zendaya', 'zend@oregonstate.edu', 'pswd')
ON CONFLICT (email) DO NOTHING;

INSERT INTO posts (user_id, sport_id, skill_level)
VALUES
    (1, 1, 2),
    (2, 1, 3),
    (3, 1, 4),
    (3, 2, 2),
    (2, 2, 5),
    (1, 2, 2);
