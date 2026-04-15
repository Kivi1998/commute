# 外部 API 调研：高德 & 豆包

| 字段 | 内容 |
|------|------|
| 创建日期 | 2026-04-15 |
| 作者 | Hao Jia |
| 关联 | [API v1.0](./API-v1.0.md) / [ADR-001](../decisions/ADR-001-数据库设计.md) |

> 用于后端对接外部服务的字段映射与调用契约。

---

## 一、高德地图 Web Service API

### 1.1 认证与通用

- **主域名**：`https://restapi.amap.com`
- **认证**：所有请求必须携带 `key=<WEB_SERVICE_KEY>` 参数（URL Query）
- **签名（可选）**：高安全场景可加 `sig`（MD5 签名）
- **返回格式**：JSON（默认）
- **坐标系**：GCJ-02（所有输入输出）
- **坐标格式**：`"经度,纬度"`，小数点后 **≤ 6 位**，如 `116.481488,39.990464`

### 1.2 通用响应结构

```json
{
  "status": "1",         // 1=成功，0=失败
  "info": "OK",          // 错误描述
  "infocode": "10000",   // 详细状态码
  "count": "1",
  "route": { ... }       // 不同接口字段不同
}
```

**关键状态码**：
| infocode | 含义 | 应对 |
|----------|------|------|
| 10000 | 成功 | - |
| 10001 | Key 不正确或过期 | 检查 .env |
| 10003 | ACCESS_OVERFLOW（超配额） | 降级 + 告警 |
| 10021 | IP 限制 | 检查白名单 |
| 20000 | 参数错误 | 修正参数 |

**错误码映射（我方）**：
- `10000` → 正常返回
- `10001 / 10021` → `50201`（配置错误，我方 5xx）
- `10003` → `50201` + 特殊日志标记
- `20000` → `40001`（我方参数透传问题）
- 其他 → `50201`

---

### 1.3 驾车路径规划 v5

**Endpoint**：`GET https://restapi.amap.com/v5/direction/driving`

#### 请求参数
| 参数 | 必填 | 示例 | 说明 |
|------|------|------|------|
| `key` | 是 | `xxx` | API Key |
| `origin` | 是 | `116.481488,39.990464` | 起点经纬度 |
| `destination` | 是 | `116.434446,39.90816` | 终点经纬度 |
| `strategy` | 否 | `32` | 算路策略，默认 32（高德推荐） |
| `waypoints` | 否 | - | 途经点 |
| `departure_time` | 否 | `1713139200` | 预计出发时间戳（秒） |
| `cartype` | 否 | `0` | 0=普通车 1=纯电 2=插混 |
| `show_fields` | 否 | `cost,tmcs,navi,polyline` | **建议带上**，否则响应简略 |

**strategy 策略取值**：
| 值 | 含义 |
|----|------|
| 32 | 默认推荐 ✅ |
| 33 | 躲避拥堵 |
| 34 | 高速优先 |
| 35 | 不走高速 |
| 36 | 少收费 |
| 38 | 速度最快 |
| 43 | 躲避拥堵+少收费+不走高速 |

**我方默认**：`strategy=32`。

#### 响应结构（核心字段）
```json
{
  "status": "1",
  "info": "OK",
  "route": {
    "origin": "116.481488,39.990464",
    "destination": "116.434446,39.90816",
    "paths": [
      {
        "distance": "12345",           // 米
        "restriction": "0",
        "steps": [
          {
            "instruction": "沿建国路向西行驶 500 米",
            "road_name": "建国路",
            "orientation": "西",
            "step_distance": "500"
          }
        ],
        "cost": {
          "duration": "2100",          // 秒
          "tolls": "0",                // 元
          "toll_distance": "0",        // 米
          "toll_road": "",
          "taxi_fee": "35.00",
          "traffic_lights": "12"
        }
      }
    ]
  }
}
```

#### 字段映射（→ 我方 commute_result）
| 高德字段 | 我方字段 | 转换 |
|---------|---------|------|
| `paths[0].cost.duration` (秒) | `duration_min_raw` | `⌈seconds / 60⌉` |
| `paths[0].distance` (米) | `distance_km` | `meters / 1000`, 2 位小数 |
| `paths[0].cost.taxi_fee` | `cost_yuan` | 直接取值 |
| - | `transfer_count` | 驾车=NULL |
| `paths[0]` 全部 | `route_detail` | JSON 存储 |

