-- =========================================================
-- 0001_init_schema.up.sql
-- 初始化 Commute 项目数据库 Schema
-- 关联文档：docs/decisions/ADR-001-数据库设计.md
-- =========================================================

-- ---------- 枚举类型 ----------
CREATE TYPE company_type_enum AS ENUM ('big_tech', 'mid_tech', 'startup', 'foreign', 'other');

CREATE TYPE company_status_enum AS ENUM (
    'watching',
    'applied',
    'interviewing',
    'offered',
    'rejected',
    'archived'
);

CREATE TYPE company_source_enum AS ENUM ('ai_recommend', 'manual');

CREATE TYPE transport_mode_enum AS ENUM ('transit', 'driving', 'cycling', 'walking');

CREATE TYPE time_strategy_enum AS ENUM ('depart_at', 'arrive_by');

CREATE TYPE commute_direction_enum AS ENUM ('to_work', 'to_home');

-- ---------- 用户表（预留） ----------
CREATE TABLE app_user (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(64),
    phone       VARCHAR(20) UNIQUE,
    email       VARCHAR(128) UNIQUE,
    avatar_url  VARCHAR(512),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

COMMENT ON TABLE app_user IS '用户表（预留多用户扩展，MVP 阶段使用 id=1 的虚拟用户）';

-- ---------- 用户画像 ----------
CREATE TABLE user_profile (
    id                       BIGSERIAL PRIMARY KEY,
    user_id                  BIGINT NOT NULL DEFAULT 1 REFERENCES app_user(id),
    current_city             VARCHAR(64) NOT NULL,
    current_city_code        VARCHAR(16),
    target_position          VARCHAR(128) NOT NULL,
    experience_years         SMALLINT,
    preferred_company_types  company_type_enum[],
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id)
);

COMMENT ON TABLE user_profile IS '用户画像（城市、目标岗位、经验、偏好）';
COMMENT ON COLUMN user_profile.current_city_code IS '高德 adcode，如 110000';

