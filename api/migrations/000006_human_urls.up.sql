ALTER TABLE users ADD COLUMN username TEXT;
ALTER TABLE wishlists ADD COLUMN slug TEXT;

UPDATE users
SET username = lower(regexp_replace(trim(display_name), '[^a-zA-Z0-9]+', '', 'g'))
WHERE username IS NULL OR username = '';

UPDATE users
SET username = 'user' || substr(replace(id::text, '-', ''), 1, 8)
WHERE username IS NULL OR username = '';

WITH ranked AS (
    SELECT id,
           username,
           ROW_NUMBER() OVER (PARTITION BY username ORDER BY created_at, id) AS rn
    FROM users
)
UPDATE users u
SET username = ranked.username || ranked.rn::text
FROM ranked
WHERE u.id = ranked.id
  AND ranked.rn > 1;

ALTER TABLE users ALTER COLUMN username SET NOT NULL;
CREATE UNIQUE INDEX idx_users_username ON users (lower(username));

UPDATE wishlists
SET slug = lower(
    regexp_replace(
        regexp_replace(trim(title), '\s+', '-', 'g'),
        '[^a-zA-Z0-9-]+',
        '',
        'g'
    )
)
WHERE slug IS NULL OR slug = '';

UPDATE wishlists
SET slug = trim(both '-' from slug)
WHERE slug IS NOT NULL;

UPDATE wishlists
SET slug = 'wishlist'
WHERE slug IS NULL OR slug = '';

WITH ranked AS (
    SELECT id,
           slug,
           ROW_NUMBER() OVER (PARTITION BY slug ORDER BY created_at, id) AS rn
    FROM wishlists
)
UPDATE wishlists w
SET slug = ranked.slug || '-' || ranked.rn::text
FROM ranked
WHERE w.id = ranked.id
  AND ranked.rn > 1;

ALTER TABLE wishlists ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX idx_wishlists_slug ON wishlists (slug);
