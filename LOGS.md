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

2026-07-17 00:00:04

1. 补充 `README.md` 安装使用说明，明确 Windows release 包内包含 `git-muse.exe`。
2. 新增 Windows、macOS、Linux 的 release 包下载、解压、加入 `PATH` 和验证命令说明。
3. 补充源码运行和源码构建命令，避免 README 只说明开发入口。

2026-07-17 00:00:05

1. 新增 `.build_run` 发版开关文件，作为 GitHub Actions 打包发布的显式触发条件。
2. 调整 CI 工作流，普通 push 只运行格式检查、测试和 smoke build，只有 `.build_run` 改动时才执行 release 打包。
3. 更新 `README.md` 和 `DEVOPS.md`，补充通过修改 `.build_run` 触发发版的操作说明。

2026-07-17 00:00:06

1. 按 `plan` 完成 MVP：新增 Git config 配置读取、Ollama / OpenAI-compatible provider、prompt 构造和严格 JSON 候选解析。
2. 接通 `git muse generate`、`git muse inspect`、`git muse doctor` 和默认交互提交流程，最终通过 `git commit -F` 提交。
3. 新增 staged diff 过滤、secret 行脱敏、超长 diff 截断和相关单元测试，并更新 README、DEVOPS 与计划状态。
