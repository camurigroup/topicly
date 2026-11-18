-- Etapa 3: núcleo do fórum (categorias, tópicos, respostas, likes, follows).

CREATE TABLE IF NOT EXISTS categories (
    id          CHAR(36) PRIMARY KEY,
    parent_id   CHAR(36) REFERENCES categories(id) ON DELETE CASCADE,
    name        VARCHAR(120) NOT NULL,
    slug        VARCHAR(140) NOT NULL UNIQUE,
    description VARCHAR(500),
    icon        VARCHAR(60),
    cover_url   VARCHAR(500),
    position    INTEGER NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_categories_parent ON categories (parent_id);

CREATE TABLE IF NOT EXISTS topics (
    id               CHAR(36) PRIMARY KEY,
    category_id      CHAR(36) NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    author_id        CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title            VARCHAR(200) NOT NULL,
    slug             VARCHAR(240) NOT NULL,
    content          TEXT NOT NULL,
    is_pinned        BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked        BOOLEAN NOT NULL DEFAULT FALSE,
    is_featured      BOOLEAN NOT NULL DEFAULT FALSE,
    status           VARCHAR(20) NOT NULL DEFAULT 'published',
    view_count       INTEGER NOT NULL DEFAULT 0,
    reply_count      INTEGER NOT NULL DEFAULT 0,
    like_count       INTEGER NOT NULL DEFAULT 0,
    follower_count   INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL,
    updated_at       TIMESTAMPTZ NOT NULL,
    last_activity_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_topics_category ON topics (category_id);
CREATE INDEX IF NOT EXISTS idx_topics_author ON topics (author_id);
CREATE INDEX IF NOT EXISTS idx_topics_last_activity ON topics (last_activity_at);

CREATE TABLE IF NOT EXISTS replies (
    id         CHAR(36) PRIMARY KEY,
    topic_id   CHAR(36) NOT NULL REFERENCES topics(id) ON DELETE CASCADE,
    author_id  CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    parent_id  CHAR(36) REFERENCES replies(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    like_count INTEGER NOT NULL DEFAULT 0,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    edited_at  TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_replies_topic ON replies (topic_id);
CREATE INDEX IF NOT EXISTS idx_replies_author ON replies (author_id);

CREATE TABLE IF NOT EXISTS likes (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(10) NOT NULL,
    target_id   CHAR(36) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_like UNIQUE (user_id, target_type, target_id)
);
CREATE INDEX IF NOT EXISTS idx_likes_target ON likes (target_type, target_id);

CREATE TABLE IF NOT EXISTS follows (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type VARCHAR(12) NOT NULL,
    target_id   CHAR(36) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL,
    CONSTRAINT uq_follow UNIQUE (user_id, target_type, target_id)
);
CREATE INDEX IF NOT EXISTS idx_follows_target ON follows (target_type, target_id);

CREATE TABLE IF NOT EXISTS content_edits (
    id               CHAR(36) PRIMARY KEY,
    entity_type      VARCHAR(10) NOT NULL,
    entity_id        CHAR(36) NOT NULL,
    editor_id        CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    previous_content TEXT NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_content_edits_entity ON content_edits (entity_type, entity_id);
