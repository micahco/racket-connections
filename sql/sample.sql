INSERT INTO user_ (name_, email_, password_hash_)
VALUES
    ('John Doe', 'johnd@oregonstate.edu', 'pswd'),
    ('Sarah Jane', 'sarahd@oregonstate.edu', 'pswd'),
    ('Chad Thunder', 'chadt@oregonstate.edu', 'pswd'),
    ('Johnny Appleseed', 'johna@oregonstate.edu', 'pswd'),
    ('Serena Williams', 'belala@oregonstate.edu', 'pswd'),
    ('Free Thinker', 'frreee@oregonstate.edu', 'pswd'),
    ('Hal 3000', 'hal@oregonstate.edu', 'pswd'),
    ('Dewey Decimal', 'deweu@oregonstate.edu', 'pswd'),
    ('Chad Ochocinco', 'chado@oregonstate.edu', 'pswd'),
    ('Mowgli', 'mow@oregonstate.edu', 'pswd'),
    ('Lip', 'lip@oregonstate.edu', 'pswd'),
    ('Master Mumbai', 'mumbai@oregonstate.edu', 'pswd'),
    ('Bruce Lee', 'bl@oregonstate.edu', 'pswd'),
    ('Johhny Sack', 'ginny@oregonstate.edu', 'pswd'),
    ('Tony Soprano', 'boss@oregonstate.edu', 'pswd'),
    ('Chrissy', 'nose@oregonstate.edu', 'pswd');

INSERT INTO contact_ (user_id_, contact_method_id_, value_)
VALUES
    (1, 1, 'johnny@gmail.com'),
    (2, 1, 'sardubs@gmail.com'),
    (3, 2, '123-456-7890'),
    (4, 1, 'johnny@gmail.com'),
    (5, 1, 'sardubs@gmail.com'),
    (6, 2, '123-456-7890'),
    (7, 1, 'johnny@gmail.com'),
    (8, 1, 'sardubs@gmail.com'),
    (9, 2, '123-456-7890'),
    (10, 1, 'johnny@gmail.com'),
    (11, 1, 'sardubs@gmail.com'),
    (12, 2, '123-456-7890'),
    (13, 1, 'johnny@gmail.com'),
    (14, 1, 'sardubs@gmail.com'),
    (15, 2, '123-456-7890');

INSERT INTO timeslot_ (user_id_, day_id_, time_id_)
VALUES
    (1, 1, 1),
    (2, 2, 2),
    (3, 3, 3),
    (2, 1, 2),
    (4, 4, 3),
    (5, 1, 1),
    (6, 2, 2),
    (7, 3, 3),
    (8, 1, 2),
    (9, 4, 3),
    (10, 1, 1),
    (11, 2, 2),
    (12, 3, 3),
    (13, 1, 2),
    (14, 4, 3),
    (15, 5, 3);

INSERT INTO post_ (user_id_, sport_id_, skill_level_id_)
VALUES
    (1, 1, 2),
    (2, 1, 3),
    (3, 1, 4),
    (3, 2, 2),
    (2, 2, 5),
    (1, 2, 2),
    (4, 1, 2),
    (5, 1, 3),
    (6, 1, 4),
    (7, 2, 2),
    (8, 2, 5),
    (9, 2, 2),
    (10, 1, 2),
    (11, 1, 3),
    (12, 1, 4),
    (13, 2, 2),
    (14, 2, 5),
    (15, 2, 2),
    (15, 3, 4);
