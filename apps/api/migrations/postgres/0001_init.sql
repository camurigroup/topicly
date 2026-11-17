-- Etapa 1: tabela de instalação (singleton) para controle do instalador.
CREATE TABLE IF NOT EXISTS installations (
    id           SMALLINT PRIMARY KEY DEFAULT 1,
    installed    BOOLEAN NOT NULL DEFAULT FALSE,
    version      VARCHAR(32) NOT NULL DEFAULT '0.1.0',
    installed_at TIMESTAMPTZ NULL,
    CONSTRAINT installations_singleton CHECK (id = 1)
);

INSERT INTO installations (id, installed)
VALUES (1, FALSE)
ON CONFLICT (id) DO NOTHING;
