# AIGuard 打包执行文件完整指南

这份文档面向 Windows 和 macOS 本地构建，目标是把当前源码打成可直接分发的桌面程序。

## 1. 构建目标

### Windows
- 调试包：`AIGuard.exe`
- 安装包：`AIGuard-Setup.exe` 或 NSIS 生成的安装程序

### macOS
- 调试包：`AIGuard.app`
- 发布包：签名后的 `.app`，再按团队习惯封装成 `.zip` 或 `.dmg`

## 2. 必备依赖

### 通用
- Go 1.21+
- Node.js 18+
- npm
- Git
- Wails CLI

### Windows
- WebView2 Runtime
- PowerShell 5.1+ 或 PowerShell 7+
- 可选：NSIS（如果需要安装包）
- 可选：Windows SDK SignTool（如果需要代码签名）

### macOS
- Xcode Command Line Tools
- 可选：gon（如果需要签名与 notarization）

## 3. 安装 Wails CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

安装后建议执行：

```bash
wails doctor
```

## 4. 首次构建前检查

### Windows

```powershell
./scripts/preflight.ps1
```

### macOS

```bash
chmod +x ./scripts/*.sh
./scripts/preflight.sh
```

## 5. 准备配置文件

```bash
cp examples/config.production.yaml config.yaml
```

修改以下项目：

- `openai.base_url`
- `openai.api_key`
- `openai.default_model`

如果你不想把配置写死到文件里，也可以走环境变量：

```bash
OPENAI_BASE_URL
OPENAI_API_KEY
OPENAI_DEFAULT_MODEL
```

## 6. Windows 本地打包

### 6.1 先构建 EXE

```powershell
./scripts/build_windows.ps1
```

脚本会执行：

1. 检查 Go / Node / npm / Git / Wails
2. 安装前端依赖并构建前端
3. 执行 `go mod tidy`
4. 执行 `wails doctor`
5. 执行 `wails build -platform windows/amd64`
6. 把产物归档到 `release/`

### 6.2 生成 NSIS 安装包

先安装 NSIS：

```powershell
winget install NSIS.NSIS --silent
```

然后执行：

```powershell
./scripts/build_windows.ps1 -Nsis
```

### 6.3 对 EXE 或安装包签名

如果已安装 Windows SDK 并有 `.pfx` 证书：

```powershell
./scripts/build_windows.ps1 -Nsis -Sign -SignPfxPath "C:\cert\codesign.pfx" -SignPfxPassword "你的密码"
```

脚本会尝试对 `build/bin` 下的 `.exe` 执行签名。

## 7. macOS 本地打包

### 7.1 安装 Xcode Command Line Tools

```bash
xcode-select --install
```

### 7.2 构建 universal `.app`

```bash
./scripts/build_macos.sh
```

脚本会执行：

1. 检查 Go / Node / npm / Git / Wails
2. 检查 `xcode-select`
3. 安装前端依赖并构建前端
4. 执行 `go mod tidy`
5. 执行 `wails doctor`
6. 执行 `wails build -platform darwin/universal`
7. 归档产物到 `release/`

### 7.3 签名与 notarization

如果需要上线分发，建议补签名。

#### 安装 gon

```bash
brew install Bearer/tap/gon
```

#### 准备配置

- `build/darwin/Info.plist`
- `build/darwin/entitlements.plist`
- `build/darwin/gon-sign.sample.json`

把 `gon-sign.sample.json` 复制为 `gon-sign.json`，然后填入：

- Apple ID
- app-specific password
- team/provider ID
- Developer ID Application identity
- bundle identifier

#### 执行签名

```bash
gon -log-level=info ./build/darwin/gon-sign.json
```

## 8. 产物位置

### 默认 Wails 构建目录

```text
build/bin/
```

### 本项目归档目录

```text
release/
```

## 9. 推荐的发版流程

### 团队内测版

1. 在目标系统本机构建
2. 不做签名，先验证主功能
3. 使用 HTML 报告链路做端到端验证

### 正式发布版

1. 固定版本号与提交 SHA
2. 在干净环境构建
3. 生成 Windows EXE / NSIS 包
4. 生成 macOS `.app`
5. 完成代码签名与 notarization
6. 归档到 release 目录
7. 产物、配置样例、文档一起上传

## 10. 常见问题

### Q1：`wails doctor` 提示缺少 WebView2
在 Windows 安装 WebView2 Runtime 后重试。

### Q2：`wails doctor` 提示 macOS 依赖缺失
先执行 `xcode-select --install`，然后重试。

### Q3：构建时前端依赖报错
删除 `frontend/node_modules` 后重新 `npm install`。

### Q4：没有 `go.sum`
执行：

```bash
go mod tidy
```

### Q5：本地模型网关不校验 API Key
可以将 `api_key` 设成 `EMPTY`。

## 11. 建议同时保留的文件

发布给团队时，除了可执行文件，建议同时提供：

- `config.production.yaml` 模板
- `AIGUARD.md` 样例
- `report.html` 示例
- `docs/BUILD_EXECUTABLE_GUIDE.md`