-- ---------- 家庭住址 ----------
CREATE TABLE home_address (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL DEFAULT 1 REFERENCES app_user(id),
    alias       VARCHAR(64) NOT NULL,
    address     VARCHAR(512) NOT NULL,
    province    VARCHAR(32),
    city        VARCHAR(32),
    district    VARCHAR(32),
    longitude   NUMERIC(10, 7) NOT NULL,
    latitude    NUMERIC(10, 7) NOT NULL,
    is_default  BOOLEAN NOT NULL DEFAULT FALSE,
    note        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

COMMENT ON TABLE home_address IS '家庭住址（支持多个，用于租房选址对比）';
COMMENT ON COLUMN home_address.longitude IS 'GCJ-02 高德坐标系';
COMMENT ON COLUMN home_address.latitude IS 'GCJ-02 高德坐标系';

CREATE INDEX idx_home_address_user ON home_address (user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uniq_home_address_default
    ON home_address (user_id)
    WHERE is_default = TRUE AND deleted_at IS NULL;

-- ---------- 公司 ----------
CREATE TABLE company (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL DEFAULT 1 REFERENCES app_user(id),
    name        VARCHAR(128) NOT NULL,
    address     VARCHAR(512) NOT NULL,
    province    VARCHAR(32),
    city        VARCHAR(32),
    district    VARCHAR(32),
    longitude   NUMERIC(10, 7) NOT NULL,
    latitude    NUMERIC(10, 7) NOT NULL,
    category    company_type_enum,
    industry    VARCHAR(64),
    status      company_status_enum NOT NULL DEFAULT 'watching',
    source      company_source_enum NOT NULL DEFAULT 'manual',
    ai_reason   TEXT,
    note        TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

COMMENT ON TABLE company IS '公司（AI 推荐勾选或手动添加）';

CREATE INDEX idx_company_user ON company (user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_company_user_status ON company (user_id, status) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX uniq_company_user_name_addr
    ON company (user_id, name, address)
    WHERE deleted_at IS NULL;

-- ---------- 通勤查询会话 ----------
CREATE TABLE commute_query (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL DEFAULT 1 REFERENCES app_user(id),
    home_id           BIGINT NOT NULL REFERENCES home_address(id),
    transport_modes   transport_mode_enum[] NOT NULL,
    morning_strategy  time_strategy_enum NOT NULL DEFAULT 'depart_at',
    morning_time      TIME NOT NULL DEFAULT '08:00',
    evening_strategy  time_strategy_enum NOT NULL DEFAULT 'depart_at',
    evening_time      TIME NOT NULL DEFAULT '17:30',
    weekday           SMALLINT NOT NULL DEFAULT 1,
    buffer_minutes    SMALLINT NOT NULL DEFAULT 5,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE commute_query IS '通勤查询会话（一次批量计算的参数集）';
COMMENT ON COLUMN commute_query.weekday IS 'ISO weekday：1=周一，7=周日';

CREATE INDEX idx_commute_query_user ON commute_query (user_id, created_at DESC);

-- ---------- 通勤计算结果（含 7 天缓存） ----------
CREATE TABLE commute_result (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL DEFAULT 1 REFERENCES app_user(id),
    query_id          BIGINT REFERENCES commute_query(id) ON DELETE SET NULL,
    home_id           BIGINT NOT NULL REFERENCES home_address(id),
    company_id        BIGINT NOT NULL REFERENCES company(id),
    direction         commute_direction_enum NOT NULL,
    transport_mode    transport_mode_enum NOT NULL,
    depart_time       TIME NOT NULL,
    arrive_time       TIME NOT NULL,
    weekday           SMALLINT NOT NULL DEFAULT 1,
    duration_min      INT NOT NULL,
    duration_min_raw  INT NOT NULL,
    distance_km       NUMERIC(8, 2) NOT NULL,
    cost_yuan         NUMERIC(8, 2),
    transfer_count    SMALLINT,
    route_detail      JSONB NOT NULL,
    calculated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at        TIMESTAMPTZ NOT NULL,
    is_failed         BOOLEAN NOT NULL DEFAULT FALSE,
    error_message     TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE commute_result IS '通勤计算结果（含 7 天缓存，命中条件见 idx_commute_result_cache）';
COMMENT ON COLUMN commute_result.duration_min IS '含 5 分钟 buffer';
COMMENT ON COLUMN commute_result.duration_min_raw IS '高德返回原值，不含 buffer';

CREATE INDEX idx_commute_result_cache
    ON commute_result (home_id, company_id, transport_mode, direction, depart_time, expires_at);
CREATE INDEX idx_commute_result_query ON commute_result (query_id);
CREATE INDEX idx_commute_result_user ON commute_result (user_id, calculated_at DESC);

-- ---------- AI 推荐缓存 ----------
CREATE TABLE ai_recommendation_cache (
    id                BIGSERIAL PRIMARY KEY,
    user_id           BIGINT NOT NULL DEFAULT 1 REFERENCES app_user(id),
    city              VARCHAR(64) NOT NULL,
    position          VARCHAR(128) NOT NULL,
    experience_years  SMALLINT,
    company_types     company_type_enum[],
    cache_key         VARCHAR(256) NOT NULL,
    raw_response      JSONB NOT NULL,
    company_count     SMALLINT NOT NULL,
    requested_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at        TIMESTAMPTZ NOT NULL,
    token_input       INT,
    token_output      INT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE ai_recommendation_cache IS '豆包 AI 公司推荐缓存（24h TTL）';
COMMENT ON COLUMN ai_recommendation_cache.cache_key IS 'md5(user_id+city+position+exp+types)';

CREATE INDEX idx_ai_cache_key ON ai_recommendation_cache (cache_key, expires_at);
CREATE INDEX idx_ai_cache_user ON ai_recommendation_cache (user_id, created_at DESC);

-- ---------- 查询历史快照 ----------
CREATE TABLE query_history (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL DEFAULT 1 REFERENCES app_user(id),
    title       VARCHAR(128),
    query_id    BIGINT REFERENCES commute_query(id) ON DELETE SET NULL,
    snapshot    JSONB NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at  TIMESTAMPTZ
);

COMMENT ON TABLE query_history IS '查询历史快照（保存完整 UI 状态，便于一键恢复）';

CREATE INDEX idx_query_history_user
    ON query_history (user_id, created_at DESC) WHERE deleted_at IS NULL;
