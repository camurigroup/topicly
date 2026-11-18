-- Etapa 2: autenticação, usuários, cargos, permissões e refresh tokens.

CREATE TABLE IF NOT EXISTS users (
    id             CHAR(36) PRIMARY KEY,
    username       VARCHAR(32) NOT NULL UNIQUE,
    email          VARCHAR(255) NOT NULL UNIQUE,
    password_hash  VARCHAR(255) NOT NULL,
    display_name   VARCHAR(80) NOT NULL,
    bio            VARCHAR(500),
    avatar_url     VARCHAR(500),
    banner_url     VARCHAR(500),
    reputation     INTEGER NOT NULL DEFAULT 0,
    status         VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at     TIMESTAMPTZ NOT NULL,
    updated_at     TIMESTAMPTZ NOT NULL,
    last_active_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_users_status ON users (status);

CREATE TABLE IF NOT EXISTS roles (
    id         CHAR(36) PRIMARY KEY,
    slug       VARCHAR(50) NOT NULL UNIQUE,
    name       VARCHAR(80) NOT NULL,
    rank_order INTEGER NOT NULL DEFAULT 0,
    is_system  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS permissions (
    id          CHAR(36) PRIMARY KEY,
    slug        VARCHAR(80) NOT NULL UNIQUE,
    name        VARCHAR(120) NOT NULL,
    description VARCHAR(255),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       CHAR(36) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id CHAR(36) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id CHAR(36) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(64) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    replaced_by CHAR(36),
    user_agent  VARCHAR(255),
    ip          VARCHAR(64),
    created_at  TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user ON refresh_tokens (user_id);
