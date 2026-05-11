# API 文档

## 认证

所有 API（除 `/api/auth/*` 外）需要在 Header 中携带 token：

```
Authorization: {token}
```

## 账号管理

### 列表账号
```
GET /api/accounts
```

### 搜索账号
```
GET /api/accounts/search?q={query}
```

### 创建账号
```
POST /api/accounts
Content-Type: application/json

{
  "categoryId": 1,
  "title": "GitHub",
  "website": "https://github.com",
  "username": "user",
  "password": "secret",
  "email": "user@example.com",
  "reminderEmail": "notify@example.com",
  "remindAt": 1893456000,
  "phone": "",
  "notes": "",
  "isFavorite": 0
}
```

### 更新账号
```
PUT /api/accounts/{id}
```

### 删除账号
```
DELETE /api/accounts/{id}
```

## 笔记管理

### 列表笔记
```
GET /api/files
```

### 创建笔记
```
POST /api/files
Content-Type: multipart/form-data

title: 笔记标题
body: 笔记内容
bodyFormat: text | markdown
```

### 更新笔记
```
PUT /api/files/{id}
Content-Type: application/json

{
  "title": "新标题",
  "body": "新内容",
  "bodyFormat": "text"
}
```

### 删除笔记
```
DELETE /api/files/{id}
```

## 笔记附件

### 上传附件
```
POST /api/files/{id}/attachments
Content-Type: multipart/form-data

files: [多个文件]
```

### 列表附件
```
GET /api/files/{id}/attachments
```

### 下载附件
```
GET /api/note-attachments/{id}
```

### 预览附件
```
GET /api/note-attachments/{id}/preview
```

### 删除附件
```
DELETE /api/note-attachments/{id}
```

## 分类管理

### 列表分类
```
GET /api/categories
```

### 创建分类
```
POST /api/categories
{
  "name": "工作",
  "icon": "briefcase"
}
```

## 提醒

### 列表提醒
```
GET /api/reminders
```

### 获取待发送提醒
```
GET /api/reminders/pending
```

### 发送到期提醒
```
POST /api/reminders/send-due
```

## 设置

### 列表设置
```
GET /api/settings
```

### 获取单个设置
```
GET /api/settings/{key}
```

### 更新设置
```
PUT /api/settings/{key}
{
  "value": "setting-value"
}
```

邮件相关设置 key：
- `mail.smtp_host`
- `mail.smtp_port`
- `mail.smtp_username`
- `mail.smtp_password`
- `mail.from_address`
- `mail.from_name`

## 导入导出

### 导出数据
```
GET /api/export
```

### 导入数据
```
POST /api/import
Content-Type: multipart/form-data

file: [json或csv文件]
```