---

### 1.4 公交路径规划 v5

**Endpoint**：`GET https://restapi.amap.com/v5/direction/transit/integrated`

#### 请求参数
| 参数 | 必填 | 示例 | 说明 |
|------|------|------|------|
| `key` | 是 | | |
| `origin` | 是 | `116.48,39.99` | |
| `destination` | 是 | `116.43,39.90` | |
| `city1` | 是 | `010` | 起点 citycode |
| `city2` | 是 | `010` | 终点 citycode（跨城时不同） |
| `strategy` | 否 | `0` | 换乘策略 |
| `date` | 否 | `2026-04-20` | 出发日期（YYYY-MM-DD） |
| `time` | 否 | `08:00` | 出发时间（HH:MM） |
| `max_trans` | 否 | `4` | 最大换乘次数 |
| `nightflag` | 否 | `0` | 是否考虑夜班车 |
| `AlternativeRoute` | 否 | `3` | 返回方案数（1-10） |
| `show_fields` | 否 | `cost,navi,polyline` | **建议带上** |

**strategy**：
| 值 | 含义 |
|----|------|
| 0 | 推荐 ✅ |
| 1 | 最经济 |
| 2 | 最少换乘 |
| 3 | 最少步行 |
| 5 | 不乘地铁 |

**我方默认**：`strategy=0`, `date=下周一`, `time=08:00`。

#### 响应结构
```json
{
  "status": "1",
  "route": {
    "transits": [
      {
        "cost": {
          "duration": "3420",         // 秒（总耗时）
          "transit_fee": "6.00"       // 元
        },
        "distance": "15200",          // 米
        "walking_distance": "1200",   // 米
        "nightflag": "0",
        "segments": [
          {
            "walking": {
              "origin": "...",
              "destination": "...",
              "distance": "400",
              "cost": { "duration": "300" },
              "steps": [...]
            }
          },
          {
            "bus": {
              "buslines": [
                {
                  "name": "10 号线(巴沟--劲松)",
                  "type": "地铁线路",
                  "departure_stop": { "name": "国贸", ... },
                  "arrival_stop": { "name": "海淀黄庄", ... },
                  "via_num": "8",
                  "distance": "12000",
                  "cost": { "duration": "2100" }
                }
              ]
            }
          },
          {
            "walking": { ... }
          }
        ]
      }
    ]
  }
}
```

#### 字段映射（→ 我方 commute_result）
| 高德字段 | 我方字段 | 转换 |
|---------|---------|------|
| `transits[0].cost.duration` | `duration_min_raw` | `⌈seconds / 60⌉` |
| `transits[0].distance` | `distance_km` | `meters / 1000` |
| `transits[0].cost.transit_fee` | `cost_yuan` | 直接取值 |
| `transits[0].segments[].bus.buslines.length` | `transfer_count` | 数组长度（地铁/公交段计数） |
| `transits[0]` 全部 | `route_detail` | JSON 存储 |

**citycode 获取**：用户画像的 `current_city_code`，或通过 `/v3/config/district` 查询。北京=`010`，上海=`021`，广州=`020`，深圳=`0755`。

---

### 1.5 骑行路径规划 v5

**Endpoint**：`GET https://restapi.amap.com/v5/direction/bicycling`

**请求**：`key`, `origin`, `destination`, `show_fields=cost,navi`

**响应简化版**：
```json
{
  "data": {
    "paths": [
      {
        "distance": "8500",
        "duration": "1800",
        "steps": [...]
      }
    ]
  }
}
```

注意：**骑行字段在 `data.paths` 而非 `route.paths`**（与驾车不同）。

#### 字段映射
| 高德字段 | 我方字段 |
|---------|---------|
| `data.paths[0].duration` | `duration_min_raw` |
| `data.paths[0].distance` | `distance_km` |
| `cost_yuan` | NULL |
| `transfer_count` | NULL |

---

### 1.6 步行路径规划 v5

**Endpoint**：`GET https://restapi.amap.com/v5/direction/walking`

结构与骑行类似（`route.paths[0]`）。适用距离 ≤ 100km，我方实际使用场景 < 10km。

字段映射同骑行。

---

### 1.7 地理编码 / 逆地理编码

