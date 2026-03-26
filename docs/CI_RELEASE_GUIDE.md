# AIGuard CI / 发布说明

这份文档给你一个“能持续出 Windows / macOS 包”的参考思路。

## 1. 目标

- 在 Windows Runner 构建 `windows/amd64`
- 在 macOS Runner 构建 `darwin/universal`
- 可选加入签名 / notarization
- 把产物上传为 CI artifacts

## 2. 推荐做法

### 方案 A：先不签名
适合团队内部先稳定产物。

### 方案 B：再补签名
适合对外分发或减少安全告警。

## 3. 示例工作流位置

```text
.github/workflows/release.yml
```

## 4. 建议准备的 Secrets

### Windows
- `WIN_SIGNING_CERT_BASE64`
- `WIN_SIGNING_CERT_PASSWORD`

### macOS
- `APPLE_DEVELOPER_CERTIFICATE_P12_BASE64`
- `APPLE_DEVELOPER_CERTIFICATE_PASSWORD`
- `APPLE_ID`
- `APPLE_TEAM_ID`
- `APPLE_APP_PASSWORD`

## 5. 版本管理建议

- Git Tag 作为版本号
- 把 Tag 注入包名与 release 目录名
- 在报告中记录版本号、提交 SHA、构建时间

## 6. CI 落地建议

1. 先跑 `go mod tidy`
2. 安装前端依赖并构建
3. 跑 `wails doctor`
4. 调用 `wails build`
5. 上传 `build/bin/*`
6. 需要签名时再在平台专属 step 中处理

## 7. 注意事项

- Windows 与 macOS 的签名流程不同，建议拆 step。
- macOS notarization 可能耗时较久。
- 对内分发先跑未签名包，确认功能后再补签名，更省时间。
