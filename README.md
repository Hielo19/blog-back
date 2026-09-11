# Gin Blog API

基于 Gin、pgx 和 PostgreSQL 的 Blog 后端项目，当前提供文章的创建、列表、详情、修改和软删除接口。

## 环境

- Go 1.26.5
- Gin v1.12.0
- pgx v5.11.0
- PostgreSQL 14+

## 目录结构

```text
.
├── cmd/server/                 # HTTP 服务入口
├── configs/                    # YAML 运行配置
├── database/                   # 数据库设计与 PostgreSQL 建表脚本
├── internal/
│   ├── config/                 # 配置文件读取与校验
│   ├── database/               # PostgreSQL 连接池
│   ├── dto/                    # 请求参数结构与校验规则
│   ├── handler/                # Gin HTTP 处理器
│   ├── model/                  # 领域模型
│   ├── repository/             # PostgreSQL 数据访问层
│   ├── response/               # 统一 HTTP 响应
│   ├── router/                 # 路由注册
│   └── service/                # 文章业务逻辑
├── go.mod
└── go.sum
```

## 启动

先编辑 `configs/config.yaml`，填写真实的 PostgreSQL 信息：

```yaml
server:
  port: 8080

database:
  host: 127.0.0.1
  port: 5432
  user: postgres
  password: "你的数据库密码"
  name: blog
  ssl_mode: disable
```

然后在 Windows CMD 中启动：

```bat
cd /d C:\Users\nieve\Desktop\back
go run ./cmd/server
```

程序默认读取 `configs/config.yaml`。本地开发可使用 `ssl_mode: disable`；生产环境应根据 PostgreSQL 部署方式启用 TLS。应用启动时会检查配置和数据库连接，失败时不会启动 HTTP 服务。

## 构建

```bat
go build -o bin\gin-blog.exe ./cmd/server
```

## 接口一览

| 方法 | 路径 | 说明 |
| --- | --- | --- |
| `POST` | `/api/v1/posts` | 创建文章 |
| `GET` | `/api/v1/posts` | 获取文章列表 |
| `GET` | `/api/v1/posts/:id` | 获取文章详情 |
| `PATCH` | `/api/v1/posts/:id` | 修改文章 |
| `DELETE` | `/api/v1/posts/:id` | 软删除文章 |
| `GET` | `/ping` | 服务健康检查 |

以下示例默认服务地址为 `http://127.0.0.1:8080`。

## 接口示例

### 1. 服务健康检查

请求：

```http
GET /ping HTTP/1.1
Host: 127.0.0.1:8080
```

成功响应：`200 OK`

```json
{
  "message": "pong"
}
```

### 2. 创建文章

请求：

```http
POST /api/v1/posts HTTP/1.1
Host: 127.0.0.1:8080
Content-Type: application/json

{
  "author_id": 1,
  "category_id": 1,
  "title": "我的第一篇文章",
  "slug": "my-first-post",
  "summary": "文章摘要",
  "content": "# Hello Blog",
  "content_format": "markdown",
  "cover_image_url": "https://example.com/images/cover.jpg",
  "status": "draft",
  "is_featured": false
}
```

`author_id` 对应的用户必须存在，非空的 `category_id` 也必须对应已有分类。`content_format` 默认为 `markdown`，`status` 默认为 `draft`。

成功响应：`201 Created`

```json
{
  "data": {
    "id": 1,
    "author_id": 1,
    "category_id": 1,
    "title": "我的第一篇文章",
    "slug": "my-first-post",
    "summary": "文章摘要",
    "content": "# Hello Blog",
    "content_format": "markdown",
    "cover_image_url": "https://example.com/images/cover.jpg",
    "status": "draft",
    "is_featured": false,
    "published_at": null,
    "created_at": "2026-09-11T08:00:00Z",
    "updated_at": "2026-09-11T08:00:00Z"
  }
}
```

### 3. 查看文章列表

请求：

```http
GET /api/v1/posts?page=1&page_size=10&status=published HTTP/1.1
Host: 127.0.0.1:8080
```