#### 地理编码（地址 → 坐标）
**Endpoint**：`GET https://restapi.amap.com/v3/geocode/geo`
```
?key=xxx&address=北京市朝阳区建国路88号&city=010
```
**响应**：`geocodes[0].location = "116.48,39.99"`（字符串）

#### 逆地理编码（坐标 → 地址）
**Endpoint**：`GET https://restapi.amap.com/v3/geocode/regeo`
```
?key=xxx&location=116.48,39.99
```

### 1.8 POI 搜索

**Endpoint**：`GET https://restapi.amap.com/v5/place/text`
```
?key=xxx&keywords=字节跳动&region=北京&page_size=10
```
**响应**：`pois[].location`, `pois[].name`, `pois[].address`, `pois[].id`

---

### 1.9 高德最佳实践（我方约束）

1. **所有 Web Service 请求走后端**，Key 从未暴露到前端
2. **前端 JS SDK Key 单独申请**，域名白名单限制
3. **请求超时**：5s 连接 + 10s 读取
4. **重试**：幂等请求（GET）失败时指数退避重试 2 次
5. **show_fields**：所有路径接口显式指定，避免默认返回过大
6. **departure_time**：周一早 8:00 换算为时间戳；晚 17:30 同理（用下周一的日期避免历史路况为空）
7. **并发控制**：每用户单次通勤计算并发 ≤ 5，全局并发 ≤ 50
8. **配额监控**：每小时采样 `infocode=10003` 计数，超阈值告警

---

## 二、豆包 API（火山方舟）

### 2.1 认证与通用

- **主域名**：`https://ark.cn-beijing.volces.com`
- **认证**：`Authorization: Bearer <ARK_API_KEY>`
- **API Key 获取**：火山引擎控制台 → 方舟大模型 → API Key 管理
- **计费**：按 token 计费（input/output 分别计价）

### 2.2 模型选型

| 模型 | 上下文 | 适用场景 | 结构化输出 |
|------|-------|---------|-----------|
| `doubao-seed-1.6` | 256K | 综合推荐 ✅ | ✅ JSON Object + JSON Schema |
| `doubao-seed-1.6-flash` | 256K | 速度优先 | ✅ |
| `doubao-seed-1.6-thinking` | 256K | 复杂推理 | ✅ |
| `doubao-1.5-pro-32k` | 32K | 性价比 | 仅 json_object |
| `doubao-1.5-lite-32k` | 32K | 轻量 | 仅 json_object |

**我方选型**：
- **生产**：`doubao-seed-1.6-flash`（公司推荐场景对推理深度要求不高，响应速度优先）
- **开发**：`doubao-1.5-lite-32k`（成本最低，便于调试）

> **注意**：实际调用时使用**接入点 ID（endpoint id）**，形如 `ep-20260415xxxxxx-xxxxx`，由控制台创建接入点时生成。建议配置到 `.env` 避免硬编码模型名。

### 2.3 Chat Completions API

**Endpoint**：`POST https://ark.cn-beijing.volces.com/api/v3/chat/completions`

#### 请求 Body
```json
{
  "model": "ep-20260415xxxxxx-xxxxx",
  "messages": [
    { "role": "system", "content": "..." },
    { "role": "user",   "content": "..." }
  ],
  "temperature": 0.7,
  "top_p": 0.9,
  "max_tokens": 4096,
  "stream": false,
  "response_format": {
    "type": "json_object"
  }
}
```

#### 响应结构
```json
{
  "id": "chat-xxxxx",
  "object": "chat.completion",
  "created": 1713139200,
  "model": "doubao-seed-1.6-flash",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "{\"companies\": [...]}"
      },
      "finish_reason": "stop"
    }
  ],
  "usage": {
    "prompt_tokens": 350,
    "completion_tokens": 1200,
    "total_tokens": 1550,
    "prompt_tokens_details": { "cached_tokens": 0 }
  }
}
```

#### 字段映射（→ 我方 ai_recommendation_cache）
| 豆包字段 | 我方字段 |
|---------|---------|
| `choices[0].message.content` (JSON 字符串) | `raw_response` (解析后存 JSONB) |
| `usage.prompt_tokens` | `token_input` |
| `usage.completion_tokens` | `token_output` |

---

### 2.4 结构化输出（JSON Mode）

