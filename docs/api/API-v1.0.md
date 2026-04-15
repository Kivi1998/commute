# Commute 后端 API 设计文档

| 字段 | 内容 |
|------|------|
| 文档版本 | v1.0 |
| 创建日期 | 2026-04-15 |
| 作者 | Hao Jia |
| 状态 | 已定稿 |
| Base URL | `http://localhost:8080/api/v1` |
| 关联 | [PRD](../prd/PRD-v1.0.md) / [DB Schema](../decisions/ADR-001-数据库设计.md) |

---

## 1. 通用规范

### 1.1 协议与编码
- 协议：HTTP/1.1
- Content-Type：`application/json; charset=utf-8`
- 字符集：UTF-8
- 时间格式：ISO 8601 (`2026-04-15T08:00:00+08:00`)
- 时区：默认 `Asia/Shanghai`
- 坐标系：GCJ-02（高德坐标系），所有经纬度统一

### 1.2 统一响应格式

```json
{
  "code": 0,
  "message": "ok",
  "data": { },
  "request_id": "req_xxxxxxxxxxxxx"
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 业务码，**0 = 成功**，非 0 表示业务错误 |
| `message` | string | 错误描述（中文） |
| `data` | object \| array \| null | 实际数据 |
| `request_id` | string | 请求追踪 ID（日志关联） |

### 1.3 分页响应

```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "list": [],
    "pagination": {
      "page": 1,
      "page_size": 20,
      "total": 100,
      "total_pages": 5
    }
  }
}
```

### 1.4 HTTP 状态码使用
| HTTP | 含义 | 业务 code 是否覆盖 |
|------|------|-------------------|
| 200 | 成功 | code=0 |
| 400 | 请求参数错误 | code=400xx |
| 401 | 未认证（预留） | code=401xx |
| 403 | 无权限（预留） | code=403xx |
| 404 | 资源不存在 | code=404xx |
| 409 | 资源冲突 | code=409xx |
| 422 | 业务校验失败 | code=422xx |
| 500 | 服务器异常 | code=500xx |
| 502 | 上游错误（高德/豆包） | code=502xx |

### 1.5 业务错误码

| code | HTTP | 含义 |
|------|------|------|
| 0 | 200 | 成功 |
| 40001 | 400 | 参数缺失或格式错误 |
| 40002 | 400 | 参数取值非法 |
| 40401 | 404 | 资源不存在 |
| 40901 | 409 | 资源已存在（如同名公司） |
| 42201 | 422 | 业务规则校验失败 |
| 42202 | 422 | 公司勾选数量超限（>20） |
| 50001 | 500 | 内部错误 |
| 50201 | 502 | 高德 API 调用失败 |
| 50202 | 502 | 高德返回数据异常 |
| 50203 | 502 | 豆包 API 调用失败 |
| 50204 | 502 | 豆包返回格式异常 |

### 1.6 鉴权（预留）
当前 MVP **无鉴权**，所有请求自动归属 `user_id=1`。

未来扩展方案：
```
Header: Authorization: Bearer <jwt_token>
```
后端中间件解析 token → 注入 `user_id` 到 context。

### 1.7 通用 Header

| Header | 必填 | 说明 |
|--------|------|------|
| `Content-Type` | 是（POST/PUT） | `application/json` |
| `X-Request-Id` | 否 | 客户端可传入用于追踪；后端缺省自动生成 |

---

## 2. API 资源总览

| 资源 | 路径前缀 | 模块 |
|------|---------|------|
| 用户画像 | `/profile` | M1 |
| 家庭住址 | `/addresses` | M2 |
| AI 推荐 | `/ai/recommend` | M3 |
| 公司 | `/companies` | M4 |
| 通勤计算 | `/commute` | M5 |
| 历史记录 | `/history` | M7 |
| 导出 | `/export` | M9 |
| 地图代理 | `/map` | M6 / 高德代理 |
| 元数据 | `/meta` | 字典/枚举 |
| 健康检查 | `/health` | 运维 |

---

## 3. 详细接口定义

### M1 用户画像 `/profile`

#### GET `/profile`
获取当前用户画像。

**响应 200**
```json
{
  "code": 0,
  "message": "ok",
  "data": {
    "id": 1,
    "user_id": 1,
    "current_city": "北京",
    "current_city_code": "110000",
    "target_position": "后台开发",
    "experience_years": 5,
    "preferred_company_types": ["big_tech", "mid_tech"],
    "created_at": "2026-04-15T10:00:00+08:00",
    "updated_at": "2026-04-15T10:00:00+08:00"
  }
}
```
> 若不存在返回 `data: null`。

#### PUT `/profile`
创建或更新（upsert）用户画像。

**请求 Body**
```json
{
  "current_city": "北京",
  "current_city_code": "110000",
  "target_position": "后台开发",
  "experience_years": 5,
  "preferred_company_types": ["big_tech", "mid_tech"]
}
```

**响应 200**：返回更新后的完整 profile（结构同 GET）。

**校验**：
- `current_city` 必填，长度 1-64
- `target_position` 必填，长度 1-128
- `experience_years` 0-30
- `preferred_company_types` 元素必须在 `[big_tech, mid_tech, startup, foreign, other]` 中

---

### M2 家庭住址 `/addresses`

#### GET `/addresses`
列出当前用户所有家庭住址。

**Query 参数**
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `include_deleted` | bool | 否 | 默认 false |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "alias": "我家",
        "address": "北京市朝阳区建国路 88 号",
        "province": "北京市",
        "city": "北京市",
        "district": "朝阳区",
        "longitude": 116.4500000,
        "latitude": 39.9080000,
        "is_default": true,
        "note": "现居",
        "created_at": "2026-04-15T10:00:00+08:00"
      }
    ]
  }
}
```

