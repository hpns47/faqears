INSERT INTO auth_users (id, email, password_hash, roles, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@faqears.local',
    '$2a$10$NtHz4phZ5KskH5DXIT3MSOCdDpCyCV4MLN2D/iL3S7GCUvUThGd3.',
    ARRAY['user','admin'],
    NOW(),
    NOW()
)
ON CONFLICT (email) DO UPDATE
SET roles = ARRAY['user','admin'],
    updated_at = NOW();
