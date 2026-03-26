# AI代码监视（AIGuard）

一个面向团队内部使用的本地桌面程序，用于对 GitHub PR / GitLab MR / 本地仓库改动做 AI 代码审计。

## 这一版增强了什么

这次增强包在首版工程骨架基础上，补了更接近生产落地的交付内容：

- 更完整的 Windows / macOS 打包执行文件指导文档
- 本地预检脚本（检查 Go、Node、Wails、Git、WebView2 / Xcode CLT 等依赖）
- Windows 一键构建脚本（可选 NSIS 安装包、可选签名）
- macOS 一键构建脚本（可选 universal 包、可选签名与 notarization 模板）
- GitHub Actions 跨平台发布工作流样例
- macOS 签名模板：`Info.plist`、`entitlements.plist`、`gon-sign.sample.json`
- 生产环境配置样例 `examples/config.production.yaml`
- 配置文件环境变量覆盖能力

## 功能概览

- 粘贴 PR/MR 链接，填写源分支和目标分支
- 本地自动 clone / fetch / worktree / diff
- 审计前先做项目画像
- 规则预扫 + LLM 分片审计
- 输出 HTML / Markdown / JSON 报告
- 支持刷新后的问题对比（新增 / 已修复 / 仍存在）
- 默认兼容 OpenAI-compatible 本地模型网关

## 目录重点

```text
aiguard/
  main.go
  app.go
  internal/
  frontend/
  docs/
  scripts/
  build/darwin/
  .github/workflows/
  examples/
```

## 文档入口

- `docs/BUILD_EXECUTABLE_GUIDE.md`：Windows / macOS 打包执行文件完整指南
- `docs/CI_RELEASE_GUIDE.md`：CI 发布与签名思路
- `docs/RELEASE_CHECKLIST.md`：发版检查清单
- `docs/PRODUCTION_ENHANCEMENTS.md`：本增强包增加了哪些内容

## 运行前准备

- Go 1.21+（macOS 15+ 建议 Go 1.23.3+）
- Node.js 18+
- Wails CLI
- Git
- 可访问的 OpenAI-compatible 模型服务
- Windows：建议确认 WebView2 Runtime 已安装
- macOS：先安装 Xcode Command Line Tools

## 快速开始

### 1. 安装前端依赖

```bash
cd frontend
npm install
npm run build
cd ..
```

### 2. 准备配置文件

```bash
cp examples/config.production.yaml config.yaml
```

然后修改：

- `openai.base_url`
- `openai.api_key`
- `openai.default_model`

### 3. 运行开发模式

```bash
wails dev
```

### 4. 本地构建桌面二进制

```bash
wails build
```

或按平台脚本执行：

```bash
# Windows PowerShell
./scripts/build_windows.ps1

# macOS / Linux Shell
./scripts/build_macos.sh
```

## 构建脚本说明

### Windows

- `scripts/preflight.ps1`：依赖检查
- `scripts/build_windows.ps1`：构建 EXE，可选生成 NSIS 安装包，可选签名
- `scripts/package_release.ps1`：把 `build/bin` 归档到 `release/`

### macOS

- `scripts/preflight.sh`：依赖检查
- `scripts/build_macos.sh`：构建 `.app`，支持 `darwin/universal`
- `scripts/package_release.sh`：把 `.app` 归档到 `release/`

## 环境变量覆盖

除 `config.yaml` 外，还支持这些环境变量覆盖：

- `OPENAI_BASE_URL`
- `OPENAI_API_KEY`
- `OPENAI_DEFAULT_MODEL`
- `AIGUARD_OPENAI_BASE_URL`
- `AIGUARD_OPENAI_API_KEY`
- `AIGUARD_OPENAI_DEFAULT_MODEL`
- `AIGUARD_WORKSPACE_DIR`
- `AIGUARD_LOG_LEVEL`
- `AIGUARD_CONCURRENCY`
- `AIGUARD_SAFE_INPUT_TOKENS`
- `AIGUARD_RESERVED_OUTPUT_TOKENS`

适合本地开发、CI/CD、以及多套模型网关切换。

## 配置文件说明

示例见：

- `examples/config.yaml`
- `examples/config.production.yaml`
- `examples/AIGUARD.md`

## 建议的 V1 使用方式

必填项：

- MR/PR 链接
- 源分支
- 目标分支
- 配置文件路径

可选项：

- 仓库地址（若链接无法自动识别）
- 本地工作区路径
- 本地仓库路径（本地模式）

## 生产发布建议

1. 先在目标平台本机运行 `wails doctor`
2. 先出未签名调试包验证功能
3. 再走签名 / notarization / 安装包流程
4. 统一保留 `report.json` 作为审计归档的结构化数据源
5. 发版前执行 `docs/RELEASE_CHECKLIST.md`

## 重要说明

- 这个包已经把工程骨架、打包脚本、打包文档、CI 样例和签名模板都放齐了。
- 由于当前交付物是源码增强包，不包含已经签名好的最终二进制；你需要在自己的 Windows/macOS 机器上按文档执行打包。
- 当前环境未为你实际下载 Go/Wails 前端依赖并完成整机出包，因此首次本地构建时建议先执行预检脚本。
