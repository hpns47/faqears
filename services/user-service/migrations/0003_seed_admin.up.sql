INSERT INTO users (id, email, display_name, avatar_url, country, language, tier, created_at, updated_at)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'admin@faqears.local',
    'Admin',
    '',
    '',
    'en',
    'free',
    now(),
    now()
)
ON CONFLICT (id) DO NOTHING;
