# git-muse

`git-muse` 是一个 Go 编写的 AI commit message helper。安装到 `PATH` 后，Git 会把它识别为子命令：

```bash
git muse
```

目标不是替代 Git，而是读取 staged changes，生成可编辑的 commit message，再交给原生 `git commit -F` 完成提交。

当前状态：MVP 已接入 staged diff、配置、Ollama / OpenAI-compatible provider、生成、确认编辑和 `git commit -F` 提交。

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
git muse init
git muse help
git-muse --help
git muse --version
git muse --proxy http://127.0.0.1:7890
git muse --proxy 127.0.0.1:7890
```

说明：`git muse --help` 由 Git 自己处理，会打开 `git-muse.html`；请先运行一次 `git muse init` 安装该页面。日常也可用 `git muse help` 或 `git-muse --help`。

当前可直接运行：

```bash
go run . help
go run . --version
go run . inspect
go run . doctor
```

`inspect` 和 `doctor` 需要仓库里已有 staged changes。

## 安装使用

### Windows

从 GitHub Release 下载：

```text
git-muse-vX.Y.Z-windows-amd64.zip
```

解压后得到：

```text
git-muse.exe
README.md
```

把 `git-muse.exe` 放到任意已加入 `PATH` 的目录，例如：

```text
C:\Users\<you>\bin\git-muse.exe
```

确认安装成功：

```bash
git-muse help
git muse help
```

在任意 Git 仓库中使用：

```bash
git add <files>
git muse inspect
git muse generate
git muse
```

使用前需要先配置 provider 和 model。

### macOS / Linux

从 GitHub Release 下载对应压缩包：

```text
git-muse-vX.Y.Z-darwin-arm64.tar.gz
git-muse-vX.Y.Z-linux-amd64.tar.gz
```

解压并放进 `PATH`：

```bash
tar -xzf git-muse-vX.Y.Z-linux-amd64.tar.gz
chmod +x git-muse
sudo mv git-muse /usr/local/bin/git-muse
git muse help
```

### 从源码运行

```bash
go run . help
go run . inspect
go run . generate
```

### 从源码构建

Windows：

```bash
go build -trimpath -ldflags "-s -w" -o git-muse.exe .
```

macOS / Linux：

```bash
go build -trimpath -ldflags "-s -w" -o git-muse .
```

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

## 配置

优先使用 `git muse init` 写入的 `~/.git-muse.json`。其中：

- `output.language`：仅支持 `en` / `zh`（默认 `en`；改为 `zh` 生成中文 commit message）
- `llm`：单个对象或对象数组。数组按顺序尝试；上游返回 429 / RPM 限流时自动切下一个，并在 stderr 打印切换日志
- `llm.proxy`：LLM HTTP 代理（同时用于访问 http/https 目标；也可用 CLI `--proxy`）
- `diff.ignore`：额外排除路径 glob（匹配完整路径或 basename）
- `diff.allow`：强制纳入路径（优先于 ignore 与内置 secret/vendor 规则）

示例：

```json
{
  "llm": [
    {
      "provider": "compatible",
      "baseUrl": "https://token.sensenova.cn/v1/chat/completions",
      "model": "deepseek-v4-flash",
      "apiKey": "sk-primary",
      "proxy": "http://127.0.0.1:7890"
    },
    {
      "provider": "compatible",
      "baseUrl": "https://api.openai.com/v1",
      "model": "gpt-4.1-mini",
      "apiKey": "sk-fallback"
    }
  ],
  "output": {
    "language": "zh"
  },
  "diff": {
    "maxBytes": 100000,
    "ignore": ["*.lock", "coverage/*"],
    "allow": [".env.example"]
  }
}
```

仍支持把 `llm` 写成单个对象。数组中未写的 `proxy` / `maxTokens` / `timeoutSeconds` / `temperature` 会继承第一条。限流切换只针对 RPM / 429 / `rate_limit` 一类错误，认证失败不会切到下一个。

代理优先级：`--proxy` > `GIT_MUSE_PROXY` > `~/.git-muse.json` 的 `llm.proxy` > `muse.proxy`。

也可读取这些 Git config：

```bash
git config --global muse.provider ollama
git config --global muse.baseUrl http://localhost:11434
git config --global muse.model qwen2.5-coder
git config --global muse.format conventional
git config --global muse.language en
git config --global muse.proxy http://127.0.0.1:7890
git config --global muse.diffIgnore "*.lock,coverage/*"
git config --global muse.diffAllow ".env.example"
```

API key 只从环境变量读取：

```text
OPENAI_API_KEY
OPENROUTER_API_KEY
ANTHROPIC_API_KEY
GEMINI_API_KEY
```

不会提供把 API key 写入项目配置的命令。

### Ollama 示例

```bash
git config --global muse.provider ollama
git config --global muse.baseUrl http://localhost:11434
git config --global muse.model qwen2.5-coder
git add <files>
git muse generate
```

### OpenAI-compatible 示例

```bash
git config --global muse.provider compatible
git config --global muse.baseUrl https://api.openai.com/v1
git config --global muse.model gpt-4.1-mini
```

设置环境变量：

```bash
OPENAI_API_KEY=...
```

然后运行：

```bash
git add <files>
git muse
```

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

## 发版

CI 每次 push 到 `ambisonic` 都会跑测试，但只有 `.build_run` 变更时才会打包并创建 GitHub Release。

触发发版：

```bash
date -u +%Y-%m-%dT%H:%M:%SZ > .build_run
git add .build_run
git commit -m "chore: trigger release"
git push origin ambisonic
```

普通代码提交不会打包。

## 文档

- `plan`：实现计划和里程碑
- `AGENTS.md`：工程协作规则
- `DESIGN.md`：CLI 输出和交互规则
- `DEVOPS.md`：构建、安装、发布和排障
- `LOGS.md`：更改记录，只追加
