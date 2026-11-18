-- Etapa 4: social, notificações, busca/descoberta e home administrável.

CREATE TABLE IF NOT EXISTS notifications (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    actor_id    CHAR(36) NULL,
    type        VARCHAR(30) NOT NULL,
    entity_type VARCHAR(12) NULL,
    entity_id   CHAR(36) NULL,
    message     VARCHAR(500) NOT NULL,
    link        VARCHAR(500) NULL,
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  DATETIME NOT NULL,
    INDEX idx_notifications_user (user_id, created_at),
    INDEX idx_notifications_unread (user_id, is_read),
    CONSTRAINT fk_notifications_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_notifications_actor FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS site_settings (
    setting_key VARCHAR(80) PRIMARY KEY,
    value       TEXT NOT NULL,
    updated_at  DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Ordenações de busca/descoberta.
CREATE INDEX IF NOT EXISTS idx_topics_created ON topics (created_at);
CREATE INDEX IF NOT EXISTS idx_topics_likes ON topics (like_count);
CREATE INDEX IF NOT EXISTS idx_replies_created ON replies (created_at);
CREATE INDEX IF NOT EXISTS idx_likes_created ON likes (target_type, target_id, created_at);