#### 方案 A：json_object（宽松模式）

```json
{
  "response_format": { "type": "json_object" }
}
```

**关键约束**：**prompt 中必须包含 "JSON" 关键字**（大小写不敏感），否则报错。

#### 方案 B：json_schema（严格模式，仅 Seed 1.6+）

```json
{
  "response_format": {
    "type": "json_schema",
    "json_schema": {
      "name": "company_list",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": {
          "companies": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "name":     { "type": "string" },
                "category": { "type": "string", "enum": ["big_tech","mid_tech","startup","foreign","other"] },
                "industry": { "type": "string" },
                "address_hint": { "type": "string" },
                "reason":   { "type": "string" }
              },
              "required": ["name", "category", "industry", "address_hint", "reason"],
              "additionalProperties": false
            }
          }
        },
        "required": ["companies"],
        "additionalProperties": false
      }
    }
  }
}
```

**我方选择**：**方案 B（json_schema + strict: true）**
- 优势：模型必定返回符合 schema 的结构，无需再做格式校验与 retry
- 注意：仅 `doubao-seed-1.6` 系列支持，需在部署接入点时选对模型

---

### 2.5 Prompt 模板（AI 公司推荐）

#### System Prompt
```
你是一位专业的互联网行业分析师，熟悉国内各城市的科技公司分布。

请根据用户提供的城市和求职岗位，推荐本地真实存在的公司列表。要求：
1. 公司必须真实存在，且在目标城市有办公地点。
2. 均衡覆盖大厂、中厂、创业公司三个层级。
3. 每家公司给出真实办公地点的简要描述（区/街道级别）。
4. 推荐理由需结合公司业务与目标岗位的契合度，控制在 50 字以内。
5. 严格输出 JSON，不要包含任何解释性文字。
6. 禁止虚构不存在的公司名称。
```

#### User Prompt 模板
```
城市：{{city}}
求职岗位：{{position}}
{{#if experience_years}}工作经验：{{experience_years}} 年{{/if}}
{{#if company_types}}偏好公司类型：{{company_types | 中文化逗号分隔}}{{/if}}

请推荐 {{count}} 家符合条件的公司，按 JSON 格式返回：

{
  "companies": [
    {
      "name": "公司名",
      "category": "big_tech | mid_tech | startup | foreign | other",
      "industry": "所属行业",
      "address_hint": "办公地点的区/街道级别描述",
      "reason": "推荐理由（50字内）"
    }
  ]
}
```

#### category 对应中文
| 枚举 | 中文 |
|------|------|
| big_tech | 大厂 |
| mid_tech | 中厂 |
| startup | 创业公司 |
| foreign | 外企 |
| other | 其他 |

#### 参数建议
| 参数 | 值 | 说明 |
|------|-----|------|
| `temperature` | `0.7` | 保证推荐多样性 |
| `top_p` | `0.9` | |
| `max_tokens` | `4096` | 足够 30 家公司 |

#### 示例 curl
```bash
curl -X POST https://ark.cn-beijing.volces.com/api/v3/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ARK_API_KEY" \
  -d '{
    "model": "ep-xxx",
    "messages": [
      {"role": "system", "content": "你是一位专业的互联网行业分析师..."},
      {"role": "user", "content": "城市：北京\n求职岗位：后台开发\n请推荐 20 家公司，按 JSON 返回 {\"companies\": [...]}"}
    ],
    "temperature": 0.7,
    "max_tokens": 4096,
    "response_format": {
      "type": "json_schema",
      "json_schema": { ... }
    }
  }'
```

---

### 2.6 豆包错误处理

| HTTP | 错误类型 | 我方处理 |
|------|---------|---------|
| 400 | InvalidRequest | → `40002`（参数层 bug） |
| 401 | AuthenticationError | → `50203` + 配置告警 |
| 429 | RateLimitExceeded | → `50203` + 退避重试 |
| 500/503 | ServerError | → `50203` + 指数退避重试 2 次 |

**Schema 校验失败兜底**：即使使用 json_schema，仍需在后端用 Go 的 `json.Unmarshal` 二次校验，失败则 → `50204`。

---

### 2.7 豆包最佳实践（我方约束）

