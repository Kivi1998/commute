-- 用户表增加密码哈希列（bcrypt）
ALTER TABLE app_user ADD COLUMN password_hash TEXT;

COMMENT ON COLUMN app_user.password_hash IS 'bcrypt 哈希；NULL 表示此账号未启用登录';
