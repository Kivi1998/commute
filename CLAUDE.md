# Commute - 通勤查询 Web App

## 项目简介

一个面向**求职者**的通勤查询工具：输入家庭住址和目标公司，可视化展示通勤距离、时间，辅助求职决策。

## 技术栈

| 层级 | 技术 |
|------|------|
| 前端 | **Vue 3** + **Ant Design Vue** + TailwindCSS + Vite + TypeScript |
| 状态/路由 | Pinia + Vue Router 4 + VueUse |
| 后端 | Golang (Gin) |
| 数据库 | PostgreSQL |
| 地图 | 高德地图开放平台 API |
| AI | 豆包（字节火山引擎） |
| 部署 | Docker / Docker Compose（本地运行） |
| 平台 | Web + 移动端响应式（后续可能扩展小程序） |

## 目录结构（规划）

```
Commute/
├── CLAUDE.md                    # 项目协作规则（本文件）
├── README.md                    # 项目说明
├── docker-compose.yml           # 本地一键启动
├── backend/                     # Go 后端
│   ├── cmd/
│   ├── internal/
│   │   ├── handler/             # HTTP handler
│   │   ├── service/             # 业务逻辑
│   │   ├── repository/          # 数据访问
│   │   ├── model/               # 数据模型
│   │   └── pkg/
│   │       ├── amap/            # 高德 API 封装
│   │       └── doubao/          # 豆包 AI 封装
│   ├── migrations/              # 数据库迁移
│   ├── go.mod
│   └── Dockerfile
├── frontend/                    # React 前端
│   ├── src/
│   │   ├── components/
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── api/
│   │   └── lib/
│   ├── package.json
│   └── Dockerfile
├── docs/
│   ├── conversations/           # 对话记录（按日期归档）
│   ├── prd/                     # 产品需求文档
│   ├── api/                     # 接口文档
│   └── decisions/               # 关键决策记录 (ADR)
└── .env.example                 # 环境变量示例
```

## 核心功能模块

### 1. 个人信息（暂无登录，预留扩展点）
- 家庭住址（支持多个，可对比选址）
- 所在城市
- 工作岗位

### 2. AI 公司推荐
- 输入"城市 + 岗位" → 豆包 AI 返回公司列表
- 用户勾选关注的公司

### 3. 通勤可视化
- **出行方式**：用户可选（公交/地铁/驾车/骑行/步行），多选时同时展示
- **时间策略**：
  - 默认：周一早 8:00 出发 / 17:30 返程
  - 支持自定义：出发时间 → 推断到达 / 到达时间 → 推断出发
- **路线详情**：换乘、距离、耗时
- 地图标注 + 列表展示

### 4. 数据持久化
- 历史通勤查询记录
- 多公司对比视图（图表）
- 导出 Excel / PDF

## Skill / Subagent 调用指引（强制）

> ⚠️ **凡是下表中能匹配到的场景，必须优先用对应的 skill / subagent，而不是纯手写**。
> 只有 skill 产出无法满足需求（如已经存在高度定制内容），才手工补齐。
> AI 在每个回合开始前应先判断："这个场景能不能用 skill？"——能则主动调用。

### 一、按阶段匹配 Skill

| 研发阶段 | 场景 | 应调用的 skill |
|---------|------|---------------|
| **需求** | 拿到一份 PRD/需求文档要解析 | `/parse-requirements` |
| 需求 | 把功能点转成开发任务清单 | `/generate-tasks` |
| 需求 | 协作撰写 PRD / 设计文档 / ADR | `/doc-coauthoring` |
| **设计** | 设计新 API 接口文档 | `/api-design` |
| 设计 | 设计数据库表结构 / 生成建表 SQL | `/db-schema` |
| **开发** | 生成 Spring Boot / Go Controller-Service-Mapper 骨架 | `/be-service` |
| 开发 | 生成 Vue 3 组件（AntDV / Element Plus） | `/fe-component` |
| 开发 | 创建高设计感的前端页面 / 组件库 | `/frontend-design` |
| 开发 | 对接 Claude API / Anthropic SDK | `/claude-api` |
| **部署** | 生成 Dockerfile / docker-compose / Nginx / CI | `/deploy` |
| **代码检视** | 刚写完一段代码，优化可读性与一致性 | `/simplify` |
| 代码检视 | PR 完整审查（多视角） | `/review-pr` |
| 代码检视 | 安全视角审查 | `/security-review` |
| 代码检视 | 单独审查代码 | `/review` |
| **运维** | 配置 Claude Code 行为（hooks / 权限 / env） | `/update-config` |
| 运维 | 自定义键位 | `/keybindings-help` |
| 运维 | 定时/循环任务 | `/schedule` 或 `/loop` |
| **办公** | 处理 .docx / .pptx / .xlsx / .pdf | `/docx` `/pptx` `/xlsx` `/pdf` |
| **Jira/Confluence** | 基于会议纪要建 Jira 任务 | `/atlassian:capture-tasks-from-meeting-notes` |
| Jira/Confluence | 搜索公司内部知识库 | `/atlassian:search-company-knowledge` |
| Jira/Confluence | Confluence 规格 → Jira 积压 | `/atlassian:spec-to-backlog` |
| Jira/Confluence | 生成项目状态报告 | `/atlassian:generate-status-report` |
| Jira/Confluence | Bug 去重 / 工单分诊 | `/atlassian:triage-issue` |
| **元** | 创建/修改/优化 skill 本身 | `/skill-creator` |

