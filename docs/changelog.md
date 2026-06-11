# 更新日志

## 未发布

- 暂无。

## v1.0.5 (2026-06-11)

### 新增功能

- **Android 翻译悬浮窗**：设置兼容 OpenAI Chat Completions 的 Base URL、API Key 和 Model 后，可从 Android 顶部「翻译」按钮开启可拖拽、可收起的中英互译悬浮窗。
- **音频附件预览**：笔记附件支持 mp3、wav、ogg、m4a、aac、flac、opus、webm 在线播放，并使用音频图标区分。
- **Android PDF 外部打开**：Android 端预览 PDF 时优先通过系统查看器打开，减少 WebView 内嵌 PDF 兼容性问题。

### 改进

- **Android 局域网访问**：Host 设为 `0.0.0.0` 后，应用会重启服务并提示手机可访问 IP，WebView 仍使用本机 `127.0.0.1` 访问内置服务。
- **翻译复制体验**：悬浮窗复制结果使用独立短生命周期 Activity，避免把 Passworder 拉回前台任务。
- **翻译悬浮窗稳定性**：优化前台服务保活、收起/展开、点击恢复和输入焦点处理。
- **附件预览稳定性**：统一通过 `window.JSZip`、`window.mammoth`、`window.XLSX` 使用动态加载库，降低全局变量解析问题。

## v1.0.4 (2026-06-08)

### 改进

- 将嵌入式服务与移动端 gomobile 绑定整理到 `internal/embedded` 和 `mobile/bridge`，减少根目录散落 Go 文件。
- 新增 Linux/macOS/Windows amd64 构建脚本，以及 Android arm32/arm64 AAR/APK 构建脚本。
- GitHub Actions 增加 Android arm32/arm64 APK 构建，APK 版本名使用 `tag+6位commit`，版本号使用 `5000 + main 分支提交数`。
- `.gitignore` 补充 Android、Gradle、gomobile 生成产物，避免提交中间构建文件。

## v1.1.0 (2026-05-03)

### 新增功能

- **账号状态管理**：支持设置账号为"有效"或"无效"，无效账号不参与邮件提醒
- **账号注册信息**：添加"注册时间"和"注册备注"字段
- **笔记回收站**：删除的笔记移入回收站，支持恢复和清空
- **笔记浏览模式**：笔记条目右上角可直接浏览正文内容（Markdown 渲染或纯文本）
- **笔记备注**：笔记弹窗添加备注输入框
- **数据导入导出**：支持 ZIP 格式备份，包含账号、分类、笔记及附件
- **提醒邮箱列表**：设置页面可配置多个提醒邮箱，账号编辑时下拉选择

### 改进

- **弹窗交互**：点击黑色背景不再关闭弹窗，防止误触导致数据丢失
- **笔记保存后**：保持在笔记面板，不再跳转到账号面板
- **账号搜索**：支持通过"备注"和"注册备注"搜索
- **命令行参数**：长参数使用 `--`，短参数使用 `-`
- **Windows 支持**：提供无 CGO 依赖的 Windows 64 位二进制包

### 技术变更

- 账号表新增 `status`、`registration_time`、`registration_notes` 字段
- 笔记表新增 `deleted_at` 字段支持软删除
- 引入 `modernc.org/sqlite` 纯 Go SQLite 驱动用于 Windows 构建
- 导入导出功能支持 settings（SMTP 配置、提醒邮箱等）

## v1.0.0 (2026-04-20)

### 新增功能

- **多文件附件**：笔记支持上传多个文件，可预览、下载、删除单个附件
- **Markdown 编辑器**：集成 Vditor，支持所见即所得编辑模式
- **邮件提醒系统**：
  - 账号可设置登录提醒日期和提醒邮箱
  - 支持 SMTP 配置
  - 到期账号汇总发送邮件
- **文件预览**：
  - 图片直接预览
  - ZIP 压缩包显示文件树
  - DOCX 文档预览
  - XLSX 表格预览
  - PDF 全屏预览
  - TXT/CSV 文本预览
- **中文文件名支持**：下载时正确编码中文文件名

### 改进

- 前端分页支持
- 笔记搜索功能
- 账号/笔记导航整合
- 导出数据加载状态提示

### 技术变更

- 新增 `note_attachments` 表支持多文件
- 账号表新增 `reminder_email` 和 `remind_at` 字段
- 引入 Vditor 替换 EasyMDE
- 引入 JSZip、mammoth.js、SheetJS 用于文件预览
