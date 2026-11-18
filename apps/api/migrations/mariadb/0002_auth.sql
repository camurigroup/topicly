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
    reputation     INT NOT NULL DEFAULT 0,
    status         VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at     DATETIME NOT NULL,
    updated_at     DATETIME NOT NULL,
    last_active_at DATETIME NULL,
    INDEX idx_users_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS roles (
    id         CHAR(36) PRIMARY KEY,
    slug       VARCHAR(50) NOT NULL UNIQUE,
    name       VARCHAR(80) NOT NULL,
    rank_order INT NOT NULL DEFAULT 0,
    is_system  BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS permissions (
    id          CHAR(36) PRIMARY KEY,
    slug        VARCHAR(80) NOT NULL UNIQUE,
    name        VARCHAR(120) NOT NULL,
    description VARCHAR(255),
    is_active   BOOLEAN NOT NULL DEFAULT TRUE,
    created_at  DATETIME NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       CHAR(36) NOT NULL,
    permission_id CHAR(36) NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_perm FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS user_roles (
    user_id CHAR(36) NOT NULL,
    role_id CHAR(36) NOT NULL,
    PRIMARY KEY (user_id, role_id),
    CONSTRAINT fk_ur_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT fk_ur_role FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS refresh_tokens (
    id          CHAR(36) PRIMARY KEY,
    user_id     CHAR(36) NOT NULL,
    token_hash  VARCHAR(64) NOT NULL UNIQUE,
    expires_at  DATETIME NOT NULL,
    revoked_at  DATETIME NULL,
    replaced_by CHAR(36) NULL,
    user_agent  VARCHAR(255),
    ip          VARCHAR(64),
    created_at  DATETIME NOT NULL,
    INDEX idx_refresh_tokens_user (user_id),
    CONSTRAINT fk_rt_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