1. **API Key 仅存后端 `.env`**，前端禁止直接调用
2. **推荐使用接入点（endpoint id）**，而非写死模型名
3. **超时**：10s 连接 + 60s 读取（LLM 可能慢）
4. **Token 追踪**：每次调用写入 `ai_recommendation_cache.token_input/output`，便于成本分析
5. **缓存优先**：24h 内同输入复用缓存
6. **地址二次校验**：AI 返回的 `address_hint` 不可直接使用，需调高德 POI 搜索确认精确坐标

---

## 三、字段映射速查表（外部 → 内部）

### 通勤计算流程
```
用户请求 commute_calculate
  → 后端拿到 home + companies + modes
  → 对每个 [home × company × mode × direction]：
     1. 查 commute_result 缓存
     2. 未命中则调高德对应 API
     3. 字段转换（见下表）
     4. + 5 min buffer
     5. 写入 commute_result
```

### 统一转换规则
| 我方字段 | 驾车源字段 | 公交源字段 | 骑行/步行源字段 |
|---------|-----------|-----------|----------------|
| `duration_min_raw` | `route.paths[0].cost.duration ÷ 60` | `route.transits[0].cost.duration ÷ 60` | `data.paths[0].duration ÷ 60` |
| `duration_min` | `duration_min_raw + 5` | 同 | 同 |
| `distance_km` | `route.paths[0].distance ÷ 1000` | `route.transits[0].distance ÷ 1000` | `data.paths[0].distance ÷ 1000` |
| `cost_yuan` | `cost.taxi_fee` 或 `cost.tolls` | `cost.transit_fee` | NULL |
| `transfer_count` | NULL | 统计 `segments[].bus.buslines` 数量 | NULL |
| `route_detail` | `paths[0]` 整体 JSON | `transits[0]` 整体 JSON | `paths[0]` 整体 JSON |

### AI 推荐流程
```
用户请求 recommend
  → 构造 cache_key（md5 哈希）
  → 查 ai_recommendation_cache
     命中 → 直接返回
     未命中：
       1. 构造 messages（system + user prompt）
       2. 调豆包 chat/completions（response_format=json_schema）
       3. 解析 choices[0].message.content
       4. 对每家公司调高德 POI 搜索确认坐标
       5. 写入 ai_recommendation_cache
       6. 返回前端（不直接入库为公司，由用户勾选后批量入库）
```

---

## 四、环境变量规范

```bash
# .env.example

# 高德 Web Service（后端）
AMAP_WS_KEY=xxxxxxxxxxxxxxxxxxxxxxxx
AMAP_WS_BASE=https://restapi.amap.com
AMAP_TIMEOUT_MS=10000

# 高德 JS SDK（前端，公开可见）
VITE_AMAP_JS_KEY=xxxxxxxxxxxxxxxxxxxxxxxx
VITE_AMAP_JS_SECURITY=xxxxxxxxxxxxxxxxxxxxxxxx   # JS 安全密钥

# 豆包
DOUBAO_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxx
DOUBAO_BASE=https://ark.cn-beijing.volces.com
DOUBAO_MODEL=ep-20260415xxxxxx-xxxxx         # 接入点 ID
DOUBAO_TIMEOUT_MS=60000
```

---

## 五、TODO / 进一步验证

- [ ] 实际调用高德驾车接口，确认 `show_fields=cost,navi` 的返回结构与文档一致
- [ ] 实际调用高德公交接口，验证 `transits[].segments[].bus.buslines` 的换乘计数逻辑
- [ ] 申请豆包接入点，实测 `doubao-seed-1.6-flash` + `json_schema` 的稳定性（是否永远返回合法 JSON）
- [ ] 实测豆包推荐的"address_hint" 走高德 POI 搜索的命中率
- [ ] 收集各城市 citycode 清单（北京=010、上海=021、广州=020、深圳=0755 等）

---

## 附录：参考资料

- 高德路径规划 2.0 文档: https://lbs.amap.com/api/webservice/guide/api/newroute
- 高德驾车 API: https://amap.apifox.cn/api-14580571
- 高德公交 API: https://amap.apifox.cn/api-14610908
- 豆包 Chat Completions: https://www.volcengine.com/docs/82379/1494384
- 豆包结构化输出: https://www.volcengine.com/docs/82379/1568221
- 火山方舟模型列表: https://www.volcengine.com/docs/82379/1330310
