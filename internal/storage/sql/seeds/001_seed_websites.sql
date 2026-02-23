INSERT INTO websites (name, domain, status, created_at, updated_at)
VALUES
    ('hosseinibrothers', 'hosseinibrothers.ir', 'active', NOW(), NOW()),
    ('badomjip', 'badomjip.com', 'active', NOW(), NOW())
ON CONFLICT (domain) DO NOTHING;