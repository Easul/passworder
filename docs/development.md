# 开发文档

## 环境要求

- Go 1.25+
- SQLite3 开发库（Linux/macOS 需要 CGO，Windows 使用纯 Go 驱动）
- Node.js（可选，用于前端语法检查）

## 本地开发

```bash
# 安装依赖
go mod download

# 运行服务
go run ./cmd/server

# 或带参数
go run ./cmd/server --port 18080 --db ./dev.db --storage ./dev-storage

# 构建
go build -o dist/passworder ./cmd/server

# 运行测试
go test ./...
```

## 测试指南

### 测试目录结构

所有测试数据（数据库和存储目录）应放在 `test/` 目录下，按日期组织：

```
test/
├── 20260202/              # 2026年2月2日的测试数据
│   ├── test.db            # 测试数据库
│   └── storage/           # 测试文件存储
└── 20260203/              # 2026年2月3日的测试数据
    ├── qa.db
    └── storage/
```

### 运行测试

```bash
# 创建日期格式的测试目录
mkdir -p test/$(date +%Y%m%d)

# 使用测试数据库和存储路径运行
./dist/passworder --db test/$(date +%Y%m%d)/test.db --storage test/$(date +%Y%m%d)/storage

# 或使用完整参数形式
./dist/passworder --host 127.0.0.1 --port 18080 --db test/20260202/test.db --storage test/20260202/storage

# 针对特定测试场景
./dist/passworder --db test/qa-$(date +%Y%m%d).db --storage test/qa-storage-$(date +%Y%m%d)
```

### 禁止行为

**不要**：
- 在项目根目录创建 `test_*.db` 或 `test_*` 文件夹
- 将测试数据提交到 git
- 混合测试数据和生产数据

## 数据库迁移

迁移文件位于 `internal/embedded/assets/migrations/`，按序号执行：

- `001_init.up.sql` — 初始表结构
- `002_personal_files.up.sql` — 笔记文件表
- `003_personal_files_text_first.up.sql` — 笔记文本字段
- `004_account_reminder_defaults_and_sender.up.sql` — 账号提醒字段
- `005_note_attachments.up.sql` — 笔记多附件表
- `006_reminder_period.up.sql` — 提醒周期支持
- `007_add_account_status.up.sql` — 账号状态字段
- `008_add_note_trash.up.sql` — 笔记回收站功能
- `009_add_note_remarks.up.sql` — 笔记备注字段
- `010_add_account_registration_fields.up.sql` — 账号注册信息字段

服务启动时自动执行迁移。

## 添加新功能的一般流程

1. **模型**：在 `internal/model/model.go` 添加 struct
2. **迁移**：在 `internal/embedded/assets/migrations/` 添加 SQL
3. **数据层**：在 `internal/repository/` 添加 CRUD
4. **业务层**：在 `internal/service/` 添加逻辑
5. **接口层**：在 `internal/handler/` 添加 HTTP handler
6. **路由**：在 `internal/handler/router.go` 注册路由
7. **前端**：修改 `internal/embedded/assets/static/` 下的 HTML/JS/CSS

## 前端技术说明

- 纯原生 JavaScript，无框架
- 使用 MDUI 风格的自定义 CSS
- Markdown 编辑器使用 Vditor（CDN 动态加载）
- 文件预览使用 JSZip、mammoth.js、SheetJS（CDN 动态加载）
- 所有静态文件通过 `//go:embed` 嵌入二进制

## 文件存储

附件存储在 `-storage` 指定的目录下，按文件类型自动分类：
- `personal/` — 笔记附件
- `attachments/` — 账号附件

文件名格式：`{timestamp}_{originalName}`

## 安全注意事项

- 主密码使用 bcrypt 哈希存储
- 账号密码使用 AES 加密存储
- Session token 使用随机字符串
- 建议在生产环境使用 HTTPS 反向代理

## 构建发布

### 常用平台构建脚本

```bash
./scripts/build-linux-amd64.sh
./scripts/build-macos-amd64.sh
./scripts/build-windows-amd64.sh
```

**说明**：
- `-ldflags="-s -w"` 用于移除调试信息和符号表，减小二进制体积
- Linux 和 macOS 需要 CGO 支持 SQLite3
- Windows 使用纯 Go 的 SQLite 驱动，无需 CGO

### Android 构建

```bash
# 只构建 gomobile AAR
./scripts/build-android-server.sh arm32
./scripts/build-android-server.sh arm64

# 构建 APK
./scripts/build-android-apk.sh arm32
./scripts/build-android-apk.sh arm64
./scripts/build-android-apk.sh all
```

- `arm32` 对应 `armeabi-v7a`，`arm64` 对应 `arm64-v8a`。
- gomobile 绑定入口在 `mobile/bridge`。
- AAR、APK、Gradle build 目录和 keystore 都不提交。
- APK `versionName` 使用 `tag+6位commit`，例如 `v1.0.4+ab123a`。
- APK `versionCode` 使用 `5000 + main 分支提交数`，保证 adb 升级时单调递增。

## 版本规范

项目使用语义化版本控制：
- 主版本号：大功能更新和不兼容变更
- 次版本号：新功能添加
- 修订版本号：bug 修复和小改进

当前版本：v1.0.2
