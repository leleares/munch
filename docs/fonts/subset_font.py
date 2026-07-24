import glob, os, subprocess, sys

SRC = "LXGWWenKaiTC-Regular.ttf"
PROJ = "/Users/luojilab/wanglele/projects/munch/miniprogram/src"

def gb2312(level2=False):
    hi_end = 0xF8 if level2 else 0xD8   # 0xB0-0xD7 一级(常用) / 到 0xF7 含二级
    s = set()
    for b1 in range(0xB0, hi_end):
        for b2 in range(0xA1, 0xFF):
            try:
                s.add(bytes([b1, b2]).decode("gb2312"))
            except Exception:
                pass
    return s

# 项目内出现的所有字符（界面固定文案）
proj = set()
for p in glob.glob(os.path.join(PROJ, "**", "*.*"), recursive=True):
    if p.endswith((".vue", ".js", ".json", ".scss")):
        try:
            proj |= set(open(p, encoding="utf-8").read())
        except Exception:
            pass

ascii_p = set(chr(i) for i in range(0x20, 0x7F))
cjk_punct = set("，。！？；：、“”‘’（）《》〈〉【】—…·～「」『』￥％")

for name, level2 in (("l1", False), ("full", True)):
    chars = gb2312(level2) | proj | ascii_p | cjk_punct
    chars = {c for c in chars if ord(c) > 0x1F and not (0x1F300 <= ord(c) <= 0x1FAFF)}  # 去掉 emoji（系统字体渲染）
    txt = f"chars_{name}.txt"
    open(txt, "w", encoding="utf-8").write("".join(sorted(chars)))
    out = f"lxgw-{name}.woff2"
    subprocess.run([
        sys.executable, "-m", "fontTools.subset", SRC,
        f"--text-file={txt}", "--flavor=woff2",
        "--layout-features=", "--no-hinting", "--desubroutinize",
        "--drop-tables+=GSUB,GPOS,GDEF,morx,kern",
        f"--output-file={out}",
    ], check=True)
    print(f"{name}: 字符数={len(chars)}  体积={os.path.getsize(out)/1024:.0f}KB")
