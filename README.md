# git-muse

`git-muse` 是一个 Go 编写的 AI commit message helper。安装到 `PATH` 后，Git 会把它识别为子命令：

```bash
git muse
```

目标不是替代 Git，而是读取 staged changes，生成可编辑的 commit message，再交给原生 `git commit -F` 完成提交。

当前状态：Go CLI 工程骨架。LLM provider 和交互提交仍在计划中。

## 能力目标

- 读取 `git diff --cached`，只分析已暂存改动。
- 生成 Conventional Commit 候选。
- 支持用户确认、编辑、重新生成或取消。
- 使用 `git commit -F <message-file>` 提交，保留 hooks、签名和 editor 行为。
- 支持 OpenAI-compatible API 与 Ollama。
- 提供 `inspect` 查看实际将发送给 LLM 的上下文。
- 提供 `doctor` 诊断 Git、配置、模型和网络状态。

## 命令

```bash
git muse
git muse generate
git muse inspect
git muse doctor
git muse help
```

当前可直接运行：

```bash
go run . help
go run . inspect
go run . doctor
```

`inspect` 和 `doctor` 需要仓库里已有 staged changes。

## 目录结构

```text
cmd/                 CLI 命令路由
internal/app/        应用流程编排
internal/git/        系统 git 封装
internal/diffctx/    staged diff 上下文收集
internal/commitmsg/  commit message 候选与渲染
plan                 技术计划
AGENTS.md            协作规则
DESIGN.md            CLI 体验设计
DEVOPS.md            构建、安装、发布和故障排查
LOGS.md              更改记录
```

## 技术栈

| 层 | 技术 |
| --- | --- |
| CLI | Go 标准库 |
| Git | `os/exec` 调系统 `git` |
| 测试 | `go test` |
| 配置 | Git config + env，后续实现 |
| LLM | OpenAI-compatible / Ollama，后续实现 |

第一版不引入 Cobra / Viper。标准库够用时不加依赖。

## 配置计划

后续读取这些 Git config：

```bash
git config --global muse.provider ollama
git config --global muse.baseUrl http://localhost:11434
git config --global muse.model qwen2.5-coder
git config --global muse.format conventional
git config --global muse.language en
```

API key 只从环境变量读取：

```text
OPENAI_API_KEY
OPENROUTER_API_KEY
ANTHROPIC_API_KEY
GEMINI_API_KEY
```

不会提供把 API key 写入项目配置的命令。

## 安全原则

- 默认只读 staged diff。
- 默认不自动 `git add`。
- 默认不自动安装 Git hook。
- 默认不跳过 Git hooks。
- 不拼接 shell 字符串执行 commit。
- secrets、证书、`.env*`、生成文件和大 diff 必须过滤或截断。
- 提交前必须允许用户确认或编辑。

## 开发

要求：

```text
Go >= 1.26
Git >= 2.x
```

常用命令：

```bash
go test ./...
go run . help
```

如果有 `make`：

```bash
make check
```

## 文档

- `plan`：实现计划和里程碑
- `AGENTS.md`：工程协作规则
- `DESIGN.md`：CLI 输出和交互规则
- `DEVOPS.md`：构建、安装、发布和排障
- `LOGS.md`：更改记录，只追加
