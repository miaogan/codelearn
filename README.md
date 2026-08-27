# CodeLearn - 编程语言学习平台

一个类似多邻国的在线编程语言学习平台，支持技能树闯关、AI 习题生成、在线代码运行与评判。

## 技术栈

### 后端
- **Go 语言** + **Gin** (HTTP 框架)
- **Eino** (字节跳动 Go LLM 框架，用于 AI 习题生成)
- **GORM + SQLite** (ORM + 数据库)
- **JWT** 认证

### 前端
- **Vue 3 + TypeScript**
- **Vite** (构建工具)
- **Pinia** (状态管理)
- **Vue Router** (路由)

## 功能特性

- **技能树学习路径**: 课程 -> 单元 -> 课程节点，解锁式闯关
- **多题型练习**: 选择题、填空题、代码编写题
- **AI 习题生成**: 使用 Eino 框架调用 LLM 动态生成习题
- **在线代码运行**: 支持 Go 和 Python 代码执行与测试用例评判
- **游戏化系统**: XP 经验值、连续打卡 streak、心数/生命

## 项目结构

```
codelearn/
├── backend/              # Go 后端
│   ├── main.go           # 入口
│   ├── config/           # 配置
│   ├── model/            # 数据模型
│   ├── repository/       # 数据访问层
│   ├── service/          # 业务逻辑层
│   ├── handler/          # HTTP 处理器
│   ├── router/           # 路由配置
│   ├── middleware/       # JWT 认证中间件
│   ├── eino/             # Eino AI 习题生成
│   ├── sandbox/          # 代码运行沙箱
│   └── seed.go           # 种子数据
├── frontend/             # Vue3 前端
│   ├── src/
│   │   ├── views/        # 页面组件
│   │   ├── components/   # 习题组件
│   │   ├── stores/       # Pinia 状态
│   │   ├── api/          # API 客户端
│   │   └── types/        # TypeScript 类型
│   └── vite.config.ts
└── README.md
```

## 快速开始

### 1. 启动后端

```bash
cd backend
go mod tidy          # 下载依赖
go run .             # 启动服务 (默认端口 8080)
```

### 2. 启动前端 (开发模式)

```bash
cd frontend
npm install          # 安装依赖
npm run dev          # 开发服务器 (端口 3000, 代理 API 到 8080)
```

访问 http://localhost:3000 开始使用。

### 3. 生产模式 (后端直接服务前端)

```bash
cd frontend && npm run build   # 构建前端
cd ../backend && go run .      # 启动后端
```

访问 http://localhost:8080 即可。

## 配置 AI 习题生成

AI 习题生成使用 Eino 框架的 ChatModel 组件，支持任何 OpenAI 兼容的 LLM 服务。

设置环境变量启用:

```bash
# 豆包/Ark (推荐)
export LLM_API_KEY="your-ark-api-key"
export LLM_BASE_URL="https://ark.cn-beijing.volces.com/api/v3"
export LLM_MODEL="doubao-1-5-pro-32k-250115"

# 或 OpenAI
export LLM_API_KEY="your-openai-key"
export LLM_BASE_URL="https://api.openai.com/v1"
export LLM_MODEL="gpt-4o-mini"
```

未配置时平台正常运行，仅 AI 生成功能不可用。

## API 接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/auth/register | 注册 | 否 |
| POST | /api/auth/login | 登录 | 否 |
| GET | /api/courses | 课程列表 | 是 |
| GET | /api/courses/:id | 技能树(学习路径) | 是 |
| GET | /api/lessons/:id | 课程详情 | 是 |
| GET | /api/lessons/:id/exercises | 习题列表 | 是 |
| POST | /api/exercises/:id/submit | 提交答案 | 是 |
| POST | /api/lessons/:id/generate | AI 生成习题 | 是 |
| POST | /api/exercises/hint | AI 提示 | 是 |
| POST | /api/code/run | 运行代码 | 是 |
| POST | /api/code/judge | 评判代码 | 是 |
| POST | /api/lessons/:id/complete | 完成课程 | 是 |
| GET | /api/users/me/stats | 用户统计 | 是 |

## Eino 框架说明

本项目使用 Eino 框架的核心组件:

- **ChatModel**: 通过 `openai.NewChatModel()` 创建，调用 `Generate()` 方法与 LLM 交互
- **schema.Message**: 构造对话消息 (`schema.System` / `schema.User` 角色)
- **ChatModelConfig**: 配置 API Key、Base URL、模型名称

关键代码位于 `backend/eino/exercise_gen.go`，展示了如何使用 Eino 调用 LLM 生成结构化习题。

## 课程内容

内置两门课程:
- **Go 语言入门**: 3 个单元 (基础语法/控制流与函数/数据结构), 8 课, 16 道习题
- **Python 编程入门**: 3 个单元 (基础语法/控制流与函数/数据结构), 8 课, 16 道习题
