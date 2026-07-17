# 更改记录

2026-07-17 00:00:00

1. 初始化 Go CLI 工程骨架，新增 `main.go`、`cmd`、`internal/git`、`internal/diffctx` 和 `internal/commitmsg`。
2. 新增 `plan`，记录 `git-muse` 的 MVP 命令、架构、配置、Provider、安全边界和里程碑。
3. 新增最小 commit message 渲染测试，并验证 `go test ./...` 与 `go run . help` 可运行。

2026-07-17 00:00:01

1. 参考 `aphelios - mono` 的协作、日志和工程文档风格，新增 `AGENTS.md`、`DESIGN.md`、`DEVOPS.md` 和 `LOGS.md`。
2. 新增 `.editorconfig`、`.gitignore` 和 `Makefile`，补齐 Go CLI 项目的基础工程约束。
3. 完善 `README.md`，补充项目定位、命令、架构、配置、安全原则和开发命令。

2026-07-17 00:00:02

1. 新增 GitHub Actions CI 工作流，仅在 `ambisonic` 分支 push、pull request 或手动触发时运行。
2. CI 增加 `gofmt` 检查、`go test ./...` 和 Linux/macOS/Windows 多平台 `go build`。
3. 更新 `DEVOPS.md`，补充当前 CI 触发条件、构建目标和后续发布计划。

2026-07-17 00:00:03

1. 将 GitHub Actions 升级为 `CI and Release`，在 `ambisonic` 分支 push 且 CI 通过后自动发布。
2. 参考 `oolaf-capture` 的版本策略，新增 `vX.Y.Z` tag 自动递增、tag 推送和 GitHub Release 创建流程。
3. Release 构建 Linux、macOS 和 Windows 多平台压缩包，并生成 `checksums.txt`。