#### POST `/addresses`
新增家庭住址。

**请求 Body**
```json
{
  "alias": "候选 A",
  "address": "北京市海淀区中关村大街 1 号",
  "province": "北京市",
  "city": "北京市",
  "district": "海淀区",
  "longitude": 116.3100000,
  "latitude": 39.9830000,
  "is_default": false,
  "note": "月租 5000"
}
```

**响应 201**：返回创建后的对象。

**校验**：
- `alias` 必填
- 经纬度必填，范围合法
- `is_default=true` 时自动取消其他地址的默认状态

#### GET `/addresses/:id`
查询单个住址详情。

#### PUT `/addresses/:id`
更新住址。Body 同 POST，所有字段可选（局部更新）。

#### DELETE `/addresses/:id`
软删除住址。

**业务规则**：若为默认地址且还有其他地址，**自动将最早创建的设为默认**。

#### POST `/addresses/:id/set-default`
设为默认。

---

### M3 AI 公司推荐 `/ai/recommend`

#### POST `/ai/recommend/companies`
基于"城市 + 岗位"调用豆包 AI 推荐公司列表。

**请求 Body**
```json
{
  "city": "北京",
  "position": "后台开发",
  "experience_years": 5,
  "company_types": ["big_tech", "mid_tech"],
  "count": 20,
  "force_refresh": false
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `city` | 是 | |
| `position` | 是 | |
| `experience_years` | 否 | 影响推荐精准度 |
| `company_types` | 否 | 偏好筛选 |
| `count` | 否 | 期望数量，默认 20，范围 5-50 |
| `force_refresh` | 否 | 默认 false，true 时绕过 24h 缓存 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "from_cache": true,
    "cached_at": "2026-04-15T08:00:00+08:00",
    "expires_at": "2026-04-16T08:00:00+08:00",
    "companies": [
      {
        "name": "字节跳动",
        "category": "big_tech",
        "industry": "互联网",
        "address_hint": "北京市海淀区中关村",
        "reason": "国内顶级互联网大厂，后台开发岗位需求量大"
      }
    ]
  }
}
```

