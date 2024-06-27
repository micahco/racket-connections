INSERT INTO user_ (name_, email_, password_hash_)
VALUES
    ('John Doe', 'johnd@oregonstate.edu', 'pswd'),
    ('Sarah Jane', 'sarahd@oregonstate.edu', 'pswd'),
    ('Chad Thunder', 'chadt@oregonstate.edu', 'pswd'),
    ('Zendaya', 'zend@oregonstate.edu', 'pswd');

INSERT INTO contact_ (user_id_, contact_method_id_, value_)
VALUES
    (1, 1, 'johnny@gmail.com'),
    (2, 1, 'sardubs@gmail.com'),
    (3, 2, '123-456-7890'),
    (4, 3, 'Discord user#1234');

INSERT INTO timeslot_ (user_id_, day_id_, time_id_)
VALUES
    (1, 1, 1),
    (2, 2, 2),
    (3, 3, 3),
    (2, 1, 2),
    (4, 4, 3),
    (4, 5, 3);

INSERT INTO post_ (user_id_, sport_id_, skill_level_id_)
VALUES
    (1, 1, 2),
    (2, 1, 3),
    (3, 1, 4),
    (3, 2, 2),
    (2, 2, 5),
    (1, 2, 2),
    (4, 3, 4);
