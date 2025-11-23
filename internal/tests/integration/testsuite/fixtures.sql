TRUNCATE pull_request_reviewers, pull_requests, users, teams RESTART IDENTITY CASCADE;

INSERT INTO teams (team_name) VALUES
    ('team-alpha'),
    ('team-beta');

INSERT INTO users (user_id, username, team_name, is_active) VALUES
    ('alpha-author', 'Alice Author', 'team-alpha', TRUE),
    ('alpha-reviewer-1', 'Rita Reviewer', 'team-alpha', TRUE),
    ('alpha-reviewer-2', 'Ryan Reviewer', 'team-alpha', TRUE),
    ('alpha-reviewer-3', 'Ralph Reserve', 'team-alpha', TRUE),
    ('alpha-inactive', 'Ivan Inactive', 'team-alpha', FALSE),
    ('beta-member-1', 'Bob Beta', 'team-beta', TRUE),
    ('beta-member-2', 'Bella Beta', 'team-beta', TRUE);