支持的查询参数：

| 参数 | 默认值 | 说明 |
| --- | --- | --- |
| `page` | `1` | 页码，从 1 开始 |
| `page_size` | `10` | 每页数量，最大 100 |
| `status` | 空 | 可选：`draft`、`published`、`archived` |

成功响应：`200 OK`

```json
{
  "data": {
    "items": [
      {
        "id": 1,
        "author_id": 1,
        "category_id": 1,
        "title": "我的第一篇文章",
        "slug": "my-first-post",
        "summary": "文章摘要",
        "content": "# Hello Blog",
        "content_format": "markdown",
        "cover_image_url": "https://example.com/images/cover.jpg",
        "status": "published",
        "is_featured": false,
        "published_at": "2026-09-11T08:10:00Z",
        "created_at": "2026-09-11T08:00:00Z",
        "updated_at": "2026-09-11T08:10:00Z"
      }
    ],
    "page": 1,
    "page_size": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

没有匹配文章时，`items` 返回空数组，`total` 和 `total_pages` 返回 `0`。

### 4. 查看文章详情

请求：

```http
GET /api/v1/posts/1 HTTP/1.1
Host: 127.0.0.1:8080
```

成功响应：`200 OK`

```json
{
  "data": {
    "id": 1,
    "author_id": 1,
    "category_id": 1,
    "title": "我的第一篇文章",
    "slug": "my-first-post",
    "summary": "文章摘要",
    "content": "# Hello Blog",
    "content_format": "markdown",
    "cover_image_url": "https://example.com/images/cover.jpg",
    "status": "published",
    "is_featured": false,
    "published_at": "2026-09-11T08:10:00Z",
    "created_at": "2026-09-11T08:00:00Z",
    "updated_at": "2026-09-11T08:10:00Z"
  }
}
```

文章不存在或已经被软删除时返回 `404 Not Found`。

### 5. 修改文章

修改接口只更新请求中提供的字段：

```http
PATCH /api/v1/posts/1 HTTP/1.1
Host: 127.0.0.1:8080
Content-Type: application/json

{
  "title": "修改后的标题",
  "status": "published"
}
```

创建或修改为 `published` 时，如果没有提供 `published_at`，服务会自动写入当前 UTC 时间。

成功响应：`200 OK`

```json
{
  "data": {
    "id": 1,
    "author_id": 1,
    "category_id": 1,
    "title": "修改后的标题",
    "slug": "my-first-post",
    "summary": "文章摘要",
    "content": "# Hello Blog",
    "content_format": "markdown",
    "cover_image_url": "https://example.com/images/cover.jpg",
    "status": "published",
    "is_featured": false,
    "published_at": "2026-09-11T08:10:00Z",
    "created_at": "2026-09-11T08:00:00Z",
    "updated_at": "2026-09-11T08:10:00Z"
  }
}
```

### 6. 删除文章

请求：

```http
DELETE /api/v1/posts/1 HTTP/1.1
Host: 127.0.0.1:8080
```

成功响应：`204 No Content`，响应体为空。该操作会填写文章的 `deleted_at`，不会物理删除数据库记录；后续列表和详情接口不再返回这篇文章。

## 错误响应

统一格式：

```json
{
  "error": {
    "code": "POST_NOT_FOUND",
    "message": "文章不存在"
  }
}
```

| HTTP 状态码 | 错误代码 | 常见原因 |
| --- | --- | --- |
| `400` | `INVALID_ARGUMENT` | JSON、分页、状态、作者或分类参数不正确 |
| `404` | `POST_NOT_FOUND` | 文章不存在或已经软删除 |
| `409` | `SLUG_ALREADY_USED` | 文章 slug 已被其他文章使用 |
| `500` | `INTERNAL_ERROR` | 未预期的服务器或数据库错误 |

## 生产提示

- 设置 `GIN_MODE=release` 关闭 Gin 调试模式。
- 不要将数据库密码写入代码或提交到版本控制。
- 应根据实际部署网络配置 PostgreSQL TLS 和 Gin 可信代理。
