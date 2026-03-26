# AIGUARD

## 项目用途
用于说明项目的业务背景、关键链路和边界。

## 高敏感目录
- src/auth
- src/payment
- internal/security
- internal/export

## 禁止用法
- 禁止 SQL 字符串拼接
- 禁止关闭 TLS 校验
- 禁止明文打印 token / password / secret
- 禁止在事务中执行外部 RPC / HTTP 调用
- 禁止硬编码密钥

## 框架约定
- Spring 事务应尽量缩小边界
- React Hooks 必须遵守 hooks 规则
- Go HTTP 客户端必须设置 Timeout
- 所有导出接口必须做权限校验和审计日志

## 严重级别升级规则
- 若命中权限、认证、支付、导出链路，则问题等级上调一级
- 若命中注入、越权、敏感信息泄露，则至少判定为严重
