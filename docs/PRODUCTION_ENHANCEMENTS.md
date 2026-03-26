# 生产可用增强包说明

本次增强包相较于前一版，重点补的是“如何更稳地交付给团队使用”。

## 已新增内容

### 1. 打包执行文件文档
- `docs/BUILD_EXECUTABLE_GUIDE.md`
- `docs/CI_RELEASE_GUIDE.md`
- `docs/RELEASE_CHECKLIST.md`

### 2. 构建与发版脚本
- `scripts/preflight.ps1`
- `scripts/preflight.sh`
- `scripts/build_windows.ps1`
- `scripts/build_macos.sh`
- `scripts/package_release.ps1`
- `scripts/package_release.sh`

### 3. 签名与 notarization 模板
- `build/darwin/Info.plist`
- `build/darwin/entitlements.plist`
- `build/darwin/gon-sign.sample.json`

### 4. CI 发布样例
- `.github/workflows/release.yml`

### 5. 配置增强
- 支持环境变量覆盖模型地址、模型名、API Key、并发、上下文预算等关键配置。

## 建议落地顺序

1. 先在开发机本地跑通未签名版本。
2. 再在 Windows 机器上跑 EXE 与 NSIS 安装包。
3. 再在 macOS 机器上跑 universal `.app`。
4. 通过 CI 做可重复构建。
5. 最后再接签名和 notarization。

## 当前仍建议优先补强的点

- AST/语言专项规则更细化
- 单测与集成测试
- 更多语言与框架识别
- 报告交互式筛选
- 误报抑制白名单
- 更完整的构建版本号注入
