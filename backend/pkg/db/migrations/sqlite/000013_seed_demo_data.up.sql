-- Seed migration (bonus): fills the database with demo users, a follow
-- relationship, a public post and a group so the app is testable immediately.
-- All demo accounts share the password: password123

INSERT OR IGNORE INTO users (id, email, password_hash, first_name, last_name, date_of_birth, nickname, about_me, is_public) VALUES
('11111111-1111-1111-1111-111111111111', 'bob@demo.local',   '$2a$10$xevPn32rv40oOaV5UeltoO.BGduMxPGj1rQXgejoQ/8s4spIh4R9K', 'Bob',   'Johnson', '1992-05-12', 'bobby', 'Coffee lover and weekend hiker.', 1),
('22222222-2222-2222-2222-222222222222', 'carol@demo.local', '$2a$10$xevPn32rv40oOaV5UeltoO.BGduMxPGj1rQXgejoQ/8s4spIh4R9K', 'Carol', 'Williams', '1995-09-30', NULL, 'Private account.', 0),
('33333333-3333-3333-3333-333333333333', 'dave@demo.local',  '$2a$10$xevPn32rv40oOaV5UeltoO.BGduMxPGj1rQXgejoQ/8s4spIh4R9K', 'Dave',  'Brown', '1988-02-18', 'd_brown', 'Photographer.', 1);

-- Bob follows Dave (both public -> auto accepted).
INSERT OR IGNORE INTO followers (id, follower_id, following_id, status) VALUES
('66666666-6666-6666-6666-666666666666', '11111111-1111-1111-1111-111111111111', '33333333-3333-3333-3333-333333333333', 'accepted');

-- A public post by Bob.
INSERT OR IGNORE INTO posts (id, user_id, content, privacy) VALUES
('55555555-5555-5555-5555-555555555555', '11111111-1111-1111-1111-111111111111', 'Welcome to the demo social network! This is a public post.', 'public');

-- A group created by Bob, with Bob as the first member.
INSERT OR IGNORE INTO groups (id, creator_id, title, description) VALUES
('44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'Hiking Club', 'Plan weekend hikes and share trail photos.');

INSERT OR IGNORE INTO group_members (id, group_id, user_id, status) VALUES
('77777777-7777-7777-7777-777777777777', '44444444-4444-4444-4444-444444444444', '11111111-1111-1111-1111-111111111111', 'member');