### 二、按场景匹配 Subagent

| 场景 | 应调用的 subagent |
|------|-----------------|
| **探索代码库**（批量搜索、多轮 grep/glob，超过 3 次查询时） | `Explore` |
| **制定实现方案**（设计实施步骤、权衡架构选项） | `Plan` |
| 开放性多步任务（研究 + 代码 + 验证） | `general-purpose` |
| Claude Code 使用问题（hooks / slash / MCP） | `claude-code-guide` |
| **PR 测试覆盖分析** | `pr-review-toolkit:pr-test-analyzer` |
| **代码简化重构** | `pr-review-toolkit:code-simplifier` |
| **注释准确性与技术债** | `pr-review-toolkit:comment-analyzer` |
| **代码规范审查**（对齐 CLAUDE.md 风格） | `pr-review-toolkit:code-reviewer` |
| **静默失败 / catch 黑洞检测** | `pr-review-toolkit:silent-failure-hunter` |
| **新类型设计审查**（封装性/不变量） | `pr-review-toolkit:type-design-analyzer` |

### 三、本项目强制规则

1. **新增业务表** → 先 `/db-schema` 生成 DDL，再手工调整，不直接手写
2. **新增后端端点** → 先 `/api-design` 生成 API 文档，再 `/be-service` 生成骨架
3. **新增前端页面/组件** → 用 `/fe-component`（AntDV）生成骨架
4. **PR 提交前** → 依次 `/simplify` → `/review-pr`
5. **代码库跨多目录搜索** → 派 `Explore` subagent，不在主对话里 grep
6. **给出实施方案前** → 派 `Plan` subagent，不在主对话里空想
7. 主对话中如果已经用 skill 产出，**不要重复手写同一份内容**

### 四、官方 subagent 情况说明

Anthropic **没有**官方维护的"subagent 应用市场 + 预装清单"。现状：

