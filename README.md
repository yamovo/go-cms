# 🌪️ VortexCMS

> 一个基于 **Go + Vue 3** 构建的全功能内容管理系统，涵盖内容创作、权限管理、SEO 优化、数据分析等完整功能，开箱即用。

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev)
[![Vue](https://img.shields.io/badge/Vue-3.4-4FC08D?logo=vuedotjs)](https://vuejs.org)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

---

## 📸 预览

```
   ┌────────────────────────────────────────┐   
   │  📊 Dashboard      📝 文章管理          │   
   │  🗂 分类管理        🏷 标签管理         │   
   │  💬 评论审核        🖼 媒体库            │   
   │  👥 用户 & 角色     ⚙ 系统设置          │   
   │  🔍 SEO & 站点地图  📈 访问统计          │   
   └────────────────────────────────────────┘   
```

---

## ✨ 核心功能

| 模块 | 功能 |
|------|------|
| 🔐 认证授权 | JWT 登录注册、双 Token 刷新、RBAC 角色权限、IP 限流 |
| 📝 内容管理 | 文章 CRUD、Markdown/富文本双编辑器、版本历史、批量操作 |
| 🗃 媒体库 | 拖拽上传、缩略图生成、文件夹管理、批量删除 |
| 💬 评论系统 | 多级嵌套评论、审核/垃圾过滤、邮件通知 |
| 🏷 分类标签 | 无限层级分类树、标签合并、slug 友好链接 |
| 🔍 SEO | Meta 管理、XML Sitemap、Robots.txt、301 重定向 |
| 📊 数据分析 | 页面浏览量、来源统计、设备分布、ECharts 图表 |
| 📡 RSS / API | 自动生成 RSS 2.0 源、80+ RESTful 接口 |
| 🧩 扩展系统 | 插件架构、多主题切换、邮件通知、定时备份 |
| 🎨 前端体验 | 暗色/亮色主题、响应式布局、全局搜索、实时预览 |

---

## 🏗 架构

```
  Browser / API Client
         │
         ▼
  ┌─────────────┐
  │  Middleware  │  ← CORS、限流、JWT 鉴权、活动日志
  ├─────────────┤
  │   Handler    │  ← HTTP 解析 & 响应（薄层）
  ├─────────────┤
  │   Service    │  ← 全部业务逻辑，独立可测试
  ├─────────────┤
  │    GORM      │  ← 数据库抽象
  ├─────────────┤
  │ PostgreSQL │ MySQL │ SQLite │  ← 多数据库支持
  └─────────────┘
```

**三层架构** — Handler → Service → Repository，职责清晰，Service 层 100% 可单独测试（SQLite `:memory:`）。

---

## 🛠 技术栈

| 层级 | 技术选型 |
|------|----------|
| 后端语言 | Go 1.22+ |
| Web 框架 | Gin |
| ORM | GORM |
| 数据库 | PostgreSQL 16 / MySQL 8 / SQLite |
| 认证 | JWT HS256 + 黑名单 |
| 前端框架 | Vue 3.4 + TypeScript |
| UI 组件库 | Element Plus |
| 状态管理 | Pinia |
| 构建工具 | Vite 5 |
| 图表 | ECharts 5 |
| 样式 | SCSS + Tailwind CSS |

---

## 🚀 5 分钟快速开始

### Docker Compose（推荐）

```bash
git clone https://github.com/yamovo/go-cms.git
cd go-cms
docker-compose up -d

# 前台 http://localhost
# 后台 http://localhost/login
# 账号 admin / admin123
```

### 本地运行（仅需 Go）

```bash
git clone https://github.com/yamovo/go-cms.git
cd go-cms
go run cmd/server/main.go

# 打开 http://localhost:8080
# 账号 admin / admin123
```

> 💡 默认使用 SQLite，零配置即刻运行。生产环境切换 PostgreSQL 只需改环境变量。

---

## 📁 项目结构

```
go-cms/
├── cmd/server/main.go         # 应用入口
├── internal/
│   ├── config/                # 配置（30+ 环境变量驱动）
│   ├── database/              # 数据库连接、迁移、种子数据
│   ├── models/                # 20+ 数据模型
│   ├── auth/                  # JWT & 密码加密
│   ├── middleware/             # 认证/限流/CORS/日志/恢复
│   ├── services/              # 业务逻辑层（含单元测试）
│   └── handlers/              # HTTP 处理器（11 个 Handler）
├── web/                       # Vue 3 管理后台 & 前台博客
│   └── src/views/             # 30+ 页面组件
├── deploy/                    # Docker & Nginx 配置
├── docker-compose.yml
└── README.md
```

---

## 📡 API 概览

| 分组 | 接口数 | 说明 |
|------|--------|------|
| Auth | 5 | 登录、注册、Token 刷新、个人信息 |
| Articles | 7 | CRUD、批量操作、版本历史、点赞 |
| Categories | 5 | 树形分类、拖拽排序 |
| Tags | 5 | 增删改查、标签合并 |
| Comments | 7 | 审核、垃圾标记、批量操作 |
| Media | 7 | 上传（单/批量）、文件夹管理 |
| Users & Roles | 8 | 用户管理、角色权限分配 |
| Settings | 4 | 站点设置 |
| SEO | 5 | Meta、Sitemap、重定向 |
| Menus | 6 | 多级菜单管理 |
| Analytics | 4 | 仪表盘、访问趋势、来源分析 |
| Plugins & Themes | 7 | 插件启用/禁用、主题切换 |
| System | 3 | 系统信息、健康检查、活动日志 |

> 完整 API 文档：`docs/api.md`

---

## 🔧 配置

所有参数通过 `.env` 文件或环境变量设置：

```env
DB_DRIVER=sqlite             # postgres | mysql | sqlite
SERVER_PORT=8080
JWT_SECRET=your-secret-key
UPLOAD_MAX_SIZE=20971520     # 20MB
ANALYTICS_ENABLED=true
CACHE_DRIVER=memory          # memory | redis
```

---

## 🧪 测试

```bash
go test ./...                        # 全部测试
go test ./internal/services/ -v      # Service 层（39 用例）
```

Service 层覆盖 Auth、Article、Category、Tag、Comment，使用 SQLite 内存数据库隔离运行。

---

## 📄 License

MIT © 2024 VortexCMS

---

> 🌀 This project is feature-complete and archived. Feel free to fork and adapt.
