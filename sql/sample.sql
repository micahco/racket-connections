INSERT INTO user_ (name_, email_, password_hash_)
VALUES
    ('Briggs McBeath', 'bmcbeath0@forbes.com', '*******'),
    ('Travus Tomczykowski', 'ttomczykowski1@vimeo.com', '*******'),
    ('Iormina McGlew', 'imcglew2@webnode.com', '*******'),
    ('Murielle Audry', 'maudry3@upenn.edu', '*******'),
    ('Cassi Tulloch', 'ctulloch4@google.com.au', '*******'),
    ('Rickie Costy', 'rcosty5@hud.gov', '*******'),
    ('Mervin Satteford', 'msatteford6@purevolume.com', '*******'),
    ('Moises Mauser', 'mmauser7@addthis.com', '*******'),
    ('Louie Doddrell', 'ldoddrell8@amazon.co.jp', '*******'),
    ('Ardath Maciaszek', 'amaciaszek9@thetimes.co.uk', '*******'),
    ('Carlynn Peschka', 'cpeschkaa@washingtonpost.com', '*******'),
    ('Lark Burrel', 'lburrelb@samsung.com', '*******'),
    ('Anne-marie Waggatt', 'awaggattc@stumbleupon.com', '*******'),
    ('Obidiah Ramey', 'orameyd@webnode.com', '*******'),
    ('Graig Doeg', 'gdoege@alibaba.com', '*******'),
    ('Farlie Freear', 'ffreearf@engadget.com', '*******'),
    ('Wittie Gilliland', 'wgillilandg@pcworld.com', '*******'),
    ('Cleve Grollmann', 'cgrollmannh@xrea.com', '*******'),
    ('Bobina Markey', 'boi@oakley.com', '*******'),
    ('Dominik Morena', 'dmorenaj@dyndns.org', '*******'),
    ('Jake Skeldinge', 'jskeldingek@jimdo.com', '*******'),
    ('Conrade Christin', 'cchristinl@vk.com', '*******'),
    ('Tudor Amps', 'tampsm@techcrunch.com', '*******'),
    ('Thorsten Holtham', 'tholthamn@google.ca', '*******'),
    ('Cassi Alred', 'calredo@google.it', '*******'),
    ('Rich De Vaar', 'rdep@zimbio.com', '*******'),
    ('Waverley Damato', 'wdamatoq@soup.io', '*******'),
    ('Luelle Sends', 'lsendsr@shareasale.com', '*******'),
    ('Pauly Fassmann', 'pfassmanns@simplemachines.org', '*******'),
    ('Ursula Jobb', 'ujobbt@dropbox.com', '*******'),
    ('Vernen Totaro', 'vtotarou@altervista.org', '*******'),
    ('Mara Rea', 'mreav@diigo.com', '*******'),
    ('Matias Kosel', 'mkoselw@github.io', '*******'),
    ('Marsha Oneal', 'monealx@europa.eu', '*******'),
    ('Rocky Corryer', 'rcorryery@w3.org', '*******'),
    ('Sondra Edworthy', 'sedworthyz@auda.org.au', '*******'),
    ('Edith Izaks', 'eizaks10@seesaa.net', '*******'),
    ('Octavia Kemmey', 'okemmey11@squarespace.com', '*******'),
    ('Noll Olliff', 'nolliff12@themeforest.net', '*******'),
    ('Hakeem Nissle', 'hnissle13@arstechnica.com', '*******')
ON CONFLICT (email_) DO NOTHING;

