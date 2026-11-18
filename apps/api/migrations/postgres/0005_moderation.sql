-- Etapa 5: moderação (banimentos, mutes, advertências, denúncias) e auditoria.

CREATE TABLE IF NOT EXISTS sanctions (
    id            CHAR(36) PRIMARY KEY,
    user_id       CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type          VARCHAR(10) NOT NULL,
    reason        VARCHAR(500) NOT NULL,
    issued_by     CHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    expires_at    TIMESTAMPTZ,
    revoked_at    TIMESTAMPTZ,
    revoked_by    CHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    revoke_reason VARCHAR(500),
    created_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_sanctions_user ON sanctions (user_id, type, revoked_at);
CREATE INDEX IF NOT EXISTS idx_sanctions_created ON sanctions (created_at);

CREATE TABLE IF NOT EXISTS warnings (
    id         CHAR(36) PRIMARY KEY,
    user_id    CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason     VARCHAR(500) NOT NULL,
    issued_by  CHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_warnings_user ON warnings (user_id, created_at);

CREATE TABLE IF NOT EXISTS reports (
    id              CHAR(36) PRIMARY KEY,
    reporter_id     CHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    target_type     VARCHAR(10) NOT NULL,
    target_id       CHAR(36) NOT NULL,
    reason          VARCHAR(40) NOT NULL,
    details         VARCHAR(1000),
    status          VARCHAR(12) NOT NULL DEFAULT 'pending',
    handled_by      CHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    resolution_note VARCHAR(1000),
    created_at      TIMESTAMPTZ NOT NULL,
    updated_at      TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_reports_status ON reports (status, created_at);
CREATE INDEX IF NOT EXISTS idx_reports_target ON reports (target_type, target_id);

CREATE TABLE IF NOT EXISTS audit_logs (
    id          CHAR(36) PRIMARY KEY,
    actor_id    CHAR(36) REFERENCES users(id) ON DELETE SET NULL,
    action      VARCHAR(60) NOT NULL,
    target_type VARCHAR(20),
    target_id   VARCHAR(80),
    details     TEXT,
    ip          VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs (action, created_at);
CREATE INDEX IF NOT EXISTS idx_audit_actor ON audit_logs (actor_id, created_at);
