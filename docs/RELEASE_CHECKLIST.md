# 发版检查清单

## 构建前

- [ ] `config.yaml` 已确认指向正确模型网关
- [ ] `AIGUARD.md` 已补充当前项目规则
- [ ] `frontend` 依赖可以正常安装
- [ ] 本机 `git` / `go` / `npm` / `wails` 可用
- [ ] 执行过 `wails doctor`

## Windows 包

- [ ] 已生成 `AIGuard.exe`
- [ ] 如需安装包，已生成 NSIS 安装包
- [ ] 如需签名，已完成证书签名
- [ ] 已在干净 Windows 环境验证启动

## macOS 包

- [ ] 已生成 `AIGuard.app`
- [ ] 如需分发，已完成签名
- [ ] 如需对外分发，已完成 notarization
- [ ] 已在目标 macOS 版本验证启动

## 审计链路

- [ ] 可正常拉仓 / fetch / worktree
- [ ] 可正确生成 merge-base diff
- [ ] 可正确生成 HTML / Markdown / JSON 报告
- [ ] 可正确展示新增 / 已修复 / 仍存在问题对比

## 归档

- [ ] 已保留构建日志
- [ ] 已保留最终产物
- [ ] 已保留样例配置文件
- [ ] 已保留打包文档
