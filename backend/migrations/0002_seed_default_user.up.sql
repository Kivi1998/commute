-- =========================================================
-- 0002_seed_default_user.up.sql
-- 插入虚拟默认用户（id=1），MVP 阶段所有数据归属此用户
-- =========================================================

INSERT INTO app_user (id, name) VALUES (1, '默认用户')
ON CONFLICT (id) DO NOTHING;

-- 确保 BIGSERIAL 序列从 2 开始
SELECT setval('app_user_id_seq', GREATEST((SELECT MAX(id) FROM app_user), 1), true);
