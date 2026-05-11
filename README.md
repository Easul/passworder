# Passworder

个人密码与笔记管理工具。单二进制部署，内置 Web UI，支持账号管理、笔记记录、文件附件、邮件提醒等功能。

## 项目简介

**Passworder** 是一个轻量级的个人数据管理工具，采用 Go 语言编写，前端静态文件嵌入到后端二进制中，单文件即可部署运行，无需依赖外部服务。

### 核心特点

- 🔐 **安全存储**：数据存储于本地 SQLite 数据库，密码加密存储
- 📬 **单文件部署**：仅需一个可执行文件，前端资源嵌入后端
- 📝 **笔记管理**：支持 Markdown 和纯文本格式，集成 Vditor 编辑器
- 📂 **多文件附件**：每条笔记可关联多个文件，支持图片/压缩包/文档预览
- ⏰ **邮件提醒**：为账号设置登录提醒日期，到期自动发送邮件通知
- 🔄 **数据导入导出**：支持 JSON/CSV 格式备份与恢复

## 快速开始

### 环境要求

- Go 1.25 或更高版本（用于编译）
- 或直接下载预编译好的二进制文件

### 编译与运行

```bash
# 克隆代码
git clone https://github.com/yourusername/passworder.git
cd passworder

# 编译
go build -o dist/passworder ./cmd/server

# 运行（默认端口 8080）
./dist/passworder

# 指定参数运行
./dist/passworder --host 0.0.0.0 --port 8080 --db ./password.db --storage ./storage
```

### 首次使用

1. 打开浏览器访问 `http://localhost:8080`
2. 设置主密码（至少 8 位）完成初始化
3. 开始添加账号或笔记

## 功能详情

### 账号管理

- **分类存储**：支持网站/工作/社交/金融/游戏/其他分类
- **密码生成**：内置密码生成器，可配置长度、字符类型
- **搜索筛选**：按标题、用户名、网站搜索，按分类筛选
- **收藏标记**：重要账号可标记为收藏
- **登录提醒**：为账号设置定期更新提醒

### 笔记系统

- **双格式支持**：纯文本或 Markdown 格式
- **Vditor 编辑器**：支持实时预览、快捷键、上传图片
- **回收站**：删除的笔记进入回收站，可恢复或清空
- **附件管理**：每条笔记可关联多个文件

### 文件附件预览

- **图片**：支持 JPEG、PNG、GIF 等格式预览
- **压缩包**：ZIP 文件可列出内部文件结构
- **文档**：支持 Word (docx)、Excel (xlsx)、PDF、文本文件在线预览
- **下载功能**：所有附件都可下载到本地

### 邮件提醒

- **SMTP 配置**：支持任意 SMTP 服务器
- **定时提醒**：为账号设置登录提醒日期
- **批量发送**：到期账号自动汇总发送邮件

## 配置说明

支持命令行参数和环境变量两种方式：

| 参数 | 环境变量 | 默认值 | 说明 |
|------|----------|--------|-------|
| `--host` | `PASSORDER_HOST` | `0.0.0.0` | HTTP 服务地址 |
| `--port` | `PASSORDER_PORT` | `8080` | HTTP 服务端口 |
| `--db` | - | `./password.db` | SQLite 数据库路径 |
| `--storage` | - | `./storage` | 文件存储目录 |

示例：
```bash
# 使用环境变量
PASSORDER_HOST=127.0.0.1 PASSORDER_PORT=9000 ./dist/passworder

# 使用命令行参数
./dist/passworder --host 0.0.0.0 --port 8080 --db ./password.db --storage ./storage
```

## 数据备份

### 导出数据
1. 进入「设置」页面
2. 点击「导出数据」
3. 选择 JSON 或 CSV 格式，下载备份文件

### 导入数据
1. 进入「设置」页面
2. 点击「导入数据」
3. 选择之前导出的 JSON 或 CSV 文件
4. 确认导入

## 项目结构

```
.
├── cmd/server/           # 服务入口
│   ├── main.go              # 主程序
│   ├── static/              # 前端资源（嵌入二进制）
│   │   ├── index.html       # 主页面
│   │   ├── css/style.css    # 样式文件
│   │   └── js/app.js        # 前端逻辑
│   └── migrations/        # 数据库迁移文件
├── internal/
│   ├── config/            # 配置管理
│   ├── database/          # 数据库连接与迁移
│   ├── handler/           # HTTP 接口处理
│   ├── service/           # 业务逻辑
│   ├── repository/        # 数据访问层
│   ├── model/             # 数据模型
│   └── storage/           # 文件存储
├── docs/                  # 文档
└── dist/                  # 构建输出（不提交到git）
```

## 技术栈

- **后端**：Go 1.20 + Gorilla Mux + SQLite (sqlx)
- **前端**：原生 JavaScript + MDUI 组件库
- **编辑器**：Vditor (Markdown 编辑器)
- **文档预览**：mammoth.js (docx) + SheetJS (xlsx) + 浏览器原生 PDF

## 数据库迁移

项目使用数据库迁移管理模式，所有迁移文件位于 `cmd/server/migrations/` 目录：

- 文件名格式：`XXX_description.up.sql`
- 服务启动时自动检查并执行未应用的迁移
- 迁移状态保存在 `schema_migrations` 表中
- 仅支持向前迁移，不支持回滚

## 常见问题

### 如何更改端口？
```bash
./dist/passworder --port 9000
```

### 如何在公网访问？
```bash
./dist/passworder --host 0.0.0.0 --port 8080
```
**注意**：公网访问时建议配置防火墙和 HTTPS。

### 数据存储在哪里？
- 数据库：默认 `./password.db`（可通过 `--db` 指定）
- 文件存储：默认 `./storage/` 目录（可通过 `--storage` 指定）

## 开发

```bash
# 安装依赖
go mod download

# 运行测试
go test ./...

# 开发运行
go run ./cmd/server

# 构建发布版本
# Linux
CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o dist/passworder-linux-amd64 ./cmd/server

# macOS
CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o dist/passworder-darwin-amd64 ./cmd/server

# Windows（无 CGO 依赖）
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o dist/passworder-windows-amd64.exe ./cmd/server
```

## 文档

- [用户指南](docs/user-guide.md) - 详细使用说明
- [API 文档](docs/api.md) - 接口参考
- [开发文档](docs/development.md) - 开发相关说明
- [更新日志](docs/changelog.md) - 版本历史

## 贡献

欢迎提交 Issue 和 Pull Request！

1. Fork 本仓库
2. 创建你的功能分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开 Pull Request

## 许可证

MIT License

```
MIT License

Copyright (c) 2024 Passworder

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

---

**提示**：请定期备份您的数据库和文件存储目录，以防数据丢失。
