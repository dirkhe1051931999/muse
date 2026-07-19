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

2026-07-17 21:10:34

1. 新增 `git muse init`，用于在用户目录创建示例 `.git-muse.json` 配置文件。
2. 配置读取增加 `~/.git-muse.json` 支持，可从 JSON 读取 provider、baseUrl、model 和 apiKey。
3. OpenAI-compatible provider 支持 `baseUrl` 直接填写完整 `/chat/completions` 地址，避免重复拼接路径。

2026-07-17 21:15:33

1. 新增 `maxTokens` 和 `timeoutSeconds` 配置项，用于控制大 diff 场景下的模型输出长度和 HTTP 等待时间。
2. OpenAI-compatible 请求写入 `max_tokens`，Ollama 请求写入 `num_predict`。
3. `git muse doctor` 增加 maxDiffBytes、maxTokens 和 timeoutSeconds 输出，便于确认实际参数。

2026-07-17 21:19:02

1. 将 `.git-muse.json` 示例改为树形结构，按 `llm`、`output` 和 `diff` 分类配置。
2. 配置加载逻辑改为读取分类 JSON 字段，避免所有参数平铺在根节点。
3. 更新配置测试，覆盖树形 JSON 中的 LLM、输出和 diff 参数。

2026-07-17 21:25:32

1. 调整 `git muse init` 行为：配置文件已存在时直接覆盖写入示例配置。
2. 更新初始化提示文案，从创建改为写入，避免覆盖场景描述不准确。
3. 更新配置初始化测试，覆盖重复执行 `InitFile` 的覆盖行为。

2026-07-17 21:39:02

1. 交互选择新增 `[p] commit and push`，确认后先提交再执行 `git push`。
2. 新增 Git `Push` 封装，推送时继承终端输入输出，保留认证提示和 Git 原生输出。
3. 确认提示区分普通提交和提交并推送，避免用户误操作。

2026-07-17 21:43:05

1. 调整 `git muse init`：`.git-muse.json` 已存在时先询问是否覆盖。
2. `git muse init` 增加用户 PATH 写入，使用当前 `git-muse.exe` 所在目录并放到 PATH 最前。
3. `go run . init` 场景跳过 PATH 写入，避免把临时构建目录写进用户环境变量。

2026-07-17 21:43:56

1. 新增 PATH 前置逻辑测试，覆盖移动后的 `git-muse.exe` 目录优先级。

2026-07-17 21:48:26

1. 修复 `git muse init` 写入用户 PATH 时路径包含空格导致 PowerShell 参数解析失败的问题。
2. 用户 PATH 写入改为通过 `reg add HKCU\Environment` 完成，避免命令参数被空格拆分。

2026-07-17 21:56:28

1. 移除交互界面的 view diff 选项和相关参数传递，简化提交选择流程。
2. 修复 Windows 下 Git editor 配置为不可用命令时直接报错的问题，自动回退到 Notepad。
3. 新增编辑器命令解析测试，覆盖带空格的引号路径。

2026-07-17 22:01:59

1. 修复 Windows 编辑器启动方式，改为直接执行 editor 程序并传入临时文件路径，不再通过 `cmd /C` 拼接命令。
2. 新增 editor 参数拆分逻辑，支持带空格的引号路径和附加参数。
3. 扩展 UI 测试，覆盖编辑器程序和参数拆分。

2026-07-18 14:54:00

1. 支持 output.language 仅 en/zh：prompt 按语言生成中英文 commit message，inspect/doctor 展示 Language。
2. 生成期间 stderr 显示 Generating... 进度，完成后只输出渲染后的 message。
3. 配置 diff.ignore / diff.allow（及 git/env 对应项），过滤优先级 allow > ignore > 内置规则。
4. 新增全局 --help / --version，内嵌 version 包；Release 构建通过 ldflags 注入版本号。

2026-07-18 15:02:00

1. 超长 commit subject 改为自动截断到 72 字符，不再直接报错退出。
2. 强化 prompt，要求中英文 subject 保持简短，细节放到 body。

2026-07-18 15:05:00

1. 新增全局 `--proxy`，支持 `http://127.0.0.1:7890` 或 `127.0.0.1:7890`，供 LLM 请求走代理。
2. 配置增加 `llm.proxy` / `muse.proxy` / `GIT_MUSE_PROXY`，inspect 与 doctor 展示当前代理。

2026-07-18 15:08:00

1. OpenAI-compatible 请求默认关闭 thinking，避免 deepseek-v4 等模型把 max_tokens 耗在推理上导致 content 为空。
2. 空响应错误补充 finish_reason 提示与响应片段；默认 maxTokens 调整为 2048。

2026-07-18 15:20:00

1. `git muse init` 更新用户 PATH 时，识别并替换已有 git-muse 安装目录条目，不再每个版本追加一条。

2026-07-18 15:22:00

1. 说明并修复 `git muse --help`：Git 会查找 git-muse.html，不进入子命令；`init` 写入 `git --html-path`/git-muse.html。
2. 帮助文案抽到 `internal/docs`，推荐 `git muse help` / `git-muse --help` 作为不依赖手册页的用法。

2026-07-19 18:14:41

1. 新增蓝奏上传 GitHub Actions 工作流，参考 `json-prettier` 的接口、client key 和上传流程。
2. 工作流在 `ambisonic` 分支 push 时构建 `git-muse.exe` 并打包上传，`LANZOU_FOLDER_ID` 暂用占位值。

2026-07-19 18:17:30

1. 调整蓝奏上传工作流，先检查 push 中 `.build_run` 是否变化。
2. 仅当 `.build_run` 发生变化时执行构建打包和蓝奏上传任务。

2026-07-19 18:19:19

1. 调整蓝奏上传包名为 `git-muse-v0.2.10-windows-amd64-[10位时间戳].zip`，避免多次 CI 上传冲突。
2. 蓝奏上传工作流仅构建并上传 Windows amd64 包，包内包含 `git-muse.exe`、README 和可选 LICENSE。

2026-07-19 18:46:41

1. 精简蓝奏上传包内容，仅压缩 `git-muse.exe`，降低触发服务端 413 限制的概率。
2. 上传前输出压缩包大小，便于判断 nginx 请求体限制是否仍需调整。

2026-07-19 19:10:43

1. 蓝奏上传工作流增加 `x-lanzou-ylogin` 和 `x-lanzou-phpdisk-info` 请求头，配合服务端按指定蓝奏账号凭证上传。

2026-07-19 19:23:21

1. 撤回蓝奏上传工作流中的蓝奏账号凭证请求头，恢复仅传 `x-oolaf-client-key`、`folderId` 和文件。

2026-07-19 19:36:05

1. 调整蓝奏上传工作流的 HTTP 状态处理，打印响应体并保留非 2xx 错误判断。
2. 针对蓝奏文件已上传但上游网关返回 504 的场景，将 504 视为非致命结果，避免 CI 误报失败。
