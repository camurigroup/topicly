-- Etapa 3: núcleo do fórum (categorias, tópicos, respostas, likes, follows).

CREATE TABLE IF NOT EXISTS categories (
    id          CHAR(36) PRIMARY KEY,
    parent_id   CHAR(36) NULL,
    name        VARCHAR(120) NOT NULL,
    slug        VARCHAR(140) NOT NULL UNIQUE,
    description VARCHAR(500),
    icon        VARCHAR(60),
    cover_url   VARCHAR(500),
    position    INT NOT NULL DEFAULT 0,
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  DATETIME NOT NULL,
    updated_at  DATETIME NOT NULL,
    INDEX idx_categories_parent (parent_id),
    CONSTRAINT fk_categories_parent FOREIGN KEY (parent_id) REFERENCES categories(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS topics (
    id               CHAR(36) PRIMARY KEY,
    category_id      CHAR(36) NOT NULL,
    author_id        CHAR(36) NOT NULL,
    title            VARCHAR(200) NOT NULL,
    slug             VARCHAR(240) NOT NULL,
    content          TEXT NOT NULL,
    is_pinned        BOOLEAN NOT NULL DEFAULT FALSE,
    is_locked        BOOLEAN NOT NULL DEFAULT FALSE,
    is_featured      BOOLEAN NOT NULL DEFAULT FALSE,
    status           VARCHAR(20) NOT NULL DEFAULT 'published',
    view_count       INT NOT NULL DEFAULT 0,
    reply_count      INT NOT NULL DEFAULT 0,
    like_count       INT NOT NULL DEFAULT 0,
    follower_count   INT NOT NULL DEFAULT 0,
    created_at       DATETIME NOT NULL,
    updated_at       DATETIME NOT NULL,
    last_activity_at DATETIME NOT NULL,
    INDEX idx_topics_category (category_id),
    INDEX idx_topics_author (author_id),
    INDEX idx_topics_last_activity (last_activity_at),
    CONSTRAINT fk_topics_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE,
    CONSTRAINT fk_topics_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS replies (
    id         CHAR(36) PRIMARY KEY,
    topic_id   CHAR(36) NOT NULL,
    author_id  CHAR(36) NOT NULL,
    parent_id  CHAR(36) NULL,
    content    TEXT NOT NULL,
    like_count INT NOT NULL DEFAULT 0,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    edited_at  DATETIME NULL,
    INDEX idx_replies_topic (topic_id),
    INDEX idx_replies_author (author_id),
    CONSTRAINT fk_replies_topic FOREIGN KEY (topic_id) REFERENCES topics(id) ON DELETE CASCADE,
    CONSTRAINT fk_replies_author FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_replies_parent FOREIGN KEY (parent_id) REFERENCES replies(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS likes (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    target_type VARCHAR(10) NOT NULL,
    target_id   CHAR(36) NOT NULL,
    created_at  DATETIME NOT NULL,
    UNIQUE KEY uq_like (user_id, target_type, target_id),
    INDEX idx_likes_target (target_type, target_id),
    CONSTRAINT fk_likes_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS follows (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    target_type VARCHAR(12) NOT NULL,
    target_id   CHAR(36) NOT NULL,
    created_at  DATETIME NOT NULL,
    UNIQUE KEY uq_follow (user_id, target_type, target_id),
    INDEX idx_follows_target (target_type, target_id),
    CONSTRAINT fk_follows_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS content_edits (
    id               CHAR(36) PRIMARY KEY,
    entity_type      VARCHAR(10) NOT NULL,
    entity_id        CHAR(36) NOT NULL,
    editor_id        CHAR(36) NOT NULL,
    previous_content TEXT NOT NULL,
    created_at       DATETIME NOT NULL,
    INDEX idx_content_edits_entity (entity_type, entity_id),
    CONSTRAINT fk_content_edits_editor FOREIGN KEY (editor_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
