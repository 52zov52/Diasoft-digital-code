-- ФАЗА 2: Добавляем поддержку аутентификации и сессий
BEGIN;

-- 1. Обновляем таблицу users (добавляем login и last_login_at)
ALTER TABLE users ADD COLUMN IF NOT EXISTS login VARCHAR(50) UNIQUE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMPTZ;
ALTER TABLE users ALTER COLUMN email DROP NOT NULL;

-- Индекс для login
CREATE INDEX IF NOT EXISTS idx_user_login ON users(login);

-- 2. Таблица сессий (для JWT blacklist / управления сессиями)
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE, -- Хеш JWT токена для blacklist
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    revoked_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_session_token ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_session_user ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_session_expires ON user_sessions(expires_at);

-- 3. Таблица университетов уже существует, но убедимся что она есть
-- (если миграция 001 еще не применена, эта таблица будет создана там)

COMMIT;
