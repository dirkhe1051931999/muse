# 工程协作规则

本文档约束 `git-muse` 仓库的日常修改。项目是 Go 编写的 Git 子命令 CLI，不套用前端 / Node / 数据库项目规则。

## 工作流硬性规则

- 修改前先确认影响范围，优先读相关文件和调用链，不做无目的全仓库改写。
- 工作区可能已有用户改动；禁止回滚、覆盖或重排无关变更。
- 每次代码修改后运行 `go test ./...`。
- 修改 Go 代码后必须运行 `gofmt`。
- 每次完成代码或工程文档改动后，追加更新 `LOGS.md`，不要把更改记录写进 `README.md`。
- `LOGS.md` 只能追加到文末，按时间正序排列；禁止插入到文件开头、中间或打乱已有记录。
- 中文 Markdown 文件必须保持 UTF-8。禁止用 Windows 默认编码追加中文内容。
- README 只记录项目能力、入口、安装、使用和架构摘要；细节计划写 `plan`，变更流水写 `LOGS.md`。

## Go 规则

- 标准库优先。没有明确收益，不引入依赖。
- 系统 Git 行为必须通过 `os/exec` 调用 `git`，不要用 Go 库重造 Git 提交、hook、签名和编辑器行为。
- CLI 默认只读取 staged changes，即 `git diff --cached`。
- 提交必须使用 `git commit -F <message-file>`，禁止拼接 shell 字符串执行 `git commit -m "..."`。
- API key 只能从环境变量读取，禁止写入 Git config、项目配置或日志。
- LLM 输出只能当作候选数据，最终 commit message 由程序渲染和校验。
- 单个 Go 源文件建议控制在 300 行内；超过 500 行必须拆分。
- 错误信息面向 CLI 用户，必须直接给出下一步命令或原因。

## 安全边界

- 默认不读取 unstaged diff。
- 默认不自动 `git add`、`git reset`、`git commit --no-verify`。
- 默认不安装 Git hook。
- 发送给 LLM 前必须能通过 `git muse inspect` 查看真实上下文。
- secrets、证书、`.env*`、大体积生成文件必须过滤或截断。
- 不记录 prompt 全量内容到普通日志，除非用户显式启用调试模式。

## LOGS.md 格式

若文件已有标题 `# 更改记录`，只追加时间戳和条目：

```text
2026-07-17 15:04:05

1. 完成 xxx。
2. 修复 xxx。
```

## 验证命令

日常改动后至少运行：

```bash
go test ./...
```

可选：

```bash
go run . help
go run . doctor
```