**错误**：
- `50203` 豆包调用失败
- `50204` 豆包返回非 JSON 或字段缺失

---

### M4 公司 `/companies`

#### GET `/companies`
列出关注公司。

**Query 参数**
| 参数 | 类型 | 默认 | 说明 |
|------|------|------|------|
| `status` | string | - | 按状态筛选：`watching/applied/interviewing/...` |
| `category` | string | - | 按类型筛选 |
| `keyword` | string | - | 名称模糊搜索 |
| `page` | int | 1 | |
| `page_size` | int | 50 | 最大 100 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "name": "字节跳动",
        "address": "北京市海淀区中关村大街 1 号",
        "province": "北京市",
        "city": "北京市",
        "district": "海淀区",
        "longitude": 116.3100000,
        "latitude": 39.9830000,
        "category": "big_tech",
        "industry": "互联网",
        "status": "watching",
        "source": "ai_recommend",
        "ai_reason": "...",
        "note": "HR: 张三 13800138000",
        "created_at": "2026-04-15T10:00:00+08:00"
      }
    ],
    "pagination": { "page": 1, "page_size": 50, "total": 12, "total_pages": 1 }
  }
}
```

#### POST `/companies`
手动添加公司。

**Body**：除 `id`/`source` 外的字段。`source` 自动设为 `manual`。

#### POST `/companies/batch`
批量添加（用于 AI 推荐勾选场景）。

**Body**
```json
{
  "companies": [
    {
      "name": "字节跳动",
      "address": "北京市海淀区中关村大街 1 号",
      "longitude": 116.3100000,
      "latitude": 39.9830000,
      "category": "big_tech",
      "industry": "互联网",
      "ai_reason": "..."
    }
  ]
}
```

**响应 200**
```json
{
  "code": 0,
  "data": {
    "created": [{ "id": 10, "name": "字节跳动" }],
    "skipped": [{ "name": "美团", "reason": "duplicate" }]
  }
}
```

**业务规则**：
- 自动 `source = ai_recommend`
- 同名+同地址重复时跳过（不报错）
- 当前用户公司总数 + 新增数 > 20 时，仍允许但响应中包含 `warning: "soft_limit_exceeded"`

#### GET `/companies/:id`
公司详情。

#### PUT `/companies/:id`
更新（含状态变更）。

#### PATCH `/companies/:id/status`
专用快捷接口：仅修改状态。
```json
{ "status": "interviewing" }
```

#### DELETE `/companies/:id`
软删除。

---

### M5 通勤计算 `/commute`

#### POST `/commute/calculate`
**核心接口**：批量计算通勤数据。

**请求 Body**
```json
{
  "home_id": 1,
  "company_ids": [10, 11, 12],
  "transport_modes": ["transit", "driving"],
  "morning": {
    "strategy": "depart_at",
    "time": "08:00"
  },
  "evening": {
    "strategy": "depart_at",
    "time": "17:30"
  },
  "weekday": 1,
  "buffer_minutes": 5,
  "force_refresh": false,
  "save_query": true
}
```

| 字段 | 必填 | 说明 |
|------|------|------|
| `home_id` | 是 | 起点（家） |
| `company_ids` | 是 | 公司列表，1-20 |
| `transport_modes` | 是 | 至少一种 |
| `morning.strategy` | 否 | `depart_at`（默认）/ `arrive_by` |
| `morning.time` | 否 | HH:MM，默认 `08:00` |
| `evening.strategy` | 否 | 默认 `depart_at` |
| `evening.time` | 否 | 默认 `17:30` |
| `weekday` | 否 | 1-7，默认 1（周一） |
| `buffer_minutes` | 否 | 默认 5 |
| `force_refresh` | 否 | true 时跳过 7 天缓存 |
| `save_query` | 否 | true 时保存 commute_query 供历史回溯 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "query_id": 100,
    "home": {
      "id": 1,
      "alias": "我家",
      "longitude": 116.4500000,
      "latitude": 39.9080000
    },
    "weekday": 1,
    "buffer_minutes": 5,
    "results": [
      {
        "company_id": 10,
        "company_name": "字节跳动",
        "company_longitude": 116.3100000,
        "company_latitude": 39.9830000,
        "items": [
          {
            "direction": "to_work",
            "transport_mode": "transit",
            "depart_time": "08:00",
            "arrive_time": "08:55",
            "duration_min": 55,
            "duration_min_raw": 50,
            "distance_km": 12.30,
            "cost_yuan": 6.00,
            "transfer_count": 1,
            "from_cache": false,
            "result_id": 200
          },
          {
            "direction": "to_home",
            "transport_mode": "transit",
            "depart_time": "17:30",
            "arrive_time": "18:32",
            "duration_min": 62,
            "duration_min_raw": 57,
            "distance_km": 12.30,
            "cost_yuan": 6.00,
            "transfer_count": 1,
            "from_cache": true,
            "result_id": 201
          }
        ],
        "errors": []
      }
    ],
    "summary": {
      "total_companies": 1,
      "total_calculations": 4,
      "cache_hits": 1,
      "failures": 0
    }
  }
}
```

