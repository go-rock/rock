# Changelog

本文件记录 rock 核心框架的版本变更。

## [Unreleased]

### 工程
- 核心测试覆盖率从 56% 提升到 80%（新增 5 个测试文件）
- `Context` 接口补全 `Attachment`/`Inline`（此前仅 `*Ctx` 上可用，handler 中无法调用）

### 健壮性
- trie：防御畸形路由模式（`"/:"`、`":+json"` 等）导致的下标越界 panic，改为干净的 `invalid pattern` 报错

### 示例（example/）
- 监听端口读取 `config.json` 的 `port`，缺省回退 `:8989`
- `auth()` 中间件改为 `?token=admin` 演示认证（原先硬编码未认证，/admin 永远 401）
- 清理死代码

## [v0.3.1] - 2026-08-03

### 安全
- 升级全部依赖：`go-playground/validator` v10.11 → v10.30、`golang.org/x/{crypto,sys,text}`、`kataras/golog` 等，清理 GitHub 报告的 21 个依赖漏洞
- Go 版本要求提升到 1.25

### 修复
- `c.String` 动态错误消息被当格式串的问题（含 `%` 会乱码）

## [v0.3.0] - 2026-08-03

### 安全
- 修复 `sync.Pool` 复用 Ctx 时 `values`/`ViewData` 跨请求数据泄漏
- `ClientIP` 不再无条件信任 `X-Real-IP`/`X-Forwarded-For`（新增 `SetTrustProxy` 开关）
- 错误响应不再向客户端泄露内部错误细节（仅调试模式）
- 文件上传：`http.MaxBytesReader` 限制请求体，防磁盘/内存 DoS；文本文件 MIME 误拒修复
- `ShouldBind` 校验错误不再误报为 500（映射为 400 Validation failed）

### 新增
- per-group `NoRoute`/`NoMethod`（各分组可有自己的 404/405，按前缀就近匹配）
- HEAD 请求自动回退到 GET
- `RunTLS` 与 SIGINT/SIGTERM 优雅关闭
- 请求日志默认开启；`SetDebug`/`SetTrustProxy` 配置
- 中间件收集改为免排序（按前缀长度有序拼接）
- `GetQuery` 对空值查询参数（`?foo=`）的准确判定

### 修复
- `Status()` 懒写头，消除重复 `WriteHeader` 与"先设后改"失效
- `Next()` 的 63 处理器链长上限
- 中间件前缀跨段误匹配（`/admin` 不再套到 `/administrator`）
- 404/405/OPTIONS/重定向跳过中间件的问题
- `Decode` 默认 body 上限从 10KB 提升到 10MB；支持同一请求重复 `ShouldBind`
- `Store` 并发竞态；`Entry.Value()` nil 防护；`ResetRequest` 清空残留状态
- TSR/FPR 重定向不再污染原始 `req.URL.Path`

### 工程
- trie / binding 单元测试，覆盖率 0% → 82%/95%
- GitHub Actions CI
- 完整 README

## [v0.2.0] - 2022-06

上一个发布版本（基线）。包含基础路由、中间件、请求绑定、上传、视图引擎插拔等核心能力。