- **内置（无需安装）**：`Explore`、`Plan`、`general-purpose`、`statusline-setup`、`claude-code-guide`
- **官方 plugin 自带**：Anthropic 发布的 `pr-review-toolkit` 插件带 6 个审查类 subagent（已装）
- **社区主流集合**：[wshobson/agents](https://github.com/wshobson/agents)（182 个 agents / 77 个 plugins，覆盖语言/架构/运维/安全），可按需 `/plugin install` 引入
- 官方文档：[Claude Code Sub-agents](https://docs.claude.com/en/docs/claude-code/sub-agents)

> **结论**：除内置 + pr-review-toolkit 外，如需更多（如 vue-developer / go-pro / postgres-pro），走 wshobson/agents 社区插件。

## 协作规则（与 AI 协作时遵守）

### 语言
- **所有交互、注释、文档、Commit Message 使用中文**
- 代码标识符（变量、函数名）使用英文

### 编码规范
- **Go**：遵循官方规范 + `golangci-lint`，error 不能吞，函数控制在 50 行内
- **React**：函数组件 + Hooks，TypeScript 严格模式，组件单文件 < 300 行
- **数据库**：表名/字段名 `snake_case`，每张表必含 `id, created_at, updated_at`
- **API**：RESTful 风格，统一响应格式 `{ code, message, data }`

### 安全
- API Key（高德、豆包）必须放在 `.env`，禁止硬编码
- `.env` 加入 `.gitignore`，仓库提供 `.env.example`

### 工作流（强制规则）
- ⚠️ **每一次对话结束都必须保存对话记录到 `docs/conversations/`**，无一例外
  - **归档粒度：每日单文件**，文件名 `YYYY-MM-DD.md`
  - 当日首次对话：创建文件，写入 `# Session 01 — 主题`
  - 当日后续对话：**读取**当日文件后**追加** `# Session 02`、`# Session 03` ...
  - **禁止**为同一天创建多个文件
  - 每个 Session 必须包含：决策摘要、关键问答、产出物、原文精华
  - 必须同步更新 `docs/conversations/INDEX.md` 索引
- 重大架构决策额外写入 `docs/decisions/ADR-NNN-标题.md`
- 需求变更同步更新本文件
- AI 协作时若用户未明确要求保存，AI 也必须主动在回合末尾保存对话记录

## 关键决策记录

| 日期 | 决策 | 理由 |
|------|------|------|
| 2026-04-15 | 暂不做登录系统 | 初期单人使用，但代码结构预留 user_id 字段 |
| 2026-04-15 | AI 选用豆包 | 国内访问稳定，中文场景效果好 |
| 2026-04-15 | 通勤时间只算周一 | 周一最堵，作为最坏情况估算 |
| 2026-04-15 | 公司数据由 AI 实时生成 | 不维护静态公司库，降低数据成本 |
| 2026-04-15 | AI 推荐结果缓存到数据库 | 节省豆包 token，相同输入 24h 内复用 |
| 2026-04-15 | 高德 API 采用混合方案 | 地图渲染用 JS SDK，路径计算/地理编码走后端 Web Service 保护 Key |
| 2026-04-15 | 通勤时间增加 5 分钟 buffer | 容错冗余，避免估算过于乐观 |
| 2026-04-15 | 单次勾选公司软上限 20 家 | 避免地图标注过密、AI 响应膨胀 |
| 2026-04-15 | 暂不做国际化 | 中文为唯一语言，英文版本未来再考虑 |

## 待开发清单

- [x] PRD 完整版 → [`docs/prd/PRD-v1.0.md`](docs/prd/PRD-v1.0.md)
- [x] 数据库 Schema 设计 → [`docs/decisions/ADR-001-数据库设计.md`](docs/decisions/ADR-001-数据库设计.md)
- [x] 数据库迁移文件 → `backend/migrations/`
- [x] 本地 PostgreSQL `commute` 库已创建并初始化
- [x] 后端 API 接口设计 → [`docs/api/API-v1.0.md`](docs/api/API-v1.0.md)
- [x] 高德 / 豆包 API 调研 → [`docs/api/external-API-research.md`](docs/api/external-API-research.md)
- [x] Go 后端脚手架（Gin + pgx/v5，/health 端点已跑通，端口 8090）
- [x] Vue 前端脚手架（Vue3 + Vite + AntDV + Tailwind + Pinia + Vue Router，端口 5173，已代理后端）
- [x] Settings 页完整闭环：个人画像 + 家庭住址 CRUD（/profile, /addresses, /meta/enums）
- [x] Companies 页完整闭环：7 个端点 + 筛选/分页/状态切换/批量（/companies）
- [x] 通勤计算后端完整实现：amap client + 4 种出行 + 7 天缓存 + 并发 5 + /commute + /map 代理
- [x] 前端 AmapPicker（地图点选回填地址+坐标+城市），升级 Address/Company FormModal
- [x] Commute 页完整闭环：Config + Map + Results（家/公司/模式/时间 → 地图标注 + 表格排序）
- [x] 豆包 AI 推荐接入：后端 Client + 24h 缓存 + POI 二次校验；前端 AIRecommendDialog + 批量入库
- [x] History 查询历史（M7）：列表 + 一键恢复参数 + 删除；Commute 页支持 ?from_query=N 恢复
- [ ] 后端 API 接口设计
- [ ] 高德 API 调用封装
- [ ] 豆包 AI Prompt 设计
- [ ] 前端页面原型
- [ ] Docker 部署配置

## 本地环境

### 数据库
```
Host:     127.0.0.1
Port:     5432
Database: commute
User:     postgres
Password: postgres
DSN:      postgres://postgres:postgres@127.0.0.1:5432/commute?sslmode=disable
```

### 后端服务
```
Port:     8090（8080 被其他 Java 进程占用）
Health:   http://127.0.0.1:8090/api/v1/health
启动：    cd backend && go run ./cmd/server
构建：    cd backend && go build -o bin/server ./cmd/server
```

### 前端服务
```
Port:      5173（Vite 默认）
Local:     http://localhost:5173/
代理：     /api → http://127.0.0.1:8090（vite.config.ts server.proxy）
启动 dev： cd frontend && pnpm dev
构建：     cd frontend && pnpm build
包管理器： pnpm 10.14.0（Node 20.19.4）
```

### 常用命令
```bash
# 连接数据库
PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d commute

# 应用迁移
PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d commute -v ON_ERROR_STOP=1 -f backend/migrations/XXXX.up.sql

# 回滚迁移
PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d commute -v ON_ERROR_STOP=1 -f backend/migrations/XXXX.down.sql
```

## 联系信息

- **开发者**: Hao Jia
- **邮箱**: jiahao@diit.cn