**业务规则**：
- 并发调用高德 API，**最大并发 5**
- 每条 [home × company × mode × direction] 独立缓存命中
- 任一组合失败不影响其他，记录在 `errors` 中
- `company_ids.length > 20` 返回 `42202` 但仍执行（软上限）

**错误**：
- `40001` 缺少必填字段
- `40002` 时间格式错误 / weekday 越界
- `40401` home_id 或 company_id 不存在
- `42202` 公司数超过 20（仍执行，警告）

#### GET `/commute/results/:id`
查询单条结果详情（含完整 `route_detail`）。

**响应 200**
```json
{
  "code": 0,
  "data": {
    "id": 200,
    "home": { ... },
    "company": { ... },
    "direction": "to_work",
    "transport_mode": "transit",
    "depart_time": "08:00",
    "arrive_time": "08:55",
    "duration_min": 55,
    "distance_km": 12.30,
    "cost_yuan": 6.00,
    "transfer_count": 1,
    "route_detail": {
      "segments": [
        { "type": "walking", "duration": 5, "distance": 0.4, "instruction": "步行至 国贸地铁站" },
        { "type": "subway", "line": "10 号线", "from": "国贸", "to": "海淀黄庄", "duration": 35, "stops": 8 },
        { "type": "walking", "duration": 10, "distance": 0.8, "instruction": "步行至 字节跳动" }
      ]
    },
    "calculated_at": "2026-04-15T10:00:00+08:00"
  }
}
```

#### GET `/commute/queries/:id`
查询某次会话的所有结果（用于历史还原）。

---

### M7 历史记录 `/history`

#### GET `/history`
列出查询历史。

**Query**：`page`, `page_size`

**响应 200**
```json
{
  "code": 0,
  "data": {
    "list": [
      {
        "id": 1,
        "title": "2026-04-15 朝阳-后台开发",
        "query_id": 100,
        "summary": {
          "company_count": 5,
          "transport_modes": ["transit", "driving"]
        },
        "created_at": "2026-04-15T10:00:00+08:00"
      }
    ],
    "pagination": { ... }
  }
}
```

#### POST `/history`
保存当前查询为历史快照。

**Body**
```json
{
  "title": "2026-04-15 朝阳-后台开发",
  "query_id": 100,
  "snapshot": { ... }
}
```

#### GET `/history/:id`
查询历史详情（返回完整 snapshot）。

#### PATCH `/history/:id`
更新标题。

#### DELETE `/history/:id`
软删除。

#### POST `/history/:id/restore`
返回历史快照内容供前端还原 UI 状态。响应即 snapshot。

