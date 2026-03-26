CREATE TABLE IF NOT EXISTS shortcuts (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    slug        TEXT NOT NULL,
    target_url  TEXT NOT NULL,
    type        TEXT NOT NULL CHECK(type IN ('subdomain', 'path')),
    click_count INTEGER NOT NULL DEFAULT 0,
    created_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at  TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_shortcuts_type_slug ON shortcuts(type, slug);

CREATE TABLE IF NOT EXISTS homepage_config (
    id           INTEGER PRIMARY KEY CHECK (id = 1),
    mode         TEXT NOT NULL DEFAULT 'carousel' CHECK(mode IN ('redirect', 'carousel')),
    redirect_url TEXT NOT NULL DEFAULT '',
    updated_at   TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
INSERT OR IGNORE INTO homepage_config (id, mode, redirect_url) VALUES (1, 'carousel', '');

CREATE TABLE IF NOT EXISTS carousel_images (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    filename   TEXT NOT NULL,
    original   TEXT NOT NULL,
    mime_type  TEXT NOT NULL,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);
