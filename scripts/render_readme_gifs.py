"""Render README demo GIFs into docs/."""

from __future__ import annotations

from pathlib import Path

from PIL import Image, ImageDraw, ImageFont

ROOT = Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"

W, H = 960, 540
BG = (11, 17, 32)
PANEL = (22, 27, 40)
TITLEBAR = (28, 35, 51)
BORDER = (48, 58, 78)
TEXT = (230, 237, 243)
DIM = (139, 148, 158)
GREEN = (63, 185, 80)
BLUE = (88, 166, 255)
CYAN = (57, 197, 187)
YELLOW = (227, 179, 65)
ORANGE = (240, 136, 62)
RED = (248, 81, 73)
WHITE = (255, 255, 255)
CODE_BG = (13, 17, 23)
HIGHLIGHT = (56, 45, 24)

MONO = "C:\\Windows\\Fonts\\cascadiamono.ttf"
MONO_FALLBACK = "C:\\Windows\\Fonts\\consola.ttf"
CN = "C:\\Windows\\Fonts\\msyh.ttc"
CN_BOLD = "C:\\Windows\\Fonts\\msyhbd.ttc"


def font(path: str, size: int, fallback: str | None = None) -> ImageFont.FreeTypeFont:
    try:
        return ImageFont.truetype(path, size)
    except OSError:
        if fallback:
            return ImageFont.truetype(fallback, size)
        return ImageFont.truetype(MONO_FALLBACK, size)


F_TITLE = font(CN_BOLD, 26, CN)
F_STEP = font(CN, 18, CN)
F_MONO = font(MONO, 16, MONO_FALLBACK)
F_MONO_SM = font(MONO, 15, MONO_FALLBACK)
F_CAPTION = font(CN, 15, CN)
F_BADGE = font(CN_BOLD, 14, CN)


class Gif:
    def __init__(self) -> None:
        self.frames: list[Image.Image] = []
        self.durations: list[int] = []

    def add(self, im: Image.Image, ms: int) -> None:
        self.frames.append(im.convert("RGB"))
        self.durations.append(ms)

    def save(self, path: Path) -> None:
        path.parent.mkdir(parents=True, exist_ok=True)
        base = self.frames[0].quantize(colors=64, method=Image.Quantize.MEDIANCUT)
        quantized = [base] + [
            frame.quantize(palette=base, dither=Image.Dither.NONE) for frame in self.frames[1:]
        ]
        quantized[0].save(
            path,
            save_all=True,
            append_images=quantized[1:],
            duration=self.durations,
            loop=0,
            optimize=True,
            disposal=2,
        )


def new_canvas(step: str, title: str) -> tuple[Image.Image, ImageDraw.ImageDraw]:
    im = Image.new("RGB", (W, H), BG)
    d = ImageDraw.Draw(im)
    d.rounded_rectangle((24, 18, 936, 70), 12, fill=TITLEBAR)
    d.ellipse((44, 34, 60, 50), fill=RED)
    d.ellipse((70, 34, 86, 50), fill=YELLOW)
    d.ellipse((96, 34, 112, 50), fill=GREEN)
    d.text((136, 30), title, font=F_TITLE, fill=WHITE)
    d.text((24, 86), step, font=F_STEP, fill=BLUE)
    return im, d


def draw_window(d: ImageDraw.ImageDraw, box: tuple[int, int, int, int], title: str) -> None:
    x0, y0, x1, y1 = box
    d.rounded_rectangle(box, 14, fill=PANEL, outline=BORDER, width=1)
    d.rounded_rectangle((x0, y0, x1, y0 + 34), 14, fill=TITLEBAR)
    d.rectangle((x0, y0 + 20, x1, y0 + 34), fill=TITLEBAR)
    d.ellipse((x0 + 14, y0 + 10, x0 + 26, y0 + 22), fill=RED)
    d.ellipse((x0 + 32, y0 + 10, x0 + 44, y0 + 22), fill=YELLOW)
    d.ellipse((x0 + 50, y0 + 10, x0 + 62, y0 + 22), fill=GREEN)
    d.text((x0 + 78, y0 + 8), title, font=F_BADGE, fill=DIM)