---

### M9 导出 `/export`

#### POST `/export/excel`
导出 Excel。

**Body**
```json
{
  "query_id": 100
}
```
或
```json
{
  "snapshot": { ... }
}
```

**响应**：
- Content-Type: `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet`
- Content-Disposition: `attachment; filename="commute-2026-04-15.xlsx"`
- Body: 二进制流

#### POST `/export/pdf`
同上，返回 PDF。

---

### M6 / 地图相关 `/map`

> 高德 API 走后端代理，避免前端泄露 Web Service Key。
> JS SDK Key（前端使用）由前端 `.env` 直接配置。

#### GET `/map/geocode`
地址 → 经纬度。

**Query**
| 参数 | 必填 | 说明 |
|------|------|------|
| `address` | 是 | 地址字符串 |
| `city` | 否 | 限定城市，提升精度 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "results": [
      {
        "formatted_address": "北京市朝阳区建国路 88 号",
        "province": "北京市",
        "city": "北京市",
        "district": "朝阳区",
        "longitude": 116.4500000,
        "latitude": 39.9080000,
        "level": "门牌号"
      }
    ]
  }
}
```

#### GET `/map/regeocode`
经纬度 → 地址（逆地理编码）。

**Query**：`longitude`, `latitude`

#### GET `/map/poi/search`
POI 搜索（用于公司地址定位）。

**Query**
| 参数 | 必填 | 说明 |
|------|------|------|
| `keyword` | 是 | 公司名等 |
| `city` | 否 | 限定城市 |
| `page_size` | 否 | 默认 10 |

**响应 200**
```json
{
  "code": 0,
  "data": {
    "results": [
      {
        "id": "B0FFFAB6J2",
        "name": "字节跳动（中关村）",
        "address": "北京市海淀区中关村大街 1 号",
        "longitude": 116.3100000,
        "latitude": 39.9830000,
        "type": "公司企业",
        "city": "北京市"
      }
    ]
  }
}
```

#### GET `/map/cities`
城市列表（高德支持的主要城市）。

---

### 元数据 `/meta`

#### GET `/meta/enums`
返回所有枚举字典，供前端下拉选择。

**响应 200**
```json
{
  "code": 0,
  "data": {
    "company_type": [
      { "value": "big_tech", "label": "大厂" },
      { "value": "mid_tech", "label": "中厂" },
      { "value": "startup", "label": "创业公司" },
      { "value": "foreign", "label": "外企" },
      { "value": "other", "label": "其他" }
    ],
    "company_status": [
      { "value": "watching", "label": "关注" },
      { "value": "applied", "label": "已投递" },
      { "value": "interviewing", "label": "面试中" },
      { "value": "offered", "label": "已 offer" },
      { "value": "rejected", "label": "已拒/被拒" },
      { "value": "archived", "label": "归档" }
    ],
    "transport_mode": [
      { "value": "transit", "label": "公交/地铁", "icon": "🚇" },
      { "value": "driving", "label": "驾车", "icon": "🚗" },
      { "value": "cycling", "label": "骑行", "icon": "🚴" },
      { "value": "walking", "label": "步行", "icon": "🚶" }
    ],
    "time_strategy": [
      { "value": "depart_at", "label": "指定出发时间" },
      { "value": "arrive_by", "label": "指定到达时间" }
    ]
  }
}
```

---

### 健康检查 `/health`

#### GET `/health`
**响应 200**
```json
{
  "code": 0,
  "data": {
    "status": "ok",
    "version": "1.0.0",
    "uptime_seconds": 12345,
    "dependencies": {
      "database": "ok",
      "amap": "ok",
      "doubao": "ok"
    }
  }
}
```

---

## 4. 速率限制（预留）

MVP 阶段不强制，但后端中间件已预留：
- 全局：100 req/s（防误用）
- AI 推荐：每用户 10 req/min（防 token 浪费）
- 通勤计算：每用户 30 req/min（防高德配额耗尽）

---

## 5. 缓存策略汇总

| 接口 | 缓存表 | 命中条件 | TTL | 绕过方式 |
|------|--------|---------|-----|---------|
| `/ai/recommend/companies` | `ai_recommendation_cache` | city + position + experience + types | 24h | `force_refresh=true` |
| `/commute/calculate` 单条 | `commute_result` | home_id + company_id + mode + direction + depart_time | 7d | `force_refresh=true` |
| `/map/geocode` | 内存 LRU | 完整 query | 1h | 无 |
| `/meta/enums` | 内存 | 全局 | 启动期 | 重启 |

---

## 6. 错误处理示例

### 示例 1：参数缺失
```json
HTTP 400
{
  "code": 40001,
  "message": "current_city 不能为空",
  "data": {
    "field": "current_city"
  },
  "request_id": "req_abc"
}
```

### 示例 2：高德失败
```json
HTTP 502
{
  "code": 50201,
  "message": "高德地图服务暂时不可用，请稍后重试",
  "data": {
    "amap_code": "10003",
    "amap_message": "ACCESS_OVERFLOW"
  },
  "request_id": "req_abc"
}
```

### 示例 3：公司数量软上限
```json
HTTP 200
{
  "code": 0,
  "message": "已计算 25 家公司，建议不超过 20 家以保证体验",
  "data": {
    "warning": "soft_limit_exceeded",
    "results": [...]
  },
  "request_id": "req_abc"
}
```

---

## 7. 接口路径速查

```
GET    /api/v1/health
GET    /api/v1/meta/enums

