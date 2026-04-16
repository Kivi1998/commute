-- 用户画像新增联系人字段（姓名/电话/邮箱），用于地址复制、信息展示等
ALTER TABLE user_profile
  ADD COLUMN full_name VARCHAR(32),
  ADD COLUMN phone VARCHAR(20),
  ADD COLUMN email VARCHAR(128);

COMMENT ON COLUMN user_profile.full_name IS '用户真实姓名（用于地址卡片展示）';
COMMENT ON COLUMN user_profile.phone IS '联系电话';
COMMENT ON COLUMN user_profile.email IS '联系邮箱';
