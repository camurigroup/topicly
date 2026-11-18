-- Etapa 4: social, notificações, busca/descoberta e home administrável.

CREATE TABLE IF NOT EXISTS notifications (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    actor_id    CHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    type        VARCHAR(30) NOT NULL,
    entity_type VARCHAR(12),
    entity_id   CHAR(36),
    message     VARCHAR(500) NOT NULL,
    link        VARCHAR(500),
    is_read     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications (user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_notifications_unread ON notifications (user_id, is_read);

CREATE TABLE IF NOT EXISTS site_settings (
    setting_key VARCHAR(80) PRIMARY KEY,
    value       TEXT NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL
);

-- Ordenações de busca/descoberta.
CREATE INDEX IF NOT EXISTS idx_topics_created ON topics (created_at);
CREATE INDEX IF NOT EXISTS idx_topics_likes ON topics (like_count);
CREATE INDEX IF NOT EXISTS idx_replies_created ON replies (created_at);
CREATE INDEX IF NOT EXISTS idx_likes_created ON likes (target_type, target_id, created_at);