GET    /api/v1/profile
PUT    /api/v1/profile

GET    /api/v1/addresses
POST   /api/v1/addresses
GET    /api/v1/addresses/:id
PUT    /api/v1/addresses/:id
DELETE /api/v1/addresses/:id
POST   /api/v1/addresses/:id/set-default

POST   /api/v1/ai/recommend/companies

GET    /api/v1/companies
POST   /api/v1/companies
POST   /api/v1/companies/batch
GET    /api/v1/companies/:id
PUT    /api/v1/companies/:id
PATCH  /api/v1/companies/:id/status
DELETE /api/v1/companies/:id

POST   /api/v1/commute/calculate
GET    /api/v1/commute/results/:id
GET    /api/v1/commute/queries/:id

GET    /api/v1/history
POST   /api/v1/history
GET    /api/v1/history/:id
PATCH  /api/v1/history/:id
DELETE /api/v1/history/:id
POST   /api/v1/history/:id/restore

POST   /api/v1/export/excel
POST   /api/v1/export/pdf

GET    /api/v1/map/geocode
GET    /api/v1/map/regeocode
GET    /api/v1/map/poi/search
GET    /api/v1/map/cities
```

---

## 8. 后续扩展

| 版本 | 新增 |
|------|------|
| v1.1 | 多家庭 + 历史完整：无新增端点（已预留） |
| v1.2 | 对比视图：`POST /commute/compare` 聚合接口 |
| v1.3 | 导出已包含 |
| v2.0 | 鉴权：登录/注册端点 + 中间件 |

---

## 附录 A：请求示例（curl）

### 设置 profile
```bash
curl -X PUT http://localhost:8080/api/v1/profile \
  -H "Content-Type: application/json" \
  -d '{
    "current_city": "北京",
    "target_position": "后台开发",
    "experience_years": 5
  }'
```

### AI 推荐
```bash
curl -X POST http://localhost:8080/api/v1/ai/recommend/companies \
  -H "Content-Type: application/json" \
  -d '{
    "city": "北京",
    "position": "后台开发"
  }'
```

### 通勤计算
```bash
curl -X POST http://localhost:8080/api/v1/commute/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "home_id": 1,
    "company_ids": [10, 11],
    "transport_modes": ["transit", "driving"]
  }'
```
