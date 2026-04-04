-- ФАЗА 1: Базовая схема с учетом 152-ФЗ и криптографических требований
BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_stat_statements"; -- Для мониторинга нагрузки

-- 1. Таблица ВУЗов
CREATE TABLE universities (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(16) NOT NULL UNIQUE,
    official_name VARCHAR(255) NOT NULL,
    public_key_pem TEXT NOT NULL, -- Ed25519 или RSA публичный ключ для верификации подписи
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_uni_code ON universities(code);

-- 2. Пользователи (RBAC)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL, -- bcrypt/scrypt
    role VARCHAR(20) CHECK (role IN ('university', 'student', 'hr')),
    university_id UUID REFERENCES universities(id) ON DELETE SET NULL,
    consent_recorded_at TIMESTAMPTZ, -- Фиксация согласия на обработку ПДн (152-ФЗ ст.9)
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_user_role ON users(role);

-- 3. Реестр дипломов (основная таблица)
CREATE TABLE diplomas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    university_id UUID NOT NULL REFERENCES universities(id) ON DELETE RESTRICT,
    serial_number VARCHAR(32) NOT NULL,
    specialty VARCHAR(255) NOT NULL,
    issue_year INTEGER NOT NULL CHECK (issue_year >= 1990 AND issue_year <= 2030),
    
    -- 🔒 Шифрование ПДн (AES-256-GCM, ciphertext в base64)
    encrypted_fio_full TEXT NOT NULL,
    encrypted_passport_sn TEXT, -- Серия/номер паспорта (опционально, по согласованию)
    
    -- 🔐 Целостность и подпись
    document_hash VARCHAR(64) NOT NULL, -- SHA-256 канонического JSON payload
    digital_signature TEXT NOT NULL,    -- Подпись ВУЗа (Ed25519/RSA) base64
    signing_algorithm VARCHAR(16) DEFAULT 'Ed25519',
    
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'revoked', 'suspended')),
    revoked_at TIMESTAMPTZ,
    revocation_reason TEXT,
    
    issued_at TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    CONSTRAINT uq_diploma_serial UNIQUE (university_id, serial_number)
);

CREATE INDEX idx_diploma_hash ON diplomas(document_hash);
CREATE INDEX idx_diploma_status ON diplomas(status) WHERE status != 'active'; -- Partial index
CREATE INDEX idx_diploma_year ON diplomas(issue_year);

-- 4. QR-токены (временный доступ)
CREATE TABLE qr_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    diploma_id UUID NOT NULL REFERENCES diplomas(id) ON DELETE CASCADE,
    token VARCHAR(64) NOT NULL UNIQUE, -- JTI или криптографический рандом
    payload_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    max_uses INTEGER DEFAULT 50,
    current_uses INTEGER DEFAULT 0 CHECK (current_uses <= max_uses),
    is_revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_qr_token ON qr_tokens(token);
CREATE INDEX idx_qr_expires ON qr_tokens(expires_at);

-- 5. Журнал верификаций (Аудит 152-ФЗ ст.18.1)
CREATE TABLE verification_logs (
    id BIGSERIAL PRIMARY KEY,
    diploma_id UUID REFERENCES diplomas(id),
    qr_token_id UUID REFERENCES qr_tokens(id),
    verifier_ip INET NOT NULL,
    user_agent TEXT,
    result VARCHAR(10) CHECK (result IN ('valid', 'invalid', 'expired', 'revoked')),
    verified_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_log_verified_at ON verification_logs(verified_at DESC);
CREATE INDEX idx_log_diploma ON verification_logs(diploma_id);

COMMIT;