# DEVOPS

本文档记录 `git-muse` 的本地验证、构建、发布和故障排查方式。项目是单文件 Go CLI，不需要常驻服务、数据库或服务器部署。

## 架构

- 产物：`git-muse` 可执行文件。
- Git 子命令入口：把 `git-muse` 放进 `PATH` 后可执行 `git muse`。
- Git 行为：调用系统 `git`，保留用户本地 hooks、签名、editor、credential helper。
- LLM：后续通过 OpenAI-compatible API 或 Ollama 调用。

## 本地环境

要求：

```text
Go >= 1.26
Git >= 2.x
```

检查：

```bash
go version
git --version
```

## 常用命令

```bash
go test ./...
go run . help
go run . inspect
go run . doctor
```

若系统有 `make`：

```bash
make check
```

Windows PowerShell 下若没有 `make`，直接用 Go 命令。

## 构建

本机构建：

```bash
go build -o bin/git-muse.exe .
```

Linux：

```bash
GOOS=linux GOARCH=amd64 go build -o dist/git-muse-linux-amd64 .
```

macOS：

```bash
GOOS=darwin GOARCH=arm64 go build -o dist/git-muse-darwin-arm64 .
```

Windows：

```bash
GOOS=windows GOARCH=amd64 go build -o dist/git-muse-windows-amd64.exe .
```

## 安装验证

把二进制放到 `PATH` 后：

```bash
git muse help
```

如果提示：

```text
git: 'muse' is not a git command
```

说明 `git-muse` 不在 `PATH`，或文件名不是 `git-muse` / `git-muse.exe`。

## CI

文件：`.github/workflows/ci.yml`

触发：

- push 到 `main`
- pull request 到 `main`
- 手动 `workflow_dispatch`

当前 CI 流程：

1. `gofmt` 检查
2. `go test ./...`
3. 本机构建 smoke binary

push 到 `main` 且 CI 成功后，只有本次 push 包含 `.build_run` 变更，才会继续执行 release：

1. 读取现有 `vX.Y.Z` tag。
2. 若当前 commit 已有版本 tag，复用该 tag。
3. 否则自动递增版本号并创建 tag。
4. 构建多平台压缩包。
5. 生成 `checksums.txt`。
6. 发布 GitHub Release。

普通代码提交不会打包发版。需要发版时，修改根目录 `.build_run` 后提交即可。

例如：

```bash
date -u +%Y-%m-%dT%H:%M:%SZ > .build_run
git add .build_run
git commit -m "chore: trigger release"
git push origin main
```

当前构建目标：

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`

## 发布计划

当前已自动创建 GitHub Release。版本号递增规则：

- 无历史 tag：`v0.1.1`
- patch 小于 `10`：patch + 1
- patch 达到 `10`：minor + 1，patch 回到 `1`
- minor 达到 `20`：major + 1，minor 回到 `0`，patch 回到 `1`

Release 包：

- `git-muse-vX.Y.Z-linux-amd64.tar.gz`
- `git-muse-vX.Y.Z-linux-arm64.tar.gz`
- `git-muse-vX.Y.Z-darwin-amd64.tar.gz`
- `git-muse-vX.Y.Z-darwin-arm64.tar.gz`
- `git-muse-vX.Y.Z-windows-amd64.zip`
- `checksums.txt`

后续再加：

1. Homebrew tap
2. Scoop manifest
3. winget manifest

## Provider 配置

Ollama 示例：

```bash
git config --global muse.provider ollama
git config --global muse.baseUrl http://localhost:11434
git config --global muse.model qwen2.5-coder
```

OpenAI-compatible 示例：

```bash
git config --global muse.provider compatible
git config --global muse.baseUrl https://api.openai.com/v1
git config --global muse.model gpt-4.1-mini
```

API key 只放环境变量：

```bash
OPENAI_API_KEY=...
OPENROUTER_API_KEY=...
```

不要写进仓库配置文件。

## 故障排查

### `No staged changes found`

先暂存文件：

```bash
git add <files>
```

### `not inside Git repository`

进入 Git 仓库目录再执行。

### `git muse` 找不到

检查二进制名称和 `PATH`：

```bash
git-muse help
git muse help
```

### LLM 请求失败

优先检查：

- `muse.provider`
- `muse.model`
- `muse.baseUrl`
- API key 环境变量
- 本机网络或 Ollama 服务状态

Ollama 本地验证：

```bash
ollama list
curl http://localhost:11434/api/tags
```

OpenAI-compatible 验证：

```bash
echo $OPENAI_API_KEY
```

### Commit 失败

`git-muse` 不绕过 Git。失败通常来自：

- commit hook 拒绝
- GPG / SSH 签名失败
- user.name / user.email 未配置
- message 被编辑为空

直接看 Git 返回的 stderr。
