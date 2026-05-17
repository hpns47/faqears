INSERT INTO artists (id, name, country, biography, created_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'User Uploads',
    '',
    'System artist for user-generated tracks',
    now()
)
ON CONFLICT (id) DO NOTHING;

INSERT INTO albums (id, artist_id, title, year, cover_url, created_at)
VALUES (
    '22222222-2222-2222-2222-222222222222',
    '11111111-1111-1111-1111-111111111111',
    'User Uploads',
    0,
    '',
    now()
)
ON CONFLICT (id) DO NOTHING;
