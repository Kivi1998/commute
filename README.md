# Commute · 通勤查询系统

> 面向**求职者**的通勤成本可视化工具。输入家庭住址和目标公司，基于真实地图数据可视化对比各家通勤距离、时间、费用，辅助求职与租房决策。

![Stack](https://img.shields.io/badge/stack-Vue3%20%7C%20Go%20%7C%20PostgreSQL-blue)
![Map](https://img.shields.io/badge/map-AMap-green)
![AI](https://img.shields.io/badge/AI-Doubao-orange)

## ✨ 核心功能

- 🏠 **多家庭住址**：地图点选保存多个候选住址，支持租房选址对比
- 🏢 **公司管理**：手动添加 / **AI 推荐**（豆包）一键入库，投递状态追踪
- 🚇 **通勤可视化**：基于高德真实数据计算 4 种出行方式（公交/驾车/骑行/步行）的早晚高峰通勤
- 🗺️ **地图总览**：家与公司位置一张图，按耗时自动着色
- 📜 **查询历史**：每次计算自动保存，一键恢复参数
- ⚡ **缓存优化**：通勤计算 7 天缓存，AI 推荐 24 小时缓存，大幅降低 API 成本

## 🛠 技术栈

| 层 | 技术 |
|----|------|
| 前端 | Vue 3 + TypeScript + Vite + Ant Design Vue + Tailwind CSS + Pinia + Vue Router + VueUse |
| 后端 | Go 1.26 + Gin + pgx/v5 |
| 数据库 | PostgreSQL 17 |
| 地图 | 高德 Web Service API（后端代理）+ 高德 JS SDK（前端渲染） |
| AI | 豆包（火山方舟 Ark） |

## 📁 目录结构

```
Commute/
├── backend/                # Go 后端
│   ├── cmd/server/         # 入口
│   ├── internal/
│   │   ├── config/         # godotenv 配置
│   │   ├── database/       # pgxpool
│   │   ├── handler/        # HTTP 层
│   │   ├── service/        # 业务逻辑
│   │   ├── repository/     # 数据访问
│   │   ├── model/          # 类型定义
│   │   ├── middleware/     # request_id / recovery
│   │   ├── router/         # 路由注册
│   │   └── pkg/
│   │       ├── amap/       # 高德 Client
│   │       ├── doubao/     # 豆包 Client
│   │       └── citydict/   # 城市 adcode 字典
│   ├── migrations/         # SQL 迁移
│   └── pkg/response/       # 统一响应格式
├── frontend/               # Vue 3 前端
│   └── src/
│       ├── api/            # axios 封装 + 各模块 API
│       ├── components/     # 可复用组件（AmapPicker / CommuteMap / AIRecommendDialog ...）
│       ├── layouts/        # 应用布局
│       ├── pages/          # 路由页面
│       ├── router/
│       ├── store/          # Pinia
│       └── lib/amap.ts     # 高德 JS SDK 单例 loader
├── docs/
│   ├── prd/                # 产品需求文档
│   ├── api/                # API 接口文档 + 外部 API 调研
│   └── decisions/          # 架构决策记录 (ADR)
├── .env.example            # 环境变量模板
└── CLAUDE.md               # 项目协作规则
```

## 🚀 快速开始

### 前置要求
- Node.js 20+
- Go 1.22+
- PostgreSQL 15+
- pnpm

### 1. 克隆 & 配置

```bash
git clone https://github.com/Kivi1998/commute.git
cd commute

# 复制环境变量模板
cp .env.example backend/.env
# 编辑 backend/.env 填入你的数据库与 API 密钥

# 前端环境变量
cat > frontend/.env <<EOF
VITE_AMAP_JS_KEY=你的高德_JS_Key
VITE_AMAP_JS_SECURITY=你的高德_JS_安全密钥
EOF
```

### 2. 获取 API 密钥

- **高德**：[lbs.amap.com](https://lbs.amap.com) → 实名认证 → 创建「Web 服务」+「Web 端 JS API」两个 Key
- **豆包**：[volcengine.com](https://console.volcengine.com) → 火山方舟 → 创建 API Key + 推理接入点

### 3. 初始化数据库

```bash
# 创建 commute 数据库
PGPASSWORD=postgres psql -h 127.0.0.1 -U postgres -d postgres \
  -c "CREATE DATABASE commute WITH ENCODING 'UTF8';"

# 执行迁移
cd backend
psql -h 127.0.0.1 -U postgres -d commute -f migrations/0001_init_schema.up.sql
psql -h 127.0.0.1 -U postgres -d commute -f migrations/0002_seed_default_user.up.sql
```

### 4. 启动服务

**推荐：一键启停脚本**

```bash
./start.sh          # 等价于 ./dev.sh start
./stop.sh           # 等价于 ./dev.sh stop
./dev.sh status     # 查看状态
./dev.sh logs       # 跟踪两边日志
./dev.sh restart    # 重启
```

脚本会：前置检查环境 → 编译后端 → 后台启动两端 → 等待就绪 → 返回 URL。
日志落在 `.dev/logs/`，PID 落在 `.dev/pid/`（都在 `.gitignore` 里）。

**手动启动（可选）**

```bash
cd backend && go run ./cmd/server      # 后端 :8090
cd frontend && pnpm install && pnpm dev # 前端 :5173
```

打开浏览器访问 http://localhost:5173

## 📋 使用流程

1. **设置** → 填个人画像 + 地图点选家庭住址
2. **公司** → AI 推荐 / 手动添加公司
3. **通勤** → 选家、公司、出行方式 → 查看地图总览和时长排序
4. **历史** → 查看过往查询，一键恢复参数重算

## 🎨 截图

（TODO：填入生产界面截图）

## 🔑 API 接口

完整接口文档见 [docs/api/API-v1.0.md](docs/api/API-v1.0.md)。

核心端点：
- `POST /api/v1/commute/calculate` — 批量通勤计算
- `POST /api/v1/ai/recommend/companies` — AI 公司推荐
- `GET /api/v1/map/{geocode,regeocode,poi/search}` — 高德代理

## 🧱 关键设计决策

- [数据库设计 ADR-001](docs/decisions/ADR-001-数据库设计.md)
- [PRD v1.0](docs/prd/PRD-v1.0.md)
- [高德 & 豆包 API 调研](docs/api/external-API-research.md)

## 📜 License

MIT
