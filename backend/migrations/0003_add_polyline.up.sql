-- 为通勤计算结果添加真实路线坐标串
ALTER TABLE commute_result ADD COLUMN polyline TEXT NOT NULL DEFAULT '';

COMMENT ON COLUMN commute_result.polyline IS '高德返回的路线点串，格式 "lng,lat;lng,lat;..."';
