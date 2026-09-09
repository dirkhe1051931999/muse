# CLI 体验设计

`git-muse` 是命令行工具，不做 Web UI。设计目标是：输出清楚、确认明确、失败可恢复。

## 体验原则

- 默认安全：生成建议，不直接提交。
- 默认只看 staged diff，避免描述未提交内容。
- 重要动作二次确认：提交、hook 安装、跳过 hook 等必须明确确认。
- 输出短，不刷屏；需要细节时提供 `inspect` 和 `doctor`。
- 机器可读输出后续用 `--json` 提供，不混入人类提示。

## 默认交互

```text
Suggested commit message:

feat(auth): add refresh token rotation

[c] commit  [e] edit  [r] regenerate  [v] view diff  [q] quit
>
```

提交前：

```text
Commit with this message?

feat(auth): add refresh token rotation

[y/N]
```

## 命令输出规则

- 成功输出写 stdout。
- 错误输出写 stderr。
- 默认不输出 debug prompt。
- `generate --quiet` 只输出 commit message，不输出解释文字。
- `inspect` 输出将要发送给 LLM 的摘要和过滤原因。

## Commit Message 规则

默认 Conventional Commits：

```text
<type>[optional scope]: <subject>
```

允许 type：

```text
feat fix docs style refactor perf test build ci chore revert
```

规则：

- subject 必须非空。
- subject 建议不超过 72 字符。
- subject 使用英文时偏向 imperative mood。
- scope 只能来自目录、包、模块或模型判断出的明确区域。
- 模型不能直接决定最终格式；程序负责渲染。

## Inspect 输出

目标：让用户知道什么会离开本机。

```text
Provider: ollama
Model: qwen2.5-coder
Base URL: http://localhost:11434
Fallback 2: compatible  gpt-4.1-mini  https://api.openai.com/v1
Files changed: 5
Included bytes: 8420
Excluded:
  .env.local       secret pattern
  dist/app.js      generated file
  pnpm-lock.yaml   large lockfile
```

## Doctor 输出

```text
Git repository      OK
Staged changes      OK, 5 files
Provider            ollama
Model               qwen2.5-coder
API key             not required
Network             OK
```

## 错误文案

无 staged changes：

```text
No staged changes found.

Stage files first:

  git add <files>
```

未配置 provider：

```text
No LLM provider configured.

Run:

  git config --global muse.provider ollama
  git config --global muse.model qwen2.5-coder
```

## 不做的设计

- 不用彩色输出承载关键信息；颜色只能增强，不能成为唯一语义。
- 不默认联网安装 hook。
- 不默认读取整个仓库文件内容。
- 不把 AI 生成内容直接当 shell 命令参数。
