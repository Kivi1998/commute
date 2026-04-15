-- =========================================================
-- 0001_init_schema.down.sql
-- 回滚初始化 Schema
-- =========================================================

DROP TABLE IF EXISTS query_history CASCADE;
DROP TABLE IF EXISTS ai_recommendation_cache CASCADE;
DROP TABLE IF EXISTS commute_result CASCADE;
DROP TABLE IF EXISTS commute_query CASCADE;
DROP TABLE IF EXISTS company CASCADE;
DROP TABLE IF EXISTS home_address CASCADE;
DROP TABLE IF EXISTS user_profile CASCADE;
DROP TABLE IF EXISTS app_user CASCADE;

DROP TYPE IF EXISTS commute_direction_enum;
DROP TYPE IF EXISTS time_strategy_enum;
DROP TYPE IF EXISTS transport_mode_enum;
DROP TYPE IF EXISTS company_source_enum;
DROP TYPE IF EXISTS company_status_enum;
DROP TYPE IF EXISTS company_type_enum;