def line_color(text: str) -> tuple[int, int, int]:
    stripped = text.lstrip()
    if stripped.startswith("PS ") or stripped.startswith("$ "):
        return GREEN
    if stripped.startswith(">") or stripped.startswith("[y/N]") or stripped.startswith("[c]"):
        return CYAN
    if stripped.startswith("feat(") or stripped.startswith("fix("):
        return ORANGE
    if stripped.startswith("Changes") or stripped.startswith("new file") or stripped.startswith("[main"):
        return GREEN
    if stripped.startswith("Wrote") or stripped.startswith("Edit ") or stripped.startswith("Installed") or stripped.startswith("PATH"):
        return DIM
    if stripped.startswith("Suggested") or stripped.startswith("Commit"):
        return BLUE
    if stripped.startswith("Generating"):
        return YELLOW
    if stripped.startswith("#"):
        return DIM
    return TEXT


def has_cjk(text: str) -> bool:
    return any("\u4e00" <= ch <= "\u9fff" for ch in text)


def draw_lines(
    d: ImageDraw.ImageDraw,
    origin: tuple[int, int],
    lines: list[str],
    *,
    size_font: ImageFont.FreeTypeFont | None = None,
    lh: int = 24,
) -> None:
    x, y = origin
    used = size_font or F_MONO
    for raw in lines:
        if raw is None:
            continue
        text = raw
        color = line_color(text)
        line_font = F_CAPTION if has_cjk(text) else used
        if text.startswith("PS ") and not has_cjk(text):
            prompt, _, rest = text.partition(">")
            d.text((x, y), prompt + ">", font=used, fill=GREEN)
            if rest:
                d.text((x + used.getlength(prompt + ">"), y), rest, font=used, fill=WHITE)
        else:
            d.text((x, y), text, font=line_font, fill=color)
        y += lh


def callout(d: ImageDraw.ImageDraw, box: tuple[int, int, int, int], text: str) -> None:
    x0, y0, x1, y1 = box
    d.rounded_rectangle(box, 10, fill=(36, 28, 14), outline=YELLOW, width=2)
    for i, part in enumerate(text.split("\n")):
        d.text((x0 + 12, y0 + 8 + i * 22), part, font=F_CAPTION, fill=YELLOW)


def type_command(gif: Gif, render, prompt: str, command: str, *, char_ms: int = 28, pause: int = 360) -> str:
    line = prompt
    for i in range(0, len(command) + 1, 2):
        line = prompt + command[:i]
        gif.add(render([line]), char_ms)
    done = prompt + command
    gif.add(render([done]), pause)
    return done


def render_init() -> None:
    gif = Gif()
    term = (36, 122, 924, 508)

    def frame(lines: list[str], step: str, extra=None) -> Image.Image:
        im, d = new_canvas(step, "git-muse  ·  第一次配置")
        draw_window(d, term, "PowerShell")
        draw_lines(d, (56, 170), lines)
        if extra:
            extra(d)
        return im

    step1 = "步骤 1 / 3    生成用户配置"
    gif.add(frame(["PS C:\\demo>"], step1), 500)
    typed = type_command(
        gif,
        lambda current: frame(current, step1),
        "PS C:\\demo> ",
        "git muse init",
    )
    built = [typed]
    for line in (
        "Wrote C:\\Users\\you\\.git-muse.json",
        "Edit apiKey before running git muse.",
        "Updated user PATH git-muse entry to C:\\Users\\you\\bin",
        "Installed Git help page git-muse.html",
        "PS C:\\demo> _",
    ):
        built.append(line)
        gif.add(frame(built, step1), 380)
    gif.add(frame(built, step1), 900)

    step2 = "步骤 2 / 3    打开 ~/.git-muse.json，填入模型和密钥"
    json_lines = [
        "{",
        '  "llm": [{',
        '    "provider": "compatible",',
        '    "baseUrl": "https://api.example.com/v1/chat/completions",',
        '    "model": "your-model",',
        '    "apiKey": "sk-xxxxxxxx",',
        '    "proxy": "http://127.0.0.1:7890"',
        "  }],",
        '  "output": { "language": "zh" }',
        "}",
    ]

    def json_frame(highlight: int | None, tip: str | None) -> Image.Image:
        im, d = new_canvas(step2, "git-muse  ·  第一次配置")
        draw_window(d, term, "~/.git-muse.json")
        x, y = 56, 168
        for i, line in enumerate(json_lines):
            if highlight is not None and i == highlight:
                d.rounded_rectangle((48, y - 2, 700, y + 22), 6, fill=HIGHLIGHT)
            color = DIM if line.strip() in {"{", "}", "}],"} else TEXT
            if "apiKey" in line or "language" in line or "model" in line or "baseUrl" in line:
                color = ORANGE if highlight == i else CYAN
            d.text((x, y), line, font=F_MONO_SM, fill=color)
            y += 26
        if tip:
            callout(d, (720, 250, 920, 330), tip)
        return im

    gif.add(json_frame(None, None), 500)
    gif.add(json_frame(3, "填接口地址"), 1300)
    gif.add(json_frame(4, "填模型名"), 1300)
    gif.add(json_frame(5, "填入你的 apiKey"), 1600)
    gif.add(json_frame(8, 'language: "zh"\n生成中文 message'), 1800)

    step3 = "步骤 3 / 3    配置完成，以后改动暂存后即可 git muse"
    done_lines = [
        "PS C:\\demo> git muse doctor",
        "Git repository      OK",
        "Staged changes      OK, 1840 diff bytes",
        "LLM 1               compatible  your-model",
        "Language            zh",
        "API key             OK, from ~/.git-muse.json",
        "Config file         C:\\Users\\you\\.git-muse.json",
    ]
    gif.add(frame(done_lines, step3), 2200)
    gif.save(DOCS / "init.gif")


