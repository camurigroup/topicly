-- Etapa 5: moderação (banimentos, mutes, advertências, denúncias) e auditoria.

CREATE TABLE IF NOT EXISTS sanctions (
    id            CHAR(36) PRIMARY KEY,
    user_id       CHAR(36) NOT NULL,
    type          VARCHAR(10) NOT NULL,
    reason        VARCHAR(500) NOT NULL,
    issued_by     CHAR(36) NULL,
    expires_at    DATETIME NULL,
    revoked_at    DATETIME NULL,
    revoked_by    CHAR(36) NULL,
    revoke_reason VARCHAR(500) NULL,
    created_at    DATETIME NOT NULL,
    INDEX idx_sanctions_user (user_id, type, revoked_at),
    INDEX idx_sanctions_created (created_at),
    CONSTRAINT fk_sanctions_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_sanctions_issuer FOREIGN KEY (issued_by) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_sanctions_revoker FOREIGN KEY (revoked_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS warnings (
    id         CHAR(36) PRIMARY KEY,
    user_id    CHAR(36) NOT NULL,
    reason     VARCHAR(500) NOT NULL,
    issued_by  CHAR(36) NULL,
    created_at DATETIME NOT NULL,
    INDEX idx_warnings_user (user_id, created_at),
    CONSTRAINT fk_warnings_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_warnings_issuer FOREIGN KEY (issued_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS reports (
    id              CHAR(36) PRIMARY KEY,
    reporter_id     CHAR(36) NULL,
    target_type     VARCHAR(10) NOT NULL,
    target_id       CHAR(36) NOT NULL,
    reason          VARCHAR(40) NOT NULL,
    details         VARCHAR(1000) NULL,
    status          VARCHAR(12) NOT NULL DEFAULT 'pending',
    handled_by      CHAR(36) NULL,
    resolution_note VARCHAR(1000) NULL,
    created_at      DATETIME NOT NULL,
    updated_at      DATETIME NOT NULL,
    INDEX idx_reports_status (status, created_at),
    INDEX idx_reports_target (target_type, target_id),
    CONSTRAINT fk_reports_reporter FOREIGN KEY (reporter_id) REFERENCES users(id) ON DELETE SET NULL,
    CONSTRAINT fk_reports_handler FOREIGN KEY (handled_by) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS audit_logs (
    id          CHAR(36) PRIMARY KEY,
    actor_id    CHAR(36) NULL,
    action      VARCHAR(60) NOT NULL,
    target_type VARCHAR(20) NULL,
    target_id   VARCHAR(80) NULL,
    details     TEXT NULL,
    ip          VARCHAR(64) NULL,
    created_at  DATETIME NOT NULL,
    INDEX idx_audit_created (created_at),
    INDEX idx_audit_action (action, created_at),
    INDEX idx_audit_actor (actor_id, created_at),
    CONSTRAINT fk_audit_actor FOREIGN KEY (actor_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