INSERT INTO contact_ (user_id_, contact_method_id_, value_)
VALUES
    (1,	1, 'wcastenda0@economist.com'),
    (2,	1, 'gstairmand1@constantcontact.com'),
    (3,	1, 'htelfer2@fotki.com'),
    (4,	1, 'pblagdon3@hc360.com'),
    (5,	1, 'javeray4@apache.org'),
    (6,	1, 'tdelagua5@usda.gov'),
    (7,	1, 'rpedersen6@yandex.ru'),
    (8,	1, 'pdarbishire7@princeton.edu'),
    (9,	1, 'jmalling8@seesaa.net'),
    (10, 1,	'wrosenstiel9@webnode.com'),
    (11, 1,	'lturna@ameblo.jp'),
    (12, 1,	'rtrownsonb@miitbeian.gov.cn'),
    (13, 1,	'gargilec@goo.gl'),
    (14, 1,	'lsevittd@usnews.com'),
    (15, 1,	'ebricknere@themeforest.net'),
    (16, 1,	'edissmanf@diigo.com'),
    (17, 1,	'dpragnellg@columbia.edu'),
    (18, 1,	'srawleh@digg.com'),
    (19, 1,	'erowbottomi@elegantthemes.com'),
    (20, 1,	'icapstakej@cbsnews.com'),
    (21, 1,	'correllk@virginia.edu'),
    (22, 1,	'nliveleyl@tinyurl.com'),
    (23, 1,	'khargatem@wp.com'),
    (24, 1,	'lthorleyn@nymag.com'),
    (25, 1,	'flepereo@example.com'),
    (26, 1,	'mbehningp@hibu.com'),
    (27, 1,	'wbarrowsq@engadget.com'),
    (28, 1,	'mroylancer@china.com.cn'),
    (29, 1,	'uruspines@spotify.com'),
    (30, 1,	'ntoffaninit@goo.ne.jp'),
    (31, 1,	'jnorvalu@cbc.ca'),
    (32, 1,	'laskinv@elpais.com'),
    (33, 1,	'bnapletonw@globo.com'),
    (34, 1,	'jbeltznerx@baidu.com'),
    (35, 1,	'hdaneluty@friendfeed.com'),
    (36, 1,	'ccridgez@is.gd'),
    (37, 1,	'lpristnor10@sourceforge.net'),
    (38, 1,	'epostance11@seattletimes.com'),
    (39, 1,	'shousen12@tinypic.com'),
    (40, 1,	'gstife13@odnoklassniki.ru');

INSERT INTO timeslot_ (user_id_, day_id_, time_id_)
VALUES
    (31,6,3),
    (9,5,2),
    (21,6,1),
    (33,4,3),
    (26,5,2),
    (6,1,1),
    (21,2,3),
    (32,5,2),
    (12,6,2),
    (17,5,2),
    (14,2,1),
    (2,1,1),
    (35,7,3),
    (3,3,1),
    (18,4,2),
    (38,6,2),
    (32,3,1),
    (21,4,1),
    (32,3,3),
    (27,6,2),
    (29,4,3),
    (26,3,1),
    (28,7,3),
    (21,3,3),
    (6,5,1),
    (24,2,1),
    (8,2,2),
    (3,2,3),
    (32,1,2),
    (35,3,3),
    (38,7,2),
    (35,5,3),
    (10,7,3),
    (28,2,1),
    (24,6,1),
    (22,2,2),
    (22,5,1),
    (35,4,2),
    (2,7,2),
    (31,5,3),
    (37,1,2),
    (30,1,1),
    (14,3,1),
    (28,6,3),
    (5,2,3),
    (8,1,2),
    (27,2,1),
    (35,1,1),
    (7,6,2),
    (37,3,1),
    (18,5,1),
    (31,7,2),
    (15,2,1),
    (19,2,1),
    (39,6,1),
    (29,1,2),
    (34,4,1),
    (15,6,2),
    (35,4,3),
    (39,2,1),
    (6,7,2),
    (5,5,1),
    (29,5,1),
    (22,7,2),
    (38,1,1),
    (2,2,2),
    (11,2,2),
    (5,1,1),
    (10,3,2),
    (27,6,3),
    (22,6,2),
    (37,6,3),
    (32,5,3),
    (30,6,2),
    (6,6,1),
    (34,6,1),
    (25,7,3),
    (29,7,1),
    (29,3,1),
    (19,6,2),
    (38,4,1),
    (13,7,3),
    (14,2,2),
    (15,3,3),
    (36,3,2),
    (35,2,2),
    (10,2,1),
    (17,6,2),
    (14,2,2),
    (25,1,1),
    (27,5,1),
    (3,4,2),
    (3,4,2),
    (19,1,1),
    (28,4,2),
    (19,6,2),
    (5,4,2),
    (20,4,3),
    (12,6,3)
ON CONFLICT (user_id_, day_id_, time_id_) DO NOTHING;

INSERT INTO post_ (user_id_, sport_id_, skill_level_id_)
VALUES
    (1,6,5),
    (2,3,1),
    (3,4,2),
    (4,2,2),
    (5,2,1),
    (6,2,4),
    (7,1,5),
    (8,3,1),
    (9,6,4),
    (10,3,4),
    (11,3,3),
    (12,2,5),
    (13,6,2),
    (14,1,3),
    (15,1,4),
    (16,2,1),
    (17,6,3),
    (18,4,4),
    (19,4,5),
    (20,1,3),
    (21,5,4),
    (22,4,4),
    (23,2,4),
    (24,4,4),
    (25,5,3),
    (26,6,4),
    (27,1,5),
    (28,2,4),
    (29,2,2),
    (30,6,1),
    (31,1,5),
    (32,6,4),
    (33,4,5),
    (34,6,3),
    (35,5,5),
    (36,6,2),
    (37,5,1),
    (38,3,1),
    (39,1,4),
    (40,4,3);