def render_usage() -> None:
    gif = Gif()
    term = (36, 122, 924, 508)

    def frame(lines: list[str], step: str) -> Image.Image:
        im, d = new_canvas(step, "git-muse  ·  日常使用")
        draw_window(d, term, "PowerShell")
        draw_lines(d, (56, 170), lines, lh=23)
        return im

    step1 = "步骤 1 / 3    先 git add，只分析已暂存改动"
    gif.add(frame(["PS C:\\demo>"], step1), 400)
    typed = type_command(
        gif,
        lambda current: frame(current, step1),
        "PS C:\\demo> ",
        "git add internal/llm/ratelimit.go",
    )
    built = [typed, "PS C:\\demo> git status"]
    gif.add(frame(built, step1), 400)
    for line in (
        "On branch main",
        "Changes to be committed:",
        "  new file:   internal/llm/ratelimit.go",
        "PS C:\\demo> _",
    ):
        built.append(line)
        gif.add(frame(built, step1), 320)
    gif.add(frame(built, step1), 800)

    step2 = "步骤 2 / 3    运行 git muse，生成 commit message"
    gif.add(frame(["PS C:\\demo>"], step2), 300)
    typed = type_command(gif, lambda current: frame(current, step2), "PS C:\\demo> ", "git muse")
    prompt = [typed]
    for spin in ("|", "/", "-", "\\", "|", "/"):
        gif.add(frame(prompt + [f"Generating... {spin}"], step2), 90)
    suggested = [
        "PS C:\\demo> git muse",
        "Suggested commit message:",
        "",
        "feat(llm): retry rate-limited requests before failover",
        "",
        "[c] commit  [p] commit and push  [e] edit  [r] regenerate  [q] quit",
        "> ",
    ]
    gif.add(frame(suggested, step2), 1600)
    gif.add(frame(suggested[:-1] + ["> c"], step2), 1000)

    step3 = "步骤 3 / 3    确认后交给原生 git commit -F 提交"
    confirm = [
        "Commit with this message?",
        "",
        "feat(llm): retry rate-limited requests before failover",
        "",
        "[y/N]",
        "> ",
    ]
    gif.add(frame(confirm, step3), 1400)
    gif.add(frame(confirm[:-1] + ["> y"], step3), 900)
    done = [
        "[main a3f91c2] feat(llm): retry rate-limited requests before failover",
        " 1 file changed, 94 insertions(+)",
        " create mode 100644 internal/llm/ratelimit.go",
        "",
        "提交完成。使用 git commit -F，保留 hooks 和签名。",
    ]
    gif.add(frame(done, step3), 2600)
    gif.save(DOCS / "usage.gif")


if __name__ == "__main__":
    render_init()
    render_usage()
    for name in ("init.gif", "usage.gif"):
        path = DOCS / name
        print(f"{path}  {path.stat().st_size / 1024:.1f} KB")
