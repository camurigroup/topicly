-- Etapa 1: tabela de instalação (singleton) para controle do instalador.
CREATE TABLE IF NOT EXISTS installations (
    id           TINYINT PRIMARY KEY DEFAULT 1,
    installed    BOOLEAN NOT NULL DEFAULT FALSE,
    version      VARCHAR(32) NOT NULL DEFAULT '0.1.0',
    installed_at DATETIME NULL,
    CONSTRAINT installations_singleton CHECK (id = 1)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT IGNORE INTO installations (id, installed) VALUES (1, FALSE);
